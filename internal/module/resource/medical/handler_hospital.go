package medical

import (
	"strconv"

	"rare_backend/internal/module/resource/medical/domain"
	"rare_backend/internal/module/resource/medical/service"

	"github.com/gin-gonic/gin"
)

func parseHospitalID(c *gin.Context) (uint64, bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		respondInvalidHospitalID(c)
		return 0, false
	}
	return id, true
}

func GetHospitalList(c *gin.Context) {
	filter := service.BuildHospitalListFilter(
		c.DefaultQuery("keyword", ""),
		c.DefaultQuery("provinceCode", ""),
		c.DefaultQuery("cityCode", ""),
		c.DefaultQuery("districtCode", ""),
		c.DefaultQuery("level", ""),
		c.DefaultQuery("isRareNetwork", ""),
		c.DefaultQuery("auditStatus", ""),
		c.DefaultQuery("page", "1"),
		c.DefaultQuery("pageSize", "20"),
	)

	result, err := hospitalSvc.List(filter)
	if err != nil {
		if err == domain.ErrCountList {
			respondHospitalListCountError(c)
			return
		}
		respondHospitalListError(c)
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	respondPage(c, result.List, result.Total, page, pageSize)
}

func GetHospitalDetail(c *gin.Context) {
	id, ok := parseHospitalID(c)
	if !ok {
		return
	}

	result, err := hospitalSvc.GetByID(id)
	if err != nil {
		if err == domain.ErrHospitalNotFound {
			respondNotFound(c, "医院不存在")
			return
		}
		respondHospitalDetailError(c)
		return
	}

	respondOK(c, gin.H{
		"id":            result.ID,
		"name":          result.Name,
		"provinceCode":  result.ProvinceCode,
		"cityCode":      result.CityCode,
		"districtCode":  result.DistrictCode,
		"provinceName":  result.ProvinceName,
		"cityName":      result.CityName,
		"districtName":  result.DistrictName,
		"level":         result.Level,
		"address":       result.Address,
		"phone":         result.Phone,
		"hospitalUrl":   result.HospitalURL,
		"treatScope":    result.TreatScope,
		"isRareNetwork": result.IsRareNetwork,
		"auditStatus":   result.AuditStatus,
		"rejectReason":  result.RejectReason,
		"diseases":      result.Diseases,
		"createdAt":     result.CreatedAt,
	})
}

func GetHospitalOptions(c *gin.Context) {
	options, err := hospitalSvc.ListOptions(c.DefaultQuery("keyword", ""))
	if err != nil {
		respondHospitalOptionsError(c)
		return
	}
	respondOK(c, options)
}

func CreateHospital(c *gin.Context) {
	var req CreateHospitalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondBindError(c, err)
		return
	}

	id, err := hospitalSvc.Create(toCreateHospitalInput(req))
	if err != nil {
		respondHospitalCreateError(c, err)
		return
	}
	respondOK(c, gin.H{"id": id})
}

func UpdateHospital(c *gin.Context) {
	id, ok := parseHospitalID(c)
	if !ok {
		return
	}

	var req UpdateHospitalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondBindErrorSimple(c)
		return
	}

	if err := hospitalSvc.Update(id, toUpdateHospitalInput(req)); err != nil {
		if err == domain.ErrHospitalNotFound {
			respondNotFound(c, "医院不存在")
			return
		}
		respondHospitalUpdateError(c)
		return
	}
	respondOK(c, nil)
}

func DeleteHospital(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		respondInvalidHospitalID(c)
		return
	}

	if err := hospitalSvc.Delete(id); err != nil {
		handleHospitalDeleteError(c, err)
		return
	}
	respondOK(c, nil)
}
