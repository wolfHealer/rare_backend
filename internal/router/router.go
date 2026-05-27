package router

import (
	"rare_backend/internal/middleware"
	"rare_backend/internal/module/auth"
	"rare_backend/internal/module/community"
	"rare_backend/internal/module/knowledge"
	"rare_backend/internal/module/region"
	"rare_backend/internal/module/resource"
	"rare_backend/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

func Register(r *gin.Engine) {
	// 健康检查（最先）
	r.GET("/health", func(c *gin.Context) {
		response.OK(c, gin.H{"status": "ok"})
	})

	// API 分组
	api := r.Group("/api")

	// ===== auth 模块 =====
	auth.Register(api)

	// ===== 用户中心（我的发布等）=====
	user := api.Group("/user")
	user.Use(middleware.AuthRequired())
	community.RegisterUserRoutes(user)

	// ===== post 模块 =====
	community.Register(api)

	// ===== knowledge 模块 =====
	knowledge.Register(api)

	//======= resource模块 =====
	resource.SetupResourceRoutes(api)

	region.Register(api)
}
