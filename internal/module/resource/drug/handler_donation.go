package drug

import (
	"net/http"
	"strconv"

	"rare_backend/internal/module/resource/drug/domain"
	"rare_backend/internal/module/resource/drug/service"

	"github.com/gin-gonic/gin"
)

func parseDonationID(c *gin.Context) (uint, bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		respondBadRequest(c, "无效的项目 ID")
		return 0, false
	}
	return uint(id), true
}

func toCreateDonationInput(req CreateDonationRequest) domain.CreateDonationInput {
	return domain.CreateDonationInput{
		DrugID:         req.DrugID,
		DiseaseValue:   req.DiseaseValue,
		Name:           req.Name,
		Organizer:      req.Organizer,
		ApplyCondition: req.ApplyCondition,
		ReliefCycle:    req.ReliefCycle,
		DrugDosage:     req.DrugDosage,
		ApplyForm:      req.ApplyForm,
		ApplyGuide:     req.ApplyGuide,
		MaterialList:   req.MaterialList,
		ProgressQuery:  req.ProgressQuery,
	}
}

func toUpdateDonationInput(req UpdateDonationRequest) domain.UpdateDonationInput {
	return domain.UpdateDonationInput{
		DrugID:         req.DrugID,
		DiseaseValue:   req.DiseaseValue,
		Name:           req.Name,
		Organizer:      req.Organizer,
		ApplyCondition: req.ApplyCondition,
		ReliefCycle:    req.ReliefCycle,
		DrugDosage:     req.DrugDosage,
		ApplyForm:      req.ApplyForm,
		ApplyGuide:     req.ApplyGuide,
		MaterialList:   req.MaterialList,
		ProgressQuery:  req.ProgressQuery,
	}
}

func GetDonationList(c *gin.Context) {
	filter := service.BuildDonationListFilter(
		c.DefaultQuery("keyword", ""),
		c.DefaultQuery("diseaseId", "0"),
		c.DefaultQuery("drugId", "0"),
		c.DefaultQuery("auditStatus", ""),
		c.DefaultQuery("organizer", ""),
		c.DefaultQuery("page", "1"),
		c.DefaultQuery("pageSize", "10"),
	)

	result, err := donationSvc.List(filter)
	if err != nil {
		if err == domain.ErrCountList {
			respondDonationListCountError(c, err)
			return
		}
		respondDonationListError(c)
		return
	}
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	respondPage(c, result.List, result.Total, result.Page, pageSize)
}

func GetDonationDetail(c *gin.Context) {
	id, ok := parseDonationID(c)
	if !ok {
		return
	}

	result, err := donationSvc.GetByID(id)
	if err != nil {
		respondDonationError(c, err)
		return
	}
	respondOK(c, result)
}

func CreateDonation(c *gin.Context) {
	var req CreateDonationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondBindErrorSimple(c)
		return
	}
	if req.Name == "" {
		respondBadRequest(c, "项目名称不能为空")
		return
	}
	if req.Organizer == "" {
		respondBadRequest(c, "主办方不能为空")
		return
	}
	if req.DrugID == 0 {
		respondBadRequest(c, "关联药品 ID 不能为空")
		return
	}

	id, err := donationSvc.Create(toCreateDonationInput(req))
	if err != nil {
		respondDonationError(c, err)
		return
	}
	respondOK(c, gin.H{"id": id})
}

func UpdateDonation(c *gin.Context) {
	id, ok := parseDonationID(c)
	if !ok {
		return
	}

	var req UpdateDonationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondBindErrorSimple(c)
		return
	}

	if err := donationSvc.Update(id, toUpdateDonationInput(req)); err != nil {
		respondDonationError(c, err)
		return
	}
	respondOK(c, nil)
}

func DeleteDonation(c *gin.Context) {
	id, ok := parseDonationID(c)
	if !ok {
		return
	}

	if err := donationSvc.Delete(id); err != nil {
		respondDonationError(c, err)
		return
	}
	respondOK(c, nil)
}

func ApplyDonation(c *gin.Context) {
	projectID, ok := parseDonationID(c)
	if !ok {
		return
	}

	var req ApplyDonationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondBindErrorSimple(c)
		return
	}
	if req.PatientName == "" || req.PatientIdCard == "" || req.ContactPhone == "" {
		respondBadRequest(c, "请填写完整信息")
		return
	}
	if len(req.PatientIdCard) != 18 {
		respondBadRequest(c, "身份证号格式错误")
		return
	}

	result, err := donationSvc.Apply(projectID, domain.ApplyDonationInput{
		UserID:         req.UserID,
		PatientName:    req.PatientName,
		PatientIdCard:  req.PatientIdCard,
		DiagnosisProof: req.DiagnosisProof,
		IncomeProof:    req.IncomeProof,
		ContactPhone:   req.ContactPhone,
	})
	if err != nil {
		respondDonationError(c, err)
		return
	}
	respondDonationApplySuccess(c, result)
}

func GetDonationProgress(c *gin.Context) {
	result, err := donationSvc.GetProgress(c.Param("id"))
	if err != nil {
		respondDonationError(c, err)
		return
	}
	respondOK(c, result)
}

func DownloadDonationGuide(c *gin.Context) {
	projectID, ok := parseDonationID(c)
	if !ok {
		return
	}

	result, err := donationSvc.GetGuide(projectID)
	if err != nil {
		respondDonationError(c, err)
		return
	}

	if result.IsLocal {
		filename := result.Name + "_申请指南.pdf"
		c.Header("Content-Type", "application/pdf")
		c.Header("Content-Disposition", "attachment; filename=\""+filename+"\"")
		c.File(result.FilePath)
		return
	}
	c.Redirect(http.StatusTemporaryRedirect, result.GuideURL)
}

func GetDonationOptions(c *gin.Context) {
	result, err := donationSvc.Options()
	if err != nil {
		respondDonationOptionsDiseaseError(c)
		return
	}
	respondOK(c, result)
}
