package knowledge

import (
	"strconv"

	"rare_backend/internal/module/knowledge/domain"

	"github.com/gin-gonic/gin"
)

func parseCategoryID(c *gin.Context) (uint, bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		respondBadRequest(c, "无效的分类 ID")
		return 0, false
	}
	return uint(id), true
}

func CreateCategory(c *gin.Context) {
	var req CreateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondBadRequest(c, "参数错误")
		return
	}
	id, err := categorySvc.Create(domain.CreateCategoryInput{
		ParentID: req.ParentID, Level: req.Level, Name: req.Name, Code: req.Code,
		Description: req.Description, IconURL: req.IconURL, SortOrder: req.SortOrder,
	})
	if err != nil {
		respondCategoryError(c, err, "")
		return
	}
	respondOK(c, gin.H{"id": id})
}

func UpdateCategory(c *gin.Context) {
	id, ok := parseCategoryID(c)
	if !ok {
		return
	}
	var req UpdateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondBadRequest(c, "参数错误")
		return
	}
	err := categorySvc.Update(id, domain.UpdateCategoryInput{
		ParentID: req.ParentID, Level: req.Level, Name: req.Name, Code: req.Code,
		Description: req.Description, IconURL: req.IconURL, SortOrder: req.SortOrder, Status: req.Status,
	})
	if err != nil {
		respondCategoryError(c, err, "分类不存在")
		return
	}
	respondOK(c, nil)
}

func DeleteCategory(c *gin.Context) {
	id, ok := parseCategoryID(c)
	if !ok {
		return
	}
	if err := categorySvc.Delete(id); err != nil {
		respondCategoryError(c, err, "分类不存在或已删除")
		return
	}
	respondOK(c, nil)
}

func GetCategories(c *gin.Context) {
	filter := domain.CategoryListFilter{}
	parentIDStr := c.DefaultQuery("parentId", "-1")
	if parentIDStr != "-1" {
		if parentID, err := strconv.Atoi(parentIDStr); err == nil {
			filter.ParentID = &parentID
		}
	}
	if levelStr := c.DefaultQuery("level", ""); levelStr != "" {
		if level, err := strconv.Atoi(levelStr); err == nil {
			filter.Level = &level
		}
	}
	list, err := categorySvc.List(filter)
	if err != nil {
		respondInternalError(c, "查询分类列表失败")
		return
	}
	respondOK(c, list)
}

func GetCategoryTree(c *gin.Context) {
	filter := domain.CategoryTreeFilter{Keyword: c.DefaultQuery("keyword", "")}
	statusStr := c.DefaultQuery("status", "1")
	if statusStr != "" {
		if s, err := strconv.Atoi(statusStr); err == nil {
			filter.Status = &s
		}
	}
	tree := categorySvc.Tree(filter)
	respondOK(c, tree)
}
