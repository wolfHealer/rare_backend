package repo

import (
	"database/sql"
	"log"
	"strconv"

	"rare_backend/internal/module/resource/medical/domain"
	"rare_backend/internal/pkg/db"
	"rare_backend/internal/pkg/search"
)

type DoctorRepo struct {
	rel        *DiseaseRelRepo
	hospitalRepo *HospitalRepo
}

func NewDoctorRepo(rel *DiseaseRelRepo, hospitalRepo *HospitalRepo) *DoctorRepo {
	return &DoctorRepo{rel: rel, hospitalRepo: hospitalRepo}
}

type doctorListRow struct {
	ID            uint64
	Name          string
	Title         string
	Department    string
	GoodAt        sql.NullString
	ClinicTime    sql.NullString
	Contact       sql.NullString
	Score         float64
	CommentNum    int
	HospitalName  string
	ProvinceCode  string
	CityCode      string
	DistrictCode  string
	ProvinceName  string
	CityName      string
	DistrictName  string
	Level         string
	IsRareNetwork int8
	AuditStatus   int8
}

type doctorDetailRow struct {
	ID            uint64
	Name          string
	Title         string
	Department    string
	GoodAt        sql.NullString
	ClinicTime    sql.NullString
	Contact       sql.NullString
	Score         float64
	CommentNum    int
	AuditStatus   int8
	RejectReason  sql.NullString
	HospitalID    uint64
	HospitalName  string
	ProvinceCode  string
	CityCode      string
	DistrictCode  string
	ProvinceName  string
	CityName      string
	DistrictName  string
	Address       string
	Phone         string
	Level         string
	IsRareNetwork int8
}

func (r *DoctorRepo) buildListWhere(filter domain.DoctorListFilter) (string, []interface{}, string) {
	whereClause := "WHERE 1=1"
	args := []interface{}{}

	if filter.AuditStatus != "" {
		if auditStatus, err := strconv.Atoi(filter.AuditStatus); err == nil {
			whereClause += " AND d.audit_status = ?"
			args = append(args, auditStatus)
		}
	}

	if filter.Keyword != "" {
		if clause, arg, ok := search.MatchClause("d.name, d.good_at", filter.Keyword); ok {
			whereClause += clause
			args = append(args, arg)
		}
	}

	if filter.Title != "" {
		whereClause += " AND d.title = ?"
		args = append(args, filter.Title)
	}

	if filter.HospitalID > 0 {
		whereClause += " AND d.hospital_id = ?"
		args = append(args, filter.HospitalID)
	}

	if filter.Level != "" {
		whereClause += " AND h.level = ?"
		args = append(args, filter.Level)
	}

	joinClause := "JOIN hospital h ON d.hospital_id = h.id"

	if filter.DiseaseID > 0 {
		joinClause += " JOIN doctor_disease_rel ddr ON d.id = ddr.doctor_id"
		whereClause += " AND ddr.disease_id = ?"
		args = append(args, filter.DiseaseID)
	}

	return whereClause, args, joinClause
}

func (r *DoctorRepo) CountList(filter domain.DoctorListFilter) (int64, error) {
	whereClause, args, joinClause := r.buildListWhere(filter)
	var total int64
	countQuery := "SELECT COUNT(*) FROM doctor d " + joinClause + " " + whereClause
	err := db.MySQL.QueryRow(countQuery, args...).Scan(&total)
	return total, err
}

