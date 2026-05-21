package community

import (
	"errors"

	"rare_backend/internal/module/community/domain"
	"rare_backend/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

func respondOK(c *gin.Context, data any) {
	response.OK(c, data)
}

func respondCreated(c *gin.Context, message string, data any) {
	response.OKMessage(c, message, data)
}

func respondBadRequest(c *gin.Context, message string) {
	response.BadRequest(c, message)
}

func respondNotFound(c *gin.Context, message string) {
	response.NotFound(c, message)
}

func respondForbidden(c *gin.Context, message string) {
	response.Forbidden(c, message)
}

func respondInternalError(c *gin.Context, message string) {
	response.InternalError(c, message)
}

func respondPage(c *gin.Context, list any, total int64, page, pageSize int) {
	response.Page(c, list, total, page, pageSize)
}

func respondServiceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrInvalidID):
		response.BadRequest(c, "无效的 ID")
	case errors.Is(err, domain.ErrPostNotFound):
		response.NotFound(c, "帖子不存在")
	case errors.Is(err, domain.ErrCommentNotFound):
		response.NotFound(c, "评论不存在")
	case errors.Is(err, domain.ErrParentComment):
		response.NotFound(c, "父评论不存在或不属于当前帖子")
	case errors.Is(err, domain.ErrForbidden):
		response.Forbidden(c, "无权限")
	case errors.Is(err, domain.ErrNoUpdateFields):
		response.BadRequest(c, "未提供有效更新字段")
	default:
		response.InternalError(c, "服务器错误")
	}
}
