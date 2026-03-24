// package charity

// import (
// 	"github.com/gin-gonic/gin"
// )

// // 参数类型改为 *gin.RouterGroup 以支持嵌套
// func SetupCharityRoutes(r *gin.RouterGroup) {
// 	// 1. 定义子模块前缀，最终路径变为 /api/resource/medical
// 	charity := r.Group("/charity")
// 	// 2. 注册具体路由
// 	// 注意：确保 handler 函数已正确初始化或为包级函数
// 	// 救助项目相关
// 	charity.GET("/projects", ListProjects)
// 	charity.GET("/projects/:id", GetProjectDetail)
// 	charity.GET("/filters", GetFilters)

// 	//医保政策解读
// 	charity.GET("/policies", GetPolicyList)               //获取政策列表
// 	charity.GET("/policies/regions", GetRegions)          //获取地区选项
// 	charity.GET("/policies/detail/:id", GetPolicyDetail)  //获取政策详情
// 	charity.GET("/policies/materials", DownloadMaterials) //下载资料

// 	// 求助通道相关（新增）
// 	charity.GET("/channels", GetChannels)                 // 获取求助渠道列表
// 	charity.GET("/channels/detail/:id", GetChannelDetail) // 获取求助渠道详情
// 	charity.GET("/channels/template", GetTemplates)       // 获取求助模板
// 	charity.POST("/channels/help/submit", SubmitHelp)     // 提交求助请求

// 	// 救助案例相关（新增）
// 	charity.GET("/cases", GetCases)                 // 获取案例列表
// 	charity.GET("/cases/:id", GetCaseDetail)        // 获取案例详情
// 	charity.GET("/cases/:id/pdf", GetCasePDF)       // 下载案例 PDF（新增）
// 	charity.GET("/cases/diseases", GetCaseDiseases) // 获取疾病选项（新增）

// }

package charity

import (
	"github.com/gin-gonic/gin"
)

// 参数类型改为 *gin.RouterGroup 以支持嵌套
func SetupCharityRoutes(r *gin.RouterGroup) {
	// 1. 定义子模块前缀，最终路径变为 /api/resource/charity
	charity := r.Group("/charity")
	// 2. 注册具体路由

	// 救助项目资源
	charity.GET("/projects", ListProjects)         // 项目列表
	charity.GET("/projects/:id", GetProjectDetail) // 项目详情
	charity.POST("/projects", CreateProject)       // 新增项目（补充）
	charity.PUT("/projects/:id", UpdateProject)    // 更新项目（补充）
	charity.DELETE("/projects/:id", DeleteProject) // 删除项目（补充）
	charity.GET("/projects/filters", GetFilters)   // 筛选选项

	// 医保政策资源
	charity.GET("/policies", GetPolicyList)                   // 政策列表
	charity.GET("/policies/regions", GetRegions)              // 地区选项
	charity.GET("/policies/:id", GetPolicyDetail)             // 政策详情（路径优化）
	charity.GET("/policies/:id/materials", DownloadMaterials) // 下载资料（路径优化）

	// 求助通道资源
	charity.GET("/channels", GetChannels)            // 渠道列表
	charity.GET("/channels/:id", GetChannelDetail)   // 渠道详情（路径优化）
	charity.GET("/channels/templates", GetTemplates) // 求助模板
	charity.POST("/channels/:id/help", SubmitHelp)   // 提交求助（路径优化）

	// 救助案例资源
	charity.GET("/cases", GetCases)                 // 案例列表
	charity.GET("/cases/:id", GetCaseDetail)        // 案例详情
	charity.GET("/cases/:id/pdf", GetCasePDF)       // 下载案例 PDF
	charity.GET("/cases/diseases", GetCaseDiseases) // 疾病选项
	charity.POST("/cases", CreateCase)              // 新增案例（补充）
	charity.PUT("/cases/:id", UpdateCase)           // 更新案例（补充）
	charity.DELETE("/cases/:id", DeleteCase)        // 删除案例（补充）
}
