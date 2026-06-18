package knowledge

import (
	"rare_backend/internal/middleware"

	"github.com/gin-gonic/gin"
)

func Register(r *gin.RouterGroup) {
	k := r.Group("/knowledge")

	// 公开读
	k.GET("/diseases", GetDiseases)
	k.GET("/diseases/search", SearchDiseases)
	k.GET("/diseases/options", GetDiseaseOptions)
	k.GET("/disease/:id", GetDiseaseByID)
	k.GET("/categories", GetCategories)
	k.GET("/categories/tree", GetCategoryTree)
	k.GET("/category/:categoryId/diseases", GetDiseasesByCategory)
	k.GET("/tags", GetTags)
	k.GET("/articles/tags", ListArticleTags)
	k.GET("/articles/tags/:id", GetArticleTagByID)
	k.GET("/articles", GetArticles)
	k.GET("/article/:id", GetArticleByID)

	// 后台写（管理员）
	admin := k.Group("")
	admin.Use(middleware.AdminWrite()...)
	admin.POST("/disease", CreateDisease)
	admin.PUT("/disease/:id", UpdateDisease)
	admin.DELETE("/disease/:id", DeleteDisease)
	admin.POST("/category", CreateCategory)
	admin.PUT("/category/:id", UpdateCategory)
	admin.DELETE("/category/:id", DeleteCategory)
	admin.POST("/tag", CreateTag)
	admin.PUT("/tag/:id", UpdateTag)
	admin.DELETE("/tag/:id", DeleteTag)
	admin.POST("/article", CreateArticle)
	admin.PUT("/article/:id", UpdateArticle)
	admin.DELETE("/article/:id", DeleteArticle)
	admin.POST("/articles/tags", CreateArticleTag)
	admin.PUT("/articles/tags/:id", UpdateArticleTag)
	admin.DELETE("/articles/tags/:id", DeleteArticleTag)
}
