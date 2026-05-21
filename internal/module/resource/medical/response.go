package medical

import (
	"errors"
	"fmt"

	"rare_backend/internal/module/resource/medical/domain"
	"rare_backend/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

func respondOK(c *gin.Context, data any) {
	response.OK(c, data)
}

func respondNotFound(c *gin.Context, message string) {
	response.NotFound(c, message)
}

func respondInternalError(c *gin.Context, message string) {
	response.InternalError(c, message)
}

func respondPage(c *gin.Context, list any, total int64, page, pageSize int) {
	response.Page(c, list, total, page, pageSize)
}

func respondBindError(c *gin.Context, err error) {
	response.BadRequest(c, "参数错误")
}

func respondBindErrorSimple(c *gin.Context) {
	response.BadRequest(c, "参数错误")
}

func respondInvalidHospitalID(c *gin.Context) {
	response.BadRequest(c, "无效的医院 ID")
}

func respondInvalidDoctorID(c *gin.Context) {
	response.BadRequest(c, "无效ID")
}

func respondInvalidExaminationID(c *gin.Context) {
	response.BadRequest(c, "无效ID")
}

func respondHospitalListCountError(c *gin.Context) {
	response.InternalError(c, "查询总数失败")
}

func respondHospitalListError(c *gin.Context) {
	response.InternalError(c, "查询医院列表失败")
}

func respondHospitalDetailError(c *gin.Context) {
	response.InternalError(c, "查询失败")
}

func respondHospitalCreateTxError(c *gin.Context) {
	response.InternalError(c, "数据库事务失败")
}

func respondHospitalCreateError(c *gin.Context, err error) {
	response.InternalError(c, "创建医院失败")
}

func respondHospitalRelError(c *gin.Context) {
	response.InternalError(c, "关联疾病失败")
}

func respondHospitalCommitError(c *gin.Context) {
	response.InternalError(c, "提交事务失败")
}

func respondHospitalUpdateTxError(c *gin.Context) {
	response.InternalError(c, "事务失败")
}

func respondHospitalUpdateError(c *gin.Context) {
	response.InternalError(c, "更新失败")
}

func respondHospitalRelCleanupError(c *gin.Context) {
	response.InternalError(c, "清理旧关联失败")
}

func respondHospitalRelUpdateError(c *gin.Context) {
	response.InternalError(c, "更新关联疾病失败")
}

func respondHospitalUpdateCommitError(c *gin.Context) {
	response.InternalError(c, "提交失败")
}

func respondHospitalDeleteTxError(c *gin.Context) {
	response.InternalError(c, "事务失败")
}

func respondHospitalDeleteDoctorCheckError(c *gin.Context) {
	response.InternalError(c, "查询医生关联失败")
}

func respondHospitalHasDoctors(c *gin.Context, count int64) {
	response.BadRequest(c, fmt.Sprintf("该医院下还有 %d 名在职医生，无法删除", count))
}

func respondHospitalDeleteRelError(c *gin.Context, err error) {
	response.InternalError(c, "清理疾病关联失败")
}

func respondHospitalDeleteError(c *gin.Context, err error) {
	response.InternalError(c, "删除医院失败")
}

func respondHospitalOptionsError(c *gin.Context) {
	response.InternalError(c, "查询医院选项失败")
}

func respondDoctorListCountError(c *gin.Context) {
	response.InternalError(c, "查询总数失败")
}

func respondDoctorListError(c *gin.Context, err error) {
	response.InternalError(c, "查询失败")
}

func respondDoctorDetailError(c *gin.Context, err error) {
	response.InternalError(c, "查询失败")
}

func respondDoctorCreateHospitalError(c *gin.Context) {
	response.BadRequest(c, "所属医院不存在或未通过审核")
}

func respondDoctorCreateTxError(c *gin.Context) {
	response.InternalError(c, "事务失败")
}

func respondDoctorCreateError(c *gin.Context, err error) {
	response.InternalError(c, "创建医生失败")
}

func respondDoctorRelError(c *gin.Context, err error) {
	response.InternalError(c, "关联疾病失败")
}

func respondDoctorUpdateHospitalError(c *gin.Context) {
	response.BadRequest(c, "新医院无效或未通过审核")
}

func respondDoctorUpdateError(c *gin.Context, err error) {
	response.InternalError(c, "更新失败")
}

func respondDoctorDeleteRelError(c *gin.Context, err error) {
	response.InternalError(c, "清理疾病关联失败")
}

func respondDoctorDeleteError(c *gin.Context, err error) {
	response.InternalError(c, "删除医生失败")
}

func respondExaminationListCountError(c *gin.Context) {
	response.InternalError(c, "查询总数失败")
}

func respondExaminationListError(c *gin.Context) {
	response.InternalError(c, "查询失败")
}

func respondExaminationCreateError(c *gin.Context, err error) {
	response.InternalError(c, "创建失败")
}

func respondExaminationUpdateError(c *gin.Context) {
	response.InternalError(c, "更新失败")
}

func respondExaminationRelCleanupError(c *gin.Context) {
	response.InternalError(c, "清理关联失败")
}

func respondExaminationRelUpdateError(c *gin.Context) {
	response.InternalError(c, "更新关联失败")
}

func respondExaminationDeleteRelError(c *gin.Context, err error) {
	response.InternalError(c, "清理疾病关联失败")
}

func respondExaminationDeleteError(c *gin.Context, err error) {
	response.InternalError(c, "删除检查手册失败")
}

func handleHospitalDeleteError(c *gin.Context, err error) {
	var hasDoctors *domain.HospitalHasDoctorsError
	if errors.As(err, &hasDoctors) {
		respondHospitalHasDoctors(c, hasDoctors.Count)
		return
	}
	if errors.Is(err, domain.ErrHospitalNotFound) {
		response.NotFound(c, "医院不存在")
		return
	}
	respondHospitalDeleteError(c, err)
}

func handleDoctorCreateError(c *gin.Context, err error) {
	if errors.Is(err, domain.ErrHospitalNotApproved) {
		respondDoctorCreateHospitalError(c)
		return
	}
	respondDoctorCreateError(c, err)
}

func handleDoctorUpdateError(c *gin.Context, err error) {
	if errors.Is(err, domain.ErrHospitalNotApproved) {
		respondDoctorUpdateHospitalError(c)
		return
	}
	respondDoctorUpdateError(c, err)
}
