package rehab

import (
	"github.com/gin-gonic/gin"
)

// SetupRehabRoutes 注册康复模块路由
func SetupRehabRoutes(r *gin.RouterGroup) {
	// 1. 定义子模块前缀，最终路径变为 /api/resource/rehab
	rehab := r.Group("/rehab")
	// 2. 注册具体路由

	// 心理咨询机构
	rehab.GET("/psychological/orgs", GetPsychologicalOrgs)          // 机构列表
	rehab.GET("/psychological/orgs/:id", GetPsychologicalOrgDetail) // 机构详情
	rehab.POST("/psychological/orgs", CreatePsychologicalOrg)       // 新增机构
	rehab.PUT("/psychological/orgs/:id", UpdatePsychologicalOrg)    // 更新机构
	rehab.DELETE("/psychological/orgs/:id", DeletePsychologicalOrg) // 删除机构

	// 康复机构
	rehab.GET("/institutions/options", GetInstitutionOptions) // 筛选选项
	rehab.GET("/institutions", GetInstitutions)               // 机构列表
	rehab.GET("/institutions/:id", GetInstitutionDetail)      // 机构详情
	rehab.POST("/institutions", CreateInstitution)            // 新增机构
	rehab.PUT("/institutions/:id", UpdateInstitution)         // 更新机构
	rehab.DELETE("/institutions/:id", DeleteInstitution)      // 删除机构
	rehab.GET("/institutions/regions", GetRegions)            // 地区选项

	// 康复训练指南资源
	rehab.GET("/trainings", GetTrainingList)                  // 训练指南列表
	rehab.GET("/trainings/:id", GetTrainingDetail)            // 训练指南详情
	rehab.POST("/trainings", CreateTraining)                  // 新增训练指南
	rehab.PUT("/trainings/:id", UpdateTraining)               // 更新训练指南
	rehab.DELETE("/trainings/:id", DeleteTraining)            // 删除训练指南
	rehab.GET("/trainings/:id/resource", GetTrainingResource) // 下载训练资源

}
