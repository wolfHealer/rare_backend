package repo

import (
	"database/sql"
	"errors"
	"strings"

	"github.com/go-sql-driver/mysql"

	"rare_backend/internal/module/knowledge/domain"
	"rare_backend/internal/pkg/db"
)

type ArticleTagRepo struct{}

func NewArticleTagRepo() *ArticleTagRepo {
	return &ArticleTagRepo{}
}

func (r *ArticleTagRepo) List(filter domain.ArticleTagListFilter) ([]domain.ArticleTagItem, error) {
	where := "WHERE 1=1"
	args := []interface{}{}
	if t := strings.TrimSpace(filter.Type); t != "" {
		where += " AND type = ?"
		args = append(args, t)
	}

	rows, err := db.MySQL.Query(
		`SELECT id, name, type FROM article_tag `+where+` ORDER BY id ASC`,
		args...,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := make([]domain.ArticleTagItem, 0)
	for rows.Next() {
		var item domain.ArticleTagItem
		if err := rows.Scan(&item.ID, &item.Name, &item.Type); err != nil {
			return nil, err
		}
		list = append(list, item)
	}
	return list, rows.Err()
}

func (r *ArticleTagRepo) GetByID(id uint) (*domain.ArticleTagItem, error) {
	var item domain.ArticleTagItem
	err := db.MySQL.QueryRow(
		`SELECT id, name, type FROM article_tag WHERE id = ?`, id,
	).Scan(&item.ID, &item.Name, &item.Type)
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *ArticleTagRepo) Create(in domain.CreateArticleTagInput) (int64, error) {
	tagType := in.Type
	if tagType == "" {
		tagType = "general"
	}
	res, err := db.MySQL.Exec(
		`INSERT INTO article_tag (name, type) VALUES (?, ?)`,
		in.Name, tagType,
	)
	if err != nil {
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
			return 0, domain.ErrArticleTagNameExists
		}
		return 0, err
	}
	return res.LastInsertId()
}

func (r *ArticleTagRepo) Update(id uint, in domain.UpdateArticleTagInput) error {
	setParts := []string{}
	args := []interface{}{}
	if in.Name != nil {
		setParts = append(setParts, "name = ?")
		args = append(args, *in.Name)
	}
	if in.Type != nil {
		setParts = append(setParts, "type = ?")
		args = append(args, *in.Type)
	}
	if len(setParts) == 0 {
		return domain.ErrNoUpdateFields
	}
	args = append(args, id)
	_, err := db.MySQL.Exec(
		`UPDATE article_tag SET `+strings.Join(setParts, ", ")+` WHERE id = ?`,
		args...,
	)
	if err != nil {
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
			return domain.ErrArticleTagNameExists
		}
		return err
	}
	return nil
}

func (r *ArticleTagRepo) Delete(id uint) error {
	var relCount int
	if err := db.MySQL.QueryRow(
		`SELECT COUNT(*) FROM article_tag_rel WHERE tag_id = ?`, id,
	).Scan(&relCount); err != nil {
		return err
	}
	if relCount > 0 {
		return domain.ErrArticleTagInUse
	}

	res, err := db.MySQL.Exec(`DELETE FROM article_tag WHERE id = ?`, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *ArticleTagRepo) Exists(id uint) (bool, error) {
	var exists uint
	err := db.MySQL.QueryRow(`SELECT id FROM article_tag WHERE id = ?`, id).Scan(&exists)
	if err == sql.ErrNoRows {
		return false, nil
	}
	return err == nil, err
}
