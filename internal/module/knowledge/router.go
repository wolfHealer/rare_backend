package knowledge

import "github.com/gin-gonic/gin"

func Register(r *gin.RouterGroup) {
	knowledge := r.Group("/knowledge")

	knowledge.GET("/diseases", GetDiseases)
	knowledge.GET("/diseases/search", SearchDiseases)
	// 【新增】疾病下拉选项接口
	knowledge.GET("/diseases/options", GetDiseaseOptions)

	knowledge.POST("/disease", CreateDisease)
	knowledge.GET("/disease/:id", GetDiseaseByID)
	knowledge.PUT("/disease/:id", UpdateDisease)
	knowledge.DELETE("/disease/:id", DeleteDisease)

	knowledge.GET("/categories", GetCategories)
	knowledge.GET("/categories/tree", GetCategoryTree)
	knowledge.POST("/category", CreateCategory)
	knowledge.PUT("/category/:id", UpdateCategory)
	knowledge.GET("/category/:categoryId/diseases", GetDiseasesByCategory)
	knowledge.DELETE("/category/:id", DeleteCategory)

	knowledge.GET("/tags", GetTags)
	knowledge.POST("/tag", CreateTag)
	knowledge.PUT("/tag/:id", UpdateTag)
	knowledge.DELETE("/tag/:id", DeleteTag)

	// --- 新增：文章管理路由 ---
	knowledge.GET("/articles", GetArticles)         // 文章列表
	knowledge.POST("/article", CreateArticle)       // 创建文章
	knowledge.GET("/article/:id", GetArticleByID)   // 文章详情
	knowledge.PUT("/article/:id", UpdateArticle)    // 更新文章
	knowledge.DELETE("/article/:id", DeleteArticle) // 删除文章
}
