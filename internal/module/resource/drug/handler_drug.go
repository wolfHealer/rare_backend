package drug

import (
	"strconv"

	"rare_backend/internal/module/resource/drug/domain"
	"rare_backend/internal/module/resource/drug/service"

	"github.com/gin-gonic/gin"
)

func parseDrugID(c *gin.Context) (uint, bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		respondBadRequest(c, "无效的药品 ID")
		return 0, false
	}
	return uint(id), true
}

func toCreateDrugInput(req CreateDrugRequest) domain.CreateDrugInput {
	return domain.CreateDrugInput{
		GenericName:      req.GenericName,
		BrandName:        req.BrandName,
		Indication:       req.Indication,
		DrugType:         req.DrugType,
		IsInsurance:      req.IsInsurance,
		DosageForm:       req.DosageForm,
		Spec:             req.Spec,
		RefPrice:         req.RefPrice,
		HasRelief:        req.HasRelief,
		IsLaunched:       req.IsLaunched,
		NeedPrescription: req.NeedPrescription,
		ManualOriginal:   req.ManualOriginal,
		ManualPopular:    req.ManualPopular,
		AuditStatus:      req.AuditStatus,
		RejectReason:     req.RejectReason,
		DiseaseIds:       req.DiseaseIds,
	}
}

func toUpdateDrugInput(req UpdateDrugRequest) domain.UpdateDrugInput {
	return domain.UpdateDrugInput{
		GenericName:      req.GenericName,
		BrandName:        req.BrandName,
		Indication:       req.Indication,
		DrugType:         req.DrugType,
		IsInsurance:      req.IsInsurance,
		DosageForm:       req.DosageForm,
		Spec:             req.Spec,
		RefPrice:         req.RefPrice,
		HasRelief:        req.HasRelief,
		IsLaunched:       req.IsLaunched,
		NeedPrescription: req.NeedPrescription,
		ManualOriginal:   req.ManualOriginal,
		ManualPopular:    req.ManualPopular,
		AuditStatus:      req.AuditStatus,
		RejectReason:     req.RejectReason,
		DiseaseIds:       req.DiseaseIds,
		HasDiseaseIds:    req.DiseaseIds != nil,
	}
}

func CreateDrug(c *gin.Context) {
	var req CreateDrugRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondBindError(c, err)
		return
	}
	if req.GenericName == "" {
		respondBadRequest(c, "通用名不能为空")
		return
	}
	if req.DrugType == "" {
		respondBadRequest(c, "药品类型不能为空")
		return
	}

	id, err := drugSvc.Create(toCreateDrugInput(req))
	if err != nil {
		if err == domain.ErrDrugRelFailed {
			respondDrugCreateError(c, "新增药品成功，但疾病关联失败")
			return
		}
		respondDrugCreateError(c, "新增药品失败")
		return
	}
	respondOK(c, gin.H{"id": id})
}

func UpdateDrug(c *gin.Context) {
	id, ok := parseDrugID(c)
	if !ok {
		return
	}

	var req UpdateDrugRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondBindError(c, err)
		return
	}

	if err := drugSvc.Update(id, toUpdateDrugInput(req)); err != nil {
		respondDrugError(c, err)
		return
	}
	respondOK(c, nil)
}

func DeleteDrug(c *gin.Context) {
	id, ok := parseDrugID(c)
	if !ok {
		return
	}

	if err := drugSvc.Delete(id); err != nil {
		respondDrugError(c, err)
		return
	}
	respondOK(c, nil)
}

func GetDrugList(c *gin.Context) {
	filter := service.BuildDrugListFilter(
		c.DefaultQuery("keyword", ""),
		c.DefaultQuery("drugType", ""),
		c.DefaultQuery("isInsurance", ""),
		c.DefaultQuery("hasRelief", ""),
		c.DefaultQuery("auditStatus", ""),
		c.DefaultQuery("page", "1"),
		c.DefaultQuery("pageSize", "10"),
	)

	result, err := drugSvc.List(filter)
	if err != nil {
		if err == domain.ErrCountList {
			respondDrugListCountError(c, err)
			return
		}
		respondDrugListError(c, err)
		return
	}
	respondPage(c, result.List, result.Total, result.Page, result.PageSize)
}

func GetDrugDetail(c *gin.Context) {
	id, ok := parseDrugID(c)
	if !ok {
		return
	}

	result, err := drugSvc.GetByID(id)
	if err != nil {
		respondDrugError(c, err)
		return
	}
	respondOK(c, result)
}

func DownloadManual(c *gin.Context) {
	id, ok := parseDrugID(c)
	if !ok {
		return
	}

	result, err := drugSvc.GetManual(id)
	if err != nil {
		respondDrugError(c, err)
		return
	}
	respondOK(c, result)
}

func ExportDrugs(c *gin.Context) {
	filter := service.BuildDrugExportFilter(
		c.DefaultQuery("disease", "0"),
		c.DefaultQuery("keyword", ""),
		c.DefaultQuery("type", ""),
		c.DefaultQuery("insurance", ""),
	)

	excel, filename, err := drugSvc.ExportExcel(filter)
	if err != nil {
		respondExportError(c, "查询药品数据失败")
		return
	}

	setExcelHeaders(c, filename)
	if err := excel.Write(c.Writer); err != nil {
		respondExcelWriteError(c)
		return
	}
}

func GetDrugOptions(c *gin.Context) {
	respondOK(c, drugSvc.Options())
}
