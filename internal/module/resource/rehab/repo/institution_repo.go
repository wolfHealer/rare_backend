package repo

import (
	"strings"

	"rare_backend/internal/module/resource/rehab/domain"
	"rare_backend/internal/pkg/db"
)

type InstitutionRepo struct{}

func NewInstitutionRepo() *InstitutionRepo {
	return &InstitutionRepo{}
}

func (r *InstitutionRepo) buildListWhereFull(filter domain.InstitutionListFilter) (string, []interface{}) {
	baseFrom := "FROM rehab_institution i"
	whereConditions := []string{}
	args := []interface{}{}

	if filter.AuditStatus != "" {
		whereConditions = append(whereConditions, "i.audit_status = ?")
		args = append(args, filter.AuditStatus)
	} else {
		whereConditions = append(whereConditions, "i.audit_status = 1")
	}

	if filter.ProvinceCode != "" && filter.ProvinceCode != "all" {
		whereConditions = append(whereConditions, "i.province_code = ?")
		args = append(args, filter.ProvinceCode)
	}
	if filter.CityCode != "" {
		whereConditions = append(whereConditions, "i.city_code = ?")
		args = append(args, filter.CityCode)
	}
	if filter.DistrictCode != "" {
		whereConditions = append(whereConditions, "i.district_code = ?")
		args = append(args, filter.DistrictCode)
	}
	if filter.DiseaseID != "" && filter.DiseaseID != "all" {
		whereConditions = append(whereConditions, "EXISTS (SELECT 1 FROM rehab_institution_disease_rel r WHERE r.institution_id = i.id AND r.disease_id = ?)")
		args = append(args, filter.DiseaseID)
	}
	if filter.Keyword != "" {
		whereConditions = append(whereConditions, "(i.name LIKE ? OR i.address LIKE ?)")
		args = append(args, "%"+filter.Keyword+"%", "%"+filter.Keyword+"%")
	}

	whereClause := ""
	if len(whereConditions) > 0 {
		whereClause = "WHERE " + strings.Join(whereConditions, " AND ")
	}
	return baseFrom + " " + whereClause, args
}

func (r *InstitutionRepo) CountList(filter domain.InstitutionListFilter) (int64, error) {
	fromWhere, args := r.buildListWhereFull(filter)
	var total int64
	err := db.MySQL.QueryRow("SELECT COUNT(*) "+fromWhere, args...).Scan(&total)
	return total, err
}

