package knowledge

import "github.com/gin-gonic/gin"

func Register(r *gin.RouterGroup) {
	knowledge := r.Group("/knowledge")

	// 静态路径优先注册
	knowledge.GET("/diseases", GetDiseases)        // 复数列表
	knowledge.GET("/disease-tree", GetDiseaseTree) // 特殊路径
	knowledge.GET("/categories", GetCategories)    // 复数列表

	// 动态参数路径后注册
	knowledge.POST("/disease", CreateDisease)
	knowledge.GET("/disease/:id", GetDiseaseByID)
	knowledge.PUT("/disease/:id", UpdateDisease)

	knowledge.POST("/category", CreateCategory)
	// knowledge.GET("/category/:id", GetCategoryByID) // 如需获取单个分类
	knowledge.PUT("/category/:id", UpdateCategory)
	knowledge.GET("/category/:categoryId/diseases", GetDiseasesByCategory)
}
