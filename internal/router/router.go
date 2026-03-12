package router

import (
	"rare_backend/internal/module/auth"
	"rare_backend/internal/module/knowledge"
	"rare_backend/internal/module/post"
	"rare_backend/internal/module/resource"

	"github.com/gin-gonic/gin"
)

func Register(r *gin.Engine) {
	// 健康检查（最先）
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// API 分组
	api := r.Group("/api")

	// ===== auth 模块 =====
	auth.Register(api)

	// ===== post 模块 =====
	post.Register(api)

	// ===== knowledge 模块 =====
	knowledge.Register(api)

	//======= resource模块 =====
	resource.SetupResourceRoutes(api)
}
