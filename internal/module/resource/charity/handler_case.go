package charity

import (
	"strconv"

	"rare_backend/internal/module/resource/charity/domain"
	"rare_backend/internal/module/resource/charity/service"

	"github.com/gin-gonic/gin"
)

func parseCaseID(c *gin.Context) (uint, bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		respondBadRequest(c, "无效的案例 ID")
		return 0, false
	}
	return uint(id), true
}

func toCreateCaseInput(req CreateCaseRequest) domain.CreateCaseInput {
	return domain.CreateCaseInput{
		DiseaseID:        req.DiseaseID,
		ProjectID:        req.ProjectID,
		CaseTitle:        req.CaseTitle,
		PatientDesc:      req.PatientDesc,
		ApplyCycle:       req.ApplyCycle,
		ActualRelief:     req.ActualRelief,
		Experience:       req.Experience,
		PitfallGuide:     req.PitfallGuide,
		CasePdf:          req.CasePdf,
		MaterialTemplate: req.MaterialTemplate,
		AuditStatus:      req.AuditStatus,
		RejectReason:     req.RejectReason,
	}
}

func toUpdateCaseInput(req UpdateCaseRequest) domain.UpdateCaseInput {
	return domain.UpdateCaseInput{
		Title:            req.Title,
		PatientDesc:      req.PatientDesc,
		DiseaseID:        req.DiseaseID,
		ActualRelief:     req.ActualRelief,
		Experience:       req.Experience,
		PitfallGuide:     req.PitfallGuide,
		ApplyCycle:       req.ApplyCycle,
		CasePdf:          req.CasePdf,
		MaterialTemplate: req.MaterialTemplate,
		ProjectID:        req.ProjectID,
		AuditStatus:      req.AuditStatus,
	}
}

func GetCases(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))

	filter := service.BuildCaseListFilter(
		c.DefaultQuery("diseaseId", ""),
		c.DefaultQuery("keyword", ""),
		c.DefaultQuery("auditStatus", ""),
		c.DefaultQuery("page", "1"),
		c.DefaultQuery("pageSize", "10"),
	)

	result, err := caseSvc.List(filter)
	if err != nil {
		if err == domain.ErrCountList {
			respondCaseListCountError(c)
			return
		}
		respondCaseListError(c)
		return
	}

	respondPage(c, result.List, result.Total, page, pageSize)
}

func GetCaseDetail(c *gin.Context) {
	id, ok := parseCaseID(c)
	if !ok {
		return
	}

	result, err := caseSvc.GetByID(id)
	if err != nil {
		if err == domain.ErrCaseNotFound {
			respondNotFound(c, "案例不存在")
			return
		}
		respondCaseDetailError(c)
		return
	}
	respondOK(c, result)
}

func GetCasePDF(c *gin.Context) {
	id, ok := parseCaseID(c)
	if !ok {
		return
	}

	result, err := caseSvc.GetPDF(id)
	if err != nil {
		switch err {
		case domain.ErrCaseNotAudited:
			respondNotFound(c, "案例不存在或未审核")
		case domain.ErrCaseNoPDF:
			respondNotFound(c, "该案例暂无 PDF 版本")
		default:
			respondCasePDFQueryError(c)
		}
		return
	}
	respondOK(c, result)
}

func GetCaseDiseases(c *gin.Context) {
	result, err := caseSvc.ListDiseaseOptions()
	if err != nil {
		respondCaseDiseaseOptionsError(c)
		return
	}
	respondOK(c, result)
}

func CreateCase(c *gin.Context) {
	var req CreateCaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondBindError(c, err)
		return
	}

	id, err := caseSvc.Create(toCreateCaseInput(req))
	if err != nil {
		respondCaseError(c, err)
		return
	}

	respondCreated(c, "创建成功", gin.H{"id": id})
}

func UpdateCase(c *gin.Context) {
	id, ok := parseCaseID(c)
	if !ok {
		return
	}

	var req UpdateCaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondBindErrorSimple(c)
		return
	}

	if err := caseSvc.Update(id, toUpdateCaseInput(req)); err != nil {
		if err == domain.ErrNoUpdateFields {
			respondBadRequest(c, "未提供更新字段")
			return
		}
		respondCaseUpdateError(c)
		return
	}

	respondOKMessage(c, "更新成功", nil)
}

func DeleteCase(c *gin.Context) {
	id, ok := parseCaseID(c)
	if !ok {
		return
	}

	if err := caseSvc.Delete(id); err != nil {
		if err == domain.ErrCaseNotFound {
			respondNotFound(c, "案例不存在")
			return
		}
		respondCaseDeleteError(c)
		return
	}

	respondOKMessage(c, "删除成功", nil)
}
