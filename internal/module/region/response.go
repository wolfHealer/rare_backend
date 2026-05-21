package region

import (
	"errors"

	"rare_backend/internal/module/region/domain"
	"rare_backend/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

func respondOK(c *gin.Context, data any) {
	response.OK(c, data)
}

func respondBadRequest(c *gin.Context, message string) {
	response.BadRequest(c, message)
}

func respondPage(c *gin.Context, list any, total int64, page, pageSize int) {
	response.Page(c, list, total, page, pageSize)
}

func respondServiceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrInvalidID):
		response.BadRequest(c, "无效的 ID")
	case errors.Is(err, domain.ErrNotFound):
		response.NotFound(c, "行政区划不存在")
	case errors.Is(err, domain.ErrCodeExists):
		response.BadRequest(c, "行政区编码已存在")
	case errors.Is(err, domain.ErrParentNotFound):
		response.BadRequest(c, "父级行政区不存在")
	case errors.Is(err, domain.ErrHasChildren):
		response.BadRequest(c, "该行政区划下存在子级，无法删除")
	case errors.Is(err, domain.ErrInvalidLevel):
		response.BadRequest(c, "层级必须在1-3之间")
	case errors.Is(err, domain.ErrNoUpdateFields):
		response.BadRequest(c, "未提供更新字段")
	default:
		response.InternalError(c, "服务器错误")
	}
}

func respondTreeError(c *gin.Context, err error) {
	response.InternalError(c, "查询失败")
}
