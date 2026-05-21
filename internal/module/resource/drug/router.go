package drug

import (
	"rare_backend/internal/middleware"

	"github.com/gin-gonic/gin"
)

func SetupDrugRoutes(r *gin.RouterGroup) {
	drug := r.Group("/drug")

	// 公开读
	drug.GET("/drugs/options", GetDrugOptions)
	drug.GET("/drugs", GetDrugList)
	drug.GET("/drugs/export", ExportDrugs)
	drug.GET("/drugs/:id", GetDrugDetail)
	drug.GET("/drugs/:id/manual", DownloadManual)
	drug.GET("/channels", GetChannelList)
	drug.GET("/channels/:id", GetChannelDetail)
	drug.GET("/donations", GetDonationList)
	drug.GET("/donations/:id", GetDonationDetail)
	drug.GET("/donations/:id/progress", GetDonationProgress)
	drug.GET("/donations/:id/guide", DownloadDonationGuide)

	// 登录用户操作
	user := drug.Group("")
	user.Use(middleware.AuthWrite()...)
	user.POST("/channels/:id/contact", ContactChannel)
	user.POST("/donations/:id/apply", ApplyDonation)

	// 后台写（管理员）
	admin := drug.Group("")
	admin.Use(middleware.AdminWrite()...)
	admin.POST("/drugs", CreateDrug)
	admin.PUT("/drugs/:id", UpdateDrug)
	admin.DELETE("/drugs/:id", DeleteDrug)
	admin.POST("/channels", CreateChannel)
	admin.PUT("/channels/:id", UpdateChannel)
	admin.DELETE("/channels/:id", DeleteChannel)
	admin.POST("/donations", CreateDonation)
	admin.PUT("/donations/:id", UpdateDonation)
	admin.DELETE("/donations/:id", DeleteDonation)
}
