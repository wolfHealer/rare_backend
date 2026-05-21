package repo

import (
	"time"

	"rare_backend/internal/module/knowledge/domain"
	"rare_backend/internal/pkg/db"
)

type TagRepo struct{}

func NewTagRepo() *TagRepo {
	return &TagRepo{}
}

func (r *TagRepo) CodeExistsActive(code string) (bool, error) {
	var exists uint
	err := db.MySQL.QueryRow("SELECT id FROM tag WHERE code = ? AND status = 1", code).Scan(&exists)
	if err == nil {
		return true, nil
	}
	if IsNoRows(err) {
		return false, nil
	}
	return false, err
}

func (r *TagRepo) CodeExistsActiveExcept(code string, id uint) (bool, error) {
	var exists uint
	err := db.MySQL.QueryRow("SELECT id FROM tag WHERE code = ? AND id != ? AND status = 1", code, id).Scan(&exists)
	if err == nil {
		return true, nil
	}
	if IsNoRows(err) {
		return false, nil
	}
	return false, err
}

func (r *TagRepo) Exists(id uint) (bool, error) {
	return ExistsByID("tag", id)
}

func (r *TagRepo) ExistsActive(id uint) (bool, error) {
	return ExistsActiveByID("tag", id)
}

func (r *TagRepo) Create(in domain.CreateTagInput) (int64, error) {
	now := time.Now()
	result, err := db.MySQL.Exec(`
		INSERT INTO tag 
		(name, code, sort_order, status, created_at, updated_at)
		VALUES (?, ?, ?, 1, ?, ?)
	`, in.Name, in.Code, in.SortOrder, now, now)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (r *TagRepo) UpdateDynamic(id uint, setParts []string, args []interface{}) error {
	if len(setParts) == 0 {
		return nil
	}
	query := "UPDATE tag SET " + JoinUpdateFields(setParts) + " WHERE id = ?"
	_, err := db.MySQL.Exec(query, args...)
	return err
}

func (r *TagRepo) SoftDelete(id uint) error {
	return SoftDelete("tag", id)
}

func (r *TagRepo) ListActive() ([]domain.TagItem, error) {
	query := `
		SELECT id, name, code, sort_order, status, created_at, updated_at
		FROM tag
		WHERE status = 1
		ORDER BY sort_order ASC, id ASC
	`
	rows, err := db.MySQL.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []domain.TagItem
	for rows.Next() {
		var t struct {
			ID        uint
			Name      string
			Code      string
			SortOrder int
			Status    int8
			CreatedAt string
			UpdatedAt string
		}
		if err := rows.Scan(&t.ID, &t.Name, &t.Code, &t.SortOrder, &t.Status, &t.CreatedAt, &t.UpdatedAt); err != nil {
			continue
		}
		list = append(list, domain.TagItem{
			ID:        t.ID,
			Name:      t.Name,
			Code:      t.Code,
			SortOrder: t.SortOrder,
			Status:    int(t.Status),
			CreatedAt: t.CreatedAt,
			UpdatedAt: t.UpdatedAt,
		})
	}
	return list, nil
}
