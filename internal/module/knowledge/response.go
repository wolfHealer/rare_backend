package knowledge

import (
	"errors"

	"rare_backend/internal/module/knowledge/domain"
	"rare_backend/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

func respondOK(c *gin.Context, data any) {
	response.OK(c, data)
}

func respondBadRequest(c *gin.Context, message string) {
	response.BadRequest(c, message)
}

func respondNotFound(c *gin.Context, message string) {
	response.NotFound(c, message)
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
	case errors.Is(err, domain.ErrNameRequired):
		response.BadRequest(c, "名称不能为空")
	case errors.Is(err, domain.ErrCodeRequired):
		response.BadRequest(c, "编码不能为空")
	case errors.Is(err, domain.ErrInvalidLevel):
		response.BadRequest(c, "层级必须在1-3之间")
	case errors.Is(err, domain.ErrCodeExists):
		response.BadRequest(c, "编码已存在")
	case errors.Is(err, domain.ErrNoUpdateFields):
		response.BadRequest(c, "未提供更新字段")
	case errors.Is(err, domain.ErrKeywordRequired):
		response.BadRequest(c, "关键词不能为空")
	case errors.Is(err, domain.ErrNotFound):
		response.NotFound(c, "资源不存在")
	default:
		response.InternalError(c, "服务器错误")
	}
}

func respondCategoryError(c *gin.Context, err error, notFoundMsg string) {
	if errors.Is(err, domain.ErrNotFound) {
		response.NotFound(c, notFoundMsg)
		return
	}
	if errors.Is(err, domain.ErrNameRequired) {
		response.BadRequest(c, "分类名称不能为空")
		return
	}
	if errors.Is(err, domain.ErrCodeRequired) {
		response.BadRequest(c, "分类编码不能为空")
		return
	}
	if errors.Is(err, domain.ErrInvalidLevel) {
		response.BadRequest(c, "层级必须在1-3之间")
		return
	}
	if errors.Is(err, domain.ErrCodeExists) {
		response.BadRequest(c, "分类编码已存在")
		return
	}
	if errors.Is(err, domain.ErrNoUpdateFields) {
		response.BadRequest(c, "未提供更新字段")
		return
	}
	response.InternalError(c, "服务器错误")
}

func respondTagError(c *gin.Context, err error, notFoundMsg string) {
	if errors.Is(err, domain.ErrNotFound) {
		response.NotFound(c, notFoundMsg)
		return
	}
	if errors.Is(err, domain.ErrNameRequired) {
		response.BadRequest(c, "标签名称不能为空")
		return
	}
	if errors.Is(err, domain.ErrCodeRequired) {
		response.BadRequest(c, "标签编码不能为空")
		return
	}
	if errors.Is(err, domain.ErrCodeExists) {
		response.BadRequest(c, "标签编码已存在")
		return
	}
	if errors.Is(err, domain.ErrNoUpdateFields) {
		response.BadRequest(c, "未提供更新字段")
		return
	}
	response.InternalError(c, "服务器错误")
}

func respondDiseaseError(c *gin.Context, err error, notFoundMsg string) {
	if errors.Is(err, domain.ErrNotFound) {
		response.NotFound(c, notFoundMsg)
		return
	}
	if errors.Is(err, domain.ErrNameRequired) {
		response.BadRequest(c, "病种名称不能为空")
		return
	}
	if errors.Is(err, domain.ErrKeywordRequired) {
		response.BadRequest(c, "关键词不能为空")
		return
	}
	if errors.Is(err, domain.ErrNoUpdateFields) {
		response.BadRequest(c, "未提供更新字段")
		return
	}
	response.InternalError(c, "服务器错误")
}

func respondArticleTagError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		response.NotFound(c, "文章标签不存在")
	case errors.Is(err, domain.ErrNameRequired):
		response.BadRequest(c, "标签名称不能为空")
	case errors.Is(err, domain.ErrNoUpdateFields):
		response.BadRequest(c, "未提供更新字段")
	case errors.Is(err, domain.ErrArticleTagNameExists):
		response.BadRequest(c, "标签名称已存在")
	case errors.Is(err, domain.ErrArticleTagInUse):
		response.BadRequest(c, "标签已被文章引用，无法删除")
	default:
		response.InternalError(c, "服务器错误")
	}
}
