package repo

import (
	"fmt"
	"strings"
	"time"

	"rare_backend/internal/module/resource/charity/domain"
	"rare_backend/internal/pkg/db"
	"rare_backend/internal/pkg/search"
)

type ProjectRepo struct{}

func NewProjectRepo() *ProjectRepo {
	return &ProjectRepo{}
}

type projectListRow struct {
	ID              uint
	Name            string
	Organizer       string
	ApplyCondition  string
	AuditStatus     int
	RejectReason    string
	Sort            int
	ReliefType      string
	ReliefStandard  string
	ApplyDifficulty string
	UpdatedAt       time.Time
}

type projectDetailRow struct {
	ID              uint
	Name            string
	ApplyProcess    string
	ApplyCondition  string
	ApplyDeadline   *time.Time
	ContactPhone    string
	ContactUrl      string
	ApplyForm       string
	ApplyGuide      string
	MaterialList    string
	ReliefType      string
	ReliefStandard  string
	ApplyDifficulty string
	Organizer       string
	CreatedAt       time.Time
	UpdatedAt       time.Time
	Sort            int
	AuditStatus     int
	RejectReason    string
}

func (r *ProjectRepo) buildListWhere(filter domain.ProjectListFilter) (string, []interface{}) {
	whereClause := "WHERE 1=1"
	args := []interface{}{}

	if filter.AuditStatus >= 0 && filter.AuditStatus <= 2 {
		whereClause += " AND p.audit_status = ?"
		args = append(args, filter.AuditStatus)
	} else {
		whereClause += " AND p.audit_status = 1"
	}

	if filter.ReliefType != "" {
		whereClause += " AND p.relief_type = ?"
		args = append(args, filter.ReliefType)
	}
	if filter.DiseaseID != 0 {
		whereClause += " AND EXISTS (SELECT 1 FROM relief_project_disease_rel r WHERE r.project_id = p.id AND r.disease_id = ?)"
		args = append(args, filter.DiseaseID)
	}
	if filter.ApplyDifficulty != "" {
		whereClause += " AND p.apply_difficulty = ?"
		args = append(args, filter.ApplyDifficulty)
	}
	if filter.Keyword != "" {
		if clause, arg, ok := search.MatchClause("p.name, p.organizer", filter.Keyword); ok {
			whereClause += clause
			args = append(args, arg)
		}
	}
	return whereClause, args
}

func (r *ProjectRepo) CountList(filter domain.ProjectListFilter) (int64, error) {
	whereClause, args := r.buildListWhere(filter)
	var total int64
	err := db.MySQL.QueryRow("SELECT COUNT(*) FROM relief_project p "+whereClause, args...).Scan(&total)
	return total, err
}

