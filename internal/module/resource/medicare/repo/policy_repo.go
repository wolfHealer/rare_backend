package repo

import (
	"rare_backend/internal/module/resource/medicare/domain"
	"rare_backend/internal/pkg/db"
)

type PolicyRepo struct{}

func NewPolicyRepo() *PolicyRepo {
	return &PolicyRepo{}
}

func (r *PolicyRepo) buildListWhere(filter domain.PolicyListFilter, mode domain.QueryMode) (string, []interface{}) {
	whereClause := "WHERE audit_status = 1"
	args := []interface{}{}

	if filter.NeedLatest {
		whereClause += " AND is_latest = 1"
	}

	switch mode {
	case domain.ModeRegionDisease:
		whereClause += " AND disease_id = ?"
		args = append(args, filter.DiseaseID)
		whereClause += " AND (scope_level = 1 OR (scope_level = 2 AND province_code = ?) OR (scope_level = 3 AND city_code = ?))"
		args = append(args, filter.ProvinceCode, filter.CityCode)
	case domain.ModeRegionOnly:
		whereClause += " AND (scope_level = 1 OR (scope_level = 2 AND province_code = ?) OR (scope_level = 3 AND city_code = ?))"
		args = append(args, filter.ProvinceCode, filter.CityCode)
	case domain.ModeDiseaseOnly:
		whereClause += " AND disease_id = ?"
		args = append(args, filter.DiseaseID)
	case domain.ModeDefaultRecommend:
		whereClause += " AND scope_level = 1"
	}

	return whereClause, args
}

func (r *PolicyRepo) CountList(filter domain.PolicyListFilter, mode domain.QueryMode) (int64, error) {
	whereClause, args := r.buildListWhere(filter, mode)
	var total int64
	err := db.MySQL.QueryRow("SELECT COUNT(*) FROM medical_insurance_policy "+whereClause, args...).Scan(&total)
	return total, err
}

func (r *PolicyRepo) List(filter domain.PolicyListFilter, mode domain.QueryMode) ([]domain.PolicyRow, error) {
	whereClause, args := r.buildListWhere(filter, mode)
	offset := (filter.Page - 1) * filter.PageSize

	listQuery := `
		SELECT id, disease_id, scope_level, province_code, city_code, province_name, city_name,
		       policy_title, policy_original, popular_interpret, is_latest, publish_date, created_at
		FROM medical_insurance_policy
		` + whereClause + `
		ORDER BY scope_level ASC, publish_date DESC, id DESC
		LIMIT ? OFFSET ?
	`
	args = append(args, filter.PageSize, offset)

	rows, err := db.MySQL.Query(listQuery, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []domain.PolicyRow
	for rows.Next() {
		var row domain.PolicyRow
		if err := rows.Scan(
			&row.ID, &row.DiseaseID, &row.ScopeLevel,
			&row.ProvinceCode, &row.CityCode, &row.ProvinceName, &row.CityName,
			&row.PolicyTitle, &row.PolicyOriginal, &row.PopularInterpret,
			&row.IsLatest, &row.PublishDate, &row.CreatedAt,
		); err != nil {
			continue
		}
		list = append(list, row)
	}
	return list, nil
}

func (r *PolicyRepo) GetByID(id uint64) (*domain.PolicyDetailRow, error) {
	query := `
		SELECT id, disease_id, scope_level, province_code, city_code, province_name, city_name,
		       policy_title, policy_original, popular_interpret, reimburse_ratio, reimburse_limit,
		       reimburse_process, reimburse_material, remote_apply_template,
		       is_latest, publish_date, effective_date, created_at, updated_at
		FROM medical_insurance_policy
		WHERE id = ? AND audit_status = 1
	`
	var row domain.PolicyDetailRow
	err := db.MySQL.QueryRow(query, id).Scan(
		&row.ID, &row.DiseaseID, &row.ScopeLevel,
		&row.ProvinceCode, &row.CityCode, &row.ProvinceName, &row.CityName,
		&row.PolicyTitle, &row.PolicyOriginal, &row.PopularInterpret,
		&row.ReimburseRatio, &row.ReimburseLimit,
		&row.ReimburseProcess, &row.ReimburseMaterial, &row.RemoteApplyTemplate,
		&row.IsLatest, &row.PublishDate, &row.EffectiveDate,
		&row.CreatedAt, &row.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *PolicyRepo) ListRelated(policyID uint64, provinceCode string, diseaseID int) ([]domain.RelatedPolicy, error) {
	query := `
		SELECT id, policy_title
		FROM medical_insurance_policy
		WHERE audit_status = 1
		  AND id != ?
		  AND (province_code = ? OR disease_id = ?)
		ORDER BY is_latest DESC, publish_date DESC
		LIMIT 5
	`
	rows, err := db.MySQL.Query(query, policyID, provinceCode, diseaseID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []domain.RelatedPolicy
	for rows.Next() {
		var item domain.RelatedPolicy
		if err := rows.Scan(&item.ID, &item.Title); err != nil {
			continue
		}
		list = append(list, item)
	}
	return list, nil
}

func (r *PolicyRepo) buildMaterialsWhere(materialType string) string {
	whereClause := "WHERE audit_status = 1"
	typeFieldMap := map[string]string{
		"flowchart": "reimburse_process",
		"guide":     "popular_interpret",
		"template":  "remote_apply_template",
		"checklist": "reimburse_material",
	}
	if materialType != "" {
		if field, ok := typeFieldMap[materialType]; ok {
			whereClause += " AND " + field + " != ''"
		}
	}
	return whereClause
}

func (r *PolicyRepo) ListMaterials(materialType string) ([]domain.MaterialRow, error) {
	query := `
		SELECT
			reimburse_process as url_flowchart,
			reimburse_material as url_checklist,
			remote_apply_template as url_template,
			updated_at
		FROM medical_insurance_policy
		` + r.buildMaterialsWhere(materialType) + `
		ORDER BY is_latest DESC, updated_at DESC
		LIMIT 50
	`

	rows, err := db.MySQL.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []domain.MaterialRow
	for rows.Next() {
		var row domain.MaterialRow
		if err := rows.Scan(&row.UrlFlowchart, &row.UrlChecklist, &row.UrlTemplate, &row.UpdateTime); err != nil {
			continue
		}
		list = append(list, row)
	}
	return list, nil
}
