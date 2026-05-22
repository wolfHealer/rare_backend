package repo

import (
	"database/sql"
	"encoding/json"
	"strings"
	"time"

	"rare_backend/internal/module/knowledge/domain"
	"rare_backend/internal/pkg/db"
	"rare_backend/internal/pkg/search"
)

type DiseaseRepo struct{}

func NewDiseaseRepo() *DiseaseRepo {
	return &DiseaseRepo{}
}

func (r *DiseaseRepo) Exists(id uint) (bool, error) {
	return ExistsByID("disease", id)
}

func (r *DiseaseRepo) Create(in domain.CreateDiseaseInput, imagesJSON string) (int64, error) {
	tx, err := db.MySQL.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	now := time.Now()
	result, err := tx.Exec(`
		INSERT INTO disease 
		(name, alias, introduction, symptoms, images, status, creator_id, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, in.Name, in.Alias, in.Introduction, in.Symptoms, imagesJSON, in.Status, in.CreatorID, now, now)
	if err != nil {
		return 0, err
	}

	diseaseID, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	if len(in.CategoryIDs) > 0 {
		categoryMap := make(map[uint]bool)
		for _, catID := range in.CategoryIDs {
			categoryMap[catID] = true
		}
		for catID := range categoryMap {
			isPrimary := 0
			if catID == in.PrimaryCategoryID {
				isPrimary = 1
			}
			if _, err := tx.Exec(`
				INSERT INTO disease_category_rel (disease_id, category_id, is_primary, created_at)
				VALUES (?, ?, ?, ?)
			`, diseaseID, catID, isPrimary, now); err != nil {
				return 0, err
			}
		}
	}

	if len(in.TagIDs) > 0 {
		tagMap := make(map[uint]bool)
		for _, tagID := range in.TagIDs {
			tagMap[tagID] = true
		}
		for tagID := range tagMap {
			if _, err := tx.Exec(`
				INSERT INTO disease_tag_rel (disease_id, tag_id, created_at)
				VALUES (?, ?, ?)
			`, diseaseID, tagID, now); err != nil {
				return 0, err
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return diseaseID, nil
}

func (r *DiseaseRepo) GetActiveByID(id uint) (*domain.DiseaseItem, error) {
	query := `
		SELECT id, name, alias, introduction, symptoms, images, status, creator_id, created_at, updated_at
		FROM disease
		WHERE id = ? AND status = 1
	`
	var disease struct {
		ID           uint
		Name         string
		Alias        sql.NullString
		Introduction sql.NullString
		Symptoms     sql.NullString
		Images       sql.NullString
		Status       int8
		CreatorID    sql.NullInt64
		CreatedAt    time.Time
		UpdatedAt    time.Time
	}
	err := db.MySQL.QueryRow(query, id).Scan(
		&disease.ID, &disease.Name, &disease.Alias, &disease.Introduction,
		&disease.Symptoms, &disease.Images, &disease.Status, &disease.CreatorID,
		&disease.CreatedAt, &disease.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	var imagesData interface{}
	if disease.Images.Valid && disease.Images.String != "" {
		json.Unmarshal([]byte(disease.Images.String), &imagesData)
	}

	categoryIDs, categories, primaryCategoryID, primaryCategoryName, err := r.loadCategories(id)
	if err != nil {
		return nil, err
	}
	tagIDs, tags, err := r.loadTags(id)
	if err != nil {
		return nil, err
	}
	articles, err := r.loadLinkedArticles(id)
	if err != nil {
		return nil, err
	}

	return &domain.DiseaseItem{
		ID:                  disease.ID,
		Name:                disease.Name,
		Alias:               disease.Alias.String,
		Introduction:        disease.Introduction.String,
		Symptoms:            disease.Symptoms.String,
		Images:              imagesData,
		PrimaryCategoryID:   primaryCategoryID,
		PrimaryCategoryName: primaryCategoryName,
		CategoryIDs:         categoryIDs,
		Categories:          categories,
		TagIDs:              tagIDs,
		Tags:                tags,
		Articles:            articles,
		Status:              int(disease.Status),
		CreatorID:           uint(disease.CreatorID.Int64),
		CreatedAt:           disease.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:           disease.UpdatedAt.Format("2006-01-02 15:04:05"),
	}, nil
}

func (r *DiseaseRepo) loadCategories(diseaseID uint) ([]uint, []domain.SimpleCategoryItem, *uint, *string, error) {
	catQuery := `
		SELECT dcr.category_id, c.name, dcr.is_primary 
		FROM disease_category_rel dcr
		JOIN category c ON dcr.category_id = c.id
		WHERE dcr.disease_id = ?
	`
	var categoryIDs []uint
	var categories []domain.SimpleCategoryItem
	var primaryCategoryID *uint
	var primaryCategoryName *string

	catRows, err := db.MySQL.Query(catQuery, diseaseID)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	defer catRows.Close()
	for catRows.Next() {
		var catID uint
		var catName string
		var isPrimary int8
		if err := catRows.Scan(&catID, &catName, &isPrimary); err == nil {
			categoryIDs = append(categoryIDs, catID)
			categories = append(categories, domain.SimpleCategoryItem{ID: catID, Name: catName})
			if isPrimary == 1 {
				pid := catID
				pName := catName
				primaryCategoryID = &pid
				primaryCategoryName = &pName
			}
		}
	}
	if categoryIDs == nil {
		categoryIDs = []uint{}
	}
	if categories == nil {
		categories = []domain.SimpleCategoryItem{}
	}
	return categoryIDs, categories, primaryCategoryID, primaryCategoryName, nil
}

func (r *DiseaseRepo) loadTags(diseaseID uint) ([]uint, []domain.SimpleTagItem, error) {
	tagQuery := `
		SELECT dtr.tag_id, t.name 
		FROM disease_tag_rel dtr
		JOIN tag t ON dtr.tag_id = t.id
		WHERE dtr.disease_id = ?
	`
	var tagIDs []uint
	var tags []domain.SimpleTagItem
	tagRows, err := db.MySQL.Query(tagQuery, diseaseID)
	if err != nil {
		return nil, nil, err
	}
	defer tagRows.Close()
	for tagRows.Next() {
		var tagID uint
		var tagName string
		if err := tagRows.Scan(&tagID, &tagName); err == nil {
			tagIDs = append(tagIDs, tagID)
			tags = append(tags, domain.SimpleTagItem{ID: tagID, Name: tagName})
		}
	}
	if tagIDs == nil {
		tagIDs = []uint{}
	}
	if tags == nil {
		tags = []domain.SimpleTagItem{}
	}
	return tagIDs, tags, nil
}

func (r *DiseaseRepo) loadLinkedArticles(diseaseID uint) ([]domain.ArticleItem, error) {
	articleQuery := `
		SELECT a.id, a.title, a.summary, a.cover_image, a.source_name, a.publish_time, a.view_count
		FROM disease_article_rel dar
		JOIN article a ON dar.article_id = a.id
		WHERE dar.disease_id = ? AND a.status = 2
		ORDER BY a.publish_time DESC
	`
	var articles []domain.ArticleItem
	artRows, err := db.MySQL.Query(articleQuery, diseaseID)
	if err != nil {
		return nil, err
	}
	defer artRows.Close()
	for artRows.Next() {
		var art domain.ArticleItem
		var publishTime sql.NullTime
		if err := artRows.Scan(
			&art.ID, &art.Title, &art.Summary, &art.CoverImage,
			&art.SourceName, &publishTime, &art.ViewCount,
		); err == nil {
			if publishTime.Valid {
				art.PublishTime = publishTime.Time.Format("2006-01-02 15:04:05")
			}
			articles = append(articles, art)
		}
	}
	if articles == nil {
		articles = []domain.ArticleItem{}
	}
	return articles, nil
}

func (r *DiseaseRepo) Update(id uint, mainSetParts []string, mainArgs []interface{}, in domain.UpdateDiseaseInput) error {
	tx, err := db.MySQL.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	now := time.Now()

	if len(mainSetParts) > 0 {
		query := "UPDATE disease SET " + strings.Join(mainSetParts, ", ") + " WHERE id = ?"
		if _, err := tx.Exec(query, mainArgs...); err != nil {
			return err
		}
	}

	if in.CategoryIDs != nil {
		if _, err := tx.Exec("DELETE FROM disease_category_rel WHERE disease_id = ?", id); err != nil {
			return err
		}
		if len(*in.CategoryIDs) > 0 {
			categoryMap := make(map[uint]bool)
			for _, catID := range *in.CategoryIDs {
				categoryMap[catID] = true
			}
			for catID := range categoryMap {
				isPrimary := 0
				if in.PrimaryCategoryID != nil && *in.PrimaryCategoryID == catID {
					isPrimary = 1
				}
				if _, err := tx.Exec(`
					INSERT INTO disease_category_rel (disease_id, category_id, is_primary, created_at)
					VALUES (?, ?, ?, ?)
				`, id, catID, isPrimary, now); err != nil {
					return err
				}
			}
		}
	} else if in.PrimaryCategoryID != nil {
		if _, err := tx.Exec("UPDATE disease_category_rel SET is_primary = 0 WHERE disease_id = ?", id); err != nil {
			return err
		}
		if *in.PrimaryCategoryID > 0 {
			if _, err := tx.Exec("UPDATE disease_category_rel SET is_primary = 1 WHERE disease_id = ? AND category_id = ?", id, *in.PrimaryCategoryID); err != nil {
				return err
			}
		}
	}

	if in.TagIDs != nil {
		if _, err := tx.Exec("DELETE FROM disease_tag_rel WHERE disease_id = ?", id); err != nil {
			return err
		}
		if len(*in.TagIDs) > 0 {
			tagMap := make(map[uint]bool)
			for _, tagID := range *in.TagIDs {
				tagMap[tagID] = true
			}
			for tagID := range tagMap {
				if _, err := tx.Exec(`
					INSERT INTO disease_tag_rel (disease_id, tag_id, created_at)
					VALUES (?, ?, ?)
				`, id, tagID, now); err != nil {
					return err
				}
			}
		}
	}

	return tx.Commit()
}

func (r *DiseaseRepo) Delete(id uint) error {
	tx, err := db.MySQL.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var exists uint
	if err := tx.QueryRow("SELECT id FROM disease WHERE id = ?", id).Scan(&exists); err != nil {
		return err
	}

	if _, err := tx.Exec("DELETE FROM disease WHERE id = ?", id); err != nil {
		return err
	}
	if _, err := tx.Exec("DELETE FROM disease_category_rel WHERE disease_id = ?", id); err != nil {
		return err
	}
	if _, err := tx.Exec("DELETE FROM disease_tag_rel WHERE disease_id = ?", id); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *DiseaseRepo) List(filter domain.DiseaseListFilter) (*domain.DiseaseListResult, error) {
	whereConditions := []string{"1=1"}
	args := []interface{}{}

	if filter.Status != nil {
		whereConditions = append(whereConditions, "d.status = ?")
		args = append(args, *filter.Status)
	} else {
		whereConditions = append(whereConditions, "d.status = 1")
	}

	if filter.Keyword != "" {
		if cond, arg, ok := search.MatchCondition("d.name, d.alias", filter.Keyword); ok {
			whereConditions = append(whereConditions, cond)
			args = append(args, arg)
		}
	}

	hasJoinCategory := false
	if filter.CategoryID != nil {
		hasJoinCategory = true
		whereConditions = append(whereConditions, "dcr.category_id = ?")
		args = append(args, *filter.CategoryID)
	}

	hasJoinTag := false
	if filter.TagID != nil {
		hasJoinTag = true
		whereConditions = append(whereConditions, "dtr.tag_id = ?")
		args = append(args, *filter.TagID)
	}

	whereClause := strings.Join(whereConditions, " AND ")

	countQuery := "SELECT COUNT(*) FROM disease d"
	if hasJoinCategory {
		countQuery += " INNER JOIN disease_category_rel dcr ON d.id = dcr.disease_id"
	}
	if hasJoinTag {
		countQuery += " INNER JOIN disease_tag_rel dtr ON d.id = dtr.disease_id"
	}
	countQuery += " WHERE " + whereClause

	var total int64
	if err := db.MySQL.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return nil, err
	}

	listQuery := "SELECT d.id, d.name, d.alias, d.introduction, d.symptoms, d.images, d.status, d.creator_id, d.created_at, d.updated_at FROM disease d"
	if hasJoinCategory {
		listQuery += " INNER JOIN disease_category_rel dcr ON d.id = dcr.disease_id"
	}
	if hasJoinTag {
		listQuery += " INNER JOIN disease_tag_rel dtr ON d.id = dtr.disease_id"
	}
	offset := (filter.Page - 1) * filter.PageSize
	listQuery += " WHERE " + whereClause + " ORDER BY d.created_at DESC LIMIT ? OFFSET ?"
	listArgs := append(args, filter.PageSize, offset)

	rows, err := db.MySQL.Query(listQuery, listArgs...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []domain.DiseaseItem
	for rows.Next() {
		var d struct {
			ID           uint
			Name         string
			Alias        sql.NullString
			Introduction sql.NullString
			Symptoms     sql.NullString
			Images       sql.NullString
			Status       int8
			CreatorID    sql.NullInt64
			CreatedAt    time.Time
			UpdatedAt    time.Time
		}
		if err := rows.Scan(
			&d.ID, &d.Name, &d.Alias, &d.Introduction,
			&d.Symptoms, &d.Images, &d.Status, &d.CreatorID,
			&d.CreatedAt, &d.UpdatedAt,
		); err != nil {
			continue
		}
		list = append(list, domain.DiseaseItem{
			ID:           d.ID,
			Name:         d.Name,
			Alias:        d.Alias.String,
			Introduction: d.Introduction.String,
			Symptoms:     d.Symptoms.String,
			Images:       d.Images.String,
			Status:       int(d.Status),
			CreatorID:    uint(d.CreatorID.Int64),
			CreatedAt:    d.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:    d.UpdatedAt.Format("2006-01-02 15:04:05"),
		})
	}
	if list == nil {
		list = []domain.DiseaseItem{}
	}

	return &domain.DiseaseListResult{
		List:     list,
		Total:    total,
		Page:     filter.Page,
		PageSize: filter.PageSize,
	}, nil
}

func (r *DiseaseRepo) ListByCategory(filter domain.DiseasesByCategoryFilter) (*domain.DiseaseListResult, error) {
	countQuery := `
		SELECT COUNT(*) 
		FROM disease_category_rel dcr
		JOIN disease d ON dcr.disease_id = d.id
		WHERE dcr.category_id = ? AND d.status = 1
	`
	var total int64
	if err := db.MySQL.QueryRow(countQuery, filter.CategoryID).Scan(&total); err != nil {
		return nil, err
	}

	offset := (filter.Page - 1) * filter.PageSize
	listQuery := `
		SELECT d.id, d.name, d.alias, d.introduction, d.symptoms, d.images, d.status, d.creator_id, d.created_at, d.updated_at
		FROM disease_category_rel dcr
		JOIN disease d ON dcr.disease_id = d.id
		WHERE dcr.category_id = ? AND d.status = 1
		ORDER BY d.created_at DESC
		LIMIT ? OFFSET ?
	`
	rows, err := db.MySQL.Query(listQuery, filter.CategoryID, filter.PageSize, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []domain.DiseaseItem
	for rows.Next() {
		var d struct {
			ID           uint
			Name         string
			Alias        sql.NullString
			Introduction sql.NullString
			Symptoms     sql.NullString
			Images       sql.NullString
			Status       int8
			CreatorID    sql.NullInt64
			CreatedAt    time.Time
			UpdatedAt    time.Time
		}
		if err := rows.Scan(
			&d.ID, &d.Name, &d.Alias, &d.Introduction,
			&d.Symptoms, &d.Images, &d.Status, &d.CreatorID,
			&d.CreatedAt, &d.UpdatedAt,
		); err != nil {
			continue
		}
		list = append(list, domain.DiseaseItem{
			ID:           d.ID,
			Name:         d.Name,
			Alias:        d.Alias.String,
			Introduction: d.Introduction.String,
			Symptoms:     d.Symptoms.String,
			Images:       d.Images.String,
			Status:       int(d.Status),
			CreatorID:    uint(d.CreatorID.Int64),
			CreatedAt:    d.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:    d.UpdatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	return &domain.DiseaseListResult{
		List:     list,
		Total:    total,
		Page:     filter.Page,
		PageSize: filter.PageSize,
	}, nil
}

func (r *DiseaseRepo) Search(keyword string, page, pageSize int) (*domain.SearchDiseasesResult, error) {
	whereClause := "WHERE status = 1"
	args := []interface{}{}
	if cond, arg, ok := search.MatchCondition("name, alias", keyword); ok {
		whereClause += " AND " + cond
		args = append(args, arg)
	} else {
		return &domain.SearchDiseasesResult{List: []domain.SimpleDiseaseItem{}}, nil
	}

	countQuery := "SELECT COUNT(*) FROM disease " + whereClause
	var total int64
	if err := db.MySQL.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return nil, err
	}
	_ = total

	offset := (page - 1) * pageSize
	listQuery := `
		SELECT id, name, alias
		FROM disease
		` + whereClause + `
		ORDER BY id DESC
		LIMIT ? OFFSET ?
	`
	args = append(args, pageSize, offset)

	rows, err := db.MySQL.Query(listQuery, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []domain.SimpleDiseaseItem
	for rows.Next() {
		var d struct {
			ID    uint
			Name  string
			Alias sql.NullString
		}
		if err := rows.Scan(&d.ID, &d.Name, &d.Alias); err != nil {
			continue
		}
		list = append(list, domain.SimpleDiseaseItem{
			ID:    d.ID,
			Name:  d.Name,
			Alias: d.Alias.String,
		})
	}
	if list == nil {
		list = []domain.SimpleDiseaseItem{}
	}
	return &domain.SearchDiseasesResult{List: list}, nil
}

func (r *DiseaseRepo) Options(keyword string) ([]domain.DiseaseOptionItem, error) {
	whereClause := "WHERE status = 1"
	args := []interface{}{}
	if keyword != "" {
		if cond, arg, ok := search.MatchCondition("name, alias", keyword); ok {
			whereClause += " AND " + cond
			args = append(args, arg)
		}
	}

	query := `
		SELECT id, name, alias 
		FROM disease 
		` + whereClause + `
		ORDER BY name ASC
	`
	rows, err := db.MySQL.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var options []domain.DiseaseOptionItem
	for rows.Next() {
		var item domain.DiseaseOptionItem
		var alias sql.NullString
		if err := rows.Scan(&item.ID, &item.Name, &alias); err != nil {
			continue
		}
		if alias.Valid {
			item.Alias = alias.String
		}
		options = append(options, item)
	}
	if options == nil {
		options = []domain.DiseaseOptionItem{}
	}
	return options, nil
}
