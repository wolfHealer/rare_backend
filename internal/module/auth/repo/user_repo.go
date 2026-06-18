package repo

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"rare_backend/internal/module/auth/domain"
	"rare_backend/internal/pkg/db"
	"rare_backend/internal/pkg/search"
)

type UserRepo struct{}

func NewUserRepo() *UserRepo {
	return &UserRepo{}
}

func (r *UserRepo) FindActiveByPhone(phone string) (*domain.User, error) {
	var u domain.User
	var nickname, avatar sql.NullString
	var lastLogin sql.NullTime

	err := db.MySQL.QueryRow(`
		SELECT id, phone, password_hash, login_count, role, display_name, avatar, status,
		       last_login_at, created_at, updated_at
		FROM user WHERE phone = ? AND status = 1
	`, phone).Scan(
		&u.ID, &u.Phone, &u.PasswordHash, &u.LoginCount, &u.Role,
		&nickname, &avatar, &u.Status, &lastLogin, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	u.DisplayName = nickname.String
	u.Avatar = avatar.String
	if lastLogin.Valid {
		t := lastLogin.Time
		u.LastLoginAt = &t
	}
	return &u, nil
}

func (r *UserRepo) CountByPhone(phone string) (int, error) {
	var count int
	err := db.MySQL.QueryRow(`SELECT COUNT(*) FROM user WHERE phone = ?`, phone).Scan(&count)
	return count, err
}

func (r *UserRepo) CreateRegister(phone, passwordHash string) error {
	_, err := db.MySQL.Exec(`
		INSERT INTO user (phone, password_hash, display_name, role, status)
		VALUES (?, ?, CONCAT('用户', LPAD(FLOOR(RAND() * 10000), 4, '0')), 1, 1)
	`, phone, passwordHash)
	return err
}

func (r *UserRepo) UpdateLoginInfo(id int64, loginCount int) error {
	_, err := db.MySQL.Exec(
		`UPDATE user SET last_login_at = NOW(), login_count = ? WHERE id = ?`,
		loginCount, id,
	)
	return err
}

func (r *UserRepo) ExistsByID(id int64) (bool, error) {
	var exists int64
	err := db.MySQL.QueryRow(`SELECT id FROM user WHERE id = ?`, id).Scan(&exists)
	if err == sql.ErrNoRows {
		return false, nil
	}
	return err == nil, err
}

func (r *UserRepo) ExistsActiveByID(id int64) (bool, error) {
	var exists int64
	err := db.MySQL.QueryRow(`SELECT id FROM user WHERE id = ? AND status = 1`, id).Scan(&exists)
	if err == sql.ErrNoRows {
		return false, nil
	}
	return err == nil, err
}

func (r *UserRepo) FindByID(id int64) (*domain.User, error) {
	var u domain.User
	var nickname, avatar sql.NullString
	var lastLogin sql.NullTime

	err := db.MySQL.QueryRow(`
		SELECT id, phone, display_name, avatar, role, status, last_login_at, login_count, created_at, updated_at
		FROM user WHERE id = ?
	`, id).Scan(
		&u.ID, &u.Phone, &nickname, &avatar, &u.Role, &u.Status,
		&lastLogin, &u.LoginCount, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	u.DisplayName = nickname.String
	u.Avatar = avatar.String
	if lastLogin.Valid {
		t := lastLogin.Time
		u.LastLoginAt = &t
	}
	return &u, nil
}

func (r *UserRepo) List(filter domain.UserListFilter) ([]domain.User, int64, error) {
	where, args := buildUserListWhere(filter)

	var total int64
	if err := db.MySQL.QueryRow(`SELECT COUNT(*) FROM user `+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	offset := (filter.Page - 1) * filter.PageSize
	listArgs := append(append([]interface{}{}, args...), filter.PageSize, offset)

	rows, err := db.MySQL.Query(`
		SELECT id, phone, display_name, avatar, role, status, last_login_at, login_count, created_at, updated_at
		FROM user `+where+` ORDER BY created_at DESC LIMIT ? OFFSET ?
	`, listArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var list []domain.User
	for rows.Next() {
		var u domain.User
		var nickname, avatar sql.NullString
		var lastLogin sql.NullTime
		if err := rows.Scan(
			&u.ID, &u.Phone, &nickname, &avatar, &u.Role, &u.Status,
			&lastLogin, &u.LoginCount, &u.CreatedAt, &u.UpdatedAt,
		); err != nil {
			continue
		}
		u.DisplayName = nickname.String
		u.Avatar = avatar.String
		if lastLogin.Valid {
			t := lastLogin.Time
			u.LastLoginAt = &t
		}
		list = append(list, u)
	}
	return list, total, nil
}

func buildUserListWhere(filter domain.UserListFilter) (string, []interface{}) {
	where := "WHERE 1=1"
	args := []interface{}{}
	if filter.Keyword != "" {
		kw := search.Trim(filter.Keyword)
		if search.Usable(kw) {
			where += " AND (phone LIKE ? OR MATCH(display_name) AGAINST(? IN NATURAL LANGUAGE MODE))"
			args = append(args, kw+"%", kw)
		} else if kw != "" {
			where += " AND phone LIKE ?"
			args = append(args, kw+"%")
		}
	}
	if filter.Status != nil {
		where += " AND status = ?"
		args = append(args, *filter.Status)
	}
	if filter.Role != nil {
		where += " AND role = ?"
		args = append(args, *filter.Role)
	}
	return where, args
}

func (r *UserRepo) UpdateDynamic(id int64, fields map[string]interface{}) error {
	if len(fields) == 0 {
		return domain.ErrNoUpdateFields
	}
	setParts := make([]string, 0, len(fields))
	args := make([]interface{}, 0, len(fields)+1)
	for k, v := range fields {
		setParts = append(setParts, k+" = ?")
		args = append(args, v)
	}
	setParts = append(setParts, "updated_at = ?")
	args = append(args, time.Now(), id)
	query := "UPDATE user SET " + strings.Join(setParts, ", ") + " WHERE id = ?"
	_, err := db.MySQL.Exec(query, args...)
	return err
}

func (r *UserRepo) SoftDelete(id int64) error {
	_, err := db.MySQL.Exec(`UPDATE user SET status = 0, updated_at = ? WHERE id = ?`, time.Now(), id)
	return err
}

// DeactivateAccount 用户自助注销：软删 + 匿名化展示信息 + 释放手机号唯一约束
func (r *UserRepo) DeactivateAccount(id int64) error {
	now := time.Now()
	deletedPhone := fmt.Sprintf("deleted_%d_%d", id, now.Unix())

	tx, err := db.MySQL.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var status int
	err = tx.QueryRow(`SELECT status FROM user WHERE id = ? FOR UPDATE`, id).Scan(&status)
	if err != nil {
		return err
	}
	if status != 1 {
		return domain.ErrAccountDeactivated
	}

	_, err = tx.Exec(`
		UPDATE user
		SET status = 0,
		    display_name = ?,
		    avatar = '',
		    phone = ?,
		    updated_at = ?
		WHERE id = ? AND status = 1
	`, domain.AnonymousDisplayName, deletedPhone, now, id)
	if err != nil {
		return err
	}
	return tx.Commit()
}

func (r *UserRepo) UpdateRole(id int64, role int) error {
	_, err := db.MySQL.Exec(`UPDATE user SET role = ?, updated_at = ? WHERE id = ?`, role, time.Now(), id)
	return err
}

func (r *UserRepo) UpdatePassword(id int64, hash string) error {
	_, err := db.MySQL.Exec(`UPDATE user SET password_hash = ?, updated_at = ? WHERE id = ?`, hash, time.Now(), id)
	return err
}

func (r *UserRepo) CreateAdmin(input domain.CreateUserInput, passwordHash string) (int64, error) {
	res, err := db.MySQL.Exec(`
		INSERT INTO user (phone, password_hash, display_name, avatar, role, status, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, NOW(), NOW())
	`, input.Phone, passwordHash, input.DisplayName, input.Avatar, input.Role, input.Status)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func ToUserItem(u domain.User) domain.UserItem {
	item := domain.UserItem{
		ID:          u.ID,
		Phone:       u.Phone,
		DisplayName: u.DisplayName,
		Avatar:      u.Avatar,
		Role:        u.Role,
		Status:      u.Status,
		LoginCount:  u.LoginCount,
		CreatedAt:   u.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:   u.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
	if u.LastLoginAt != nil {
		item.LastLoginAt = u.LastLoginAt.Format("2006-01-02 15:04:05")
	}
	return item
}

// internal/module/auth/repo/user_repo.go
func (r *UserRepo) UpdateAvatar(id int64, avatarURL string) error {
	_, err := db.MySQL.Exec(`UPDATE user SET avatar = ?, updated_at = ? WHERE id = ?`, avatarURL, time.Now(), id)
	return err
}
