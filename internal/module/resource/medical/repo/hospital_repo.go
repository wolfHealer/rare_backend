package repo

import (
	"database/sql"
	"strconv"
	"time"

	"rare_backend/internal/module/resource/medical/domain"
	"rare_backend/internal/pkg/db"
)

type HospitalRepo struct {
	rel *DiseaseRelRepo
}

func NewHospitalRepo(rel *DiseaseRelRepo) *HospitalRepo {
	return &HospitalRepo{rel: rel}
}

type hospitalListRow struct {
	ID            uint64
	Name          string
	ProvinceCode  string
	CityCode      string
	DistrictCode  string
	ProvinceName  string
	CityName      string
	DistrictName  string
	Level         string
	IsRareNetwork int8
	TreatScope    sql.NullString
	Address       string
	Phone         string
	HospitalURL   sql.NullString
	AuditStatus   int8
	CreatedAt     time.Time
}

func (r *HospitalRepo) buildListWhere(filter domain.HospitalListFilter) (string, []interface{}) {
	whereClause := "WHERE 1=1"
	args := []interface{}{}

	if filter.Keyword != "" {
		whereClause += " AND (name LIKE ? OR treat_scope LIKE ?)"
		args = append(args, "%"+filter.Keyword+"%", "%"+filter.Keyword+"%")
	}
	if filter.ProvinceCode != "" {
		whereClause += " AND province_code = ?"
		args = append(args, filter.ProvinceCode)
	}
	if filter.CityCode != "" {
		whereClause += " AND city_code = ?"
		args = append(args, filter.CityCode)
	}
	if filter.DistrictCode != "" {
		whereClause += " AND district_code = ?"
		args = append(args, filter.DistrictCode)
	}
	if filter.Level != "" {
		whereClause += " AND level = ?"
		args = append(args, filter.Level)
	}
	if filter.IsRareNetwork != "" {
		if isRareNetwork, err := strconv.Atoi(filter.IsRareNetwork); err == nil {
			whereClause += " AND is_rare_network = ?"
			args = append(args, isRareNetwork)
		}
	}
	if filter.AuditStatus != "" {
		if auditStatus, err := strconv.Atoi(filter.AuditStatus); err == nil {
			whereClause += " AND audit_status = ?"
			args = append(args, auditStatus)
		}
	}
	return whereClause, args
}

func (r *HospitalRepo) CountList(filter domain.HospitalListFilter) (int64, error) {
	whereClause, args := r.buildListWhere(filter)
	var total int64
	err := db.MySQL.QueryRow("SELECT COUNT(*) FROM hospital "+whereClause, args...).Scan(&total)
	return total, err
}

