package rehab

import (
	"errors"
	"strconv"

	"rare_backend/internal/module/resource/rehab/domain"
	"rare_backend/internal/module/resource/rehab/service"

	"github.com/gin-gonic/gin"
)

func parsePsychOrgID(c *gin.Context) (uint64, bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		respondInvalidInstitutionID(c)
		return 0, false
	}
	return id, true
}

func GetPsychologicalOrgs(c *gin.Context) {
	filter := service.BuildPsychOrgListFilter(
		c.DefaultQuery("provinceCode", ""),
		c.DefaultQuery("cityCode", ""),
		c.DefaultQuery("districtCode", ""),
		c.DefaultQuery("consultWay", ""),
		c.DefaultQuery("diseaseId", ""),
		c.DefaultQuery("isFree", ""),
		c.DefaultQuery("keyword", ""),
		c.DefaultQuery("auditStatus", ""),
		c.DefaultQuery("page", "1"),
		c.DefaultQuery("pageSize", "10"),
	)

	result, err := psychologicalSvc.List(filter)
	if err != nil {
		var countErr *domain.CountListError
		if errors.As(err, &countErr) {
			respondPsychOrgListCountError(c, countErr.Cause)
			return
		}
		var queryErr *domain.QueryListError
		if errors.As(err, &queryErr) {
			respondPsychOrgListError(c, queryErr.Cause)
			return
		}
		respondPsychOrgListCountError(c, err)
		return
	}

	var list []PsychologicalOrgItem
	for _, item := range result.List {
		list = append(list, PsychologicalOrgItem{
			ID:           item.ID,
			Name:         item.Name,
			ProvinceCode: item.ProvinceCode,
			CityCode:     item.CityCode,
			DistrictCode: item.DistrictCode,
			ProvinceName: item.ProvinceName,
			CityName:     item.CityName,
			DistrictName: item.DistrictName,
			Address:      item.Address,
			ContactPhone: item.ContactPhone,
			ContactUrl:   item.ContactUrl,
			IsFree:       item.IsFree,
			ConsultWay:   item.ConsultWay,
			ContentIntro: item.ContentIntro,
			AuditStatus:  item.AuditStatus,
			RejectReason: item.RejectReason,
			Type:         item.Type,
			TypeName:     item.TypeName,
			Region:       item.Region,
			RegionCode:   item.RegionCode,
			ServiceTime:  item.ServiceTime,
			Description:  item.Description,
			Services:     item.Services,
			Rating:       item.Rating,
			CoverUrl:     item.CoverUrl,
			Status:       item.Status,
			DiseaseIds:   item.DiseaseIds,
			Diseases:     toInstitutionDiseaseItems(item.Diseases),
			CreatedAt:    item.CreatedAt,
			UpdatedAt:    item.UpdatedAt,
		})
	}

	respondOK(c, PsychologicalOrgListResponse{
		List:     list,
		Total:    result.Total,
		Page:     result.Page,
		PageSize: result.PageSize,
	})
}

func GetPsychologicalOrgDetail(c *gin.Context) {
	id, ok := parsePsychOrgID(c)
	if !ok {
		return
	}

	result, err := psychologicalSvc.GetByID(id)
	if err != nil {
		handlePsychOrgDetailError(c, err)
		return
	}

	var counselors []CounselorItem
	for _, c := range result.Counselors {
		counselors = append(counselors, CounselorItem{Name: c.Name, Title: c.Title, Specialty: c.Specialty})
	}

	respondOK(c, PsychologicalOrgDetailResponse{
		ID:           result.ID,
		Name:         result.Name,
		ProvinceCode: result.ProvinceCode,
		CityCode:     result.CityCode,
		DistrictCode: result.DistrictCode,
		ProvinceName: result.ProvinceName,
		CityName:     result.CityName,
		DistrictName: result.DistrictName,
		Address:      result.Address,
		ContactPhone: result.ContactPhone,
		ContactUrl:   result.ContactUrl,
		IsFree:       result.IsFree,
		ConsultWay:   result.ConsultWay,
		ContentIntro: result.ContentIntro,
		AuditStatus:  result.AuditStatus,
		RejectReason: result.RejectReason,
		Type:         result.Type,
		TypeName:     result.TypeName,
		Region:       result.Region,
		RegionCode:   result.RegionCode,
		ServiceTime:  result.ServiceTime,
		Description:  result.Description,
		Services:     result.Services,
		Rating:       result.Rating,
		CoverUrl:     result.CoverUrl,
		Status:       result.Status,
		Images:       result.Images,
		Counselors:   counselors,
		DiseaseIds:   result.DiseaseIds,
		Diseases:     toInstitutionDiseaseItems(result.Diseases),
		DiseaseCount: result.DiseaseCount,
		CreatedAt:    result.CreatedAt,
		UpdatedAt:    result.UpdatedAt,
	})
}

func GetGuideTargets(c *gin.Context) {
	result := psychologicalSvc.GetGuideTargets()
	var targets []TargetItem
	for _, t := range result.Targets {
		targets = append(targets, TargetItem{Text: t.Text, Value: t.Value})
	}
	respondOK(c, TargetResponse{Targets: targets})
}

func GetPsychologicalOrgTypes(c *gin.Context) {
	result := psychologicalSvc.GetOrgTypes()
	var types []OrgTypeItem
	for _, t := range result.Types {
		types = append(types, OrgTypeItem{Text: t.Text, Value: t.Value})
	}
	respondOK(c, OrgTypeResponse{Types: types})
}

func CreatePsychologicalOrg(c *gin.Context) {
	var req CreatePsychologicalOrgRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondBindError(c, err)
		return
	}

	id, err := psychologicalSvc.Create(toCreatePsychOrgInput(req))
	if err != nil {
		respondPsychOrgCreateError(c)
		return
	}

	respondOKMessage(c, "创建成功", gin.H{"id": id})
}

func UpdatePsychologicalOrg(c *gin.Context) {
	id, ok := parsePsychOrgID(c)
	if !ok {
		return
	}

	var req UpdatePsychologicalOrgRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondPsychOrgBindError(c)
		return
	}

	if err := psychologicalSvc.Update(id, toUpdatePsychOrgInput(req)); err != nil {
		respondPsychOrgUpdateError(c)
		return
	}

	respondOKMessage(c, "更新成功", nil)
}

func DeletePsychologicalOrg(c *gin.Context) {
	id, ok := parsePsychOrgID(c)
	if !ok {
		return
	}

	if err := psychologicalSvc.Delete(id); err != nil {
		respondInstitutionDeleteRelError(c)
		return
	}

	respondOKMessage(c, "删除成功", nil)
}
