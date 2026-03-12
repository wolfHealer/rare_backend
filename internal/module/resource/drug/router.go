package drug

import (
	"github.com/gin-gonic/gin"
)

// 参数类型改为 *gin.RouterGroup 以支持嵌套
func SetupDrugRoutes(r *gin.RouterGroup) {
	// 1. 定义子模块前缀，最终路径变为 /api/resource/drug
	drug := r.Group("/drug")
	// 2. 注册具体路由
	// 注意：确保 handler 函数已正确初始化或为包级函数
	drug.GET("/list", ListDrugs)            // 药品列表
	drug.GET("/detail/:id", GetDrugDetail)  // 药品详情
	drug.GET("/manual/:id", DownloadManual) // 说明书下载
	drug.GET("/export", ExportDrugs)        // 名录导出
	drug.GET("/options", GetDrugOptions)    // 筛选选项

	drug.GET("/channel/options", GetChannelOptions)     //渠道筛选选项
	drug.GET("/channel/list", GetChannelList)           //渠道列表
	drug.GET("/channel/detail/:id", GetChannelDetail)   //渠道详情
	drug.POST("/channel/contact/:id", ContactChannel)   //联系渠道
	drug.POST("/channel/feedback/:id", FeedbackChannel) //反馈评价

	// 赠药援助路由（新增）
	drug.GET("/donation/list", GetDonationList)
	drug.POST("/donation/apply/:id", ApplyDonation)
	drug.GET("/donation/progress/:id", GetDonationProgress)
	drug.GET("/donation/guide/:id", DownloadDonationGuide)

	// 工具路由（新增）
	drug.GET("/tool/list", GetToolList)
	drug.GET("/tool/download/:id", DownloadTool)

}
