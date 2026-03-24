package drug

import (
	"github.com/gin-gonic/gin"
)

// 参数类型改为 *gin.RouterGroup 以支持嵌套
func SetupDrugRoutes(r *gin.RouterGroup) {
	// 1. 定义子模块前缀，最终路径变为 /api/resource/drug
	drug := r.Group("/drug")
	// 2. 注册具体路由

	// 药品资源
	drug.GET("/drugs/options", GetDrugOptions)    //筛选选项
	drug.GET("/drugs", ListDrugs)                 //  药品列表
	drug.GET("/drugs/:id", GetDrugDetail)         //  药品详情
	drug.POST("/drugs", CreateDrug)               // 新增药品
	drug.PUT("/drugs/:id", UpdateDrug)            // 更新药品
	drug.DELETE("/drugs/:id", DeleteDrug)         // 删除药品
	drug.GET("/drugs/:id/manual", DownloadManual) // 说明书下载
	drug.GET("/drugs/export", ExportDrugs)        // 名录导出

	// 渠道资源
	drug.GET("/channels/options", GetChannelOptions) //渠道筛选选项
	drug.GET("/channels", GetChannelList)            //渠道列表
	drug.GET("/channels/:id", GetChannelDetail)      //渠道详情
	drug.POST("/channels", CreateChannel)            // 新增渠道
	drug.PUT("/channels/:id", UpdateChannel)         // 更新渠道
	drug.DELETE("/channels/:id", DeleteChannel)      // 删除渠道

	drug.POST("/channels/:id/contact", ContactChannel)    //联系渠道
	drug.POST("/channels/:id/feedbacks", FeedbackChannel) //反馈评价


	// 赠药援助资源
	drug.GET("/donations", GetDonationList)                  // 赠药项目列表
	drug.GET("/donations/:id", GetDonationDetail)            // 赠药项目详情
	drug.POST("/donations", CreateDonation)                  // 新增赠药项目
	drug.PUT("/donations/:id", UpdateDonation)               // 更新赠药项目
	drug.DELETE("/donations/:id", DeleteDonation)            // 删除赠药项目
	drug.POST("/donations/:id/apply", ApplyDonation)         // 提交申请
	drug.GET("/donations/:id/progress", GetDonationProgress) // 查询进度
	drug.GET("/donations/:id/guide", DownloadDonationGuide)  // 下载指南

	// 工具资源
	// 工具资源
	drug.GET("/tools/options", GetToolOptions)    // 筛选选项（新增）
	drug.GET("/tools", GetToolList)               // 工具列表
	drug.GET("/tools/:id", GetToolDetail)         // 工具详情（新增）
	drug.POST("/tools", CreateTool)               // 新增工具（新增）
	drug.PUT("/tools/:id", UpdateTool)            // 更新工具（新增）
	drug.DELETE("/tools/:id", DeleteTool)         // 删除工具（新增）
	drug.GET("/tools/:id/download", DownloadTool) // 下载文件

}
