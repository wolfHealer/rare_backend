package repo

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"rare_backend/internal/module/community/domain"
	"rare_backend/internal/pkg/db"
)

// PostRepo 帖子数据访问
type PostRepo struct{}

func NewPostRepo() *PostRepo {
	return &PostRepo{}
}

func (r *PostRepo) ExistsActive(postID int64) (bool, error) {
	var count int
	err := db.MySQL.QueryRow(
		`SELECT COUNT(*) FROM post WHERE id = ? AND status = 1`, postID,
	).Scan(&count)
	return count > 0, err
}

func (r *PostRepo) GetOwnerAndStatus(postID int64) (userID int64, status int, err error) {
	err = db.MySQL.QueryRow(
		`SELECT user_id, status FROM post WHERE id = ?`, postID,
	).Scan(&userID, &status)
	return
}

// ExistsReportable 帖子存在且未删除（status != 3）
func (r *PostRepo) ExistsReportable(postID int64) (ownerID int64, ok bool, err error) {
	var status int
	err = db.MySQL.QueryRow(
		`SELECT user_id, status FROM post WHERE id = ?`, postID,
	).Scan(&ownerID, &status)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, false, nil
		}
		return 0, false, err
	}
	return ownerID, status != 3, nil
}

func (r *PostRepo) GetOwnerActive(postID int64) (userID int64, err error) {
	err = db.MySQL.QueryRow(
		`SELECT user_id FROM post WHERE id = ? AND status = 1`, postID,
	).Scan(&userID)
	return
}

func (r *PostRepo) IsLiked(postID, userID int64) (bool, error) {
	var count int
	err := db.MySQL.QueryRow(
		`SELECT COUNT(*) FROM post_like WHERE post_id = ? AND user_id = ?`, postID, userID,
	).Scan(&count)
	return count > 0, err
}

func (r *PostRepo) IsFavorited(postID, userID int64) (bool, error) {
	var count int
	err := db.MySQL.QueryRow(
		`SELECT COUNT(*) FROM post_favorite WHERE post_id = ? AND user_id = ?`, postID, userID,
	).Scan(&count)
	return count > 0, err
}

func (r *PostRepo) GetInteractionStatus(postID, userID int64) (liked, favorited bool, err error) {
	if userID <= 0 {
		return false, false, nil
	}
	err = db.MySQL.QueryRow(`
		SELECT
			EXISTS(SELECT 1 FROM post_like WHERE post_id = ? AND user_id = ?),
			EXISTS(SELECT 1 FROM post_favorite WHERE post_id = ? AND user_id = ?)
	`, postID, userID, postID, userID).Scan(&liked, &favorited)
	return liked, favorited, err
}

func (r *PostRepo) BatchLikedPostIDs(userID int64, postIDs []int64) (map[int64]bool, error) {
	return batchPostIDsByTable(userID, postIDs, "post_like")
}

func (r *PostRepo) BatchFavoritedPostIDs(userID int64, postIDs []int64) (map[int64]bool, error) {
	return batchPostIDsByTable(userID, postIDs, "post_favorite")
}

func batchPostIDsByTable(userID int64, postIDs []int64, table string) (map[int64]bool, error) {
	result := make(map[int64]bool)
	if userID <= 0 || len(postIDs) == 0 {
		return result, nil
	}

	placeholders := make([]string, len(postIDs))
	args := make([]interface{}, 0, len(postIDs)+1)
	args = append(args, userID)
	for i, id := range postIDs {
		placeholders[i] = "?"
		args = append(args, id)
	}

	query := fmt.Sprintf(
		"SELECT post_id FROM %s WHERE user_id = ? AND post_id IN (%s)",
		table, strings.Join(placeholders, ","),
	)
	rows, err := db.MySQL.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var postID int64
		if err := rows.Scan(&postID); err != nil {
			continue
		}
		result[postID] = true
	}
	return result, rows.Err()
}

func (r *PostRepo) Unlike(postID, userID int64) error {
	_, err := db.MySQL.Exec(
		`DELETE FROM post_like WHERE post_id = ? AND user_id = ?`, postID, userID,
	)
	return err
}

func (r *PostRepo) Like(postID, userID int64) error {
	_, err := db.MySQL.Exec(
		`INSERT INTO post_like (post_id, user_id) VALUES (?, ?)`, postID, userID,
	)
	return err
}

func (r *PostRepo) RefreshLikeCount(postID int64) (int, error) {
	_, err := db.MySQL.Exec(`
		UPDATE post SET like_count = (SELECT COUNT(*) FROM post_like WHERE post_id = ?)
		WHERE id = ?
	`, postID, postID)
	if err != nil {
		return 0, err
	}
	var likeCount int
	err = db.MySQL.QueryRow(`SELECT like_count FROM post WHERE id = ?`, postID).Scan(&likeCount)
	return likeCount, err
}

func (r *PostRepo) Unfavorite(postID, userID int64) error {
	_, err := db.MySQL.Exec(
		`DELETE FROM post_favorite WHERE post_id = ? AND user_id = ?`, postID, userID,
	)
	return err
}