func (r *HospitalRepo) List(filter domain.HospitalListFilter) ([]hospitalListRow, error) {
	whereClause, args := r.buildListWhere(filter)
	offset := (filter.Page - 1) * filter.PageSize

	listQuery := `
		SELECT id, name, province_code, city_code, district_code, 
               province_name, city_name, district_name,
               level, is_rare_network, treat_scope, address, phone, hospital_url, audit_status, created_at
		FROM hospital
		` + whereClause + `
		ORDER BY is_rare_network DESC, created_at DESC
		LIMIT ? OFFSET ?
	`
	listArgs := append([]interface{}{}, args...)
	listArgs = append(listArgs, filter.PageSize, offset)

	rows, err := db.MySQL.Query(listQuery, listArgs...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []hospitalListRow
	for rows.Next() {
		var h hospitalListRow
		if err := rows.Scan(&h.ID, &h.Name, &h.ProvinceCode, &h.CityCode, &h.DistrictCode, &h.ProvinceName, &h.CityName, &h.DistrictName, &h.Level, &h.IsRareNetwork, &h.TreatScope, &h.Address, &h.Phone, &h.HospitalURL, &h.AuditStatus, &h.CreatedAt); err != nil {
			continue
		}
		list = append(list, h)
	}
	return list, nil
}

func (r *HospitalRepo) GetByID(id uint64) (*domain.HospitalRow, error) {
	query := `
		SELECT id, name, province_code, city_code, district_code, 
               province_name, city_name, district_name,
               level, is_rare_network, treat_scope, address, phone, hospital_url, audit_status, reject_reason, created_at
		FROM hospital
		WHERE id = ?
	`
	var h domain.HospitalRow
	var treatScope, hospitalURL, rejectReason sql.NullString
	err := db.MySQL.QueryRow(query, id).Scan(
		&h.ID, &h.Name, &h.ProvinceCode, &h.CityCode, &h.DistrictCode,
		&h.ProvinceName, &h.CityName, &h.DistrictName,
		&h.Level, &h.IsRareNetwork, &treatScope, &h.Address, &h.Phone, &hospitalURL,
		&h.AuditStatus, &rejectReason, &h.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	h.TreatScope = nullString(treatScope)
	h.HospitalURL = nullString(hospitalURL)
	h.RejectReason = nullString(rejectReason)
	return &h, nil
}

func (r *HospitalRepo) Exists(id uint64) (bool, error) {
	var existID uint64
	err := db.MySQL.QueryRow("SELECT id FROM hospital WHERE id = ?", id).Scan(&existID)
	if err != nil {
		if IsNoRows(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (r *HospitalRepo) Create(in domain.CreateHospitalInput) (int64, error) {
	tx, err := db.MySQL.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	insertQuery := `
		INSERT INTO hospital (name, province_code, city_code, district_code, 
                              province_name, city_name, district_name,
                              level, is_rare_network, treat_scope, address, phone, hospital_url, audit_status, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 0, NOW(), NOW())
	`
	res, err := tx.Exec(insertQuery,
		in.Name, in.ProvinceCode, in.CityCode, in.DistrictCode,
		in.ProvinceName, in.CityName, in.DistrictName,
		in.Level, in.IsRareNetwork, in.TreatScope, in.Address, in.Phone, in.HospitalURL)
	if err != nil {
		return 0, err
	}

	hospitalID, _ := res.LastInsertId()

	if len(in.DiseaseIDs) > 0 {
		if err := r.rel.InsertHospitalDiseaseRel(tx, uint64(hospitalID), in.DiseaseIDs); err != nil {
			return 0, err
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return hospitalID, nil
}

func (r *HospitalRepo) Update(id uint64, in domain.UpdateHospitalInput) error {
	tx, err := db.MySQL.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	fields := []string{}
	values := []interface{}{}

	if in.Name != "" {
		fields = append(fields, "name=?")
		values = append(values, in.Name)
	}
	if in.ProvinceCode != nil {
		fields = append(fields, "province_code=?")
		values = append(values, *in.ProvinceCode)
	}
	if in.CityCode != nil {
		fields = append(fields, "city_code=?")
		values = append(values, *in.CityCode)
	}
	if in.DistrictCode != nil {
		fields = append(fields, "district_code=?")
		values = append(values, *in.DistrictCode)
	}
	if in.ProvinceName != nil {
		fields = append(fields, "province_name=?")
		values = append(values, *in.ProvinceName)
	}
	if in.CityName != nil {
		fields = append(fields, "city_name=?")
		values = append(values, *in.CityName)
	}
	if in.DistrictName != nil {
		fields = append(fields, "district_name=?")
		values = append(values, *in.DistrictName)
	}
	if in.Level != "" {
		fields = append(fields, "level=?")
		values = append(values, in.Level)
	}
	if in.IsRareNetwork != nil {
		fields = append(fields, "is_rare_network=?")
		values = append(values, *in.IsRareNetwork)
	}
	if in.TreatScope != nil {
		fields = append(fields, "treat_scope=?")
		values = append(values, *in.TreatScope)
	}
	if in.Address != "" {
		fields = append(fields, "address=?")
		values = append(values, in.Address)
	}
	if in.Phone != "" {
		fields = append(fields, "phone=?")
		values = append(values, in.Phone)
	}
	if in.HospitalURL != nil {
		fields = append(fields, "hospital_url=?")
		values = append(values, *in.HospitalURL)
	}
	if in.AuditStatus != nil {
		fields = append(fields, "audit_status=?")
		values = append(values, *in.AuditStatus)
	}
	if in.RejectReason != nil {
		fields = append(fields, "reject_reason=?")
		values = append(values, *in.RejectReason)
	}

	fields = append(fields, "updated_at=NOW()")
	values = append(values, id)

	if len(fields) > 0 {
		sqlStr := "UPDATE hospital SET " + joinUpdateFields(fields) + " WHERE id=?"
		if _, err := tx.Exec(sqlStr, values...); err != nil {
			return err
		}
	}

	if in.HasDiseaseIDs {
		if _, err := tx.Exec("DELETE FROM hospital_disease_rel WHERE hospital_id = ?", id); err != nil {
			return err
		}
		if len(in.DiseaseIDs) > 0 {
			if err := r.rel.InsertHospitalDiseaseRel(tx, id, in.DiseaseIDs); err != nil {
				return err
			}
		}
	}

	return tx.Commit()
}

func (r *HospitalRepo) Delete(id uint64) error {
	tx, err := db.MySQL.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var count int64
	if err := tx.QueryRow("SELECT COUNT(*) FROM doctor WHERE hospital_id = ? AND audit_status = 1", id).Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return &domain.HospitalHasDoctorsError{Count: count}
	}

	if _, err := tx.Exec("DELETE FROM hospital_disease_rel WHERE hospital_id = ?", id); err != nil {
		return err
	}

	result, err := tx.Exec("DELETE FROM hospital WHERE id = ?", id)
	if err != nil {
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return tx.Commit()
}

func (r *HospitalRepo) ListOptions(keyword string) ([]domain.HospitalOptionItem, error) {
	whereClause := "WHERE audit_status = 1"
	args := []interface{}{}

	if keyword != "" {
		whereClause += " AND name LIKE ?"
		args = append(args, "%"+keyword+"%")
	}

	query := `
		SELECT id, name 
		FROM hospital 
		` + whereClause + `
		ORDER BY name ASC
	`

	rows, err := db.MySQL.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var options []domain.HospitalOptionItem
	for rows.Next() {
		var item domain.HospitalOptionItem
		if err := rows.Scan(&item.ID, &item.Name); err != nil {
			continue
		}
		options = append(options, item)
	}
	if options == nil {
		options = []domain.HospitalOptionItem{}
	}
	return options, nil
}

func (r *HospitalRepo) IsApproved(id uint64) (bool, error) {
	var hospID uint64
	err := db.MySQL.QueryRow("SELECT id FROM hospital WHERE id = ? AND audit_status = 1", id).Scan(&hospID)
	if err != nil {
		if IsNoRows(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
