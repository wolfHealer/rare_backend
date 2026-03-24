package medical

import (
	"github.com/gin-gonic/gin"
)

// 参数类型改为 *gin.RouterGroup 以支持嵌套
func SetupMedicalRoutes(r *gin.RouterGroup) {
	medical := r.Group("/medical")

	// 诊疗指南
	medical.GET("/guides", GetGuideList)       // 指南列表
	medical.GET("/guides/:id", GetGuideDetail) // 指南详情
	medical.POST("/guides", CreateGuide)       // 新增指南
	medical.PUT("/guides/:id", UpdateGuide)    // 修改指南
	medical.DELETE("/guides/:id", DeleteGuide) // 删除指南

	// 检查手册
	medical.GET("/examinations", GetInspectionList)       // 检查手册列表
	medical.GET("/examinations/:id", GetInspectionDetail) // 检查手册详情
	medical.POST("/examinations", CreateInspection)       // 新增检查手册
	medical.PUT("/examinations/:id", UpdateInspection)    // 修改检查手册
	medical.DELETE("/examinations/:id", DeleteInspection) // 删除检查手册

	// 名录相关
	medical.GET("/directory", GetDirectoryList)
	medical.GET("/directory/export/doctors", ExportDoctors)
	medical.GET("/directory/export/hospitals", ExportHospitals)

	// 医生相关
	medical.POST("/doctors", CreateDoctor)       // 新增医生（新增）
	medical.GET("/doctors", GetDoctorList)       // 医生列表
	medical.GET("/doctors/:id", GetDoctorDetail) // 医生详情
	medical.PUT("/doctors/:id", UpdateDoctor)    // 修改医生（新增）
	medical.DELETE("/doctors/:id", DeleteDoctor) // 删除医生（新增）

	// 医院相关
	medical.POST("/hospitals", CreateHospital)       // 新增医院（新增）
	medical.GET("/hospitals", GetHospitalList)       // 医院列表
	medical.GET("/hospitals/:id", GetHospitalDetail) // 医院详情
	medical.PUT("/hospitals/:id", UpdateHospital)    // 修改医院（新增）
	medical.DELETE("/hospitals/:id", DeleteHospital) // 删除医院（新增）
}
