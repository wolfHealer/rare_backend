package jwt

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// RoleAdmin 管理员角色（与 user.role 一致）
const RoleAdmin = 9

var (
	secret              []byte
	tokenExpireDuration = 24 * time.Hour
)

// Claims 解析后的 Token 载荷
type Claims struct {
	UserID int64
	Role   int
}

// TokenExpireDuration 返回当前 Token 有效期
func TokenExpireDuration() time.Duration {
	return tokenExpireDuration
}

// Init 使用配置初始化 JWT
func Init(secretKey string, expire time.Duration) error {
	if secretKey == "" {
		return errors.New("jwt: secret key is empty")
	}
	secret = []byte(secretKey)
	if expire > 0 {
		tokenExpireDuration = expire
	}
	return nil
}

// GenerateToken 签发 Token（含 user_id、role）
func GenerateToken(userID int64, role int) (string, error) {
	if len(secret) == 0 {
		return "", errors.New("jwt: not initialized, call Init first")
	}

	claims := jwt.MapClaims{
		"user_id": userID,
		"role":    role,
		"exp":     time.Now().Add(tokenExpireDuration).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secret)
}

// ParseToken 解析并校验 Token
func ParseToken(tokenString string) (*Claims, error) {
	if len(secret) == 0 {
		return nil, errors.New("jwt: not initialized, call Init first")
	}

	parsed, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return secret, nil
	})
	if err != nil {
		return nil, err
	}
	if !parsed.Valid {
		return nil, errors.New("jwt: invalid token")
	}

	mapClaims, ok := parsed.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("jwt: invalid claims")
	}

	userID, err := claimsInt64(mapClaims, "user_id")
	if err != nil {
		return nil, err
	}
	role, err := claimsInt(mapClaims, "role")
	if err != nil {
		return nil, err
	}

	return &Claims{UserID: userID, Role: role}, nil
}

func claimsInt64(claims jwt.MapClaims, key string) (int64, error) {
	v, ok := claims[key]
	if !ok {
		return 0, errors.New("jwt: missing " + key)
	}
	switch n := v.(type) {
	case float64:
		return int64(n), nil
	case int64:
		return n, nil
	case int:
		return int64(n), nil
	default:
		return 0, errors.New("jwt: invalid " + key + " type")
	}
}

func claimsInt(claims jwt.MapClaims, key string) (int, error) {
	v, ok := claims[key]
	if !ok {
		return 0, errors.New("jwt: missing " + key)
	}
	switch n := v.(type) {
	case float64:
		return int(n), nil
	case int64:
		return int(n), nil
	case int:
		return n, nil
	default:
		return 0, errors.New("jwt: invalid " + key + " type")
	}
}
