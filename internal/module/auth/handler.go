package auth

import (
	"fmt"
	"log"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"

	"rare_backend/internal/config"
	"rare_backend/internal/middleware"
	"strconv"

	"rare_backend/internal/module/auth/domain"
	"rare_backend/internal/module/auth/repo"
	"rare_backend/internal/module/auth/service"
	"rare_backend/internal/pkg/storage"

	"github.com/gin-gonic/gin"
)

var userSvc = service.NewUserService(repo.NewUserRepo())

func login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondBadRequest(c, "参数错误")
		return
	}
	result, err := userSvc.Login(domain.LoginInput{Phone: req.Phone, Password: req.Password})
	if err != nil {
		respondServiceError(c, err)
		return
	}
	respondOKMessage(c, "登录成功", gin.H{
		"user_id":    result.UserID,
		"role":       result.Role,
		"nickname":   result.Nickname,
		"phone":      result.Phone,
		"avatar":     result.Avatar,
		"token":      result.Token,
		"expires_in": result.ExpiresIn,
	})
}

func register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondBadRequest(c, "参数错误")
		return
	}
	if err := userSvc.Register(domain.RegisterInput{Phone: req.Phone, Password: req.Password}); err != nil {
		respondServiceError(c, err)
		return
	}
	respondMessage(c, 200, "注册成功")
}

func GetUserList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}
	filter := domain.UserListFilter{
		Keyword:  c.DefaultQuery("keyword", ""),
		Page:     page,
		PageSize: pageSize,
	}
	if statusStr := c.DefaultQuery("status", ""); statusStr != "" {
		if status, err := strconv.Atoi(statusStr); err == nil {
			filter.Status = &status
		}
	}
	if roleStr := c.DefaultQuery("role", ""); roleStr != "" {
		if role, err := strconv.Atoi(roleStr); err == nil {
			filter.Role = &role
		}
	}
	result, err := userSvc.ListUsers(filter)
	if err != nil {
		respondServiceError(c, err)
		return
	}
	respondPage(c, result.List, result.Total, result.Page, result.PageSize)
}

func GetUserByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		respondBadRequest(c, "无效的用户 ID")
		return
	}
	item, err := userSvc.GetUser(id)
	if err != nil {
		respondServiceError(c, err)
		return
	}
	respondOK(c, item)
}

func UpdateUser(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		respondBadRequest(c, "无效的用户 ID")
		return
	}
	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondBadRequest(c, "参数错误")
		return
	}
	if err := userSvc.UpdateUser(id, domain.UpdateUserInput{
		DisplayName: req.DisplayName,
		Avatar:      req.Avatar,
		Role:        req.Role,
		Status:      req.Status,
	}); err != nil {
		respondServiceError(c, err)
		return
	}
	respondOK(c, nil)
}

func DeleteUser(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		respondBadRequest(c, "无效的用户 ID")
		return
	}
	if err := userSvc.DeleteUser(id); err != nil {
		respondServiceError(c, err)
		return
	}
	respondOK(c, nil)
}

func UpdateUserRole(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		respondBadRequest(c, "无效的用户 ID")
		return
	}
	var req UpdateUserRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondBadRequest(c, "参数错误")
		return
	}
	if err := userSvc.UpdateUserRole(id, req.Role); err != nil {
		respondServiceError(c, err)
		return
	}
	respondOK(c, nil)
}

func ResetPassword(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		respondBadRequest(c, "无效的用户 ID")
		return
	}
	var req ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondBadRequest(c, "参数错误")
		return
	}
	if err := userSvc.ResetPassword(id, req.NewPassword); err != nil {
		respondServiceError(c, err)
		return
	}
	respondMessage(c, 200, "密码重置成功")
}

func CreateUser(c *gin.Context) {
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondBadRequest(c, "参数错误")
		return
	}
	newID, err := userSvc.CreateUser(domain.CreateUserInput{
		Phone:       req.Phone,
		Password:    req.Password,
		DisplayName: req.DisplayName,
		Avatar:      req.Avatar,
		Role:        req.Role,
		Status:      req.Status,
	})
	if err != nil {
		respondServiceError(c, err)
		return
	}
	respondOK(c, gin.H{"id": newID})
}

