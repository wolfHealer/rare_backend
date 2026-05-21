package knowledge

import (
	"strconv"

	"rare_backend/internal/module/knowledge/domain"

	"github.com/gin-gonic/gin"
)

func parseTagID(c *gin.Context) (uint, bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		respondBadRequest(c, "无效的标签 ID")
		return 0, false
	}
	return uint(id), true
}

func CreateTag(c *gin.Context) {
	var req CreateTagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondBadRequest(c, "参数错误")
		return
	}
	id, err := tagSvc.Create(domain.CreateTagInput{
		Name: req.Name, Code: req.Code, SortOrder: req.SortOrder,
	})
	if err != nil {
		respondTagError(c, err, "")
		return
	}
	respondOK(c, gin.H{"id": id})
}

func UpdateTag(c *gin.Context) {
	id, ok := parseTagID(c)
	if !ok {
		return
	}
	var req UpdateTagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondBadRequest(c, "参数错误")
		return
	}
	err := tagSvc.Update(id, domain.UpdateTagInput{
		Name: req.Name, Code: req.Code, SortOrder: req.SortOrder, Status: req.Status,
	})
	if err != nil {
		respondTagError(c, err, "标签不存在")
		return
	}
	respondOK(c, nil)
}

func DeleteTag(c *gin.Context) {
	id, ok := parseTagID(c)
	if !ok {
		return
	}
	if err := tagSvc.Delete(id); err != nil {
		respondTagError(c, err, "标签不存在或已删除")
		return
	}
	respondOK(c, nil)
}

func GetTags(c *gin.Context) {
	list, err := tagSvc.List()
	if err != nil {
		respondInternalError(c, "查询标签列表失败")
		return
	}
	respondOK(c, list)
}
