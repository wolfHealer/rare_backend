package knowledge

import "github.com/gin-gonic/gin"

func Register(r *gin.RouterGroup) {
	knowledge := r.Group("/knowledge")
	// 具体疾病
	knowledge.POST("/disease", CreateDisease)     // 创建疾病
	knowledge.PUT("/disease/:id", UpdateDisease)  // 更新疾病
	knowledge.GET("/diseases", GetDiseases)       // 获取疾病列表
	knowledge.GET("/disease/:id", GetDiseaseByID) // 根据疾病 ID 获取疾病详情
	// 基本的大分类
	knowledge.POST("/category", CreateCategory)                            // 创建大分类
	knowledge.GET("/categories", GetCategories)                            // 获取大分类列表
	knowledge.GET("/category/:categoryId/diseases", GetDiseasesByCategory) // 根据大分类获取疾病列表

	knowledge.GET("/disease-tree", GetDiseaseTree) // 获取疾病分类树
}