func getUserInfo(c *gin.Context) {
	// 从 JWT 中间件获取用户 ID
	userID, exists := c.Get("user_id")
	if !exists {
		respondBadRequest(c, "用户未登录")
		return
	}
	id, ok := userID.(int64)
	if !ok {
		respondBadRequest(c, "无效的用户 ID")
		return
	}
	item, err := userSvc.GetUser(id)
	if err != nil {
		respondServiceError(c, err)
		return
	}
	// 返回用户信息，使用 nickname 作为字段名
	respondOK(c, gin.H{
		"id":       item.ID,
		"phone":    item.Phone,
		"nickname": item.DisplayName,
		"avatar":   item.Avatar,
	})
}

func uploadAvatar(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		respondBadRequest(c, "用户未登录")
		return
	}

	file, err := formFile(c, "file", "avatar")
	if err != nil {
		respondBadRequest(c, "请选择要上传的文件")
		return
	}

	contentType, ok := storage.ResolveImageContentType(file.Header.Get("Content-Type"), file.Filename)
	if !ok {
		respondBadRequest(c, "只支持 jpeg、png、webp、gif 格式的图片")
		return
	}

	maxSize := int64(5 * 1024 * 1024)
	if file.Size > maxSize {
		respondBadRequest(c, "文件大小不能超过 5MB")
		return
	}

	ext := filepath.Ext(file.Filename)
	if ext == "" {
		ext = storage.NormalizeImageExt(contentType)
	}
	ossPath := fmt.Sprintf("avatars/%d_%d%s", userID, time.Now().Unix(), ext)

	cfg := config.GetOSSConfig()
	if err := storage.ValidateOSSConfig(cfg); err != nil {
		log.Printf("[auth] avatar OSS config: %v", err)
		respondBadRequest(c, "头像上传服务未配置")
		return
	}

	srcFile, err := file.Open()
	if err != nil {
		log.Printf("[auth] avatar open file: %v", err)
		respondInternalError(c, "打开文件失败")
		return
	}
	defer srcFile.Close()

	if err := storage.Upload(cfg, ossPath, srcFile, contentType); err != nil {
		log.Printf("[auth] avatar OSS upload: %v", err)
		respondAvatarUploadError(c, err)
		return
	}

	avatarURL := storage.PublicURL(cfg, ossPath)
	if err := userSvc.UpdateAvatar(userID, avatarURL); err != nil {
		_ = storage.Delete(cfg, ossPath)
		respondServiceError(c, err)
		return
	}

	respondOK(c, gin.H{"url": avatarURL})
}

func formFile(c *gin.Context, names ...string) (*multipart.FileHeader, error) {
	for _, name := range names {
		if fh, err := c.FormFile(name); err == nil {
			return fh, nil
		}
	}
	return nil, fmt.Errorf("no file")
}

func respondAvatarUploadError(c *gin.Context, err error) {
	msg := err.Error()
	if strings.Contains(msg, "must be addressed using the specified endpoint") {
		respondInternalError(c, "OSS 区域与 Bucket 不匹配，请检查 OSS_ENDPOINT 是否与 Bucket 所在地域一致")
		return
	}
	if strings.Contains(msg, "AccessDenied") {
		respondInternalError(c, "OSS 访问被拒绝，请检查 AccessKey 权限与 Bucket 名称")
		return
	}
	respondInternalError(c, "上传文件失败")
}

var smsSvc = service.NewSMSService(repo.NewSMSRepo())

// SendSMSRequest 发送验证码请求
type SendSMSRequest struct {
	Phone string `json:"phone" binding:"required,len=11"`
	Scene string `json:"scene" binding:"required,oneof=register login"`
}

func sendSMSCode(c *gin.Context) {
	var req SendSMSRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondBadRequest(c, "参数错误")
		return
	}

	// 验证手机号格式（简单验证）
	if len(req.Phone) != 11 {
		respondBadRequest(c, "手机号格式不正确")
		return
	}

	if err := smsSvc.SendCode(req.Phone, req.Scene); err != nil {
		respondServiceError(c, err)
		return
	}

	respondOK(c, gin.H{"message": "验证码发送成功"})
}
