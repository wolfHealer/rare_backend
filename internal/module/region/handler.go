package region

import (
	"strconv"

	"rare_backend/internal/module/region/domain"

	"github.com/gin-gonic/gin"
)

func parseRegionID(c *gin.Context) (uint, bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		respondBadRequest(c, "无效的 ID")
		return 0, false
	}
	return uint(id), true
}

func parseIsEnabled(c *gin.Context) int {
	isEnabled := 1
	if isEnabledStr := c.DefaultQuery("isEnabled", "1"); isEnabledStr != "" {
		if val, err := strconv.Atoi(isEnabledStr); err == nil {
			isEnabled = val
		}
	}
	return isEnabled
}

func CreateRegion(c *gin.Context) {
	var req CreateRegionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondBadRequest(c, "参数错误")
		return
	}
	id, err := regionSvc.Create(domain.CreateRegionInput{
		Code: req.Code, Name: req.Name, FullName: req.FullName,
		ParentCode: req.ParentCode, Level: req.Level, Sort: req.Sort, IsEnabled: req.IsEnabled,
	})
	if err != nil {
		respondServiceError(c, err)
		return
	}
	respondOK(c, gin.H{"id": id})
}

func UpdateRegion(c *gin.Context) {
	id, ok := parseRegionID(c)
	if !ok {
		return
	}
	var req UpdateRegionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondBadRequest(c, "参数错误")
		return
	}
	if err := regionSvc.Update(id, domain.UpdateRegionInput{
		Name: req.Name, FullName: req.FullName, ParentCode: req.ParentCode,
		Level: req.Level, Sort: req.Sort, IsEnabled: req.IsEnabled,
	}); err != nil {
		respondServiceError(c, err)
		return
	}
	respondOK(c, nil)
}

func DeleteRegion(c *gin.Context) {
	id, ok := parseRegionID(c)
	if !ok {
		return
	}
	if err := regionSvc.Delete(id); err != nil {
		respondServiceError(c, err)
		return
	}
	respondOK(c, nil)
}

func GetRegionList(c *gin.Context) {
	var req GetRegionListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		respondBadRequest(c, "参数错误")
		return
	}
	result, err := regionSvc.List(domain.RegionListFilter{
		ParentCode: req.ParentCode,
		Level:      req.Level,
		IsEnabled:  req.IsEnabled,
		Keyword:    req.Keyword,
		Page:       req.Page,
		PageSize:   req.PageSize,
	})
	if err != nil {
		respondServiceError(c, err)
		return
	}
	respondPage(c, result.List, result.Total, result.Page, result.PageSize)
}

func GetRegionTree(c *gin.Context) {
	tree, err := regionSvc.GetTree(parseIsEnabled(c))
	if err != nil {
		respondTreeError(c, err)
		return
	}
	respondOK(c, tree)
}

func GetProvinceCityTree(c *gin.Context) {
	tree, err := regionSvc.GetProvinceCityTree(parseIsEnabled(c))
	if err != nil {
		respondTreeError(c, err)
		return
	}
	respondOK(c, tree)
}

func GetRegionDetail(c *gin.Context) {
	id, ok := parseRegionID(c)
	if !ok {
		return
	}
	item, err := regionSvc.GetByID(id)
	if err != nil {
		respondServiceError(c, err)
		return
	}
	respondOK(c, item)
}
