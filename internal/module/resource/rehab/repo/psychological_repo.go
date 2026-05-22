package repo

import (
	"strings"

	"rare_backend/internal/module/resource/rehab/domain"
	"rare_backend/internal/pkg/db"
	"rare_backend/internal/pkg/search"
)

type PsychologicalRepo struct{}

func NewPsychologicalRepo() *PsychologicalRepo {
	return &PsychologicalRepo{}
}

func (r *PsychologicalRepo) buildListWhere(filter domain.PsychOrgListFilter) (string, []interface{}) {
	whereConditions := []string{}
	args := []interface{}{}

	if filter.AuditStatus != "" {
		whereConditions = append(whereConditions, "audit_status = ?")
		args = append(args, filter.AuditStatus)
	} else {
		whereConditions = append(whereConditions, "audit_status = 1")
	}

	if filter.ProvinceCode != "" && filter.ProvinceCode != "all" {
		whereConditions = append(whereConditions, "province_code = ?")
		args = append(args, filter.ProvinceCode)
	}
	if filter.CityCode != "" {
		whereConditions = append(whereConditions, "city_code = ?")
		args = append(args, filter.CityCode)
	}
	if filter.DistrictCode != "" {
		whereConditions = append(whereConditions, "district_code = ?")
		args = append(args, filter.DistrictCode)
	}
	if filter.ConsultWay != "" {
		whereConditions = append(whereConditions, "consult_way = ?")
		args = append(args, filter.ConsultWay)
	}
	if filter.DiseaseID != "" && filter.DiseaseID != "all" {
		whereConditions = append(whereConditions, "id IN (SELECT org_id FROM psych_support_org_disease_rel WHERE disease_id = ?)")
		args = append(args, filter.DiseaseID)
	}
	if filter.IsFree != "" {
		isFree := 0
		if filter.IsFree == "true" {
			isFree = 1
		}
		whereConditions = append(whereConditions, "is_free = ?")
		args = append(args, isFree)
	}
	if filter.Keyword != "" {
		if cond, arg, ok := search.MatchCondition("name, content_intro", filter.Keyword); ok {
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

func (r *PsychologicalRepo) CountList(filter domain.PsychOrgListFilter) (int64, error) {
	whereClause, args := r.buildListWhere(filter)
	var total int64
	err := db.MySQL.QueryRow("SELECT COUNT(*) FROM psych_support_org "+whereClause, args...).Scan(&total)
	return total, err
}

func (r *PsychologicalRepo) List(filter domain.PsychOrgListFilter) ([]domain.PsychOrgRow, error) {
	whereClause, args := r.buildListWhere(filter)
	offset := (filter.Page - 1) * filter.PageSize

	listQuery := `
		SELECT id, name, province_code, city_code, district_code, 
		       province_name, city_name, district_name,
		       address, contact_phone, contact_url, is_free, consult_way, content_intro, 
		       audit_status, reject_reason, created_at, updated_at
		FROM psych_support_org
		` + whereClause + `
		ORDER BY id DESC
		LIMIT ? OFFSET ?
	`
	args = append(args, filter.PageSize, offset)

	rows, err := db.MySQL.Query(listQuery, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []domain.PsychOrgRow
	for rows.Next() {
		var t domain.PsychOrgRow
		if err := rows.Scan(
			&t.ID, &t.Name, &t.ProvinceCode, &t.CityCode, &t.DistrictCode,
			&t.ProvinceName, &t.CityName, &t.DistrictName,
			&t.Address, &t.ContactPhone, &t.ContactUrl, &t.IsFree, &t.ConsultWay,
			&t.ContentIntro, &t.AuditStatus, &t.RejectReason, &t.CreatedAt, &t.UpdatedAt,
		); err != nil {
			continue
		}
		list = append(list, t)
	}
	return list, nil
}

func (r *PsychologicalRepo) GetByID(id uint64) (*domain.PsychOrgRow, error) {
	query := `
		SELECT id, name, province_code, city_code, district_code, 
		       province_name, city_name, district_name,
		       address, contact_phone, contact_url, is_free, consult_way, content_intro, 
		       audit_status, reject_reason, created_at, updated_at
		FROM psych_support_org
		WHERE id = ?
	`
	var org domain.PsychOrgRow
	err := db.MySQL.QueryRow(query, id).Scan(
		&org.ID, &org.Name, &org.ProvinceCode, &org.CityCode, &org.DistrictCode,
		&org.ProvinceName, &org.CityName, &org.DistrictName,
		&org.Address, &org.ContactPhone, &org.ContactUrl, &org.IsFree, &org.ConsultWay,
		&org.ContentIntro, &org.AuditStatus, &org.RejectReason, &org.CreatedAt, &org.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &org, nil
}

func (r *PsychologicalRepo) BatchGetDiseaseRels(orgIDs []uint64) (map[uint64][]uint64, map[uint64][]domain.DiseaseItem, error) {
	diseaseIdsMap := make(map[uint64][]uint64)
	diseaseDetailsMap := make(map[uint64][]domain.DiseaseItem)
	if len(orgIDs) == 0 {
		return diseaseIdsMap, diseaseDetailsMap, nil
	}

	query, args := buildInQuery(`
		SELECT r.org_id, d.id, d.name, d.alias 
		FROM psych_support_org_disease_rel r
		INNER JOIN disease d ON r.disease_id = d.id
		WHERE r.org_id IN (%s)
		ORDER BY r.org_id, d.id ASC
	`, orgIDs)

	rows, err := db.MySQL.Query(query, args...)
	if err != nil {
		return diseaseIdsMap, diseaseDetailsMap, err
	}
	defer rows.Close()

	for rows.Next() {
		var oID uint64
		var dItem domain.DiseaseItem
		if err := rows.Scan(&oID, &dItem.ID, &dItem.Name, &dItem.Alias); err == nil {
			diseaseIdsMap[oID] = append(diseaseIdsMap[oID], uint64(dItem.ID))
			diseaseDetailsMap[oID] = append(diseaseDetailsMap[oID], dItem)
		}
	}
	return diseaseIdsMap, diseaseDetailsMap, nil
}

func (r *PsychologicalRepo) GetDiseasesByOrgID(id uint64) ([]uint64, []domain.DiseaseItem, error) {
	diseaseQuery := `
		SELECT d.id, d.name, d.alias 
		FROM psych_support_org_disease_rel r
		INNER JOIN disease d ON r.disease_id = d.id
		WHERE r.org_id = ?
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

func (r *PsychologicalRepo) Create(in domain.CreatePsychOrgInput) (int64, error) {
	tx, err := db.MySQL.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	insertQuery := `
		INSERT INTO psych_support_org 
		(name, province_code, city_code, district_code, address, contact_phone, contact_url,
		 is_free, consult_way, content_intro, audit_status,
		 created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 1, NOW(), NOW())
	`
	isFreeInt := 0
	if in.IsFree {
		isFreeInt = 1
	}

	result, err := tx.Exec(insertQuery,
		in.Name, in.ProvinceCode, in.CityCode, in.DistrictCode, in.Address,
		in.ContactPhone, in.ContactUrl, isFreeInt, in.ConsultWay, in.ContentIntro)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	if len(in.DiseaseIDs) > 0 {
		relQuery := "INSERT INTO psych_support_org_disease_rel (org_id, disease_id) VALUES (?, ?)"
		for _, did := range in.DiseaseIDs {
			if _, err := tx.Exec(relQuery, id, did); err != nil {
				return 0, err
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return id, nil
}

func (r *PsychologicalRepo) Update(id uint64, in domain.UpdatePsychOrgInput) error {
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
	if in.Address != "" {
		updateFields = append(updateFields, "address = ?")
		args = append(args, in.Address)
	}
	if in.ContactPhone != "" {
		updateFields = append(updateFields, "contact_phone = ?")
		args = append(args, in.ContactPhone)
	}
	if in.ContactUrl != "" {
		updateFields = append(updateFields, "contact_url = ?")
		args = append(args, in.ContactUrl)
	}
	if in.IsFree != nil {
		isFreeInt := 0
		if *in.IsFree {
			isFreeInt = 1
		}
		updateFields = append(updateFields, "is_free = ?")
		args = append(args, isFreeInt)
	}
	if in.ConsultWay != "" {
		updateFields = append(updateFields, "consult_way = ?")
		args = append(args, in.ConsultWay)
	}
	if in.ContentIntro != "" {
		updateFields = append(updateFields, "content_intro = ?")
		args = append(args, in.ContentIntro)
	}
	if in.AuditStatus != 0 {
		updateFields = append(updateFields, "audit_status = ?")
		args = append(args, in.AuditStatus)
	}

	updateFields = append(updateFields, "updated_at = NOW()")
	args = append(args, id)

	if len(updateFields) > 1 {
		updateQuery := `UPDATE psych_support_org SET ` + joinUpdateFields(updateFields) + ` WHERE id = ?`
		if _, err = tx.Exec(updateQuery, args...); err != nil {
			return err
		}
	}

	if in.HasDiseaseIDs {
		if _, err := tx.Exec("DELETE FROM psych_support_org_disease_rel WHERE org_id = ?", id); err != nil {
			return err
		}
		if len(in.DiseaseIDs) > 0 {
			relQuery := "INSERT INTO psych_support_org_disease_rel (org_id, disease_id) VALUES (?, ?)"
			for _, did := range in.DiseaseIDs {
				if _, err := tx.Exec(relQuery, id, did); err != nil {
					return err
				}
			}
		}
	}

	return tx.Commit()
}

func (r *PsychologicalRepo) Delete(id uint64) error {
	tx, err := db.MySQL.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err = tx.Exec("DELETE FROM psych_support_org_disease_rel WHERE org_id = ?", id); err != nil {
		return err
	}
	if _, err = tx.Exec("DELETE FROM psych_support_org WHERE id = ?", id); err != nil {
		return err
	}
	return tx.Commit()
}
