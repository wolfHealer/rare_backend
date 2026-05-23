package config

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// Config 应用配置（从环境变量加载）
type Config struct {
	ServerAddr       string
	MySQLDSN         string
	JWTSecret        string
	JWTExpire        time.Duration
	CORSAllowOrigins []string
	GinMode          string
}

// Load 读取配置：先加载 .env，再读环境变量。
// MYSQL_DSN、JWT_SECRET 为必填；未配置时返回错误（不在代码中提供真实密码/密钥默认值）。
// 已存在于进程环境中的变量不会被 .env 覆盖（便于生产环境注入配置）。
func Load() (*Config, error) {
	loadDotEnv()

	cfg := &Config{
		ServerAddr: getEnv("SERVER_ADDR", ":8080"),
		MySQLDSN:   strings.TrimSpace(os.Getenv("MYSQL_DSN")),
		JWTSecret:  strings.TrimSpace(os.Getenv("JWT_SECRET")),
		JWTExpire:  24 * time.Hour,
		GinMode:    getEnv("GIN_MODE", "debug"),
	}

	if cfg.MySQLDSN == "" {
		return nil, errors.New(requiredEnvHint("MYSQL_DSN",
			"root:YOUR_PASSWORD@tcp(127.0.0.1:3306)/rare_backend?parseTime=true"))
	}

	if cfg.JWTSecret == "" {
		return nil, errors.New(requiredEnvHint("JWT_SECRET",
			"请设置足够长的随机字符串（勿提交到 Git）"))
	}

	if v := os.Getenv("JWT_EXPIRE_HOURS"); v != "" {
		if h, err := time.ParseDuration(v + "h"); err == nil {
			cfg.JWTExpire = h
		}
	}

	cfg.CORSAllowOrigins = parseCSVEnv("CORS_ALLOW_ORIGINS",
		"http://localhost:5173,http://localhost:5174,http://127.0.0.1:5174")

	return cfg, nil
}

func requiredEnvHint(key, example string) string {
	return fmt.Sprintf(
		"%s 未配置。\n"+
			"  1. 在项目根目录执行: cp .env.example .env\n"+
			"  2. 编辑 .env，设置 %s（示例: %s）\n"+
			"  3. 或在启动前执行: export %s='...'",
		key, key, example, key,
	)
}

func loadDotEnv() {
	for _, path := range []string{".env", "../.env"} {
		if _, err := os.Stat(path); err != nil {
			if errors.Is(err, os.ErrNotExist) {
				continue
			}
			log.Printf("[config] 提示: 无法访问 %s（%v）", path, err)
			continue
		}
		if err := godotenv.Load(path); err != nil {
			log.Printf("[config] 提示: 加载 %s 失败（%v）", path, err)
			continue
		}
		log.Printf("[config] 已加载 %s", path)
		return
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func parseCSVEnv(key, fallback string) []string {
	raw := getEnv(key, fallback)
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

// OSSConfig OSS 配置
type OSSConfig struct {
	Endpoint        string
	AccessKeyID     string
	AccessKeySecret string
	BucketName      string
	PublicBaseURL   string // 可选，自定义 CDN/域名，如 https://cdn.example.com
}

func GetOSSConfig() OSSConfig {
	return OSSConfig{
		Endpoint:        os.Getenv("OSS_ENDPOINT"),
		AccessKeyID:     os.Getenv("OSS_ACCESS_KEY_ID"),
		AccessKeySecret: os.Getenv("OSS_ACCESS_KEY_SECRET"),
		BucketName:      os.Getenv("OSS_BUCKET_NAME"),
		PublicBaseURL:   os.Getenv("OSS_PUBLIC_BASE_URL"),
	}
}

// SMSConfig 短信服务配置
type SMSConfig struct {
	AccessKeyID     string
	AccessKeySecret string

	// 个人资质：号码认证「短信认证」SendSmsVerifyCode（控制台赠送签名+模板）
	PNVSSignName     string // SMS_PNVS_SIGN_NAME
	PNVSTemplateCode string // SMS_PNVS_TEMPLATE_CODE

	// 公司资质：短信服务 dysmsapi SendSms（需企业资质审核签名/模板）
	SignName         string // SMS_SIGN_NAME
	RegisterTemplate string // SMS_REGISTER_TPL
	LoginTemplate    string // SMS_LOGIN_TPL
}

func GetSMSConfig() SMSConfig {
	return SMSConfig{
		AccessKeyID:      os.Getenv("SMS_ACCESS_KEY_ID"),
		AccessKeySecret:  os.Getenv("SMS_ACCESS_KEY_SECRET"),
		PNVSSignName:     os.Getenv("SMS_PNVS_SIGN_NAME"),
		PNVSTemplateCode: os.Getenv("SMS_PNVS_TEMPLATE_CODE"),
		SignName:         os.Getenv("SMS_SIGN_NAME"),
		RegisterTemplate: os.Getenv("SMS_REGISTER_TPL"),
		LoginTemplate:    os.Getenv("SMS_LOGIN_TPL"),
	}
}
