package repo

import (
	"strings"
	"time"

	"rare_backend/internal/module/region/domain"
	"rare_backend/internal/pkg/db"
)

type RegionRepo struct{}

func NewRegionRepo() *RegionRepo {
	return &RegionRepo{}
}

func (r *RegionRepo) CodeExists(code string) (bool, error) {
	var exists uint
	err := db.MySQL.QueryRow("SELECT id FROM region WHERE code = ?", code).Scan(&exists)
	if err == nil {
		return true, nil
	}
	if IsNoRows(err) {
		return false, nil
	}
	return false, err
}

func (r *RegionRepo) ParentCodeExists(parentCode string) (bool, error) {
	var exists uint
	err := db.MySQL.QueryRow("SELECT id FROM region WHERE code = ?", parentCode).Scan(&exists)
	if err == nil {
		return true, nil
	}
	if IsNoRows(err) {
		return false, nil
	}
	return false, err
}

func (r *RegionRepo) Exists(id uint) (bool, error) {
	var exists uint
	err := db.MySQL.QueryRow("SELECT id FROM region WHERE id = ?", id).Scan(&exists)
	if err == nil {
		return true, nil
	}
	if IsNoRows(err) {
		return false, nil
	}
	return false, err
}

func (r *RegionRepo) Create(in domain.CreateRegionInput) (int64, error) {
	now := time.Now()
	res, err := db.MySQL.Exec(`
		INSERT INTO region (code, name, full_name, parent_code, level, sort, is_enabled, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, in.Code, in.Name, in.FullName, in.ParentCode, in.Level, in.Sort, in.IsEnabled, now, now)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (r *RegionRepo) Update(id uint, setParts []string, args []interface{}) error {
	if len(setParts) == 0 {
		return domain.ErrNoUpdateFields
	}
	setParts = append(setParts, "updated_at = ?")
	args = append(args, time.Now(), id)
	query := "UPDATE region SET " + strings.Join(setParts, ", ") + " WHERE id = ?"
	_, err := db.MySQL.Exec(query, args...)
	return err
}

func (r *RegionRepo) GetCodeByID(id uint) (string, error) {
	var code string
	err := db.MySQL.QueryRow("SELECT code FROM region WHERE id = ?", id).Scan(&code)
	return code, err
}

func (r *RegionRepo) CountChildren(parentCode string) (int64, error) {
	var count int64
	err := db.MySQL.QueryRow("SELECT COUNT(*) FROM region WHERE parent_code = ?", parentCode).Scan(&count)
	return count, err
}

func (r *RegionRepo) Delete(id uint) error {
	_, err := db.MySQL.Exec("DELETE FROM region WHERE id = ?", id)
	return err
}

func (r *RegionRepo) List(filter domain.RegionListFilter) (*domain.RegionListResult, error) {
	whereClause := "WHERE 1=1"
	args := []interface{}{}

	if filter.ParentCode != "" {
		whereClause += " AND parent_code = ?"
		args = append(args, filter.ParentCode)
	}
	if filter.Level != 0 {
		whereClause += " AND level = ?"
		args = append(args, filter.Level)
	}
	if filter.IsEnabled != 0 {
		whereClause += " AND is_enabled = ?"
		args = append(args, filter.IsEnabled)
	}
	if filter.Keyword != "" {
		whereClause += " AND (name LIKE ? OR full_name LIKE ? OR code LIKE ?)"
		args = append(args, "%"+filter.Keyword+"%", "%"+filter.Keyword+"%", "%"+filter.Keyword+"%")
	}

	var total int64
	if err := db.MySQL.QueryRow("SELECT COUNT(*) FROM region "+whereClause, args...).Scan(&total); err != nil {
		return nil, err
	}

	offset := (filter.Page - 1) * filter.PageSize
	listArgs := append(append([]interface{}{}, args...), filter.PageSize, offset)

	rows, err := db.MySQL.Query(`
		SELECT id, code, name, full_name, parent_code, level, sort, is_enabled, created_at, updated_at
		FROM region `+whereClause+`
		ORDER BY sort ASC, code ASC
		LIMIT ? OFFSET ?
	`, listArgs...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []domain.RegionItem
	for rows.Next() {
		var row domain.RegionRow
		if err := rows.Scan(
			&row.ID, &row.Code, &row.Name, &row.FullName, &row.ParentCode,
			&row.Level, &row.Sort, &row.IsEnabled, &row.CreatedAt, &row.UpdatedAt,
		); err != nil {
			continue
		}
		list = append(list, domain.ToRegionItem(row))
	}
	if list == nil {
		list = []domain.RegionItem{}
	}
	return &domain.RegionListResult{
		List: list, Total: total, Page: filter.Page, PageSize: filter.PageSize,
	}, nil
}

func (r *RegionRepo) ListFlat(isEnabled int, maxLevel int) ([]domain.FlatRegion, error) {
	query := `
		SELECT code, name, level, parent_code
		FROM region
		WHERE is_enabled = ?
	`
	args := []interface{}{isEnabled}
	if maxLevel > 0 {
		query += " AND level <= ?"
		args = append(args, maxLevel)
	}
	query += " ORDER BY sort ASC, code ASC"

	rows, err := db.MySQL.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var flatList []domain.FlatRegion
	for rows.Next() {
		var item domain.FlatRegion
		if err := rows.Scan(&item.Code, &item.Name, &item.Level, &item.ParentCode); err != nil {
			return nil, err
		}
		flatList = append(flatList, item)
	}
	return flatList, nil
}

func (r *RegionRepo) GetByID(id uint) (*domain.RegionItem, error) {
	var row domain.RegionRow
	err := db.MySQL.QueryRow(`
		SELECT id, code, name, full_name, parent_code, level, sort, is_enabled, created_at, updated_at
		FROM region WHERE id = ?
	`, id).Scan(
		&row.ID, &row.Code, &row.Name, &row.FullName, &row.ParentCode,
		&row.Level, &row.Sort, &row.IsEnabled, &row.CreatedAt, &row.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	item := domain.ToRegionItem(row)
	return &item, nil
}
