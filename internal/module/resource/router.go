package resource

import (
	"rare_backend/internal/module/resource/medical" // 引入子模块

	"rare_backend/internal/module/resource/drug"  // 未来扩展
	"rare_backend/internal/module/resource/rehab" // 未来扩展

	"github.com/gin-gonic/gin"

	"rare_backend/internal/module/resource/charity" // 未来扩展
)

func SetupResourceRoutes(r *gin.RouterGroup) {
	// 1. 定义 resource 模块的统一前缀
	resource := r.Group("/resource")

	// 2. 初始化子模块路由，将 resource 组传递进去
	medical.SetupMedicalRoutes(resource)
	charity.SetupCharityRoutes(resource) // 未来扩展
	drug.SetupDrugRoutes(resource)       // 未来扩展
	rehab.SetupRehabRoutes(resource)

}