func (r *InstitutionRepo) List(filter domain.InstitutionListFilter) ([]domain.InstitutionRow, error) {
	fromWhere, args := r.buildListWhereFull(filter)
	offset := (filter.Page - 1) * filter.PageSize

	listQuery := `
		SELECT i.id, i.name, i.province_code, i.city_code, i.district_code, 
		       i.province_name, i.city_name, i.district_name,
		       i.qualification, i.rehab_projects, i.fee_standard, 
		       i.contact_phone, i.contact_url, i.address, i.audit_status, i.created_at, i.updated_at
		` + fromWhere + `
		ORDER BY i.id DESC
		LIMIT ? OFFSET ?
	`
	listArgs := append(append([]interface{}{}, args...), filter.PageSize, offset)

	rows, err := db.MySQL.Query(listQuery, listArgs...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []domain.InstitutionRow
	for rows.Next() {
		var t domain.InstitutionRow
		if err := rows.Scan(
			&t.ID, &t.Name, &t.ProvinceCode, &t.CityCode, &t.DistrictCode,
			&t.ProvinceName, &t.CityName, &t.DistrictName,
			&t.Qualification, &t.RehabProjects, &t.FeeStandard,
			&t.ContactPhone, &t.ContactUrl, &t.Address, &t.AuditStatus, &t.CreatedAt, &t.UpdatedAt,
		); err != nil {
			continue
		}
		list = append(list, t)
	}
	return list, nil
}

func (r *InstitutionRepo) GetByID(id uint64) (*domain.InstitutionRow, error) {
	query := `
		SELECT id, name, province_code, city_code, district_code, 
		       province_name, city_name, district_name,
		       qualification, rehab_projects, fee_standard, 
		       contact_phone, contact_url, address, 
		       audit_status, reject_reason, created_at, updated_at
		FROM rehab_institution
		WHERE id = ?
	`
	var inst domain.InstitutionRow
	err := db.MySQL.QueryRow(query, id).Scan(
		&inst.ID, &inst.Name, &inst.ProvinceCode, &inst.CityCode, &inst.DistrictCode,
		&inst.ProvinceName, &inst.CityName, &inst.DistrictName,
		&inst.Qualification, &inst.RehabProjects, &inst.FeeStandard,
		&inst.ContactPhone, &inst.ContactUrl, &inst.Address,
		&inst.AuditStatus, &inst.RejectReason, &inst.CreatedAt, &inst.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &inst, nil
}

func (r *InstitutionRepo) BatchGetDiseaseIDs(institutionIDs []uint64) (map[uint64][]uint64, error) {
	result := make(map[uint64][]uint64)
	if len(institutionIDs) == 0 {
		return result, nil
	}

	query, args := buildInQuery(
		"SELECT institution_id, disease_id FROM rehab_institution_disease_rel WHERE institution_id IN (%s)",
		institutionIDs,
	)
	rows, err := db.MySQL.Query(query, args...)
	if err != nil {
		return result, err
	}
	defer rows.Close()

	for rows.Next() {
		var instID, disID uint64
		if err := rows.Scan(&instID, &disID); err == nil {
			result[instID] = append(result[instID], disID)
		}
	}
	return result, nil
}

func (r *InstitutionRepo) GetDiseasesByInstitutionID(id uint64) ([]uint64, []domain.DiseaseItem, error) {
	diseaseQuery := `
		SELECT d.id, d.name, d.alias 
		FROM rehab_institution_disease_rel r
		INNER JOIN disease d ON r.disease_id = d.id
		WHERE r.institution_id = ?
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
			if d.ID > 0 {
				diseaseIds = append(diseaseIds, uint64(d.ID))
			}
			diseases = append(diseases, d)
		}
	}
	return diseaseIds, diseases, nil
}

func (r *InstitutionRepo) Create(in domain.CreateInstitutionInput) (int64, error) {
	tx, err := db.MySQL.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	insertQuery := `
		INSERT INTO rehab_institution 
		(name, province_code, city_code, district_code, 
		 province_name, city_name, district_name,
		 qualification, rehab_projects, fee_standard, 
		 contact_phone, contact_url, address, 
		 audit_status, reject_reason,
		 created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW())
	`

	initialAuditStatus := in.AuditStatus
	if initialAuditStatus == 0 && in.AuditStatus == 0 {
		initialAuditStatus = 0
	}

	result, err := tx.Exec(insertQuery,
		in.Name, in.ProvinceCode, in.CityCode, in.DistrictCode,
		in.ProvinceName, in.CityName, in.DistrictName,
		in.Qualification, in.RehabProjects, in.FeeStandard,
		in.ContactPhone, in.ContactUrl, in.Address,
		initialAuditStatus, in.RejectReason)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	if len(in.DiseaseIDs) > 0 {
		relQuery := "INSERT INTO rehab_institution_disease_rel (institution_id, disease_id) VALUES (?, ?)"
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

func (r *InstitutionRepo) Update(id uint64, in domain.UpdateInstitutionInput) error {
	tx, err := db.MySQL.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	updateFields := []string{}
	args := []interface{}{}

	if in.Name != "" {
		updateFields = append(updateFields, "name = ?")
		args = append(args, in.Name)
	}
	if in.ProvinceCode != nil {
		updateFields = append(updateFields, "province_code = ?")
		args = append(args, *in.ProvinceCode)
	}
	if in.CityCode != nil {
		updateFields = append(updateFields, "city_code = ?")
		args = append(args, *in.CityCode)
	}
	if in.DistrictCode != nil {
		updateFields = append(updateFields, "district_code = ?")
		args = append(args, *in.DistrictCode)
	}
	if in.ProvinceName != nil {
		updateFields = append(updateFields, "province_name = ?")
		args = append(args, *in.ProvinceName)
	}
	if in.CityName != nil {
		updateFields = append(updateFields, "city_name = ?")
		args = append(args, *in.CityName)
	}
	if in.DistrictName != nil {
		updateFields = append(updateFields, "district_name= ?")
		args = append(args, *in.DistrictName)
	}
	if in.Qualification != "" {
		updateFields = append(updateFields, "qualification = ?")
		args = append(args, in.Qualification)
	}
	if in.RehabProjects != "" {
		updateFields = append(updateFields, "rehab_projects = ?")
		args = append(args, in.RehabProjects)
	}
	if in.FeeStandard != "" {
		updateFields = append(updateFields, "fee_standard = ?")
		args = append(args, in.FeeStandard)
	}
	if in.ContactPhone != "" {
		updateFields = append(updateFields, "contact_phone = ?")
		args = append(args, in.ContactPhone)
	}
	if in.ContactUrl != "" {
		updateFields = append(updateFields, "contact_url = ?")
		args = append(args, in.ContactUrl)
	}
	if in.Address != "" {
		updateFields = append(updateFields, "address = ?")
		args = append(args, in.Address)
	}
	if in.Sort != 0 {
		updateFields = append(updateFields, "sort = ?")
		args = append(args, in.Sort)
	}
	if in.AuditStatus != nil {
		updateFields = append(updateFields, "audit_status = ?")
		args = append(args, *in.AuditStatus)
	}
	if in.RejectReason != nil {
		updateFields = append(updateFields, "reject_reason = ?")
		args = append(args, *in.RejectReason)
	}

	if len(updateFields) > 0 {
		updateFields = append(updateFields, "updated_at = NOW()")
		args = append(args, id)
		updateQuery := `UPDATE rehab_institution SET ` + joinUpdateFields(updateFields) + ` WHERE id = ?`
		if _, err = tx.Exec(updateQuery, args...); err != nil {
			return err
		}
	}

	if in.HasDiseaseIDs {
		if _, err := tx.Exec("DELETE FROM rehab_institution_disease_rel WHERE institution_id = ?", id); err != nil {
			return err
		}
		if len(in.DiseaseIDs) > 0 {
			relQuery := "INSERT INTO rehab_institution_disease_rel (institution_id, disease_id) VALUES (?, ?)"
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

func (r *InstitutionRepo) Delete(id uint64) error {
	tx, err := db.MySQL.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err = tx.Exec("DELETE FROM rehab_institution_disease_rel WHERE institution_id = ?", id); err != nil {
		return err
	}
	if _, err = tx.Exec("DELETE FROM rehab_institution WHERE id = ?", id); err != nil {
		return err
	}
	return tx.Commit()
}

