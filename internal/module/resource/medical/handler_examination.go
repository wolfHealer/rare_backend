package medical

import (
	"strconv"

	"rare_backend/internal/module/resource/medical/domain"
	"rare_backend/internal/module/resource/medical/service"

	"github.com/gin-gonic/gin"
)

func parseExaminationID(c *gin.Context) (uint64, bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		respondInvalidExaminationID(c)
		return 0, false
	}
	return id, true
}

func GetExaminationList(c *gin.Context) {
	filter := service.BuildExaminationListFilter(
		c.DefaultQuery("keyword", ""),
		c.DefaultQuery("examType", ""),
		c.DefaultQuery("auditStatus", "1"),
		c.DefaultQuery("diseaseId", "0"),
		c.DefaultQuery("page", "1"),
		c.DefaultQuery("pageSize", "10"),
	)

	result, err := examinationSvc.List(filter)
	if err != nil {
		if err == domain.ErrCountList {
			respondExaminationListCountError(c)
			return
		}
		respondExaminationListError(c)
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	respondPage(c, result.List, result.Total, page, pageSize)
}

func GetExaminationDetail(c *gin.Context) {
	id, ok := parseExaminationID(c)
	if !ok {
		return
	}

	result, err := examinationSvc.GetByID(id)
	if err != nil {
		if err == domain.ErrExaminationNotFound {
			respondNotFound(c, "不存在")
			return
		}
		respondInternalError(c, "查询失败")
		return
	}

	respondOK(c, gin.H{
		"id":                result.ID,
		"examName":          result.ExamName,
		"examType":          result.ExamType,
		"examPurpose":       result.ExamPurpose,
		"referenceValue":    result.ReferenceValue,
		"abnormalInterpret": result.AbnormalInterpret,
		"sampleNotes":       result.SampleNotes,
		"institution":       result.Institution,
		"templates": gin.H{
			"excel":   result.Templates.Excel,
			"word":    result.Templates.Word,
			"compare": result.Templates.Compare,
		},
		"auditStatus":  result.AuditStatus,
		"rejectReason": result.RejectReason,
		"sort":         result.Sort,
		"diseaseIds":   result.DiseaseIDs,
		"createdAt":    result.CreatedAt,
	})
}

func CreateExamination(c *gin.Context) {
	var req CreateExaminationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondBindError(c, err)
		return
	}

	id, err := examinationSvc.Create(toCreateExaminationInput(req))
	if err != nil {
		respondExaminationCreateError(c, err)
		return
	}
	respondOK(c, gin.H{"id": id})
}

func UpdateExamination(c *gin.Context) {
	id, ok := parseExaminationID(c)
	if !ok {
		return
	}

	var req UpdateExaminationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondBindErrorSimple(c)
		return
	}

	if err := examinationSvc.Update(id, toUpdateExaminationInput(req)); err != nil {
		respondExaminationUpdateError(c)
		return
	}
	respondOK(c, nil)
}

func DeleteExamination(c *gin.Context) {
	id, ok := parseExaminationID(c)
	if !ok {
		return
	}

	if err := examinationSvc.Delete(id); err != nil {
		if err == domain.ErrExaminationNotFound {
			respondNotFound(c, "检查手册不存在")
			return
		}
		respondExaminationDeleteError(c, err)
		return
	}
	respondOK(c, nil)
}
