package auth

import (
	"rare_backend/internal/middleware"

	"github.com/gin-gonic/gin"
)

func Register(r *gin.RouterGroup) {
	auth := r.Group("/auth")
	auth.POST("/login", login)
	auth.POST("/register", register)

	// 用户信息接口（需登录）
	auth.GET("/userinfo", middleware.AuthRequired(), getUserInfo)
	// 用户更新自己的信息（昵称、头像）
	auth.PUT("/userinfo", middleware.AuthRequired(), updateProfile)
	// 头像上传接口
	auth.POST("/user/avatar", middleware.AuthRequired(), uploadAvatar)
	// 账号注销（本人操作）
	auth.POST("/account/deactivate", middleware.AuthRequired(), deactivateAccount)

	// 用户管理（需管理员）
	users := r.Group("/system/users")
	users.Use(middleware.AuthRequired(), middleware.AdminRequired())
	users.POST("", CreateUser)
	users.GET("", GetUserList)       // GET /api/auth/users?page=1&pageSize=10&keyword=xxx
	users.GET("/:id", GetUserByID)   // GET /api/auth/users/123
	users.PUT("/:id", UpdateUser)    // PUT /api/users/123
	users.DELETE("/:id", DeleteUser) // DELETE /api/users/123

	// 新增：专门更新角色的接口，适配前端 PATCH /system/users/${id}/role
	users.PATCH("/:id/role", UpdateUserRole)
	// 新增：重置密码接口
	users.POST("/:id/reset-password", ResetPassword)

	// 短信验证码
	sms := auth.Group("/sms")
	sms.POST("/send", sendSMSCode)

	// 用户收藏（需登录）
	user := r.Group("/user")
	user.Use(middleware.AuthRequired())
	user.GET("/favorites", listFavorites)
	user.DELETE("/favorites", removeFavoriteByTarget)
	user.DELETE("/favorites/:id", removeFavoriteByID)
}
