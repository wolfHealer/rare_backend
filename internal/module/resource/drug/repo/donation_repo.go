package repo

import (
	"database/sql"
	"time"

	"rare_backend/internal/module/resource/drug/domain"
	"rare_backend/internal/pkg/db"
	"rare_backend/internal/pkg/search"
)

type DonationRepo struct{}

func NewDonationRepo() *DonationRepo {
	return &DonationRepo{}
}

type donationListRow struct {
	ID               uint
	DrugID           uint
	DiseaseID        int
	Name             string
	Organizer        string
	ApplyCondition   string
	ReliefCycle      string
	ReliefDosageDesc string
	ApplyForm        sql.NullString
	ApplyGuide       sql.NullString
	MaterialList     sql.NullString
	ProgressQuery    sql.NullString
	AuditStatus      int8
	UpdatedAt        time.Time
	GenericName      string
	BrandName        sql.NullString
	DiseaseName      sql.NullString
}

type donationDetailRow struct {
	ID               uint
	DrugID           uint
	DiseaseID        int
	Name             string
	Organizer        string
	ApplyCondition   string
	ReliefCycle      string
	ReliefDosageDesc string
	ApplyForm        sql.NullString
	ApplyGuide       sql.NullString
	MaterialList     sql.NullString
	ProgressQuery    sql.NullString
	AuditStatus      int8
	RejectReason     sql.NullString
	CreatedAt        time.Time
	UpdatedAt        time.Time
	GenericName      string
	BrandName        sql.NullString
	DiseaseName      sql.NullString
}

type DonationProgressRow struct {
	ApplicationID uint64
	ProjectID     uint
	ProjectName   string
	Status        string
	SubmitTime    time.Time
	UpdateTime    time.Time
}

type DonationProgressLogRow struct {
	Status     string
	ActionDesc string
	CreatedAt  time.Time
}

func (r *DonationRepo) buildListWhere(filter domain.DonationListFilter) (string, []interface{}) {
	whereClause := "WHERE 1=1"
	args := []interface{}{}

	if filter.AuditStatusExplicit {
		whereClause += " AND p.audit_status = ?"
		args = append(args, filter.AuditStatus)
	} else {
		whereClause += " AND p.audit_status = 1"
	}
	if filter.Organizer != "" {
		if clause, arg, ok := search.MatchClause("p.organizer", filter.Organizer); ok {
			whereClause += clause
			args = append(args, arg)
		}
	}
	if filter.Keyword != "" {
		kw := search.Trim(filter.Keyword)
		if search.Usable(kw) {
			whereClause += ` AND (
				MATCH(p.name, p.organizer) AGAINST(? IN NATURAL LANGUAGE MODE)
				OR MATCH(d.generic_name, d.brand_name) AGAINST(? IN NATURAL LANGUAGE MODE)
			)`
			args = append(args, kw, kw)
		}
	}
	if filter.DiseaseID != 0 {
		whereClause += " AND p.disease_id = ?"
		args = append(args, filter.DiseaseID)
	}
	if filter.DrugID != 0 {
		whereClause += " AND p.drug_id = ?"
		args = append(args, filter.DrugID)
	}
	return whereClause, args
}

func (r *DonationRepo) CountList(filter domain.DonationListFilter) (int64, error) {
	whereClause, args := r.buildListWhere(filter)
	var total int64
	err := db.MySQL.QueryRow("SELECT COUNT(*) FROM drug_relief_project p LEFT JOIN rare_drug d ON p.drug_id = d.id "+whereClause, args...).Scan(&total)
	return total, err
}

