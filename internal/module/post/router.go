package post

import "github.com/gin-gonic/gin"

func Register(r *gin.RouterGroup) {
	post := r.Group("/community")

	// 帖子资源 - 静态路由优先
	post.GET("/posts/options", GetPostOptions) // 筛选选项（新增）
	post.GET("/posts", GetCommunityPosts)      // 帖子列表
	post.GET("/posts/:id", GetPostDetail)      // 帖子详情
	post.POST("/posts", CreatePost)            // 创建帖子
	post.PUT("/posts/:id", UpdatePost)         // 更新帖子（新增）
	post.DELETE("/posts/:id", DeletePost)      // 删除帖子（新增）

	// 点赞功能
	post.POST("/posts/:id/like", LikePost) // 点赞/取消点赞

	// 评论资源 - 静态路由优先
	post.GET("/posts/:id/comments", GetPostComments) // 获取评论列表
	post.POST("/posts/:id/comments", CreateComment)  // 创建评论
	post.PUT("/comments/:id", UpdateComment)         // 更新评论（新增）
	post.DELETE("/comments/:id", DeleteComment)      // 删除评论（新增）
}
