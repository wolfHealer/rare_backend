package knowledge

import (
	"errors"
	"strconv"

	"rare_backend/internal/module/knowledge/domain"

	"github.com/gin-gonic/gin"
)

func parseDiseaseID(c *gin.Context) (uint, bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		respondBadRequest(c, "无效的疾病 ID")
		return 0, false
	}
	return uint(id), true
}

func parseDiseaseListFilter(c *gin.Context) domain.DiseaseListFilter {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}

	filter := domain.DiseaseListFilter{
		Keyword:  c.DefaultQuery("keyword", ""),
		Page:     page,
		PageSize: pageSize,
	}

	if statusStr := c.DefaultQuery("status", ""); statusStr != "" {
		if s, err := strconv.Atoi(statusStr); err == nil {
			filter.Status = &s
		}
	}
	if categoryIdStr := c.DefaultQuery("categoryId", ""); categoryIdStr != "" {
		if id, err := strconv.ParseUint(categoryIdStr, 10, 32); err == nil {
			v := uint(id)
			filter.CategoryID = &v
		}
	}
	if tagIdStr := c.DefaultQuery("tagId", ""); tagIdStr != "" {
		if id, err := strconv.ParseUint(tagIdStr, 10, 32); err == nil {
			v := uint(id)
			filter.TagID = &v
		}
	}
	return filter
}

func CreateDisease(c *gin.Context) {
	var req CreateDiseaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondBadRequest(c, "参数错误")
		return
	}
	id, err := diseaseSvc.Create(domain.CreateDiseaseInput{
		Name: req.Name, Alias: req.Alias, Introduction: req.Introduction, Symptoms: req.Symptoms,
		Images: req.Images, Status: req.Status, CreatorID: req.CreatorID,
		PrimaryCategoryID: req.PrimaryCategoryID, CategoryIDs: req.CategoryIDs, TagIDs: req.TagIDs,
	})
	if err != nil {
		respondDiseaseError(c, err, "")
		return
	}
	respondOK(c, gin.H{"id": id})
}

func GetDiseaseByID(c *gin.Context) {
	id, ok := parseDiseaseID(c)
	if !ok {
		return
	}
	item, err := diseaseSvc.GetByID(id)
	if err != nil {
		respondDiseaseError(c, err, "疾病不存在")
		return
	}
	respondOK(c, item)
}

func UpdateDisease(c *gin.Context) {
	id, ok := parseDiseaseID(c)
	if !ok {
		return
	}
	var req UpdateDiseaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondBadRequest(c, "参数错误")
		return
	}
	err := diseaseSvc.Update(id, domain.UpdateDiseaseInput{
		Name: req.Name, Alias: req.Alias, Introduction: req.Introduction, Symptoms: req.Symptoms,
		Images: req.Images, Status: req.Status,
		PrimaryCategoryID: req.PrimaryCategoryID, CategoryIDs: req.CategoryIDs, TagIDs: req.TagIDs,
	})
	if err != nil {
		respondDiseaseError(c, err, "疾病不存在")
		return
	}
	respondOK(c, nil)
}

func DeleteDisease(c *gin.Context) {
	id, ok := parseDiseaseID(c)
	if !ok {
		return
	}
	if err := diseaseSvc.Delete(id); err != nil {
		respondDiseaseError(c, err, "疾病不存在")
		return
	}
	respondOK(c, nil)
}

func GetDiseases(c *gin.Context) {
	filter := parseDiseaseListFilter(c)
	result, err := diseaseSvc.List(filter)
	if err != nil {
		respondInternalError(c, "查询总数失败")
		return
	}
	respondPage(c, result.List, result.Total, filter.Page, filter.PageSize)
}

func GetDiseasesByCategory(c *gin.Context) {
	categoryID, err := strconv.ParseUint(c.Param("categoryId"), 10, 32)
	if err != nil {
		respondBadRequest(c, "无效的分类 ID")
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	result, err := diseaseSvc.ListByCategory(domain.DiseasesByCategoryFilter{
		CategoryID: uint(categoryID),
		Page:       page,
		PageSize:   pageSize,
	})
	if err != nil {
		respondDiseaseError(c, err, "分类不存在或已停用")
		return
	}
	respondPage(c, result.List, result.Total, page, pageSize)
}

func SearchDiseases(c *gin.Context) {
	keyword := c.DefaultQuery("keyword", "")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	if pageSize == 0 {
		pageSize, _ = strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	}
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	result, err := diseaseSvc.Search(keyword, page, pageSize)
	if err != nil {
		if errors.Is(err, domain.ErrKeywordRequired) {
			respondBadRequest(c, "关键词不能为空")
			return
		}
		respondInternalError(c, "查询列表失败")
		return
	}

	respondOK(c, result)
}

func GetDiseaseOptions(c *gin.Context) {
	options, err := diseaseSvc.Options(c.DefaultQuery("keyword", ""))
	if err != nil {
		respondInternalError(c, "查询疾病选项失败")
		return
	}
	respondOK(c, options)
}
