package rehab

import (
	"errors"
	"strconv"

	"rare_backend/internal/module/resource/rehab/domain"
	"rare_backend/internal/module/resource/rehab/service"

	"github.com/gin-gonic/gin"
)

func parseInstitutionID(c *gin.Context) (uint64, bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		respondInvalidInstitutionID(c)
		return 0, false
	}
	return id, true
}

func GetInstitutions(c *gin.Context) {
	filter := service.BuildInstitutionListFilter(
		c.DefaultQuery("provinceCode", ""),
		c.DefaultQuery("cityCode", ""),
		c.DefaultQuery("districtCode", ""),
		c.DefaultQuery("diseaseId", ""),
		c.DefaultQuery("keyword", ""),
		c.DefaultQuery("auditStatus", ""),
		c.DefaultQuery("page", "1"),
		c.DefaultQuery("pageSize", "10"),
	)

	result, err := institutionSvc.List(filter)
	if err != nil {
		var countErr *domain.CountListError
		if errors.As(err, &countErr) {
			respondInstitutionListCountError(c, countErr.Cause)
			return
		}
		var queryErr *domain.QueryListError
		if errors.As(err, &queryErr) {
			respondInstitutionListError(c, queryErr.Cause)
			return
		}
		respondInstitutionListCountError(c, err)
		return
	}

	var list []InstitutionItem
	for _, item := range result.List {
		list = append(list, InstitutionItem{
			ID:            item.ID,
			Name:          item.Name,
			ProvinceCode:  item.ProvinceCode,
			CityCode:      item.CityCode,
			DistrictCode:  item.DistrictCode,
			ProvinceName:  item.ProvinceName,
			CityName:      item.CityName,
			DistrictName:  item.DistrictName,
			Address:       item.Address,
			ContactPhone:  item.ContactPhone,
			ContactUrl:    item.ContactUrl,
			Qualification: item.Qualification,
			RehabProjects: item.RehabProjects,
			FeeStandard:   item.FeeStandard,
			DiseaseIds:    item.DiseaseIds,
			AuditStatus:   item.AuditStatus,
			Rating:        item.Rating,
			Status:        item.Status,
			UpdateAt:      item.UpdateAt,
		})
	}

	respondOK(c, InstitutionListResponse{
		List:     list,
		Total:    result.Total,
		Page:     result.Page,
		PageSize: result.PageSize,
	})
}

func GetInstitutionDetail(c *gin.Context) {
	id, ok := parseInstitutionID(c)
	if !ok {
		return
	}

	result, err := institutionSvc.GetByID(id)
	if err != nil {
		handleInstitutionDetailError(c, err)
		return
	}

	var doctors []DoctorItem
	for _, d := range result.Doctors {
		doctors = append(doctors, DoctorItem{Name: d.Name, Title: d.Title, Specialty: d.Specialty})
	}

	respondOK(c, gin.H{
		"id":            result.ID,
		"name":          result.Name,
		"type":          result.Type,
		"typeName":      result.TypeName,
		"provinceCode":  result.ProvinceCode,
		"cityCode":      result.CityCode,
		"districtCode":  result.DistrictCode,
		"provinceName":  result.ProvinceName,
		"cityName":      result.CityName,
		"districtName":  result.DistrictName,
		"address":       result.Address,
		"contactPhone":  result.ContactPhone,
		"contactUrl":    result.ContactUrl,
		"qualification": result.Qualification,
		"rehabProjects": result.RehabProjects,
		"feeStandard":   result.FeeStandard,
		"services":      result.Services,
		"diseaseIds":    result.DiseaseIds,
		"diseases":      toInstitutionDiseaseItems(result.Diseases),
		"auditStatus":   result.AuditStatus,
		"rejectReason":  result.RejectReason,
		"createdAt":     result.CreatedAt,
		"updatedAt":     result.UpdatedAt,
		"rating":        result.Rating,
		"isInsurance":   result.IsInsurance,
		"coverUrl":      result.CoverUrl,
		"images":        result.Images,
		"businessHours": result.BusinessHours,
		"facilities":    result.Facilities,
		"doctors":       doctors,
		"status":        result.Status,
	})
}

func GetInstitutionOptions(c *gin.Context) {
	result, err := institutionSvc.GetOptions()
	if err != nil {
		handleInstitutionOptionsError(c, err)
		return
	}

	var regions []RegionItem
	for _, r := range result.Regions {
		regions = append(regions, RegionItem{Text: r.Text, Value: r.Value})
	}
	var types []OrgTypeItem
	for _, t := range result.Types {
		types = append(types, OrgTypeItem{Text: t.Text, Value: t.Value})
	}
	var diseases []OptionItem
	for _, d := range result.Diseases {
		diseases = append(diseases, OptionItem{Text: d.Text, Value: d.Value})
	}

	respondOK(c, InstitutionOptionsResponse{
		Regions:  regions,
		Types:    types,
		Diseases: diseases,
	})
}

func GetRegions(c *gin.Context) {
	result := institutionSvc.GetRegions()
	var regions []RegionItem
	for _, r := range result.Regions {
		regions = append(regions, RegionItem{Text: r.Text, Value: r.Value})
	}
	respondOK(c, RegionResponse{Regions: regions})
}

func CreateInstitution(c *gin.Context) {
	var req CreateInstitutionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondBindError(c, err)
		return
	}

	id, err := institutionSvc.Create(toCreateInstitutionInput(req))
	if err != nil {
		respondInstitutionCreateError(c, err)
		return
	}

	respondOKMessage(c, "创建成功", gin.H{"id": id})
}

func UpdateInstitution(c *gin.Context) {
	id, ok := parseInstitutionID(c)
	if !ok {
		return
	}

	var req UpdateInstitutionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondInstitutionBindParseError(c, err)
		return
	}

	if err := institutionSvc.Update(id, toUpdateInstitutionInput(req)); err != nil {
		respondInstitutionUpdateError(c, err)
		return
	}

	respondOKMessage(c, "更新成功", nil)
}

func DeleteInstitution(c *gin.Context) {
	id, ok := parseInstitutionID(c)
	if !ok {
		return
	}

	if err := institutionSvc.Delete(id); err != nil {
		respondInstitutionDeleteError(c)
		return
	}

	respondOKMessage(c, "删除成功", nil)
}
