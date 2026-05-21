package repo

import (
	"database/sql"
	"time"

	"rare_backend/internal/module/community/domain"
	"rare_backend/internal/pkg/db"
)

// CommentRepo 评论数据访问
type CommentRepo struct{}

func NewCommentRepo() *CommentRepo {
	return &CommentRepo{}
}

type parentCommentRow struct {
	ID     int64
	RootID int64
}

func (r *CommentRepo) FindParentOnPost(parentID, postID int64) (*parentCommentRow, error) {
	var row parentCommentRow
	err := db.MySQL.QueryRow(
		`SELECT id, root_id FROM comment WHERE id = ? AND target_type = 'post' AND target_id = ?`,
		parentID, postID,
	).Scan(&row.ID, &row.RootID)
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *CommentRepo) CreateWithTx(in domain.CreateCommentInput, parentID, rootID int64) (*domain.CreateCommentResult, error) {
	tx, err := db.MySQL.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	now := time.Now()
	res, err := tx.Exec(`
		INSERT INTO comment (target_id, target_type, user_id, content, parent_id, root_id, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, in.PostID, "post", in.UserID, in.Content, parentID, rootID, now)
	if err != nil {
		return nil, err
	}
	commentID, _ := res.LastInsertId()

	if parentID > 0 {
		_, err = tx.Exec(`UPDATE comment SET reply_count = reply_count + 1 WHERE id = ?`, parentID)
	} else {
		_, err = tx.Exec(`UPDATE post SET comment_count = comment_count + 1 WHERE id = ?`, in.PostID)
	}
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}

	var displayName sql.NullString
	_ = db.MySQL.QueryRow(`SELECT display_name FROM user WHERE id = ?`, in.UserID).Scan(&displayName)

	return &domain.CreateCommentResult{
		ID:          commentID,
		PostID:      in.PostID,
		UserID:      in.UserID,
		DisplayName: displayName.String,
		Content:     in.Content,
		ParentID:    parentID,
		RootID:      rootID,
		CreatedAt:   now,
	}, nil
}

func (r *CommentRepo) GetOwnerActive(commentID int64) (userID int64, err error) {
	err = db.MySQL.QueryRow(
		`SELECT user_id FROM comment WHERE id = ? AND status = 1`, commentID,
	).Scan(&userID)
	return
}

func (r *CommentRepo) GetOwnerAndPost(commentID int64) (userID, postID int64, err error) {
	err = db.MySQL.QueryRow(
		`SELECT user_id, target_id FROM comment WHERE id = ? AND status = 1`, commentID,
	).Scan(&userID, &postID)
	return
}

func (r *CommentRepo) UpdateContent(commentID int64, content string) error {
	_, err := db.MySQL.Exec(
		`UPDATE comment SET content = ?, updated_at = ? WHERE id = ?`, content, time.Now(), commentID,
	)
	return err
}

func (r *CommentRepo) SoftDeleteWithDecrement(commentID, postID int64) error {
	tx, err := db.MySQL.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err = tx.Exec(
		`UPDATE comment SET status = 0, updated_at = ? WHERE id = ?`, time.Now(), commentID,
	); err != nil {
		return err
	}
	if _, err = tx.Exec(
		`UPDATE post SET comment_count = comment_count - 1 WHERE id = ? AND comment_count > 0`, postID,
	); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *CommentRepo) ErrNoRows(err error) bool {
	return err == sql.ErrNoRows
}
