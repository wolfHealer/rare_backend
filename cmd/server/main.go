package main

import (
	"log"
	"rare_backend/internal/pkg/db"
	"rare_backend/internal/router"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	// 1. 创建 Gin 实例
	r := gin.Default()

	// 2. 配置 CORS 中间件
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"http://localhost:5173"} // 允许的前端域名
	config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Type", "Authorization"}
	r.Use(cors.New(config))

	// 3. 注册路由（关键）
	router.Register(r)

	// 4. 初始化数据库
	if err := db.InitMySQL("root:love1357hb@tcp(127.0.0.1:3306)/rare_backend?parseTime=true"); err != nil {
		log.Fatalf("mysql init: %v", err)
	}
	if err := db.InitMongo("mongodb://localhost:27017"); err != nil {
		log.Fatalf("mongo init: %v", err)
	}

	// 5️. 启动服务
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("server start failed: %v", err)
	}
}
