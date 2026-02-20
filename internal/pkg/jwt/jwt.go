package jwt

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	JWTSecret           = "your-secret-key" // 替换为安全的密钥
	TokenExpireDuration = time.Hour * 24    // Token 有效期为 24 小时
)

// GenerateToken 生成 JWT Token
func GenerateToken(userID int64) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(TokenExpireDuration).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(JWTSecret))
}