func (r *ProjectRepo) List(filter domain.ProjectListFilter) ([]projectListRow, error) {
	whereClause, args := r.buildListWhere(filter)
	offset := (filter.Page - 1) * filter.PageSize

	listQuery := `
		SELECT p.id, p.name, p.organizer, p.apply_condition, p.audit_status, p.reject_reason, p.sort,
		       p.relief_type, p.relief_standard, p.apply_difficulty, p.updated_at
		FROM relief_project p
		` + whereClause + `
		ORDER BY p.sort DESC, p.id DESC
		LIMIT ? OFFSET ?
	`
	args = append(args, filter.PageSize, offset)

	rows, err := db.MySQL.Query(listQuery, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []projectListRow
	for rows.Next() {
		var row projectListRow
		if err := rows.Scan(
			&row.ID, &row.Name, &row.Organizer, &row.ApplyCondition,
			&row.AuditStatus, &row.RejectReason, &row.Sort,
			&row.ReliefType, &row.ReliefStandard, &row.ApplyDifficulty,
			&row.UpdatedAt,
		); err != nil {
			continue
		}
		list = append(list, row)
	}
	return list, nil
}

func (r *ProjectRepo) ListDiseaseIDs(projectID uint) ([]int, error) {
	rows, err := db.MySQL.Query("SELECT disease_id FROM relief_project_disease_rel WHERE project_id = ?", projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []int
	for rows.Next() {
		var did int
		if err := rows.Scan(&did); err == nil {
			ids = append(ids, did)
		}
	}
	return ids, nil
}

func (r *ProjectRepo) ListDiseaseIDsByProjectIDs(projectIDs []uint) (map[uint][]int, error) {
	result := make(map[uint][]int)
	if len(projectIDs) == 0 {
		return result, nil
	}

	placeholders := make([]string, len(projectIDs))
	args := make([]interface{}, len(projectIDs))
	for i, id := range projectIDs {
		placeholders[i] = "?"
		args[i] = id
	}

	query := fmt.Sprintf(
		"SELECT project_id, disease_id FROM relief_project_disease_rel WHERE project_id IN (%s)",
		strings.Join(placeholders, ","),
	)
	rows, err := db.MySQL.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var projectID uint
		var diseaseID int
		if err := rows.Scan(&projectID, &diseaseID); err != nil {
			continue
		}
		result[projectID] = append(result[projectID], diseaseID)
	}
	return result, rows.Err()
}

func (r *ProjectRepo) GetByID(id uint) (*projectDetailRow, error) {
	query := `
		SELECT p.id, p.name, p.apply_process, p.apply_condition, p.apply_deadline,
		       p.contact_phone, p.contact_url, p.apply_form, p.apply_guide, p.material_list,
			   p.relief_type, p.relief_standard, p.apply_difficulty, p.organizer,
			   p.created_at, p.updated_at, p.sort, p.audit_status, p.reject_reason
		FROM relief_project p
		WHERE p.id = ?
	`
	var row projectDetailRow
	err := db.MySQL.QueryRow(query, id).Scan(
		&row.ID, &row.Name, &row.ApplyProcess, &row.ApplyCondition,
		&row.ApplyDeadline, &row.ContactPhone, &row.ContactUrl,
		&row.ApplyForm, &row.ApplyGuide, &row.MaterialList,
		&row.ReliefType, &row.ReliefStandard, &row.ApplyDifficulty, &row.Organizer,
		&row.CreatedAt, &row.UpdatedAt, &row.Sort, &row.AuditStatus, &row.RejectReason,
	)
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *ProjectRepo) ListDiseaseDetails(projectID uint) ([]domain.ProjectDiseaseDetail, error) {
	query := `
		SELECT d.id, d.name, d.alias 
		FROM relief_project_disease_rel r
		INNER JOIN disease d ON r.disease_id = d.id
		WHERE r.project_id = ?
	`
	rows, err := db.MySQL.Query(query, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var diseases []domain.ProjectDiseaseDetail
	for rows.Next() {
		var d domain.ProjectDiseaseDetail
		if err := rows.Scan(&d.ID, &d.Name, &d.Alias); err == nil {
			diseases = append(diseases, d)
		}
	}
	return diseases, nil
}

func (r *ProjectRepo) Create(in domain.CreateProjectInput) (int64, error) {
	tx, err := db.MySQL.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	insertQuery := `
		INSERT INTO relief_project 
		(name, apply_process, apply_condition, relief_type,
		 relief_standard, apply_difficulty, apply_deadline, contact_phone, contact_url,
		 apply_form, apply_guide, material_list, organizer, sort,
		 audit_status, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 1, NOW(), NOW())
	`

	var deadline interface{}
	if in.ApplyDeadline != nil && *in.ApplyDeadline != "" {
		deadline = *in.ApplyDeadline
	} else {
		deadline = nil
	}

	result, err := tx.Exec(insertQuery,
		in.Name, in.ApplyProcess, in.ApplyCondition, in.ReliefType,
		in.ReliefStandard, in.ApplyDifficulty, deadline, in.ContactPhone, in.ContactURL,
		in.ApplyForm, in.ApplyGuide, in.MaterialList, in.Organizer, in.Sort)
	if err != nil {
		return 0, err
	}

	projectID, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	if len(in.DiseaseIDs) > 0 {
		relQuery := "INSERT INTO relief_project_disease_rel (project_id, disease_id) VALUES (?, ?)"
		for _, did := range in.DiseaseIDs {
			if did > 0 {
				if _, err := tx.Exec(relQuery, projectID, did); err != nil {
					return 0, err
				}
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return projectID, nil
}

func (r *ProjectRepo) UpdateWithDiseaseRels(id uint, updateFields []string, args []interface{}, diseaseIDs []int, hasDiseaseIDs bool) error {
	tx, err := db.MySQL.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if len(updateFields) > 0 {
		fields := append([]string{}, updateFields...)
		fields = append(fields, "updated_at = NOW()")
		updateArgs := append([]interface{}{}, args...)
		updateArgs = append(updateArgs, id)
		updateQuery := `UPDATE relief_project SET ` + joinUpdateFields(fields) + ` WHERE id = ?`
		if _, err := tx.Exec(updateQuery, updateArgs...); err != nil {
			return err
		}
	}

	if hasDiseaseIDs {
		if _, err := tx.Exec("DELETE FROM relief_project_disease_rel WHERE project_id = ?", id); err != nil {
			return err
		}
		if len(diseaseIDs) > 0 {
			relQuery := "INSERT INTO relief_project_disease_rel (project_id, disease_id) VALUES (?, ?)"
			for _, did := range diseaseIDs {
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

func (r *ProjectRepo) Delete(id uint) (int64, error) {
	tx, err := db.MySQL.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	if _, err := tx.Exec("DELETE FROM relief_project_disease_rel WHERE project_id = ?", id); err != nil {
		return 0, err
	}

	result, err := tx.Exec(`DELETE FROM relief_project WHERE id = ?`, id)
	if err != nil {
		return 0, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return rowsAffected, nil
}

func (r *ProjectRepo) buildOptionsWhere(filter domain.ProjectOptionsFilter) (string, []interface{}) {
	whereClause := "WHERE 1=1"
	args := []interface{}{}

	if filter.AuditStatus >= 0 && filter.AuditStatus <= 2 {
		whereClause += " AND audit_status = ?"
		args = append(args, filter.AuditStatus)
	}
	if filter.Keyword != "" {
		if clause, arg, ok := search.MatchClause("name", filter.Keyword); ok {
			whereClause += clause
			args = append(args, arg)
		}
	}
	return whereClause, args
}

func (r *ProjectRepo) ListOptions(filter domain.ProjectOptionsFilter) ([]domain.ProjectOptionItem, error) {
	whereClause, args := r.buildOptionsWhere(filter)
	offset := (filter.Page - 1) * filter.PageSize

	listQuery := `
        SELECT id, name
        FROM relief_project
        ` + whereClause + `
        ORDER BY sort DESC, id DESC
        LIMIT ? OFFSET ?
    `
	args = append(args, filter.PageSize, offset)

	rows, err := db.MySQL.Query(listQuery, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []domain.ProjectOptionItem
	for rows.Next() {
		var item domain.ProjectOptionItem
		if err := rows.Scan(&item.ID, &item.Name); err != nil {
			continue
		}
		list = append(list, item)
	}
	return list, nil
}
