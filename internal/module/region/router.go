package region

import "github.com/gin-gonic/gin"

func Register(r *gin.RouterGroup) {
	region := r.Group("/region")

	// 基础 CRUD
	region.POST("", CreateRegion)       // 新增
	region.GET("/:id", GetRegionDetail) // 【新增】查看详情
	region.PUT("/:id", UpdateRegion)    // 修改
	region.DELETE("/:id", DeleteRegion) // 删除

	// 查询接口
	region.GET("/list", GetRegionList)                     // 通用列表查询
	region.GET("/tree", GetRegionTree)                     // 三级树形结构查询
	region.GET("/province-city-tree", GetProvinceCityTree) // 【新增】省市两级树形结构查询

}