func (r *PostRepo) Favorite(postID, userID int64) error {
	_, err := db.MySQL.Exec(
		`INSERT INTO post_favorite (post_id, user_id) VALUES (?, ?)`, postID, userID,
	)
	return err
}

func (r *PostRepo) RefreshFavoriteCount(postID int64) (int, error) {
	_, err := db.MySQL.Exec(`
		UPDATE post SET favorite_count = (SELECT COUNT(*) FROM post_favorite WHERE post_id = ?)
		WHERE id = ?
	`, postID, postID)
	if err != nil {
		return 0, err
	}
	var favoriteCount int
	err = db.MySQL.QueryRow(`SELECT favorite_count FROM post WHERE id = ?`, postID).Scan(&favoriteCount)
	return favoriteCount, err
}

func (r *PostRepo) Create(in domain.CreatePostInput) (int64, error) {
	imagesJSON, _ := json.Marshal(in.Images)
	if in.Images == nil {
		imagesJSON = []byte("[]")
	}
	res, err := db.MySQL.Exec(`
		INSERT INTO post (
			user_id, disease_id, category_id, type, title, content, images,
			view_count, like_count, comment_count, favorite_count,
			is_top, is_recommend, status
		) VALUES (?, ?, ?, ?, ?, ?, ?, 0, 0, 0, 0, 0, 0, 0)
	`, in.UserID, in.DiseaseID, in.CategoryID, in.Type, in.Title, in.Content, string(imagesJSON))
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (r *PostRepo) SoftDelete(postID int64) error {
	_, err := db.MySQL.Exec(
		`UPDATE post SET status = 3, updated_at = ? WHERE id = ?`, time.Now(), postID,
	)
	return err
}

// ListMyPosts 当前用户发布的帖子（按创建时间倒序）
func (r *PostRepo) ListMyPosts(filter domain.MyPostsFilter) ([]domain.PostListItem, int64, error) {
	where := "p.user_id = ?"
	args := []interface{}{filter.UserID}

	dbStatus, hasStatus, err := domain.MyPostStatusToDB(filter.Status)
	if err != nil {
		return nil, 0, err
	}
	if hasStatus {
		where += " AND p.status = ?"
		args = append(args, dbStatus)
	} else {
		where += " AND p.status != 3"
	}

	if kw := strings.TrimSpace(filter.Keyword); kw != "" {
		where += " AND (p.title LIKE ? OR p.content LIKE ?)"
		like := "%" + kw + "%"
		args = append(args, like, like)
	}

	countQuery := "SELECT COUNT(*) FROM post p WHERE " + where
	var total int64
	if err := db.MySQL.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	offset := (filter.Page - 1) * filter.PageSize
	listArgs := append(append([]interface{}{}, args...), filter.PageSize, offset)
	query := `
		SELECT
			p.id, p.user_id, u.display_name, p.disease_id, p.category_id, p.type, p.title, p.content, p.images,
			p.view_count, p.like_count, p.comment_count, p.favorite_count,
			p.is_top, p.is_recommend, p.status, p.reject_reason, p.created_at, p.updated_at
		FROM post p
		LEFT JOIN user u ON p.user_id = u.id
		WHERE ` + where + `
		ORDER BY p.created_at DESC
		LIMIT ? OFFSET ?`

	rows, err := db.MySQL.Query(query, listArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	list := make([]domain.PostListItem, 0)
	for rows.Next() {
		item, err := scanPostListItem(rows)
		if err != nil {
			return nil, 0, err
		}
		list = append(list, item)
	}
	return list, total, rows.Err()
}

func scanPostListItem(rows *sql.Rows) (domain.PostListItem, error) {
	var item domain.PostListItem
	var imagesBytes []byte
	var displayName, title, rejectReason sql.NullString
	var isTopInt, isRecommendInt int
	err := rows.Scan(
		&item.ID, &item.UserID, &displayName, &item.DiseaseID, &item.CategoryID, &item.Type, &title, &item.Content, &imagesBytes,
		&item.ViewCount, &item.LikeCount, &item.CommentCount, &item.FavoriteCount,
		&isTopInt, &isRecommendInt, &item.Status, &rejectReason, &item.CreatedAt, &item.UpdatedAt,
	)
	if err != nil {
		return item, err
	}
	if displayName.Valid {
		item.DisplayName = displayName.String
	}
	if title.Valid {
		item.Title = title.String
	}
	if rejectReason.Valid {
		item.RejectReason = rejectReason.String
	}
	item.IsTop = isTopInt == 1
	item.IsRecommend = isRecommendInt == 1
	if len(imagesBytes) > 0 {
		_ = json.Unmarshal(imagesBytes, &item.Images)
	}
	if item.Images == nil {
		item.Images = []string{}
	}
	return item, nil
}

func (r *PostRepo) UpdateDynamic(postID int64, fields map[string]interface{}) error {
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
	args = append(args, time.Now(), postID)
	query := "UPDATE post SET " + strings.Join(setParts, ", ") + " WHERE id = ?"
	_, err := db.MySQL.Exec(query, args...)
	return err
}

func (r *PostRepo) ErrNoRows(err error) bool {
	return err == sql.ErrNoRows
}
