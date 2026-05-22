package repo

import (
	"encoding/json"
	"strings"

	"rare_backend/internal/module/resource/rehab/domain"
	"rare_backend/internal/pkg/db"
	"rare_backend/internal/pkg/search"
)

type TrainingRepo struct{}

func NewTrainingRepo() *TrainingRepo {
	return &TrainingRepo{}
}

func (r *TrainingRepo) buildListWhere(filter domain.TrainingListFilter) (string, []interface{}) {
	whereConditions := []string{}
	args := []interface{}{}

	if filter.AuditStatus != "" {
		whereConditions = append(whereConditions, "g.audit_status = ?")
		args = append(args, filter.AuditStatus)
	} else {
		whereConditions = append(whereConditions, "g.audit_status = 1")
	}

	if filter.DiseaseID != "" && filter.DiseaseID != "all" {
		whereConditions = append(whereConditions, "EXISTS (SELECT 1 FROM rehab_train_guide_disease_rel r WHERE r.guide_id = g.id AND r.disease_id = ?)")
		args = append(args, filter.DiseaseID)
	}
	if filter.RehabStage != "" {
		whereConditions = append(whereConditions, "g.rehab_stage = ?")
		args = append(args, filter.RehabStage)
	}
	if filter.Keyword != "" {
		if cond, arg, ok := search.MatchCondition("g.title", filter.Keyword); ok {
			whereConditions = append(whereConditions, cond)
			args = append(args, arg)
		}
	}

	whereClause := ""
	if len(whereConditions) > 0 {
		whereClause = "WHERE " + strings.Join(whereConditions, " AND ")
	}
	return whereClause, args
}

func (r *TrainingRepo) CountList(filter domain.TrainingListFilter) (int64, error) {
	whereClause, args := r.buildListWhere(filter)
	var total int64
	err := db.MySQL.QueryRow("SELECT COUNT(*) FROM rehab_train_guide g "+whereClause, args...).Scan(&total)
	return total, err
}

