package charity

import (
	"rare_backend/internal/middleware"

	"github.com/gin-gonic/gin"
)

func SetupCharityRoutes(r *gin.RouterGroup) {
	charity := r.Group("/charity")

	// 公开读
	charity.GET("/projects/options", GetProjectOptions)
	charity.GET("/projects", ListProjects)
	charity.GET("/projects/:id", GetProjectDetail)
	charity.GET("/channels", GetChannels)
	charity.GET("/channels/:id", GetChannelDetail)
	charity.GET("/cases", GetCases)
	charity.GET("/cases/:id", GetCaseDetail)
	charity.GET("/cases/:id/pdf", GetCasePDF)
	charity.GET("/cases/diseases", GetCaseDiseases)

	// 后台写（管理员）
	admin := charity.Group("")
	admin.Use(middleware.AdminWrite()...)
	admin.POST("/projects", CreateProject)
	admin.PUT("/projects/:id", UpdateProject)
	admin.DELETE("/projects/:id", DeleteProject)
	admin.POST("/channels", CreateChannel)
	admin.PUT("/channels/:id", UpdateChannel)
	admin.DELETE("/channels/:id", DeleteChannel)
	admin.POST("/cases", CreateCase)
	admin.PUT("/cases/:id", UpdateCase)
	admin.DELETE("/cases/:id", DeleteCase)
}
