package auth

import (
	"strconv"

	"rare_backend/internal/module/auth/domain"
	"rare_backend/internal/module/auth/repo"
	"rare_backend/internal/module/auth/service"

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
