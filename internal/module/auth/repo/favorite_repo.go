package repo

import (
	"database/sql"
	"strings"

	"rare_backend/internal/module/auth/domain"
	"rare_backend/internal/pkg/db"
)

type FavoriteRepo struct{}

func NewFavoriteRepo() *FavoriteRepo {
	return &FavoriteRepo{}
}

func (r *FavoriteRepo) ListPostFavorites(filter domain.FavoriteListFilter) ([]domain.FavoriteItem, int64, error) {
	where := "pf.user_id = ? AND p.status = 1"
	args := []interface{}{filter.UserID}

	if kw := strings.TrimSpace(filter.Keyword); kw != "" {
		where += " AND (p.title LIKE ? OR p.content LIKE ?)"
		like := "%" + kw + "%"
		args = append(args, like, like)
	}

	countQuery := `
		SELECT COUNT(*)
		FROM post_favorite pf
		INNER JOIN post p ON p.id = pf.post_id
		WHERE ` + where
	var total int64
	if err := db.MySQL.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	offset := (filter.Page - 1) * filter.PageSize
	listArgs := append(append([]interface{}{}, args...), filter.PageSize, offset)
	query := `
		SELECT pf.id, pf.post_id, pf.created_at,
		       COALESCE(p.title, ''), p.content, p.type, p.images
		FROM post_favorite pf
		INNER JOIN post p ON p.id = pf.post_id
		WHERE ` + where + `
		ORDER BY pf.created_at DESC
		LIMIT ? OFFSET ?`

	rows, err := db.MySQL.Query(query, listArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	list := make([]domain.FavoriteItem, 0)
	for rows.Next() {
		var item domain.FavoriteItem
		var content, postType string
		var imagesJSON sql.NullString
		if err := rows.Scan(
			&item.ID, &item.TargetID, &item.CreatedAt,
			&item.Title, &content, &postType, &imagesJSON,
		); err != nil {
			return nil, 0, err
		}
		item.TargetType = domain.TargetTypePost
		item.Summary = content
		item.Extra = map[string]string{"type": postType}
		if imagesJSON.Valid && imagesJSON.String != "" {
			item.Cover = firstImageURL(imagesJSON.String)
		}
		list = append(list, item)
	}
	return list, total, rows.Err()
}

func (r *FavoriteRepo) GetPostFavoriteRecord(userID, favoriteID int64) (postID int64, err error) {
	err = db.MySQL.QueryRow(
		`SELECT post_id FROM post_favorite WHERE id = ? AND user_id = ?`,
		favoriteID, userID,
	).Scan(&postID)
	return postID, err
}

func (r *FavoriteRepo) DeletePostFavoriteByID(userID, favoriteID int64) (postID int64, err error) {
	postID, err = r.GetPostFavoriteRecord(userID, favoriteID)
	if err != nil {
		return 0, err
	}
	res, err := db.MySQL.Exec(
		`DELETE FROM post_favorite WHERE id = ? AND user_id = ?`,
		favoriteID, userID,
	)
	if err != nil {
		return 0, err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return 0, sql.ErrNoRows
	}
	return postID, nil
}

func (r *FavoriteRepo) DeletePostFavoriteByTarget(userID, postID int64) error {
	res, err := db.MySQL.Exec(
		`DELETE FROM post_favorite WHERE user_id = ? AND post_id = ?`,
		userID, postID,
	)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *FavoriteRepo) RefreshPostFavoriteCount(postID int64) error {
	_, err := db.MySQL.Exec(`
		UPDATE post SET favorite_count = (
			SELECT COUNT(*) FROM post_favorite WHERE post_id = ?
		) WHERE id = ?
	`, postID, postID)
	return err
}

func firstImageURL(imagesJSON string) string {
	imagesJSON = strings.TrimSpace(imagesJSON)
	if imagesJSON == "" || imagesJSON == "[]" || imagesJSON == "null" {
		return ""
	}
	// 简单解析 ["url1","url2"]
	imagesJSON = strings.TrimPrefix(imagesJSON, "[")
	imagesJSON = strings.TrimSuffix(imagesJSON, "]")
	parts := strings.Split(imagesJSON, ",")
	if len(parts) == 0 {
		return ""
	}
	url := strings.TrimSpace(parts[0])
	return strings.Trim(url, `"`)
}
