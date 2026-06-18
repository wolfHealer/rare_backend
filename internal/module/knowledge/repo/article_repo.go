package repo

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"rare_backend/internal/module/knowledge/domain"
	"rare_backend/internal/pkg/db"
	"rare_backend/internal/pkg/search"
)

type ArticleRepo struct{}

func NewArticleRepo() *ArticleRepo {
	return &ArticleRepo{}
}

func (r *ArticleRepo) Create(in domain.CreateArticleInput) (int64, error) {
	tx, err := db.MySQL.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	now := time.Now()
	var publishTime *time.Time
	if in.PublishTime != nil && *in.PublishTime != "" {
		pt, err := time.Parse("2006-01-02 15:04:05", *in.PublishTime)
		if err == nil {
			publishTime = &pt
		}
	}

	res, err := tx.Exec(`
		INSERT INTO article (title, summary, cover_image, content_type, author_id, source_name, source_url, 
			status, publish_time, view_count, like_count, favorite_count, is_top, is_recommend, 
			seo_title, seo_keywords, seo_description, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, 0, 0, 0, ?, ?, ?, ?, ?, ?, ?)
	`, in.Title, in.Summary, in.CoverImage, in.ContentType, in.AuthorID, in.SourceName, in.SourceURL,
		in.Status, publishTime, in.IsTop, in.IsRecommend, in.SeoTitle, in.SeoKeywords, in.SeoDescription, now, now)
	if err != nil {
		return 0, err
	}

	articleID, _ := res.LastInsertId()

	for _, b := range in.Blocks {
		extraJSON, _ := json.Marshal(b.Extra)
		if _, err := tx.Exec(`
			INSERT INTO article_block (article_id, block_type, sort_no, title, content, extra, created_at)
			VALUES (?, ?, ?, ?, ?, ?, ?)
		`, articleID, b.BlockType, b.SortNo, b.Title, b.Content, extraJSON, now); err != nil {
			return 0, err
		}
	}

	for _, tagID := range in.TagIDs {
		_, _ = tx.Exec("INSERT INTO article_tag_rel (article_id, tag_id) VALUES (?, ?)", articleID, tagID)
	}

	for _, diseaseID := range in.DiseaseIDs {
		_, _ = tx.Exec("INSERT INTO disease_article_rel (disease_id, article_id) VALUES (?, ?)", diseaseID, articleID)
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return articleID, nil
}

func (r *ArticleRepo) GetByID(id uint) (*domain.ArticleDetailResponse, error) {
	return r.getByID(id, true)
}

func (r *ArticleRepo) getByID(id uint, publishedOnly bool) (*domain.ArticleDetailResponse, error) {
	var art struct {
		ID             uint
		Title          string
		Summary        sql.NullString
		CoverImage     sql.NullString
		ContentType    sql.NullString
		AuthorID       sql.NullInt64
		SourceName     sql.NullString
		SourceURL      sql.NullString
		Status         int8
		PublishTime    sql.NullTime
		ViewCount      int64
		LikeCount      int64
		FavoriteCount  int64
		IsTop          int8
		IsRecommend    int8
		SeoTitle       sql.NullString
		SeoKeywords    sql.NullString
		SeoDescription sql.NullString
		CreatedAt      time.Time
		UpdatedAt      time.Time
	}

	query := `
		SELECT id, title, summary, cover_image, content_type, author_id, source_name, source_url,
			status, publish_time, view_count, like_count, favorite_count, is_top, is_recommend,
			seo_title, seo_keywords, seo_description, created_at, updated_at
		FROM article WHERE id = ?`
	args := []interface{}{id}
	if publishedOnly {
		query += " AND status = ?"
		args = append(args, domain.ArticleStatusPublished)
	}

	err := db.MySQL.QueryRow(query, args...).Scan(
		&art.ID, &art.Title, &art.Summary, &art.CoverImage, &art.ContentType, &art.AuthorID,
		&art.SourceName, &art.SourceURL, &art.Status, &art.PublishTime, &art.ViewCount,
		&art.LikeCount, &art.FavoriteCount, &art.IsTop, &art.IsRecommend,
		&art.SeoTitle, &art.SeoKeywords, &art.SeoDescription, &art.CreatedAt, &art.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	blocks, err := r.loadArticleBlocks(id)
	if err != nil {
		return nil, err
	}

	tags, err := r.loadArticleTags(id)
	if err != nil {
		return nil, err
	}

	diseases, err := r.loadArticleDiseases(id)
	if err != nil {
		return nil, err
	}

	var authorID *uint
	var author *domain.ArticleAuthorItem
	if art.AuthorID.Valid {
		v := uint(art.AuthorID.Int64)
		authorID = &v
		author, _ = r.loadArticleAuthor(v)
	}

	var pubTimeStr string
	if art.PublishTime.Valid {
		pubTimeStr = art.PublishTime.Time.Format("2006-01-02 15:04:05")
	}

	return &domain.ArticleDetailResponse{
		ID: art.ID, Title: art.Title, Summary: art.Summary.String, CoverImage: art.CoverImage.String,
		ContentType: art.ContentType.String, AuthorID: authorID, SourceName: art.SourceName.String,
		SourceURL: art.SourceURL.String, Status: int(art.Status), PublishTime: pubTimeStr,
		ViewCount: art.ViewCount, LikeCount: art.LikeCount, FavoriteCount: art.FavoriteCount,
		IsTop: int(art.IsTop), IsRecommend: int(art.IsRecommend),
		SeoTitle: art.SeoTitle.String, SeoKeywords: art.SeoKeywords.String, SeoDescription: art.SeoDescription.String,
		Blocks: blocks, Tags: tags, Diseases: diseases, Author: author,
		CreatedAt: art.CreatedAt.Format("2006-01-02 15:04:05"), UpdatedAt: art.UpdatedAt.Format("2006-01-02 15:04:05"),
	}, nil
}

func (r *ArticleRepo) loadArticleBlocks(articleID uint) ([]domain.ArticleBlockItem, error) {
	blockRows, err := db.MySQL.Query(
		"SELECT id, block_type, sort_no, title, content, extra FROM article_block WHERE article_id = ? ORDER BY sort_no ASC",
		articleID,
	)
	if err != nil {
		return nil, err
	}
	defer blockRows.Close()

	blocks := make([]domain.ArticleBlockItem, 0)
	for blockRows.Next() {
		var b domain.ArticleBlockItem
		var extraRaw []byte
		if err := blockRows.Scan(&b.ID, &b.BlockType, &b.SortNo, &b.Title, &b.Content, &extraRaw); err != nil {
			return nil, err
		}
		if len(extraRaw) > 0 {
			_ = json.Unmarshal(extraRaw, &b.Extra)
		}
		blocks = append(blocks, b)
	}
	return blocks, blockRows.Err()
}

func (r *ArticleRepo) loadArticleTags(articleID uint) ([]domain.ArticleTagItem, error) {
	tagRows, err := db.MySQL.Query(`
		SELECT t.id, t.name, t.type
		FROM article_tag_rel atr
		JOIN article_tag t ON atr.tag_id = t.id
		WHERE atr.article_id = ?
	`, articleID)
	if err != nil {
		return nil, err
	}
	defer tagRows.Close()

	tags := make([]domain.ArticleTagItem, 0)
	for tagRows.Next() {
		var t domain.ArticleTagItem
		if err := tagRows.Scan(&t.ID, &t.Name, &t.Type); err != nil {
			return nil, err
		}
		tags = append(tags, t)
	}
	return tags, tagRows.Err()
}

func (r *ArticleRepo) loadArticleDiseases(articleID uint) ([]domain.ArticleDiseaseItem, error) {
	rows, err := db.MySQL.Query(`
		SELECT d.id, d.name
		FROM disease_article_rel dar
		JOIN disease d ON dar.disease_id = d.id
		WHERE dar.article_id = ?
		ORDER BY d.id ASC
	`, articleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	diseases := make([]domain.ArticleDiseaseItem, 0)
	for rows.Next() {
		var item domain.ArticleDiseaseItem
		if err := rows.Scan(&item.ID, &item.Name); err != nil {
			return nil, err
		}
		diseases = append(diseases, item)
	}
	return diseases, rows.Err()
}

func (r *ArticleRepo) loadArticleAuthor(authorID uint) (*domain.ArticleAuthorItem, error) {
	var displayName, avatar sql.NullString
	err := db.MySQL.QueryRow(
		`SELECT display_name, avatar FROM user WHERE id = ?`, authorID,
	).Scan(&displayName, &avatar)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &domain.ArticleAuthorItem{
		ID:          authorID,
		DisplayName: displayName.String,
		Avatar:      avatar.String,
	}, nil
}

func (r *ArticleRepo) Update(id uint, setParts []string, args []interface{}, in domain.UpdateArticleInput) error {
	tx, err := db.MySQL.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if len(setParts) > 0 {
		query := "UPDATE article SET " + strings.Join(setParts, ", ") + " WHERE id=?"
		if _, err := tx.Exec(query, args...); err != nil {
			return err
		}
	}

	if in.Blocks != nil {
		tx.Exec("DELETE FROM article_block WHERE article_id=?", id)
		for _, b := range in.Blocks {
			extraJSON, _ := json.Marshal(b.Extra)
			tx.Exec("INSERT INTO article_block (article_id, block_type, sort_no, title, content, extra, created_at) VALUES (?,?,?,?,?,?,?)",
				id, b.BlockType, b.SortNo, b.Title, b.Content, extraJSON, time.Now())
		}
	}

	if in.TagIDs != nil {
		tx.Exec("DELETE FROM article_tag_rel WHERE article_id=?", id)
		for _, tid := range *in.TagIDs {
			tx.Exec("INSERT INTO article_tag_rel (article_id, tag_id) VALUES (?,?)", id, tid)
		}
	}

	if in.DiseaseIDs != nil {
		tx.Exec("DELETE FROM disease_article_rel WHERE article_id=?", id)
		for _, did := range *in.DiseaseIDs {
			tx.Exec("INSERT INTO disease_article_rel (disease_id, article_id) VALUES (?,?)", did, id)
		}
	}

	return tx.Commit()
}

func (r *ArticleRepo) Delete(id uint) error {
	tx, err := db.MySQL.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	tx.Exec("DELETE FROM article_block WHERE article_id=?", id)
	tx.Exec("DELETE FROM article_tag_rel WHERE article_id=?", id)
	tx.Exec("DELETE FROM disease_article_rel WHERE article_id=?", id)
	if _, err := tx.Exec("DELETE FROM article WHERE id=?", id); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *ArticleRepo) List(filter domain.ArticleListFilter) (*domain.ArticleListResult, error) {
	where := "WHERE 1=1"
	args := []interface{}{}

	if filter.Keyword != "" {
		if clause, arg, ok := search.MatchClause("a.title, a.summary", filter.Keyword); ok {
			where += clause
			args = append(args, arg)
		}
	}
	if filter.Status != "" {
		where += " AND a.status=?"
		args = append(args, filter.Status)
	}
	if filter.TagID != "" {
		where += " AND EXISTS (SELECT 1 FROM article_tag_rel atr WHERE atr.article_id = a.id AND atr.tag_id = ?)"
		args = append(args, filter.TagID)
	}

	joinDisease := ""
	if filter.DiseaseID != "" {
		joinDisease = "INNER JOIN disease_article_rel dar ON a.id = dar.article_id"
		where += " AND dar.disease_id=?"
		args = append(args, filter.DiseaseID)
	}

	fromSQL := "FROM article a " + joinDisease

	var total int64
	countSQL := "SELECT COUNT(DISTINCT a.id) " + fromSQL + " " + where
	if err := db.MySQL.QueryRow(countSQL, args...).Scan(&total); err != nil {
		return nil, err
	}

	offset := (filter.Page - 1) * filter.PageSize
	listSQL := "SELECT a.id, a.title, a.summary, a.cover_image, a.source_name, a.status, a.publish_time, a.view_count, a.is_top, a.is_recommend, a.created_at, a.updated_at " +
		fromSQL + " " + where + " GROUP BY a.id ORDER BY a.is_top DESC, a.publish_time DESC, a.id DESC LIMIT ? OFFSET ?"
	listArgs := append(append([]interface{}{}, args...), filter.PageSize, offset)

	rows, err := db.MySQL.Query(listSQL, listArgs...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := make([]domain.AdminArticleItem, 0)
	articleIDs := make([]uint, 0)
	for rows.Next() {
		var item domain.AdminArticleItem
		var pubTime, createdAt, updatedAt sql.NullTime
		if err := rows.Scan(
			&item.ID, &item.Title, &item.Summary, &item.CoverImage, &item.SourceName,
			&item.Status, &pubTime, &item.ViewCount, &item.IsTop, &item.IsRecommend,
			&createdAt, &updatedAt,
		); err != nil {
			return nil, err
		}
		if pubTime.Valid {
			item.PublishTime = pubTime.Time.Format("2006-01-02 15:04:05")
		}
		if createdAt.Valid {
			item.CreatedAt = createdAt.Time.Format("2006-01-02 15:04:05")
		}
		if updatedAt.Valid {
			item.UpdatedAt = updatedAt.Time.Format("2006-01-02 15:04:05")
		}
		item.Tags = []domain.ArticleTagItem{}
		articleIDs = append(articleIDs, item.ID)
		list = append(list, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	tagMap, err := r.loadTagsByArticleIDs(articleIDs)
	if err != nil {
		return nil, err
	}
	for i := range list {
		if tags, ok := tagMap[list[i].ID]; ok {
			list[i].Tags = tags
		}
	}

	return &domain.ArticleListResult{
		List:     list,
		Total:    total,
		Page:     filter.Page,
		PageSize: filter.PageSize,
	}, nil
}

func (r *ArticleRepo) loadTagsByArticleIDs(articleIDs []uint) (map[uint][]domain.ArticleTagItem, error) {
	result := make(map[uint][]domain.ArticleTagItem)
	if len(articleIDs) == 0 {
		return result, nil
	}

	placeholders := make([]string, len(articleIDs))
	args := make([]interface{}, len(articleIDs))
	for i, id := range articleIDs {
		placeholders[i] = "?"
		args[i] = id
	}

	query := fmt.Sprintf(`
		SELECT atr.article_id, t.id, t.name, t.type
		FROM article_tag_rel atr
		JOIN article_tag t ON atr.tag_id = t.id
		WHERE atr.article_id IN (%s)
		ORDER BY t.id ASC
	`, strings.Join(placeholders, ","))

	rows, err := db.MySQL.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var articleID uint
		var tag domain.ArticleTagItem
		if err := rows.Scan(&articleID, &tag.ID, &tag.Name, &tag.Type); err != nil {
			return nil, err
		}
		result[articleID] = append(result[articleID], tag)
	}
	return result, rows.Err()
}
