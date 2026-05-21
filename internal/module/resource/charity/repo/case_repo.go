package repo

import (
	"database/sql"
	"time"

	"rare_backend/internal/module/resource/charity/domain"
	"rare_backend/internal/pkg/db"
)

type CaseRepo struct{}

func NewCaseRepo() *CaseRepo {
	return &CaseRepo{}
}

type CaseRow struct {
	ID               uint
	DiseaseID        int64
	DiseaseName      string
	ProjectID        uint
	ProjectName      string
	CaseTitle        string
	PatientDesc      string
	ApplyCycle       string
	ActualRelief     string
	Experience       string
	PitfallGuide     string
	CasePdf          sql.NullString
	MaterialTemplate sql.NullString
	AuditStatus      int
	RejectReason     sql.NullString
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type CasePDFRow struct {
	CasePDF   string
	CaseTitle string
}

type DiseaseOptionRow struct {
	DiseaseID int64
	Name      string
}

func (r *CaseRepo) buildListWhere(filter domain.CaseListFilter) (string, []interface{}) {
	whereClause := "WHERE 1=1"
	args := []interface{}{}

	if filter.AuditStatus >= 0 && filter.AuditStatus <= 2 {
		whereClause += " AND rc.audit_status = ?"
		args = append(args, filter.AuditStatus)
	}
	if filter.DiseaseID != "" && filter.DiseaseID != "all" {
		whereClause += " AND rc.disease_id = ?"
		args = append(args, filter.DiseaseID)
	}
	if filter.Keyword != "" {
		whereClause += " AND (rc.case_title LIKE ? OR rc.patient_desc LIKE ?)"
		args = append(args, "%"+filter.Keyword+"%", "%"+filter.Keyword+"%")
	}
	return whereClause, args
}

func (r *CaseRepo) CountList(filter domain.CaseListFilter) (int64, error) {
	whereClause, args := r.buildListWhere(filter)
	var total int64
	err := db.MySQL.QueryRow("SELECT COUNT(*) FROM relief_case rc "+whereClause, args...).Scan(&total)
	return total, err
}

func (r *CaseRepo) List(filter domain.CaseListFilter) ([]CaseRow, error) {
	whereClause, args := r.buildListWhere(filter)
	offset := (filter.Page - 1) * filter.PageSize

	listQuery := `
		SELECT rc.id, rc.disease_id, d.name as disease_name, 
		       rc.project_id, rp.name as project_name,
		       rc.case_title, rc.patient_desc, rc.apply_cycle, 
		       rc.actual_relief, rc.experience, rc.pitfall_guide, 
		       rc.case_pdf, rc.material_template,
		       rc.audit_status, rc.reject_reason,
		       rc.created_at, rc.updated_at
		FROM relief_case rc
		LEFT JOIN relief_project rp ON rc.project_id = rp.id
		LEFT JOIN disease d ON rc.disease_id = d.id
		` + whereClause + `
		ORDER BY rc.created_at DESC
		LIMIT ? OFFSET ?
	`
	args = append(args, filter.PageSize, offset)

	rows, err := db.MySQL.Query(listQuery, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []CaseRow
	for rows.Next() {
		var row CaseRow
		if err := rows.Scan(
			&row.ID, &row.DiseaseID, &row.DiseaseName,
			&row.ProjectID, &row.ProjectName,
			&row.CaseTitle, &row.PatientDesc, &row.ApplyCycle,
			&row.ActualRelief, &row.Experience, &row.PitfallGuide,
			&row.CasePdf, &row.MaterialTemplate,
			&row.AuditStatus, &row.RejectReason,
			&row.CreatedAt, &row.UpdatedAt,
		); err != nil {
			continue
		}
		list = append(list, row)
	}
	return list, nil
}

func (r *CaseRepo) GetByID(id uint) (*CaseRow, error) {
	query := `
		SELECT rc.id, rc.disease_id, d.name as disease_name,
		       rc.project_id, rp.name as project_name, 
			   rc.case_title, rc.patient_desc, rc.apply_cycle, 
		       rc.actual_relief, rc.experience, rc.pitfall_guide, 
		       rc.case_pdf, rc.material_template,
		       rc.audit_status, rc.reject_reason,
		       rc.created_at, rc.updated_at
		FROM relief_case rc
		LEFT JOIN relief_project rp ON rc.project_id = rp.id
		LEFT JOIN disease d ON rc.disease_id = d.id
		WHERE rc.id = ?
	`
	var row CaseRow
	err := db.MySQL.QueryRow(query, id).Scan(
		&row.ID, &row.DiseaseID, &row.DiseaseName,
		&row.ProjectID, &row.ProjectName,
		&row.CaseTitle, &row.PatientDesc, &row.ApplyCycle,
		&row.ActualRelief, &row.Experience, &row.PitfallGuide,
		&row.CasePdf, &row.MaterialTemplate,
		&row.AuditStatus, &row.RejectReason,
		&row.CreatedAt, &row.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *CaseRepo) GetApprovedPDF(id uint) (*CasePDFRow, error) {
	query := `
		SELECT case_pdf, case_title 
		FROM relief_case 
		WHERE id = ? AND audit_status = 1
	`
	var row CasePDFRow
	err := db.MySQL.QueryRow(query, id).Scan(&row.CasePDF, &row.CaseTitle)
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *CaseRepo) ListDiseaseOptions() ([]DiseaseOptionRow, error) {
	query := `
		SELECT DISTINCT rc.disease_id, d.name
		FROM relief_case rc
		INNER JOIN disease d ON rc.disease_id = d.id
		WHERE rc.audit_status = 1 AND d.status = 1
		ORDER BY d.name ASC
	`
	rows, err := db.MySQL.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []DiseaseOptionRow
	for rows.Next() {
		var row DiseaseOptionRow
		if err := rows.Scan(&row.DiseaseID, &row.Name); err != nil {
			continue
		}
		list = append(list, row)
	}
	return list, nil
}

func (r *CaseRepo) Create(in domain.CreateCaseInput, auditStatus int) (int64, error) {
	insertQuery := `
		INSERT INTO relief_case 
		(disease_id, project_id, case_title, patient_desc, apply_cycle, actual_relief, experience,
		 pitfall_guide, case_pdf, material_template, audit_status, reject_reason,
		 created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW())
	`
	result, err := db.MySQL.Exec(insertQuery,
		in.DiseaseID, in.ProjectID, in.CaseTitle, in.PatientDesc, in.ApplyCycle, in.ActualRelief, in.Experience,
		in.PitfallGuide, in.CasePdf, in.MaterialTemplate, auditStatus, in.RejectReason)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (r *CaseRepo) Update(id uint, updateFields []string, args []interface{}) error {
	if len(updateFields) == 0 {
		return nil
	}
	updateFields = append(updateFields, "updated_at = NOW()")
	args = append(args, id)
	updateQuery := `UPDATE relief_case SET ` + joinUpdateFields(updateFields) + ` WHERE id = ?`
	_, err := db.MySQL.Exec(updateQuery, args...)
	return err
}

func (r *CaseRepo) Delete(id uint) (int64, error) {
	result, err := db.MySQL.Exec(`DELETE FROM relief_case WHERE id = ?`, id)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}
