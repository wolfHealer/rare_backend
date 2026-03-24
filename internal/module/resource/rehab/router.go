package rehab

import (
	"github.com/gin-gonic/gin"
)

// SetupRehabRoutes 注册康复模块路由
func SetupRehabRoutes(r *gin.RouterGroup) {
	// 1. 定义子模块前缀，最终路径变为 /api/resource/rehab
	rehab := r.Group("/rehab")
	// 2. 注册具体路由

	// 康复训练指南资源
	rehab.GET("/trainings/options", GetTrainingOptions)       // 筛选选项（新增）
	rehab.GET("/trainings", GetTrainingList)                  // 训练指南列表
	rehab.GET("/trainings/:id", GetTrainingDetail)            // 训练指南详情
	rehab.POST("/trainings", CreateTraining)                  // 新增训练指南（新增）
	rehab.PUT("/trainings/:id", UpdateTraining)               // 更新训练指南（新增）
	rehab.DELETE("/trainings/:id", DeleteTraining)            // 删除训练指南（新增）
	rehab.GET("/trainings/:id/resource", GetTrainingResource) // 下载训练资源

	// 居家护理手册资源
	rehab.GET("/manuals/options", GetManualOptions)      // 筛选选项（新增）
	rehab.GET("/manuals", GetCareManuals)                // 护理手册列表
	rehab.GET("/manuals/:id", GetManualDetail)           // 护理手册详情（新增）
	rehab.POST("/manuals", CreateManual)                 // 新增护理手册（新增）
	rehab.PUT("/manuals/:id", UpdateManual)              // 更新护理手册（新增）
	rehab.DELETE("/manuals/:id", DeleteManual)           // 删除护理手册（新增）
	rehab.GET("/manuals/:id/checklist", GetChecklist)    // 下载护理清单
	rehab.GET("/manuals/:id/record-form", GetRecordForm) // 下载记录表
	rehab.GET("/manuals/categories", GetCategories)      // 获取分类选项

	// 心理支持资源
	// 心理咨询机构
	rehab.GET("/psychological/orgs/options", GetPsychologicalOrgOptions) // 筛选选项（新增）
	rehab.GET("/psychological/orgs", GetPsychologicalOrgs)               // 机构列表
	rehab.GET("/psychological/orgs/:id", GetPsychologicalOrgDetail)      // 机构详情
	rehab.POST("/psychological/orgs", CreatePsychologicalOrg)            // 新增机构（新增）
	rehab.PUT("/psychological/orgs/:id", UpdatePsychologicalOrg)         // 更新机构（新增）
	rehab.DELETE("/psychological/orgs/:id", DeletePsychologicalOrg)      // 删除机构（新增）
	rehab.GET("/psychological/orgs/regions", GetPsychologicalOrgRegions) // 地区选项
	rehab.GET("/psychological/orgs/types", GetPsychologicalOrgTypes)     // 类型选项

	// // 心理疏导指南
	// rehab.GET("/psychological/guides/options", GetGuideOptions) // 筛选选项（新增）
	// rehab.GET("/psychological/guides", GetPsychologicalGuides)  // 指南列表
	// // rehab.GET("/psychological/guides/:id", GetGuideDetail)                         // 指南详情（新增）
	// rehab.POST("/psychological/guides", CreateGuide)                               // 新增指南（新增）
	// rehab.PUT("/psychological/guides/:id", UpdateGuide)                            // 更新指南（新增）
	// rehab.DELETE("/psychological/guides/:id", DeleteGuide)                         // 删除指南（新增）
	// rehab.GET("/psychological/guides/:id/download", GetPsychologicalGuideDownload) // 下载指南
	// rehab.GET("/psychological/guides/targets", GetGuideTargets)                    // 目标人群选项

	// 康复资源对接
	// 康复机构
	rehab.GET("/institutions/options", GetInstitutionOptions) // 筛选选项（新增）
	rehab.GET("/institutions", GetInstitutions)               // 机构列表
	rehab.GET("/institutions/:id", GetInstitutionDetail)      // 机构详情
	rehab.POST("/institutions", CreateInstitution)            // 新增机构（新增）
	rehab.PUT("/institutions/:id", UpdateInstitution)         // 更新机构（新增）
	rehab.DELETE("/institutions/:id", DeleteInstitution)      // 删除机构（新增）
	rehab.GET("/institutions/regions", GetRegions)            // 地区选项

	// 康复器械
	rehab.GET("/devices/options", GetDeviceOptions)       // 筛选选项（新增）
	rehab.GET("/devices", GetDevices)                     // 器械列表
	rehab.GET("/devices/:id", GetDeviceDetail)            // 器械详情（新增）
	rehab.POST("/devices", CreateDevice)                  // 新增器械（新增）
	rehab.PUT("/devices/:id", UpdateDevice)               // 更新器械（新增）
	rehab.DELETE("/devices/:id", DeleteDevice)            // 删除器械（新增）
	rehab.GET("/devices/:id/guide", GetDeviceGuide)       // 下载器械指南
	rehab.GET("/devices/categories", GetDeviceCategories) // 器械类别选项
}
