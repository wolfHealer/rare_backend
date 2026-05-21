package rehab

import (
	"rare_backend/internal/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRehabRoutes(r *gin.RouterGroup) {
	rehab := r.Group("/rehab")

	// 公开读
	rehab.GET("/psychological/orgs", GetPsychologicalOrgs)
	rehab.GET("/psychological/orgs/:id", GetPsychologicalOrgDetail)
	rehab.GET("/institutions/options", GetInstitutionOptions)
	rehab.GET("/institutions", GetInstitutions)
	rehab.GET("/institutions/:id", GetInstitutionDetail)
	rehab.GET("/institutions/regions", GetRegions)
	rehab.GET("/trainings", GetTrainingList)
	rehab.GET("/trainings/:id", GetTrainingDetail)
	rehab.GET("/trainings/:id/resource", GetTrainingResource)

	// 后台写（管理员）
	admin := rehab.Group("")
	admin.Use(middleware.AdminWrite()...)
	admin.POST("/psychological/orgs", CreatePsychologicalOrg)
	admin.PUT("/psychological/orgs/:id", UpdatePsychologicalOrg)
	admin.DELETE("/psychological/orgs/:id", DeletePsychologicalOrg)
	admin.POST("/institutions", CreateInstitution)
	admin.PUT("/institutions/:id", UpdateInstitution)
	admin.DELETE("/institutions/:id", DeleteInstitution)
	admin.POST("/trainings", CreateTraining)
	admin.PUT("/trainings/:id", UpdateTraining)
	admin.DELETE("/trainings/:id", DeleteTraining)
}
