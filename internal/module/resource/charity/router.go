package charity

import (
	"github.com/gin-gonic/gin"
)

// 参数类型改为 *gin.RouterGroup 以支持嵌套
func SetupCharityRoutes(r *gin.RouterGroup) {
	// 1. 定义子模块前缀，最终路径变为 /api/resource/medical
	charity := r.Group("/charity")
	// 2. 注册具体路由
	// 注意：确保 handler 函数已正确初始化或为包级函数
	// 救助项目相关
	charity.GET("/projects", ListProjects)
	charity.GET("/projects/:id", GetProjectDetail)
	charity.GET("/filters", GetFilters)
	// charity.POST("/apply", SubmitApplication)
	// charity.GET("/projects/:id/docs", DownloadDocs)

	//医保政策解读
	charity.GET("/policies", GetPolicyList)               //获取政策列表
	charity.GET("/policies/regions", GetRegions)          //获取地区选项
	charity.GET("/policies/detail/:id", GetPolicyDetail)  //获取政策详情
	charity.GET("/policies/materials", DownloadMaterials) //下载资料

	// 求助通道相关（新增）
	charity.GET("/channels", GetChannels)                 // 获取求助渠道列表
	charity.GET("/channels/detail/:id", GetChannelDetail) // 获取求助渠道详情
	charity.GET("/channels/template", GetTemplates)       // 获取求助模板
	charity.POST("/channels/help/submit", SubmitHelp)     // 提交求助请求

	// 救助案例相关（新增）
	charity.GET("/cases", GetCases)                 // 获取案例列表
	charity.GET("/cases/:id", GetCaseDetail)        // 获取案例详情
	charity.GET("/cases/:id/pdf", GetCasePDF)       // 下载案例 PDF（新增）
	charity.GET("/cases/diseases", GetCaseDiseases) // 获取疾病选项（新增）

}
