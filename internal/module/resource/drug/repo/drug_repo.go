package repo

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"rare_backend/internal/module/resource/drug/domain"
	"rare_backend/internal/pkg/db"
	"rare_backend/internal/pkg/search"
)

type DrugRepo struct{}

func NewDrugRepo() *DrugRepo {
	return &DrugRepo{}
}

type drugListRow struct {
	ID               uint
	GenericName      string
	BrandName        sql.NullString
	Indication       string
	DrugType         string
	IsInsurance      int8
	DosageForm       string
	Spec             string
	RefPrice         sql.NullString
	HasRelief        int8
	IsLaunched       int8
	NeedPrescription int8
	ManualOriginal   sql.NullString
	ManualPopular    sql.NullString
	AuditStatus      int8
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type drugDetailRow struct {
	ID               uint
	GenericName      string
	BrandName        sql.NullString
	Indication       string
	DrugType         string
	IsInsurance      int8
	DosageForm       string
	Spec             string
	RefPrice         sql.NullString
	HasRelief        int8
	IsLaunched       int8
	NeedPrescription int8
	ManualOriginal   sql.NullString
	ManualPopular    sql.NullString
	AuditStatus      int8
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

func (r *DrugRepo) buildListWhere(filter domain.DrugListFilter) (string, []interface{}) {
	whereClause := "WHERE 1=1"
	args := []interface{}{}

	if filter.AuditStatusExplicit {
		whereClause += " AND d.audit_status = ?"
		args = append(args, filter.AuditStatus)
	} else {
		whereClause += " AND d.audit_status = 1"
	}

	if filter.Keyword != "" {
		if clause, arg, ok := search.MatchClause("d.generic_name, d.brand_name", filter.Keyword); ok {
			whereClause += clause
			args = append(args, arg)
		}
	}
	if filter.DrugType != "" {
		whereClause += " AND d.drug_type = ?"
		args = append(args, filter.DrugType)
	}
	if filter.IsInsurance != "" {
		isInsurance := 0
		if filter.IsInsurance == "true" || filter.IsInsurance == "1" {
			isInsurance = 1
		}
		whereClause += " AND d.is_insurance = ?"
		args = append(args, isInsurance)
	}
	if filter.HasRelief != "" {
		hasRelief := 0
		if filter.HasRelief == "true" || filter.HasRelief == "1" {
			hasRelief = 1
		}
		whereClause += " AND d.has_relief = ?"
		args = append(args, hasRelief)
	}
	return whereClause, args
}

func (r *DrugRepo) CountList(filter domain.DrugListFilter) (int64, error) {
	whereClause, args := r.buildListWhere(filter)
	var total int64
	err := db.MySQL.QueryRow("SELECT COUNT(*) FROM rare_drug d "+whereClause, args...).Scan(&total)
	return total, err
}

func (r *DrugRepo) List(filter domain.DrugListFilter) ([]drugListRow, error) {
	whereClause, args := r.buildListWhere(filter)
	offset := (filter.Page - 1) * filter.PageSize

	listQuery := `
		SELECT 
			d.id, 
			d.generic_name, 
			d.brand_name, 
			d.indication, 
			d.drug_type, 
			d.is_insurance,
			d.dosage_form, 
			d.spec, 
			d.ref_price, 
			d.has_relief, 
			d.is_launched, 
			d.need_prescription,
			d.manual_original, 
			d.manual_popular, 
			d.audit_status,
			d.created_at, 
			d.updated_at
		FROM rare_drug d
		` + whereClause + `
		ORDER BY d.is_launched DESC, d.created_at DESC
		LIMIT ? OFFSET ?
	`
	listArgs := append([]interface{}{}, args...)
	listArgs = append(listArgs, filter.PageSize, offset)

	rows, err := db.MySQL.Query(listQuery, listArgs...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []drugListRow
	for rows.Next() {
		var row drugListRow
		if err := rows.Scan(
			&row.ID, &row.GenericName, &row.BrandName, &row.Indication,
			&row.DrugType, &row.IsInsurance, &row.DosageForm, &row.Spec,
			&row.RefPrice, &row.HasRelief, &row.IsLaunched, &row.NeedPrescription,
			&row.ManualOriginal, &row.ManualPopular, &row.AuditStatus,
			&row.CreatedAt, &row.UpdatedAt,
		); err != nil {
			continue
		}
		list = append(list, row)
	}
	return list, rows.Err()
}

func (r *DrugRepo) ListDiseaseIDsByDrugIDs(drugIDs []uint64) (map[uint64][]uint64, error) {
	result := make(map[uint64][]uint64)
	if len(drugIDs) == 0 {
		return result, nil
	}

	placeholders := make([]string, len(drugIDs))
	queryArgs := make([]interface{}, len(drugIDs))
	for i, id := range drugIDs {
		placeholders[i] = "?"
		queryArgs[i] = id
	}

	relQuery := fmt.Sprintf("SELECT drug_id, disease_id FROM drug_disease_rel WHERE drug_id IN (%s)", strings.Join(placeholders, ","))
	relRows, err := db.MySQL.Query(relQuery, queryArgs...)
	if err != nil {
		return nil, err
	}
	defer relRows.Close()

	for relRows.Next() {
		var dID, disID uint64
		if err := relRows.Scan(&dID, &disID); err != nil {
			continue
		}
		result[dID] = append(result[dID], disID)
	}
	return result, relRows.Err()
}

func (r *DrugRepo) GetByID(id uint) (*drugDetailRow, error) {
	query := `
		SELECT id, generic_name, brand_name, indication, drug_type, is_insurance,
		       dosage_form, spec, ref_price, has_relief, is_launched, need_prescription,
		       manual_original, manual_popular, audit_status, created_at, updated_at
		FROM rare_drug
		WHERE id = ?
	`
	var row drugDetailRow
	err := db.MySQL.QueryRow(query, id).Scan(
		&row.ID, &row.GenericName, &row.BrandName, &row.Indication,
		&row.DrugType, &row.IsInsurance, &row.DosageForm, &row.Spec,
		&row.RefPrice, &row.HasRelief, &row.IsLaunched, &row.NeedPrescription,
		&row.ManualOriginal, &row.ManualPopular, &row.AuditStatus,
		&row.CreatedAt, &row.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *DrugRepo) ListDiseasesByDrugID(drugID uint) ([]domain.DiseaseSimple, error) {
	diseaseQuery := `
		SELECT d.id, d.name 
		FROM drug_disease_rel rel
		JOIN disease d ON rel.disease_id = d.id
		WHERE rel.drug_id = ?
	`
	rows, err := db.MySQL.Query(diseaseQuery, drugID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var diseases []domain.DiseaseSimple
	for rows.Next() {
		var dis domain.DiseaseSimple
		if err := rows.Scan(&dis.ID, &dis.Name); err != nil {
			continue
		}
		diseases = append(diseases, dis)
	}
	return diseases, rows.Err()
}

func (r *DrugRepo) Exists(id uint) (bool, error) {
	var exists uint
	err := db.MySQL.QueryRow("SELECT id FROM rare_drug WHERE id = ?", id).Scan(&exists)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (r *DrugRepo) Create(input domain.CreateDrugInput) (int64, error) {
	insertQuery := `
		INSERT INTO rare_drug 
		(generic_name, brand_name, indication, drug_type, is_insurance,
		 dosage_form, spec, ref_price, has_relief, is_launched, need_prescription,
		 manual_original, manual_popular, audit_status, reject_reason, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	now := time.Now()
	result, err := db.MySQL.Exec(insertQuery,
		input.GenericName, input.BrandName, input.Indication, input.DrugType,
		boolToInt(input.IsInsurance), input.DosageForm, input.Spec, input.RefPrice,
		boolToInt(input.HasRelief), boolToInt(input.IsLaunched), boolToInt(input.NeedPrescription),
		input.ManualOriginal, input.ManualPopular, input.AuditStatus, input.RejectReason, now, now,
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (r *DrugRepo) InsertDiseaseRels(drugID int64, diseaseIDs []uint64) error {
	if len(diseaseIDs) == 0 {
		return nil
	}
	insertPlaceholders := make([]string, 0, len(diseaseIDs))
	insertArgs := make([]interface{}, 0, len(diseaseIDs)*2)
	for _, diseaseID := range diseaseIDs {
		insertPlaceholders = append(insertPlaceholders, "(?, ?)")
		insertArgs = append(insertArgs, drugID, diseaseID)
	}
	insertRelQuery := fmt.Sprintf("INSERT INTO drug_disease_rel (drug_id, disease_id) VALUES %s", strings.Join(insertPlaceholders, ","))
	_, err := db.MySQL.Exec(insertRelQuery, insertArgs...)
	return err
}

func (r *DrugRepo) Update(id uint, fields []string, args []interface{}) error {
	if len(fields) == 0 {
		return nil
	}
	fields = append(fields, "updated_at = ?")
	args = append(args, time.Now(), id)
	updateQuery := "UPDATE rare_drug SET " + joinUpdateFields(fields) + " WHERE id = ?"
	_, err := db.MySQL.Exec(updateQuery, args...)
	return err
}

func (r *DrugRepo) DeleteDiseaseRels(drugID uint) error {
	_, err := db.MySQL.Exec("DELETE FROM drug_disease_rel WHERE drug_id = ?", drugID)
	return err
}

func (r *DrugRepo) Delete(id uint) error {
	_, err := db.MySQL.Exec("DELETE FROM rare_drug WHERE id = ?", id)
	return err
}

func (r *DrugRepo) GetManual(id uint) (original, popular string, err error) {
	query := `SELECT manual_original, manual_popular FROM rare_drug WHERE id = ? AND audit_status = 1`
	err = db.MySQL.QueryRow(query, id).Scan(&original, &popular)
	return
}

func (r *DrugRepo) buildExportWhere(filter domain.DrugExportFilter) (string, []interface{}) {
	whereClause := "WHERE d.audit_status = 1"
	args := []interface{}{}

	if filter.DiseaseID != 0 {
		whereClause += " AND EXISTS (SELECT 1 FROM drug_disease_rel r WHERE r.drug_id = d.id AND r.disease_id = ?)"
		args = append(args, filter.DiseaseID)
	}
	if filter.Keyword != "" {
		if clause, arg, ok := search.MatchClause("d.generic_name, d.brand_name", filter.Keyword); ok {
			whereClause += clause
			args = append(args, arg)
		}
	}
	if filter.TypeFilter != "" {
		whereClause += " AND d.drug_type = ?"
		args = append(args, filter.TypeFilter)
	}
	if filter.Insurance != "" {
		isInsurance := 0
		if filter.Insurance == "true" || filter.Insurance == "1" {
			isInsurance = 1
		}
		whereClause += " AND d.is_insurance = ?"
		args = append(args, isInsurance)
	}
	return whereClause, args
}

func (r *DrugRepo) ListForExport(filter domain.DrugExportFilter) ([]domain.DrugExportRow, error) {
	whereClause, args := r.buildExportWhere(filter)
	query := `
		SELECT d.generic_name, d.brand_name, d.indication, d.drug_type, d.is_insurance,
		       d.dosage_form, d.spec, d.ref_price, d.has_relief, d.is_launched
		FROM rare_drug d
		` + whereClause + `
		ORDER BY d.is_launched DESC, d.created_at DESC
	`
	rows, err := db.MySQL.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []domain.DrugExportRow
	for rows.Next() {
		var row struct {
			GenericName string
			BrandName   sql.NullString
			Indication  string
			DrugType    string
			IsInsurance int8
			DosageForm  string
			Spec        string
			RefPrice    sql.NullString
			HasRelief   int8
			IsLaunched  int8
		}
		if err := rows.Scan(
			&row.GenericName, &row.BrandName, &row.Indication,
			&row.DrugType, &row.IsInsurance, &row.DosageForm, &row.Spec,
			&row.RefPrice, &row.HasRelief, &row.IsLaunched,
		); err != nil {
			continue
		}
		list = append(list, domain.DrugExportRow{
			GenericName: row.GenericName,
			BrandName:   nullString(row.BrandName),
			Indication:  row.Indication,
			DrugType:    row.DrugType,
			IsInsurance: row.IsInsurance,
			DosageForm:  row.DosageForm,
			Spec:        row.Spec,
			RefPrice:    nullString(row.RefPrice),
			HasRelief:   row.HasRelief,
			IsLaunched:  row.IsLaunched,
		})
	}
	return list, rows.Err()
}

func (r *DrugRepo) ApprovedExists(id uint) (bool, error) {
	var exists uint
	err := db.MySQL.QueryRow("SELECT id FROM rare_drug WHERE id = ? AND audit_status = 1", id).Scan(&exists)
	if err != nil {
		return false, err
	}
	return true, nil
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
