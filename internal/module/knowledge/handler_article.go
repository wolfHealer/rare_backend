package knowledge

import (
	"strconv"

	"rare_backend/internal/module/knowledge/domain"

	"github.com/gin-gonic/gin"
)

func parseArticleID(c *gin.Context) (uint, bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		respondBadRequest(c, "无效ID")
		return 0, false
	}
	return uint(id), true
}

func toCreateArticleInput(req CreateArticleRequest) domain.CreateArticleInput {
	blocks := make([]domain.CreateArticleBlockInput, len(req.Blocks))
	for i, b := range req.Blocks {
		blocks[i] = domain.CreateArticleBlockInput{
			BlockType: b.BlockType, SortNo: b.SortNo, Title: b.Title, Content: b.Content, Extra: b.Extra,
		}
	}
	return domain.CreateArticleInput{
		Title: req.Title, Summary: req.Summary, CoverImage: req.CoverImage, ContentType: req.ContentType,
		AuthorID: req.AuthorID, SourceName: req.SourceName, SourceURL: req.SourceURL,
		Status: req.Status, PublishTime: req.PublishTime, IsTop: req.IsTop, IsRecommend: req.IsRecommend,
		SeoTitle: req.SeoTitle, SeoKeywords: req.SeoKeywords, SeoDescription: req.SeoDescription,
		Blocks: blocks, TagIDs: req.TagIDs, DiseaseIDs: req.DiseaseIDs,
	}
}

func toUpdateArticleInput(req UpdateArticleRequest) domain.UpdateArticleInput {
	var blocks []domain.CreateArticleBlockInput
	if req.Blocks != nil {
		blocks = make([]domain.CreateArticleBlockInput, len(req.Blocks))
		for i, b := range req.Blocks {
			blocks[i] = domain.CreateArticleBlockInput{
				BlockType: b.BlockType, SortNo: b.SortNo, Title: b.Title, Content: b.Content, Extra: b.Extra,
			}
		}
	}
	return domain.UpdateArticleInput{
		Title: req.Title, Summary: req.Summary, CoverImage: req.CoverImage, ContentType: req.ContentType,
		AuthorID: req.AuthorID, SourceName: req.SourceName, SourceURL: req.SourceURL,
		Status: req.Status, PublishTime: req.PublishTime, IsTop: req.IsTop, IsRecommend: req.IsRecommend,
		SeoTitle: req.SeoTitle, SeoKeywords: req.SeoKeywords, SeoDescription: req.SeoDescription,
		Blocks: blocks, TagIDs: req.TagIDs, DiseaseIDs: req.DiseaseIDs,
	}
}

func CreateArticle(c *gin.Context) {
	var req CreateArticleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondBadRequest(c, "参数错误")
		return
	}
	id, err := articleSvc.Create(toCreateArticleInput(req))
	if err != nil {
		respondInternalError(c, "创建文章失败")
		return
	}
	respondOK(c, gin.H{"id": id})
}

func GetArticleByID(c *gin.Context) {
	id, ok := parseArticleID(c)
	if !ok {
		return
	}
	item, err := articleSvc.GetByID(id)
	if err != nil {
		if err == domain.ErrNotFound {
			respondNotFound(c, "文章不存在")
			return
		}
		respondInternalError(c, "查询失败")
		return
	}
	respondOK(c, item)
}

func UpdateArticle(c *gin.Context) {
	id, ok := parseArticleID(c)
	if !ok {
		return
	}
	var req UpdateArticleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondBadRequest(c, "参数错误")
		return
	}
	if err := articleSvc.Update(id, toUpdateArticleInput(req)); err != nil {
		respondInternalError(c, "更新主表失败")
		return
	}
	respondOK(c, nil)
}

func DeleteArticle(c *gin.Context) {
	id, ok := parseArticleID(c)
	if !ok {
		return
	}
	if err := articleSvc.Delete(id); err != nil {
		respondInternalError(c, "删除失败")
		return
	}
	respondOK(c, nil)
}

func GetArticles(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	if page < 1 {
		page = 1
	}
	if pageSize > 100 {
		pageSize = 100
	}

	result, err := articleSvc.List(domain.ArticleListFilter{
		Keyword:   c.DefaultQuery("keyword", ""),
		Status:    c.DefaultQuery("status", ""),
		DiseaseID: c.DefaultQuery("diseaseId", ""),
		TagID:     c.Query("tagId"),
		Page:      page,
		PageSize:  pageSize,
	})
	if err != nil {
		respondInternalError(c, "查询失败")
		return
	}
	respondPage(c, result.List, result.Total, page, pageSize)
}
