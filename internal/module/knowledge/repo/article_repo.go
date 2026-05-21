package repo

import (
	"database/sql"
	"encoding/json"
	"strings"
	"time"

	"rare_backend/internal/module/knowledge/domain"
	"rare_backend/internal/pkg/db"
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

	err := db.MySQL.QueryRow("SELECT * FROM article WHERE id = ?", id).Scan(
		&art.ID, &art.Title, &art.Summary, &art.CoverImage, &art.ContentType, &art.AuthorID,
		&art.SourceName, &art.SourceURL, &art.Status, &art.PublishTime, &art.ViewCount,
		&art.LikeCount, &art.FavoriteCount, &art.IsTop, &art.IsRecommend,
		&art.SeoTitle, &art.SeoKeywords, &art.SeoDescription, &art.CreatedAt, &art.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	blocks := []domain.ArticleBlockItem{}
	blockRows, err := db.MySQL.Query("SELECT id, block_type, sort_no, title, content, extra FROM article_block WHERE article_id = ? ORDER BY sort_no ASC", id)
	if err == nil {
		defer blockRows.Close()
		for blockRows.Next() {
			var b domain.ArticleBlockItem
			var extraRaw []byte
			blockRows.Scan(&b.ID, &b.BlockType, &b.SortNo, &b.Title, &b.Content, &extraRaw)
			json.Unmarshal(extraRaw, &b.Extra)
			blocks = append(blocks, b)
		}
	}

	tags := []domain.ArticleTagItem{}
	tagRows, err := db.MySQL.Query(`
		SELECT t.id, t.name, t.type 
		FROM article_tag_rel atr 
		JOIN article_tag t ON atr.tag_id = t.id 
		WHERE atr.article_id = ?
	`, id)
	if err == nil {
		defer tagRows.Close()
		for tagRows.Next() {
			var t domain.ArticleTagItem
			tagRows.Scan(&t.ID, &t.Name, &t.Type)
			tags = append(tags, t)
		}
	}

	var authorID *uint
	if art.AuthorID.Valid {
		v := uint(art.AuthorID.Int64)
		authorID = &v
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
		Blocks: blocks, Tags: tags,
		CreatedAt: art.CreatedAt.Format("2006-01-02 15:04:05"), UpdatedAt: art.UpdatedAt.Format("2006-01-02 15:04:05"),
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
		where += " AND (title LIKE ? OR summary LIKE ?)"
		args = append(args, "%"+filter.Keyword+"%", "%"+filter.Keyword+"%")
	}
	if filter.Status != "" {
		where += " AND status=?"
		args = append(args, filter.Status)
	}

	joinDisease := ""
	if filter.DiseaseID != "" {
		joinDisease = "INNER JOIN disease_article_rel dar ON a.id = dar.article_id"
		where += " AND dar.disease_id=?"
		args = append(args, filter.DiseaseID)
	}

	var total int64
	countSQL := "SELECT COUNT(*) FROM article a " + joinDisease + " " + where
	db.MySQL.QueryRow(countSQL, args...).Scan(&total)

	offset := (filter.Page - 1) * filter.PageSize
	listSQL := "SELECT a.id, a.title, a.summary, a.cover_image, a.source_name, a.status, a.publish_time, a.view_count, a.is_top, a.is_recommend, a.created_at, a.updated_at FROM article a " + joinDisease + " " + where + " ORDER BY a.is_top DESC, a.publish_time DESC, a.id DESC LIMIT ? OFFSET ?"
	listArgs := append(args, filter.PageSize, offset)

	rows, err := db.MySQL.Query(listSQL, listArgs...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []domain.AdminArticleItem
	for rows.Next() {
		var item domain.AdminArticleItem
		var pubTime sql.NullTime
		if err := rows.Scan(&item.ID, &item.Title, &item.Summary, &item.CoverImage, &item.SourceName, &item.Status, &pubTime, &item.ViewCount, &item.IsTop, &item.IsRecommend, &item.CreatedAt, &item.UpdatedAt); err != nil {
			continue
		}
		if pubTime.Valid {
			item.PublishTime = pubTime.Time.Format("2006-01-02 15:04:05")
		}
		list = append(list, item)
	}

	return &domain.ArticleListResult{
		List:     list,
		Total:    total,
		Page:     filter.Page,
		PageSize: filter.PageSize,
	}, nil
}
