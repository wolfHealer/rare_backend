package post

import "github.com/gin-gonic/gin"

// router.go
func Register(r *gin.RouterGroup) {
	post := r.Group("/community")
	post.GET("/posts", GetCommunityPosts)            // 社区帖子列表
	post.POST("/post", CreatePost)                   // 创建帖子
	post.POST("/posts/:id/like", LikePost)           // 点赞帖子
	post.GET("/posts/:id/comments", GetPostComments) // 获取帖子评论
	post.POST("/posts/:id/comments", CreateComment)  // 创建评论
	post.GET("/posts/:id", GetPostDetail)            // 获取帖子详情

}
