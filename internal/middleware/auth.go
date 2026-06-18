package middleware

import (
	"strconv"
	"strings"

	"rare_backend/internal/pkg/db"
	"rare_backend/internal/pkg/jwt"
	"rare_backend/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

const (
	ContextUserIDKey = "user_id"
	ContextRoleKey   = "role"
)

// AuthRequired 校验 Bearer Token，写入 user_id、role
func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, ok := parseRequestClaims(c)
		if !ok {
			return
		}
		if !ensureUserActive(c, claims.UserID) {
			return
		}
		c.Set(ContextUserIDKey, claims.UserID)
		c.Set(ContextRoleKey, claims.Role)
		c.Next()
	}
}

// OptionalAuth 有 Token 则解析并写入上下文，无 Token 或无效时继续（用于可选登录态）
func OptionalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractBearerToken(c.GetHeader("Authorization"))
		if token == "" {
			c.Next()
			return
		}
		claims, err := jwt.ParseToken(token)
		if err == nil && isUserActive(claims.UserID) {
			c.Set(ContextUserIDKey, claims.UserID)
			c.Set(ContextRoleKey, claims.Role)
		}
		c.Next()
	}
}

// AdminRequired 要求管理员角色（须放在 AuthRequired 之后）
func AdminRequired() gin.HandlerFunc {
	return RequireRole(jwt.RoleAdmin)
}

// RequireRole 要求指定角色之一（须放在 AuthRequired 之后）
func RequireRole(roles ...int) gin.HandlerFunc {
	allowed := make(map[int]struct{}, len(roles))
	for _, r := range roles {
		allowed[r] = struct{}{}
	}
	return func(c *gin.Context) {
		role, ok := GetRole(c)
		if !ok {
			response.Forbidden(c, "权限不足")
			c.Abort()
			return
		}
		if _, ok := allowed[role]; !ok {
			response.Forbidden(c, "权限不足")
			c.Abort()
			return
		}
		c.Next()
	}
}

// GetUserID 从上下文读取当前用户 ID
func GetUserID(c *gin.Context) (int64, bool) {
	v, ok := c.Get(ContextUserIDKey)
	if !ok {
		return 0, false
	}
	switch id := v.(type) {
	case int64:
		return id, true
	case int:
		return int64(id), true
	case float64:
		return int64(id), true
	default:
		return 0, false
	}
}

// GetRole 从上下文读取当前用户角色
func GetRole(c *gin.Context) (int, bool) {
	v, ok := c.Get(ContextRoleKey)
	if !ok {
		return 0, false
	}
	switch r := v.(type) {
	case int:
		return r, true
	case int64:
		return int(r), true
	case float64:
		return int(r), true
	default:
		return 0, false
	}
}

// IsAdmin 是否为管理员
func IsAdmin(c *gin.Context) bool {
	role, ok := GetRole(c)
	return ok && role == jwt.RoleAdmin
}

// MustGetUserID 获取用户 ID，未登录则中止并返回 401
func MustGetUserID(c *gin.Context) (int64, bool) {
	id, ok := GetUserID(c)
	if !ok {
		response.Unauthorized(c, "请先登录")
		c.Abort()
		return 0, false
	}
	return id, true
}

func parseRequestClaims(c *gin.Context) (*jwt.Claims, bool) {
	token := extractBearerToken(c.GetHeader("Authorization"))
	if token == "" {
		response.Unauthorized(c, "未登录或 Token 缺失")
		c.Abort()
		return nil, false
	}
	claims, err := jwt.ParseToken(token)
	if err != nil {
		response.Unauthorized(c, "Token 无效或已过期")
		c.Abort()
		return nil, false
	}
	return claims, true
}

func ensureUserActive(c *gin.Context, userID int64) bool {
	if !isUserActive(userID) {
		response.Unauthorized(c, "账号已注销或不可用")
		c.Abort()
		return false
	}
	return true
}

func isUserActive(userID int64) bool {
	var status int
	err := db.MySQL.QueryRow(`SELECT status FROM user WHERE id = ?`, userID).Scan(&status)
	return err == nil && status == 1
}

func extractBearerToken(header string) string {
	header = strings.TrimSpace(header)
	if header == "" {
		return ""
	}
	const prefix = "Bearer "
	if len(header) > len(prefix) && strings.EqualFold(header[:len(prefix)], prefix) {
		return strings.TrimSpace(header[len(prefix):])
	}
	if !strings.Contains(header, " ") {
		return header
	}
	parts := strings.SplitN(header, " ", 2)
	if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
		return strings.TrimSpace(parts[1])
	}
	return ""
}

// GetUserIDString 便于日志等场景
func GetUserIDString(c *gin.Context) string {
	id, ok := GetUserID(c)
	if !ok {
		return ""
	}
	return strconv.FormatInt(id, 10)
}
