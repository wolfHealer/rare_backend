package auth

import "github.com/gin-gonic/gin"

func login(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "login ok",
	})
}

func register(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "register ok",
	})
}
