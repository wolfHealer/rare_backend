package post

import "github.com/gin-gonic/gin"

func Register(r *gin.RouterGroup) {
	post := r.Group("/post")

	post.POST("/create", createPost)
	post.GET("/list", listPost)
	post.GET("/detail/:id", getPost)
}
