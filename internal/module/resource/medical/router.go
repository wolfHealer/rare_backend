package medical

import (
	"rare_backend/internal/middleware"

	"github.com/gin-gonic/gin"
)

func SetupMedicalRoutes(r *gin.RouterGroup) {
	medical := r.Group("/medical")

	// 公开读
	medical.GET("/examinations", GetExaminationList)
	medical.GET("/examinations/:id", GetExaminationDetail)
	medical.GET("/doctors", GetDoctorList)
	medical.GET("/doctors/:id", GetDoctorDetail)
	medical.GET("/hospitals/options", GetHospitalOptions)
	medical.GET("/hospitals", GetHospitalList)
	medical.GET("/hospitals/:id", GetHospitalDetail)

	// 后台写（管理员）
	admin := medical.Group("")
	admin.Use(middleware.AdminWrite()...)
	admin.POST("/examinations", CreateExamination)
	admin.PUT("/examinations/:id", UpdateExamination)
	admin.DELETE("/examinations/:id", DeleteExamination)
	admin.POST("/doctors", CreateDoctor)
	admin.PUT("/doctors/:id", UpdateDoctor)
	admin.DELETE("/doctors/:id", DeleteDoctor)
	admin.POST("/hospitals", CreateHospital)
	admin.PUT("/hospitals/:id", UpdateHospital)
	admin.DELETE("/hospitals/:id", DeleteHospital)
}
