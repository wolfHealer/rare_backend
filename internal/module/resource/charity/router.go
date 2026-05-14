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
	charity.GET("/projects/options", GetProjectOptions)
	charity.GET("/projects", ListProjects)         // 项目列表
	charity.GET("/projects/:id", GetProjectDetail) // 项目详情
	charity.POST("/projects", CreateProject)       // 新增项目
	charity.PUT("/projects/:id", UpdateProject)    // 更新项目
	charity.DELETE("/projects/:id", DeleteProject) // 删除项目

	// 求助通道资源
	charity.GET("/channels", GetChannels)          // 渠道列表
	charity.GET("/channels/:id", GetChannelDetail) // 渠道详情
	charity.POST("/channels", CreateChannel)       // 【新增】新增渠道
	charity.PUT("/channels/:id", UpdateChannel)    // 【新增】更新渠道
	charity.DELETE("/channels/:id", DeleteChannel) // 【新增】删除渠道

	// 救助案例资源
	charity.GET("/cases", GetCases)                 // 案例列表
	charity.GET("/cases/:id", GetCaseDetail)        // 案例详情
	charity.GET("/cases/:id/pdf", GetCasePDF)       // 下载案例 PDF
	charity.GET("/cases/diseases", GetCaseDiseases) // 疾病选项
	charity.POST("/cases", CreateCase)              // 新增案例
	charity.PUT("/cases/:id", UpdateCase)           // 更新案例
	charity.DELETE("/cases/:id", DeleteCase)        // 删除案例
}
