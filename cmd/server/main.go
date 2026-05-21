package main

import (
	"log"

	"rare_backend/internal/config"
	"rare_backend/internal/pkg/db"
	"rare_backend/internal/pkg/jwt"
	"rare_backend/internal/router"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("[config] %v", err)
	}

	// 1. 初始化数据库（必须在注册路由、处理请求之前）
	if err := db.InitMySQL(cfg.MySQLDSN); err != nil {
		log.Fatalf("mysql init: %v", err)
	}

	// 2. 初始化 JWT
	if err := jwt.Init(cfg.JWTSecret, cfg.JWTExpire); err != nil {
		log.Fatalf("jwt init: %v", err)
	}

	gin.SetMode(cfg.GinMode)
	r := gin.Default()

	// 3. CORS
	corsCfg := cors.DefaultConfig()
	corsCfg.AllowOrigins = cfg.CORSAllowOrigins
	corsCfg.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"}
	corsCfg.AllowHeaders = []string{"Origin", "Content-Type", "Authorization"}
	r.Use(cors.New(corsCfg))

	// 4. 注册路由
	router.Register(r)

	// 5. 启动服务
	log.Printf("server listening on %s", cfg.ServerAddr)
	if err := r.Run(cfg.ServerAddr); err != nil {
		log.Fatalf("server start failed: %v", err)
	}
}
