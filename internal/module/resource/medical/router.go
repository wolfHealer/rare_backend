package medical

import (
	"github.com/gin-gonic/gin"
)

// 参数类型改为 *gin.RouterGroup 以支持嵌套
func SetupMedicalRoutes(r *gin.RouterGroup) {
	// 1. 定义子模块前缀，最终路径变为 /api/resource/medical
	medical := r.Group("/medical")
	// 2. 注册具体路由
	// 注意：确保 handler 函数已正确初始化或为包级函数
	medical.GET("/guides", GetGuideList)
	medical.GET("/guides/:id", GetGuideDetail)
	// 检查手册
	medical.GET("/inspections", GetInspectionList)
	medical.GET("/inspections/:id", GetInspectionDetail) // 新增

	// 名录相关
	medical.GET("/directory", GetDirectoryList)
	medical.GET("/directory/export/doctors", ExportDoctors)
	medical.GET("/directory/export/hospitals", ExportHospitals)

	medical.GET("/directory/doctor/:id", GetDoctorDetail)     // 新增医生详情
	medical.GET("/directory/hospital/:id", GetHospitalDetail) // 新增医院详情

}
