package repo

import (
	"database/sql"
	"time"

	"rare_backend/internal/module/knowledge/domain"
	"rare_backend/internal/pkg/db"
)

type CategoryRepo struct{}

func NewCategoryRepo() *CategoryRepo {
	return &CategoryRepo{}
}

func (r *CategoryRepo) CodeExistsActive(code string) (bool, error) {
	var exists uint
	err := db.MySQL.QueryRow("SELECT id FROM category WHERE code = ? AND status = 1", code).Scan(&exists)
	if err == nil {
		return true, nil
	}
	if IsNoRows(err) {
		return false, nil
	}
	return false, err
}

func (r *CategoryRepo) CodeExistsActiveExcept(code string, id uint) (bool, error) {
	var exists uint
	err := db.MySQL.QueryRow("SELECT id FROM category WHERE code = ? AND id != ? AND status = 1", code, id).Scan(&exists)
	if err == nil {
		return true, nil
	}
	if IsNoRows(err) {
		return false, nil
	}
	return false, err
}

func (r *CategoryRepo) Exists(id uint) (bool, error) {
	return ExistsByID("category", id)
}

func (r *CategoryRepo) ExistsActive(id uint) (bool, error) {
	return ExistsActiveByID("category", id)
}

func (r *CategoryRepo) Create(in domain.CreateCategoryInput) (int64, error) {
	now := time.Now()
	result, err := db.MySQL.Exec(`
		INSERT INTO category 
		(parent_id, level, name, code, description, icon_url, sort_order, status, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, 1, ?, ?)
	`, in.ParentID, in.Level, in.Name, in.Code, in.Description, in.IconURL, in.SortOrder, now, now)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (r *CategoryRepo) UpdateDynamic(id uint, setParts []string, args []interface{}) error {
	if len(setParts) == 0 {
		return nil
	}
	query := "UPDATE category SET " + JoinUpdateFields(setParts) + " WHERE id = ?"
	_, err := db.MySQL.Exec(query, args...)
	return err
}

func (r *CategoryRepo) SoftDelete(id uint) error {
	return SoftDelete("category", id)
}

func (r *CategoryRepo) List(filter domain.CategoryListFilter) ([]domain.CategoryItem, error) {
	whereClause := "WHERE status = 1"
	args := []interface{}{}

	if filter.ParentID != nil {
		whereClause += " AND parent_id = ?"
		args = append(args, *filter.ParentID)
	}
	if filter.Level != nil {
		whereClause += " AND level = ?"
		args = append(args, *filter.Level)
	}

	query := `
		SELECT id, parent_id, level, name, code, description, icon_url, sort_order, status, created_at, updated_at
		FROM category
		` + whereClause + `
		ORDER BY sort_order ASC, id ASC
	`

	rows, err := db.MySQL.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []domain.CategoryItem
	for rows.Next() {
		var cat struct {
			ID          uint
			ParentID    uint
			Level       int
			Name        string
			Code        string
			Description sql.NullString
			IconURL     sql.NullString
			SortOrder   int
			Status      int8
			CreatedAt   string
			UpdatedAt   string
		}
		if err := rows.Scan(
			&cat.ID, &cat.ParentID, &cat.Level, &cat.Name, &cat.Code,
			&cat.Description, &cat.IconURL, &cat.SortOrder, &cat.Status,
			&cat.CreatedAt, &cat.UpdatedAt,
		); err != nil {
			continue
		}
		list = append(list, domain.CategoryItem{
			ID:          cat.ID,
			ParentID:    cat.ParentID,
			Level:       cat.Level,
			Name:        cat.Name,
			Code:        cat.Code,
			Description: cat.Description.String,
			IconURL:     cat.IconURL.String,
			SortOrder:   cat.SortOrder,
			Status:      int(cat.Status),
			CreatedAt:   cat.CreatedAt,
			UpdatedAt:   cat.UpdatedAt,
		})
	}
	return list, nil
}

func (r *CategoryRepo) ListFlatForTree(filter domain.CategoryTreeFilter) ([]domain.CategoryTreeNode, error) {
	whereClause := "WHERE 1=1"
	args := []interface{}{}

	if filter.Keyword != "" {
		whereClause += " AND name LIKE ?"
		args = append(args, "%"+filter.Keyword+"%")
	}
	if filter.Status != nil {
		whereClause += " AND status = ?"
		args = append(args, *filter.Status)
	}

	query := `
		SELECT id, parent_id, level, name, code, description, icon_url, sort_order, status, created_at, updated_at
		FROM category
		` + whereClause + `
		ORDER BY sort_order ASC, id ASC
	`

	rows, err := db.MySQL.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var all []domain.CategoryTreeNode
	for rows.Next() {
		var cat struct {
			ID          uint
			ParentID    uint
			Level       int
			Name        string
			Code        string
			Description sql.NullString
			IconURL     sql.NullString
			SortOrder   int
			Status      int8
			CreatedAt   time.Time
			UpdatedAt   time.Time
		}
		if err := rows.Scan(
			&cat.ID, &cat.ParentID, &cat.Level, &cat.Name, &cat.Code,
			&cat.Description, &cat.IconURL, &cat.SortOrder, &cat.Status,
			&cat.CreatedAt, &cat.UpdatedAt,
		); err != nil {
			continue
		}
		all = append(all, domain.CategoryTreeNode{
			ID:          cat.ID,
			ParentID:    cat.ParentID,
			Level:       cat.Level,
			Name:        cat.Name,
			Code:        cat.Code,
			Description: cat.Description.String,
			IconURL:     cat.IconURL.String,
			SortOrder:   cat.SortOrder,
			Status:      int(cat.Status),
			CreatedAt:   cat.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:   cat.UpdatedAt.Format("2006-01-02 15:04:05"),
			Children:    []domain.CategoryTreeNode{},
		})
	}
	return all, nil
}
