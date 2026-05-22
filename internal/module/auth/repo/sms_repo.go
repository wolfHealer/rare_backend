// internal/module/auth/repo/sms_repo.go
package repo

import (
	"database/sql"

	"rare_backend/internal/pkg/db"
)

type SMSRepo struct{}

func NewSMSRepo() *SMSRepo {
	return &SMSRepo{}
}

// Create 创建验证码记录
func (r *SMSRepo) Create(phone, code, scene string, expireMinutes int) error {
	_, err := db.MySQL.Exec(`
        INSERT INTO sms_codes (phone, code, scene, created_at, expired_at)
        VALUES (?, ?, ?, NOW(), DATE_ADD(NOW(), INTERVAL ? MINUTE))
    `, phone, code, scene, expireMinutes)
	return err
}

// Verify 验证验证码
func (r *SMSRepo) Verify(phone, code, scene string) (bool, error) {
	var count int
	err := db.MySQL.QueryRow(`
        SELECT COUNT(*) FROM sms_codes 
        WHERE phone = ? AND code = ? AND scene = ? AND used = 0 AND expired_at > NOW()
    `, phone, code, scene).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// MarkUsed 标记验证码已使用
func (r *SMSRepo) MarkUsed(phone, code, scene string) error {
	_, err := db.MySQL.Exec(`
        UPDATE sms_codes SET used = 1, updated_at = NOW()
        WHERE phone = ? AND code = ? AND scene = ? AND used = 0
    `, phone, code, scene)
	return err
}

// CountRecent 获取最近N分钟内发送的验证码数量
func (r *SMSRepo) CountRecent(phone string, minutes int) (int, error) {
	var count int
	err := db.MySQL.QueryRow(`
        SELECT COUNT(*) FROM sms_codes 
        WHERE phone = ? AND created_at > DATE_SUB(NOW(), INTERVAL ? MINUTE)
    `, phone, minutes).Scan(&count)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, nil
		}
		return 0, err
	}
	return count, nil
}
