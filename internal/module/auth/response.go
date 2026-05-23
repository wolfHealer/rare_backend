package auth

import (
	"errors"

	"rare_backend/internal/module/auth/domain"
	"rare_backend/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

func respondOK(c *gin.Context, data any) {
	response.OK(c, data)
}

func respondBadRequest(c *gin.Context, message string) {
	response.BadRequest(c, message)
}

func respondInternalError(c *gin.Context, message string) {
	response.InternalError(c, message)
}

func respondPage(c *gin.Context, list any, total int64, page, pageSize int) {
	response.Page(c, list, total, page, pageSize)
}

func respondOKMessage(c *gin.Context, message string, data any) {
	response.OKMessage(c, message, data)
}

func respondMessage(c *gin.Context, code int, message string) {
	response.OKMessage(c, message, nil)
}

// internal/module/auth/response.go
func respondServiceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrInvalidCredentials):
		response.Unauthorized(c, "手机号或密码错误")
	case errors.Is(err, domain.ErrPhoneExists):
		response.BadRequest(c, "手机号已被注册")
	case errors.Is(err, domain.ErrUserNotFound):
		response.NotFound(c, "用户不存在")
	case errors.Is(err, domain.ErrUserDisabled):
		response.NotFound(c, "用户不存在或已禁用")
	case errors.Is(err, domain.ErrInvalidRole):
		response.BadRequest(c, "无效的角色类型")
	case errors.Is(err, domain.ErrInvalidStatus):
		response.BadRequest(c, "无效的状态值")
	case errors.Is(err, domain.ErrNoUpdateFields):
		response.BadRequest(c, "未提供需要更新的字段")
	// 添加短信相关错误
	case errors.Is(err, domain.ErrInvalidScene):
		response.BadRequest(c, "无效的场景")
	case errors.Is(err, domain.ErrTooManyRequests):
		response.BadRequest(c, "发送过于频繁，请稍后再试")
	case errors.Is(err, domain.ErrInvalidCode):
		response.BadRequest(c, "验证码无效或已过期")
	default:
		response.InternalError(c, "服务器错误")
	}
}
