package medicare

import (
	"github.com/gin-gonic/gin"
)

// 参数类型改为 *gin.RouterGroup 以支持嵌套
func SetupMedicareRoutes(r *gin.RouterGroup) {
	// 1. 定义子模块前缀，最终路径变为 /api/resource/medicare
	medicare := r.Group("/medicare")
	// 2. 注册具体路由

	// 医保政策资源
	medicare.GET("/policies", GetPolicyList)                   // 政策列表
	medicare.GET("/policies/:id", GetPolicyDetail)             // 政策详情
	medicare.GET("/policies/:id/materials", DownloadMaterials) // 下载资料
}
