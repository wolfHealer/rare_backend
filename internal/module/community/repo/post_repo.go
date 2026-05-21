package repo

import (
	"database/sql"
	"encoding/json"
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
		`UPDATE post SET status = 0, updated_at = ? WHERE id = ?`, time.Now(), postID,
	)
	return err
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