func (r *TrainingRepo) List(filter domain.TrainingListFilter) ([]domain.TrainingRow, error) {
	whereClause, args := r.buildListWhere(filter)
	offset := (filter.Page - 1) * filter.PageSize

	listQuery := `
		SELECT g.id, g.rehab_stage, g.title, g.train_purpose, g.train_content, 
		       g.forbidden_action, g.pic_urls, g.guide_pdf, g.guide_word, 
		       g.audit_status, g.reject_reason, g.sort, g.created_at, g.updated_at
		FROM rehab_train_guide g
		` + whereClause + `
		ORDER BY g.sort DESC, g.id DESC
		LIMIT ? OFFSET ?
	`
	listArgs := append(append([]interface{}{}, args...), filter.PageSize, offset)

	rows, err := db.MySQL.Query(listQuery, listArgs...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []domain.TrainingRow
	for rows.Next() {
		var t domain.TrainingRow
		if err := rows.Scan(
			&t.ID, &t.RehabStage, &t.Title, &t.TrainPurpose, &t.TrainContent,
			&t.ForbiddenAction, &t.PicUrls, &t.GuidePdf, &t.GuideWord,
			&t.AuditStatus, &t.RejectReason, &t.Sort, &t.CreatedAt, &t.UpdatedAt,
		); err != nil {
			continue
		}
		list = append(list, t)
	}
	return list, nil
}

func (r *TrainingRepo) GetByID(id uint64) (*domain.TrainingRow, error) {
	query := `
		SELECT id, rehab_stage, title, train_purpose, train_content, 
		       forbidden_action, pic_urls, guide_pdf, guide_word, 
		       audit_status, reject_reason, sort, created_at, updated_at
		FROM rehab_train_guide
		WHERE id = ?
	`
	var t domain.TrainingRow
	err := db.MySQL.QueryRow(query, id).Scan(
		&t.ID, &t.RehabStage, &t.Title, &t.TrainPurpose, &t.TrainContent,
		&t.ForbiddenAction, &t.PicUrls, &t.GuidePdf, &t.GuideWord,
		&t.AuditStatus, &t.RejectReason, &t.Sort, &t.CreatedAt, &t.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *TrainingRepo) BatchGetDiseaseRels(guideIDs []uint64) (map[uint64][]uint64, map[uint64][]domain.DiseaseItem, error) {
	diseaseIdsMap := make(map[uint64][]uint64)
	diseaseDetailsMap := make(map[uint64][]domain.DiseaseItem)
	if len(guideIDs) == 0 {
		return diseaseIdsMap, diseaseDetailsMap, nil
	}

	query, args := buildInQuery(`
		SELECT r.guide_id, d.id, d.name, d.alias 
		FROM rehab_train_guide_disease_rel r
		INNER JOIN disease d ON r.disease_id = d.id
		WHERE r.guide_id IN (%s)
		ORDER BY r.guide_id, d.id ASC
	`, guideIDs)

	rows, err := db.MySQL.Query(query, args...)
	if err != nil {
		return diseaseIdsMap, diseaseDetailsMap, err
	}
	defer rows.Close()

	for rows.Next() {
		var gID uint64
		var dItem domain.DiseaseItem
		if err := rows.Scan(&gID, &dItem.ID, &dItem.Name, &dItem.Alias); err == nil {
			diseaseIdsMap[gID] = append(diseaseIdsMap[gID], uint64(dItem.ID))
			diseaseDetailsMap[gID] = append(diseaseDetailsMap[gID], dItem)
		}
	}
	return diseaseIdsMap, diseaseDetailsMap, nil
}

func (r *TrainingRepo) GetDiseasesByGuideID(id uint64) ([]uint64, []domain.DiseaseItem, error) {
	diseaseQuery := `
		SELECT d.id, d.name, d.alias 
		FROM rehab_train_guide_disease_rel r
		INNER JOIN disease d ON r.disease_id = d.id
		WHERE r.guide_id = ?
		ORDER BY d.id ASC
	`
	rows, err := db.MySQL.Query(diseaseQuery, id)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	var diseaseIds []uint64
	var diseases []domain.DiseaseItem
	for rows.Next() {
		var d domain.DiseaseItem
		if err := rows.Scan(&d.ID, &d.Name, &d.Alias); err == nil {
			diseaseIds = append(diseaseIds, uint64(d.ID))
			diseases = append(diseases, d)
		}
	}
	return diseaseIds, diseases, nil
}

func (r *TrainingRepo) marshalPicUrls(picUrls []string) (string, error) {
	picUrlsJson := "[]"
	if len(picUrls) > 0 {
		jsonBytes, err := json.Marshal(picUrls)
		if err != nil {
			return "", err
		}
		picUrlsJson = string(jsonBytes)
	}
	return picUrlsJson, nil
}

func (r *TrainingRepo) Create(in domain.CreateTrainingInput) (int64, error) {
	tx, err := db.MySQL.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	picUrlsJson, err := r.marshalPicUrls(in.PicUrls)
	if err != nil {
		return 0, err
	}

	initialAuditStatus := 1
	if in.AuditStatus != nil {
		initialAuditStatus = *in.AuditStatus
	}

	initialRejectReason := ""
	if in.RejectReason != nil {
		initialRejectReason = *in.RejectReason
	}

	insertQuery := `
		INSERT INTO rehab_train_guide 
		(rehab_stage, title, train_purpose, train_content, forbidden_action,
		 pic_urls, guide_pdf, guide_word, sort, audit_status, reject_reason, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW())
	`
	result, err := tx.Exec(insertQuery,
		in.RehabStage, in.Title, in.TrainPurpose, in.TrainContent, in.ForbiddenAction,
		picUrlsJson, in.GuidePDF, in.GuideWord, in.Sort, initialAuditStatus, initialRejectReason)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	if len(in.DiseaseIDs) > 0 {
		relQuery := "INSERT INTO rehab_train_guide_disease_rel (guide_id, disease_id) VALUES (?, ?)"
		for _, did := range in.DiseaseIDs {
			if did > 0 {
				if _, err := tx.Exec(relQuery, id, did); err != nil {
					return 0, err
				}
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return id, nil
}

func (r *TrainingRepo) Update(id uint64, in domain.UpdateTrainingInput) error {
	tx, err := db.MySQL.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	updateFields := []string{}
	args := []interface{}{}

	if in.Title != "" {
		updateFields = append(updateFields, "title = ?")
		args = append(args, in.Title)
	}
	if in.TrainContent != "" {
		updateFields = append(updateFields, "train_content = ?")
		args = append(args, in.TrainContent)
	}
	if in.RehabStage != "" {
		updateFields = append(updateFields, "rehab_stage = ?")
		args = append(args, in.RehabStage)
	}
	if in.TrainPurpose != "" {
		updateFields = append(updateFields, "train_purpose = ?")
		args = append(args, in.TrainPurpose)
	}
	if in.ForbiddenAction != "" {
		updateFields = append(updateFields, "forbidden_action = ?")
		args = append(args, in.ForbiddenAction)
	}
	if in.HasPicUrls {
		picUrlsJson, err := json.Marshal(in.PicUrls)
		if err != nil {
			return err
		}
		updateFields = append(updateFields, "pic_urls = ?")
		args = append(args, string(picUrlsJson))
	}
	if in.GuidePDF != "" {
		updateFields = append(updateFields, "guide_pdf = ?")
		args = append(args, in.GuidePDF)
	}
	if in.GuideWord != "" {
		updateFields = append(updateFields, "guide_word = ?")
		args = append(args, in.GuideWord)
	}
	if in.Sort != 0 {
		updateFields = append(updateFields, "sort = ?")
		args = append(args, in.Sort)
	}
	if in.AuditStatus != nil {
		updateFields = append(updateFields, "audit_status = ?")
		args = append(args, *in.AuditStatus)
		if *in.AuditStatus == 1 {
			updateFields = append(updateFields, "reject_reason = NULL")
		}
	}
	if in.RejectReason != nil {
		updateFields = append(updateFields, "reject_reason = ?")
		args = append(args, *in.RejectReason)
	}

	if len(updateFields) > 0 {
		updateFields = append(updateFields, "updated_at = NOW()")
		args = append(args, id)
		updateQuery := `UPDATE rehab_train_guide SET ` + joinUpdateFields(updateFields) + ` WHERE id = ?`
		if _, err = tx.Exec(updateQuery, args...); err != nil {
			return err
		}
	}

	if in.HasDiseaseIDs {
		if _, err := tx.Exec("DELETE FROM rehab_train_guide_disease_rel WHERE guide_id = ?", id); err != nil {
			return err
		}
		if len(in.DiseaseIDs) > 0 {
			relQuery := "INSERT INTO rehab_train_guide_disease_rel (guide_id, disease_id) VALUES (?, ?)"
			for _, did := range in.DiseaseIDs {
				if did > 0 {
					if _, err := tx.Exec(relQuery, id, did); err != nil {
						return err
					}
				}
			}
		}
	}

	return tx.Commit()
}

func (r *TrainingRepo) Delete(id uint64) (int64, error) {
	tx, err := db.MySQL.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	if _, err = tx.Exec("DELETE FROM rehab_train_guide_disease_rel WHERE guide_id = ?", id); err != nil {
		return 0, err
	}

	result, err := tx.Exec("DELETE FROM rehab_train_guide WHERE id = ?", id)
	if err != nil {
		return 0, err
	}

	rowsAffected, _ := result.RowsAffected()
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return rowsAffected, nil
}

func (r *TrainingRepo) GetResource(id uint64) (*domain.TrainingResourceRow, error) {
	query := `
		SELECT title, guide_pdf, guide_word
		FROM rehab_train_guide
		WHERE id = ? AND audit_status = 1
	`
	var training domain.TrainingResourceRow
	err := db.MySQL.QueryRow(query, id).Scan(&training.Title, &training.GuidePDF, &training.GuideWord)
	if err != nil {
		return nil, err
	}
	return &training, nil
}
