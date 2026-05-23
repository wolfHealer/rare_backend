package auth

import (
	"errors"
	"strings"

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

func respondSMSError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrInvalidScene):
		response.BadRequest(c, "无效的场景")
	case errors.Is(err, domain.ErrTooManyRequests):
		response.BadRequest(c, "发送过于频繁，请稍后再试")
	default:
		respondAliyunSMSError(c, err)
	}
}

func respondAliyunSMSError(c *gin.Context, err error) {
	msg := err.Error()
	lower := strings.ToLower(msg)

	switch {
	case strings.Contains(msg, "未配置"):
		response.BadRequest(c, msg)
	case strings.Contains(msg, "签名或者模版无效"),
		strings.Contains(lower, "invalid_parameters"),
		strings.Contains(msg, "InvalidSignName"),
		strings.Contains(msg, "InvalidTemplateCode"):
		response.BadRequest(c, "短信签名或模板无效，请在号码认证控制台核对 SMS_PNVS_SIGN_NAME 与 SMS_PNVS_TEMPLATE_CODE 是否为配套的赠送资源（勿使用短信服务里的企业模板）")
	case strings.Contains(lower, "frequency"), strings.Contains(msg, "FREQUENCY"):
		response.BadRequest(c, "发送过于频繁，请稍后再试")
	case strings.Contains(msg, "FUNCTION_NOT_OPENED"):
		response.BadRequest(c, "请先在号码认证控制台开通短信认证功能")
	case strings.Contains(msg, "MOBILE_NUMBER_ILLEGAL"):
		response.BadRequest(c, "手机号格式不正确")
	case strings.Contains(msg, "AccessDenied"), strings.Contains(msg, "Forbidden"):
		response.InternalError(c, "短信服务访问被拒绝，请检查 AccessKey 权限（需 dypns:SendSmsVerifyCode）")
	default:
		response.InternalError(c, "验证码发送失败")
	}
}
