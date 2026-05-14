package medical

import (
	"github.com/gin-gonic/gin"
)

// 参数类型改为 *gin.RouterGroup 以支持嵌套
func SetupMedicalRoutes(r *gin.RouterGroup) {
	medical := r.Group("/medical")

	// 检查手册
	medical.GET("/examinations", GetExaminationList)       // 检查手册列表
	medical.GET("/examinations/:id", GetExaminationDetail) // 检查手册详情
	medical.POST("/examinations", CreateExamination)       // 新增检查手册
	medical.PUT("/examinations/:id", UpdateExamination)    // 修改检查手册
	medical.DELETE("/examinations/:id", DeleteExamination) // 删除检查手册

	// 医生相关
	medical.POST("/doctors", CreateDoctor)       // 新增医生
	medical.GET("/doctors", GetDoctorList)       // 医生列表
	medical.GET("/doctors/:id", GetDoctorDetail) // 医生详情
	medical.PUT("/doctors/:id", UpdateDoctor)    // 修改医生
	medical.DELETE("/doctors/:id", DeleteDoctor) // 删除医生

	// 医院相关
	// 新增：医院下拉选项接口
	medical.GET("/hospitals/options", GetHospitalOptions)
	medical.POST("/hospitals", CreateHospital)       // 新增医院
	medical.GET("/hospitals", GetHospitalList)       // 医院列表
	medical.GET("/hospitals/:id", GetHospitalDetail) // 医院详情
	medical.PUT("/hospitals/:id", UpdateHospital)    // 修改医院
	medical.DELETE("/hospitals/:id", DeleteHospital) // 删除医院

}
