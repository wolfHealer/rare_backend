package charity

import (
	"strconv"

	"rare_backend/internal/module/resource/charity/domain"
	"rare_backend/internal/module/resource/charity/service"

	"github.com/gin-gonic/gin"
)

func parseProjectID(c *gin.Context) (uint, bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		respondBadRequest(c, "无效的项目 ID")
		return 0, false
	}
	return uint(id), true
}

func toCreateProjectInput(req CreateProjectRequest) domain.CreateProjectInput {
	return domain.CreateProjectInput{
		Name:            req.Name,
		ApplyProcess:    req.ApplyProcess,
		ApplyCondition:  req.ApplyCondition,
		ReliefType:      req.ReliefType,
		DiseaseIDs:      req.DiseaseIDs,
		ReliefStandard:  req.ReliefStandard,
		ApplyDifficulty: req.ApplyDifficulty,
		ApplyDeadline:   req.ApplyDeadline,
		ContactPhone:    req.ContactPhone,
		ContactURL:      req.ContactURL,
		ApplyForm:       req.ApplyForm,
		ApplyGuide:      req.ApplyGuide,
		MaterialList:    req.MaterialList,
		Organizer:       req.Organizer,
		Sort:            req.Sort,
	}
}

func toUpdateProjectInput(req UpdateProjectRequest) domain.UpdateProjectInput {
	return domain.UpdateProjectInput{
		Name:            req.Name,
		ApplyProcess:    req.ApplyProcess,
		ApplyCondition:  req.ApplyCondition,
		ReliefType:      req.ReliefType,
		Type:            req.Type,
		DiseaseIDs:      req.DiseaseIDs,
		ReliefStandard:  req.ReliefStandard,
		ApplyDifficulty: req.ApplyDifficulty,
		ApplyDeadline:   req.ApplyDeadline,
		ContactPhone:    req.ContactPhone,
		ContactURL:      req.ContactURL,
		ApplyForm:       req.ApplyForm,
		ApplyGuide:      req.ApplyGuide,
		MaterialList:    req.MaterialList,
		Organizer:       req.Organizer,
		Sort:            req.Sort,
		AuditStatus:     req.AuditStatus,
		RejectReason:    req.RejectReason,
		HasDiseaseIDs:   req.DiseaseIDs != nil,
	}
}

func ListProjects(c *gin.Context) {
	filter := service.BuildProjectListFilter(
		c.DefaultQuery("reliefType", ""),
		c.DefaultQuery("diseaseId", ""),
		c.DefaultQuery("applyDifficulty", ""),
		c.DefaultQuery("auditStatus", ""),
		c.DefaultQuery("keyword", ""),
		c.DefaultQuery("page", "1"),
		c.DefaultQuery("pageSize", "10"),
	)

	result, err := projectSvc.List(filter)
	if err != nil {
		if err == domain.ErrCountList {
			respondProjectListCountError(c)
			return
		}
		respondProjectListError(c)
		return
	}
	respondPage(c, result.List, result.Total, result.Page, result.PageSize)
}

func GetProjectDetail(c *gin.Context) {
	id, ok := parseProjectID(c)
	if !ok {
		return
	}

	result, err := projectSvc.GetByID(id)
	if err != nil {
		if err == domain.ErrProjectNotFound {
			respondNotFound(c, "项目不存在")
			return
		}
		respondProjectDetailError(c, err)
		return
	}
	respondOK(c, result)
}

func CreateProject(c *gin.Context) {
	var req CreateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondBindError(c, err)
		return
	}

	id, err := projectSvc.Create(toCreateProjectInput(req))
	if err != nil {
		respondProjectCreateError(c, "创建项目失败")
		return
	}

	respondCreated(c, "创建成功", gin.H{"id": id})
}

func UpdateProject(c *gin.Context) {
	id, ok := parseProjectID(c)
	if !ok {
		return
	}

	var req UpdateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondBindError(c, err)
		return
	}

	if err := projectSvc.Update(id, toUpdateProjectInput(req)); err != nil {
		respondProjectUpdateError(c, err)
		return
	}

	respondOKMessage(c, "更新成功", nil)
}

func DeleteProject(c *gin.Context) {
	id, ok := parseProjectID(c)
	if !ok {
		return
	}

	if err := projectSvc.Delete(id); err != nil {
		if err == domain.ErrProjectNotFound {
			respondNotFound(c, "项目不存在")
			return
		}
		respondProjectDeleteError(c, "删除项目失败")
		return
	}

	respondOKMessage(c, "删除成功", nil)
}

func GetProjectOptions(c *gin.Context) {
	filter := service.BuildProjectOptionsFilter(
		c.DefaultQuery("keyword", ""),
		c.DefaultQuery("auditStatus", ""),
		c.DefaultQuery("page", "1"),
		c.DefaultQuery("pageSize", "20"),
	)

	result, err := projectSvc.ListOptions(filter)
	if err != nil {
		respondProjectOptionsError(c)
		return
	}
	respondOK(c, result)
}