func (r *DonationRepo) List(filter domain.DonationListFilter) ([]donationListRow, error) {
	whereClause, args := r.buildListWhere(filter)
	offset := (filter.Page - 1) * filter.PageSize

	listQuery := `
		SELECT 
			p.id, p.drug_id, p.disease_id, p.name, p.organizer, p.apply_condition,
		       p.relief_cycle, p.relief_dosage_desc, p.apply_form, p.apply_guide, p.material_list, p.progress_query,
			p.audit_status, p.updated_at,
			d.generic_name, d.brand_name,
			dis.name as disease_name
		FROM drug_relief_project p
		LEFT JOIN rare_drug d ON p.drug_id = d.id
		LEFT JOIN disease dis ON p.disease_id = dis.id
		` + whereClause + `
		ORDER BY p.created_at DESC
		LIMIT ? OFFSET ?
	`
	args = append(args, filter.PageSize, offset)

	rows, err := db.MySQL.Query(listQuery, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []donationListRow
	for rows.Next() {
		var row donationListRow
		if err := rows.Scan(
			&row.ID, &row.DrugID, &row.DiseaseID, &row.Name,
			&row.Organizer, &row.ApplyCondition, &row.ReliefCycle,
			&row.ReliefDosageDesc, &row.ApplyForm, &row.ApplyGuide,
			&row.MaterialList, &row.ProgressQuery,
			&row.AuditStatus, &row.UpdatedAt,
			&row.GenericName, &row.BrandName, &row.DiseaseName,
		); err != nil {
			continue
		}
		list = append(list, row)
	}
	return list, rows.Err()
}

func (r *DonationRepo) GetByID(id uint) (*donationDetailRow, error) {
	query := `
		SELECT 
			p.id, p.drug_id, p.disease_id, p.name, p.organizer, p.apply_condition,
		       p.relief_cycle, p.relief_dosage_desc, p.apply_form, p.apply_guide, p.material_list, p.progress_query,
		       p.audit_status, p.reject_reason, p.created_at, p.updated_at,
			d.generic_name, d.brand_name,
			dis.name as disease_name
		FROM drug_relief_project p
		LEFT JOIN rare_drug d ON p.drug_id = d.id
		LEFT JOIN disease dis ON p.disease_id = dis.id
		WHERE p.id = ?
	`
	var row donationDetailRow
	err := db.MySQL.QueryRow(query, id).Scan(
		&row.ID, &row.DrugID, &row.DiseaseID, &row.Name,
		&row.Organizer, &row.ApplyCondition, &row.ReliefCycle,
		&row.ReliefDosageDesc, &row.ApplyForm, &row.ApplyGuide,
		&row.MaterialList, &row.ProgressQuery,
		&row.AuditStatus, &row.RejectReason, &row.CreatedAt, &row.UpdatedAt,
		&row.GenericName, &row.BrandName, &row.DiseaseName,
	)
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *DonationRepo) Exists(id uint) (bool, error) {
	var exists uint
	err := db.MySQL.QueryRow("SELECT id FROM drug_relief_project WHERE id = ?", id).Scan(&exists)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (r *DonationRepo) ApprovedExists(id uint) (bool, error) {
	var exists uint
	err := db.MySQL.QueryRow("SELECT id FROM drug_relief_project WHERE id = ? AND audit_status = 1", id).Scan(&exists)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (r *DonationRepo) GetApprovedProject(id uint) (name string, err error) {
	var projectID uint
	err = db.MySQL.QueryRow("SELECT id, name FROM drug_relief_project WHERE id = ? AND audit_status = 1", id).Scan(&projectID, &name)
	return
}

func (r *DonationRepo) DrugApprovedExists(drugID uint) (bool, error) {
	var exists uint
	err := db.MySQL.QueryRow("SELECT id FROM rare_drug WHERE id = ? AND audit_status = 1", drugID).Scan(&exists)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (r *DonationRepo) Create(input domain.CreateDonationInput) (int64, error) {
	insertQuery := `
		INSERT INTO drug_relief_project 
		(drug_id, disease_id, name, organizer, apply_condition,
		 relief_cycle, relief_dosage_desc, apply_form, apply_guide, material_list, progress_query,
		 audit_status, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 1, ?, ?)
	`
	now := time.Now()
	result, err := db.MySQL.Exec(insertQuery,
		input.DrugID, input.DiseaseValue, input.Name, input.Organizer, input.ApplyCondition,
		input.ReliefCycle, input.DrugDosage, input.ApplyForm, input.ApplyGuide,
		input.MaterialList, input.ProgressQuery, now, now,
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (r *DonationRepo) Update(id uint, fields []string, args []interface{}) error {
	if len(fields) == 0 {
		return nil
	}
	fields = append(fields, "updated_at = ?")
	args = append(args, time.Now(), id)
	updateQuery := "UPDATE drug_relief_project SET " + joinUpdateFields(fields) + " WHERE id = ?"
	_, err := db.MySQL.Exec(updateQuery, args...)
	return err
}

func (r *DonationRepo) SoftDelete(id uint) error {
	_, err := db.MySQL.Exec("UPDATE drug_relief_project SET audit_status = 2, updated_at = ? WHERE id = ?", time.Now(), id)
	return err
}

func (r *DonationRepo) InsertApplication(projectID uint, input domain.ApplyDonationInput, idCardEnc, maskedID string) (int64, error) {
	insertQuery := `
		INSERT INTO drug_relief_application 
		(project_id, user_id, patient_name, patient_id_card_enc, patient_id_card_mask,
		 diagnosis_proof, income_proof, contact_phone, status, submit_time)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, 'pending', ?)
	`
	now := time.Now()
	result, err := db.MySQL.Exec(insertQuery,
		projectID, input.UserID, input.PatientName, idCardEnc, maskedID,
		input.DiagnosisProof, input.IncomeProof, input.ContactPhone, now,
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (r *DonationRepo) InsertApplicationLog(appID int64) error {
	logQuery := `
		INSERT INTO drug_relief_log (application_id, status, action_desc, created_at)
		VALUES (?, 'pending', '申请已提交', ?)
	`
	_, err := db.MySQL.Exec(logQuery, appID, time.Now())
	return err
}

func (r *DonationRepo) GetProgressByNumericID(appID uint64) (*DonationProgressRow, error) {
	query := `
		SELECT a.id, a.project_id, p.name, a.status, a.submit_time, a.updated_at
		FROM drug_relief_application a
		JOIN drug_relief_project p ON a.project_id = p.id
		WHERE a.id = ?
	`
	var row DonationProgressRow
	err := db.MySQL.QueryRow(query, appID).Scan(
		&row.ApplicationID, &row.ProjectID, &row.ProjectName,
		&row.Status, &row.SubmitTime, &row.UpdateTime,
	)
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *DonationRepo) GetProgressByApplicationID(applicationID string) (*DonationProgressRow, error) {
	query := `
		SELECT a.id, a.project_id, p.name, a.status, a.submit_time, a.updated_at
		FROM drug_relief_application a
		JOIN drug_relief_project p ON a.project_id = p.id
		WHERE a.application_id = ?
	`
	var row DonationProgressRow
	err := db.MySQL.QueryRow(query, applicationID).Scan(
		&row.ApplicationID, &row.ProjectID, &row.ProjectName,
		&row.Status, &row.SubmitTime, &row.UpdateTime,
	)
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *DonationRepo) ListProgressLogs(applicationID uint64) ([]DonationProgressLogRow, error) {
	logQuery := `
		SELECT status, action_desc, created_at
		FROM drug_relief_log
		WHERE application_id = ?
		ORDER BY created_at ASC
	`
	rows, err := db.MySQL.Query(logQuery, applicationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []DonationProgressLogRow
	for rows.Next() {
		var log DonationProgressLogRow
		if err := rows.Scan(&log.Status, &log.ActionDesc, &log.CreatedAt); err != nil {
			continue
		}
		logs = append(logs, log)
	}
	return logs, rows.Err()
}

func (r *DonationRepo) GetGuide(projectID uint) (guideURL, name string, err error) {
	query := `SELECT apply_guide, name FROM drug_relief_project WHERE id = ? AND audit_status = 1`
	err = db.MySQL.QueryRow(query, projectID).Scan(&guideURL, &name)
	return
}

func (r *DonationRepo) ListDiseaseOptions() ([]int, error) {
	diseaseQuery := "SELECT DISTINCT disease_id FROM drug_relief_project WHERE audit_status = 1 AND disease_id != 0"
	rows, err := db.MySQL.Query(diseaseQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []int
	for rows.Next() {
		var diseaseID int
		if err := rows.Scan(&diseaseID); err != nil {
			continue
		}
		ids = append(ids, diseaseID)
	}
	return ids, rows.Err()
}

func (r *DonationRepo) ListDrugOptions() ([]struct {
	ID          uint
	GenericName string
	BrandName   sql.NullString
}, error) {
	drugQuery := "SELECT id, generic_name, brand_name FROM rare_drug WHERE audit_status = 1"
	rows, err := db.MySQL.Query(drugQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []struct {
		ID          uint
		GenericName string
		BrandName   sql.NullString
	}
	for rows.Next() {
		var drug struct {
			ID          uint
			GenericName string
			BrandName   sql.NullString
		}
		if err := rows.Scan(&drug.ID, &drug.GenericName, &drug.BrandName); err != nil {
			continue
		}
		list = append(list, drug)
	}
	return list, rows.Err()
}
