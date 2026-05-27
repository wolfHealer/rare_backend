package community

import (
	"rare_backend/internal/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterUserRoutes(r *gin.RouterGroup) {
	r.GET("/posts", ListMyPosts)
	r.DELETE("/posts/:id", DeleteMyPost)
}

func Register(r *gin.RouterGroup) {
	post := r.Group("/community")

	// 公开读（可选登录，用于 is_liked 等）
	read := post.Group("")
	read.Use(middleware.OptionalAuth())
	read.GET("/posts/options", GetPostOptions)
	read.GET("/posts", GetCommunityPosts)
	read.GET("/posts/:id", GetPostDetail)
	read.GET("/posts/:id/comments", GetPostComments)
	read.GET("/comments/tree", GetCommentTree)
	read.GET("/comments/replies", GetCommentReplies)

	// 需登录的写操作
	write := post.Group("")
	write.Use(middleware.AuthWrite()...)
	write.POST("/posts", CreatePost)
	write.PUT("/posts/:id", UpdatePost)
	write.DELETE("/posts/:id", DeletePost)
	write.POST("/posts/:id/like", LikePost)
	write.POST("/posts/:id/favorite", FavoritePost)
	write.POST("/posts/:id/collect", FavoritePost)
	write.POST("/posts/:id/comments", CreateComment)
	write.PUT("/comments/:id", UpdateComment)
	write.DELETE("/comments/:id", DeleteComment)
}
