package rehab

import (
	"github.com/gin-gonic/gin"
)

// SetupRehabRoutes 注册康复模块路由
func SetupRehabRoutes(r *gin.RouterGroup) {
	// 定义子模块前缀，最终路径变为 /api/resource/rehab
	rehab := r.Group("/rehab")
	{
		// 训练指南相关路由
		rehab.GET("/trainings", GetTrainingList)
		rehab.GET("/trainings/:id", GetTrainingDetail)
		rehab.GET("/trainings/:id/resource", GetTrainingResource)

		// 筛选字典路由
		rehab.GET("/dictionaries", GetDictionaries)

		// 护理手册相关路由（新增）
		rehab.GET("/manuals", GetCareManuals)
		rehab.GET("/manuals/checklist", GetChecklist)    // 下载护理清单（新增）
		rehab.GET("/manuals/record-form", GetRecordForm) // 下载记录表（新增）
		rehab.GET("/manuals/categories", GetCategories)  // 获取分类（新增）

		// 康复机构相关路由（新增）
		rehab.GET("/institutions", GetInstitutions)          // 获取机构列表
		rehab.GET("/institutions/:id", GetInstitutionDetail) // 获取机构详情

		// 康复器械相关路由（新增）
		rehab.GET("/devices", GetDevices)                    // 获取器械列表
		rehab.GET("/devices/:id/guide", GetDeviceGuide)      // 下载器械指南
		rehab.GET("/regions", GetRegions)                    // 获取地区选项
		rehab.GET("/device-categories", GetDeviceCategories) // 获取器械类别

		// 心理咨询机构相关路由
		rehab.GET("/psychological/organizations", GetPsychologicalOrgs)
		rehab.GET("/psychological/organizations/:id", GetPsychologicalOrgDetail)
		rehab.GET("/psychological/organizations/regions", GetPsychologicalOrgRegions) // 获取机构地区选项（新增）
		rehab.GET("/psychological/organizations/types", GetPsychologicalOrgTypes)     // 获取机构类型选项（新增）

		// 心理疏导指南相关路由
		rehab.GET("/psychological/guides", GetPsychologicalGuides)
		rehab.GET("/psychological/guides/:id/download", GetPsychologicalGuideDownload)
		rehab.GET("/psychological/guides/targets", GetGuideTargets) // 获取目标人群选项（新增）
	}
}
