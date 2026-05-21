package middleware

import "github.com/gin-gonic/gin"

// AdminWrite 后台写操作：登录 + 管理员
func AdminWrite() []gin.HandlerFunc {
	return []gin.HandlerFunc{AuthRequired(), AdminRequired()}
}

// AuthWrite 需登录的写操作（普通用户）
func AuthWrite() []gin.HandlerFunc {
	return []gin.HandlerFunc{AuthRequired()}
}
