// internal/pkg/hash/hash.go
package hash

import (
	"golang.org/x/crypto/bcrypt"
)

// CheckPassword 验证明文密码是否与哈希值匹配
func CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// HashPassword 对明文密码进行哈希加密（用于注册场景）
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}
