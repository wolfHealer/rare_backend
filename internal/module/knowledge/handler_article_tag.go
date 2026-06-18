package knowledge

import (
	"strconv"

	"rare_backend/internal/module/knowledge/domain"

	"github.com/gin-gonic/gin"
)

func parseArticleTagID(c *gin.Context) (uint, bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil || id == 0 {
		respondBadRequest(c, "无效的标签 ID")
		return 0, false
	}
	return uint(id), true
}

// ListArticleTags 文章标签列表
// GET /api/knowledge/articles/tags?type=general
func ListArticleTags(c *gin.Context) {
	list, err := articleTagSvc.List(domain.ArticleTagListFilter{
		Type: c.Query("type"),
	})
	if err != nil {
		respondInternalError(c, "查询文章标签失败")
		return
	}
	if list == nil {
		list = []domain.ArticleTagItem{}
	}
	respondOK(c, list)
}

// GetArticleTagByID 文章标签详情
// GET /api/knowledge/articles/tags/:id
func GetArticleTagByID(c *gin.Context) {
	id, ok := parseArticleTagID(c)
	if !ok {
		return
	}
	item, err := articleTagSvc.GetByID(id)
	if err != nil {
		respondArticleTagError(c, err)
		return
	}
	respondOK(c, item)
}

// CreateArticleTag 创建文章标签（管理员）
// POST /api/knowledge/articles/tags
func CreateArticleTag(c *gin.Context) {
	var req CreateArticleTagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondBadRequest(c, "参数错误")
		return
	}
	id, err := articleTagSvc.Create(domain.CreateArticleTagInput{
		Name: req.Name,
		Type: req.Type,
	})
	if err != nil {
		respondArticleTagError(c, err)
		return
	}
	respondOK(c, gin.H{"id": id})
}

// UpdateArticleTag 更新文章标签（管理员）
// PUT /api/knowledge/articles/tags/:id
func UpdateArticleTag(c *gin.Context) {
	id, ok := parseArticleTagID(c)
	if !ok {
		return
	}
	var req UpdateArticleTagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondBadRequest(c, "参数错误")
		return
	}
	if err := articleTagSvc.Update(id, domain.UpdateArticleTagInput{
		Name: req.Name,
		Type: req.Type,
	}); err != nil {
		respondArticleTagError(c, err)
		return
	}
	respondOK(c, nil)
}

// DeleteArticleTag 删除文章标签（管理员）
// DELETE /api/knowledge/articles/tags/:id
func DeleteArticleTag(c *gin.Context) {
	id, ok := parseArticleTagID(c)
	if !ok {
		return
	}
	if err := articleTagSvc.Delete(id); err != nil {
		respondArticleTagError(c, err)
		return
	}
	respondOK(c, nil)
}
