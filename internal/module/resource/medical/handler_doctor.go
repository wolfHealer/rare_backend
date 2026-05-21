package medical

import (
	"strconv"

	"rare_backend/internal/module/resource/medical/domain"
	"rare_backend/internal/module/resource/medical/service"

	"github.com/gin-gonic/gin"
)

func parseDoctorID(c *gin.Context) (uint64, bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		respondInvalidDoctorID(c)
		return 0, false
	}
	return id, true
}

func GetDoctorList(c *gin.Context) {
	filter := service.BuildDoctorListFilter(
		c.DefaultQuery("keyword", ""),
		c.DefaultQuery("diseaseId", "0"),
		c.DefaultQuery("title", ""),
		c.DefaultQuery("hospitalId", ""),
		c.DefaultQuery("level", ""),
		c.DefaultQuery("auditStatus", "1"),
		c.DefaultQuery("page", "1"),
		c.DefaultQuery("pageSize", "20"),
	)

	result, err := doctorSvc.List(filter)
	if err != nil {
		if err == domain.ErrCountList {
			respondDoctorListCountError(c)
			return
		}
		respondDoctorListError(c, err)
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	respondPage(c, result.List, result.Total, page, pageSize)
}

func GetDoctorDetail(c *gin.Context) {
	id, ok := parseDoctorID(c)
	if !ok {
		return
	}

	result, err := doctorSvc.GetByID(id)
	if err != nil {
		if err == domain.ErrDoctorNotFound {
			respondNotFound(c, "医生不存在")
			return
		}
		respondDoctorDetailError(c, err)
		return
	}

	respondOK(c, gin.H{
		"id":           result.ID,
		"name":         result.Name,
		"title":        result.Title,
		"department":   result.Department,
		"goodAt":       result.GoodAt,
		"clinicTime":   result.ClinicTime,
		"contact":      result.Contact,
		"score":        result.Score,
		"commentNum":   result.CommentNum,
		"auditStatus":  result.AuditStatus,
		"rejectReason": result.RejectReason,
		"hospitalId":   result.HospitalID,
		"hospitalName": result.HospitalName,
		"diseases":     result.Diseases,
		"diseaseIds":   result.DiseaseIDs,
		"isRareNetwork": result.IsRareNetwork,
		"address":       result.Address,
		"cityCode":      result.CityCode,
		"cityName":      result.CityName,
		"districtCode":  result.DistrictCode,
		"districtName":  result.DistrictName,
		"level":         result.Level,
		"phone":         result.Phone,
		"provinceCode":  result.ProvinceCode,
		"provinceName":  result.ProvinceName,
	})
}

func CreateDoctor(c *gin.Context) {
	var req CreateDoctorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondBindError(c, err)
		return
	}

	id, err := doctorSvc.Create(toCreateDoctorInput(req))
	if err != nil {
		handleDoctorCreateError(c, err)
		return
	}
	respondOK(c, gin.H{"id": id})
}

func UpdateDoctor(c *gin.Context) {
	id, ok := parseDoctorID(c)
	if !ok {
		return
	}

	var req UpdateDoctorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondBindErrorSimple(c)
		return
	}

	if err := doctorSvc.Update(id, toUpdateDoctorInput(req)); err != nil {
		handleDoctorUpdateError(c, err)
		return
	}
	respondOK(c, nil)
}

func DeleteDoctor(c *gin.Context) {
	id, ok := parseDoctorID(c)
	if !ok {
		return
	}

	if err := doctorSvc.Delete(id); err != nil {
		if err == domain.ErrDoctorNotFound {
			respondNotFound(c, "医生不存在")
			return
		}
		respondDoctorDeleteError(c, err)
		return
	}
	respondOK(c, nil)
}