func (r *DoctorRepo) List(filter domain.DoctorListFilter) ([]doctorListRow, error) {
	whereClause, args, joinClause := r.buildListWhere(filter)
	offset := (filter.Page - 1) * filter.PageSize

	listQuery := `
		SELECT d.id, d.name, d.title, d.department, d.good_at, d.clinic_time, d.contact, d.score, d.comment_num, 
		       h.name as hospital_name, h.province_code, h.city_code, h.district_code, 
               h.province_name, h.city_name, h.district_name, h.level, h.is_rare_network, h.audit_status
		FROM doctor d
		` + joinClause + `
		` + whereClause + `
		ORDER BY d.score DESC, d.comment_num DESC
		LIMIT ? OFFSET ?
	`
	listArgs := append([]interface{}{}, args...)
	listArgs = append(listArgs, filter.PageSize, offset)

	rows, err := db.MySQL.Query(listQuery, listArgs...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []doctorListRow
	for rows.Next() {
		var doc doctorListRow
		if err := rows.Scan(
			&doc.ID, &doc.Name, &doc.Title, &doc.Department, &doc.GoodAt, &doc.ClinicTime, &doc.Contact, &doc.Score, &doc.CommentNum,
			&doc.HospitalName, &doc.ProvinceCode, &doc.CityCode, &doc.DistrictCode, &doc.ProvinceName, &doc.CityName, &doc.DistrictName,
			&doc.Level, &doc.IsRareNetwork, &doc.AuditStatus); err != nil {
			log.Printf("[medical] scan doctor row error: %v", err)
			continue
		}
		list = append(list, doc)
	}
	return list, nil
}

func (r *DoctorRepo) GetByID(id uint64) (*doctorDetailRow, error) {
	query := `
		SELECT d.id, d.name, d.title, d.department, d.good_at, d.clinic_time, d.contact, d.score, d.comment_num, d.audit_status, d.reject_reason, d.hospital_id,
		       h.name as hospital_name, h.province_code, h.city_code, h.district_code, 
               h.province_name, h.city_name, h.district_name, h.address, h.phone, h.level, h.is_rare_network
		FROM doctor d
		JOIN hospital h ON d.hospital_id = h.id
		WHERE d.id = ?
	`
	var doc doctorDetailRow
	if err := db.MySQL.QueryRow(query, id).Scan(
		&doc.ID, &doc.Name, &doc.Title, &doc.Department, &doc.GoodAt, &doc.ClinicTime, &doc.Contact, &doc.Score, &doc.CommentNum,
		&doc.AuditStatus, &doc.RejectReason, &doc.HospitalID,
		&doc.HospitalName, &doc.ProvinceCode, &doc.CityCode, &doc.DistrictCode, &doc.ProvinceName, &doc.CityName, &doc.DistrictName,
		&doc.Address, &doc.Phone, &doc.Level, &doc.IsRareNetwork); err != nil {
		return nil, err
	}
	return &doc, nil
}

func (r *DoctorRepo) Create(in domain.CreateDoctorInput) (int64, error) {
	approved, err := r.hospitalRepo.IsApproved(in.HospitalID)
	if err != nil {
		return 0, err
	}
	if !approved {
		return 0, domain.ErrHospitalNotApproved
	}

	tx, err := db.MySQL.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	initScore := 0.0
	initCommentNum := 0
	initAuditStatus := 0

	if in.Score != nil {
		initScore = *in.Score
	}
	if in.CommentNum != nil {
		initCommentNum = *in.CommentNum
	}
	if in.AuditStatus != nil {
		initAuditStatus = int(*in.AuditStatus)
	}

	insertQuery := `
		INSERT INTO doctor (hospital_id, name, title, department, good_at, clinic_time, contact, score, comment_num, audit_status, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW())
	`
	res, err := tx.Exec(insertQuery,
		in.HospitalID,
		in.Name,
		in.Title,
		in.Department,
		in.GoodAt,
		in.ClinicTime,
		in.Contact,
		initScore,
		initCommentNum,
		initAuditStatus,
	)
	if err != nil {
		return 0, err
	}

	doctorID, _ := res.LastInsertId()

	if len(in.DiseaseIDs) > 0 {
		if err := r.rel.InsertDoctorDiseaseRel(tx, uint64(doctorID), in.DiseaseIDs); err != nil {
			return 0, err
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return doctorID, nil
}

func (r *DoctorRepo) Update(id uint64, in domain.UpdateDoctorInput) error {
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
	if in.Title != "" {
		fields = append(fields, "title=?")
		values = append(values, in.Title)
	}
	if in.Department != "" {
		fields = append(fields, "department=?")
		values = append(values, in.Department)
	}
	if in.GoodAt != nil {
		fields = append(fields, "good_at=?")
		values = append(values, *in.GoodAt)
	}
	if in.ClinicTime != nil {
		fields = append(fields, "clinic_time=?")
		values = append(values, *in.ClinicTime)
	}
	if in.Contact != nil {
		fields = append(fields, "contact=?")
		values = append(values, *in.Contact)
	}

	if in.HospitalID != nil {
		approved, err := r.hospitalRepo.IsApproved(*in.HospitalID)
		if err != nil {
			return err
		}
		if !approved {
			return domain.ErrHospitalNotApproved
		}
		fields = append(fields, "hospital_id=?")
		values = append(values, *in.HospitalID)
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
		sqlStr := "UPDATE doctor SET " + joinUpdateFields(fields) + " WHERE id=?"
		if _, err := tx.Exec(sqlStr, values...); err != nil {
			return err
		}
	}

	if in.HasDiseaseIDs {
		if _, err := tx.Exec("DELETE FROM doctor_disease_rel WHERE doctor_id = ?", id); err != nil {
			return err
		}
		if len(in.DiseaseIDs) > 0 {
			if err := r.rel.InsertDoctorDiseaseRel(tx, id, in.DiseaseIDs); err != nil {
				return err
			}
		}
	}

	return tx.Commit()
}

func (r *DoctorRepo) Delete(id uint64) error {
	tx, err := db.MySQL.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec("DELETE FROM doctor_disease_rel WHERE doctor_id = ?", id); err != nil {
		return err
	}

	result, err := tx.Exec("DELETE FROM doctor WHERE id = ?", id)
	if err != nil {
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return tx.Commit()
}
