package main

import (
	"log"
	"rare_backend/internal/router"

	"github.com/gin-gonic/gin"
)

func main() {
	// 1️⃣ 创建 Gin 实例
	r := gin.Default()

	// 2️⃣ 注册路由（关键）
	router.Register(r)

	// 3️⃣ 启动服务
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("server start failed: %v", err)
	}
}
