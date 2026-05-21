package region

import (
	"rare_backend/internal/middleware"

	"github.com/gin-gonic/gin"
)

func Register(r *gin.RouterGroup) {
	region := r.Group("/region")

	// 公开读（固定路径须在 /:id 之前）
	region.GET("/list", GetRegionList)
	region.GET("/tree", GetRegionTree)
	region.GET("/province-city-tree", GetProvinceCityTree)
	region.GET("/:id", GetRegionDetail)

	// 后台写（管理员）
	admin := region.Group("")
	admin.Use(middleware.AdminWrite()...)
	admin.POST("", CreateRegion)
	admin.PUT("/:id", UpdateRegion)
	admin.DELETE("/:id", DeleteRegion)
}
