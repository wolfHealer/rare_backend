package repo

import (
	"rare_backend/internal/pkg/db"
)

type OptionsRepo struct{}

func NewOptionsRepo() *OptionsRepo {
	return &OptionsRepo{}
}

func (r *OptionsRepo) GetDistinctProvinces() ([]string, error) {
	provinceQuery := "SELECT DISTINCT province_name FROM rehab_institution WHERE audit_status = 1 AND province_name != '' ORDER BY province_name ASC"
	rows, err := db.MySQL.Query(provinceQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var provinces []string
	for rows.Next() {
		var province string
		if err := rows.Scan(&province); err != nil {
			continue
		}
		if province != "" {
			provinces = append(provinces, province)
		}
	}
	return provinces, nil
}

type DiseaseOptionRow struct {
	ID   int64
	Name string
}

func (r *OptionsRepo) GetDiseaseOptionRows() ([]DiseaseOptionRow, error) {
	diseaseQuery := `
		SELECT id, name
		FROM disease
		WHERE status = 1
		ORDER BY id ASC
	`
	rows, err := db.MySQL.Query(diseaseQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []DiseaseOptionRow
	for rows.Next() {
		var row DiseaseOptionRow
		if err := rows.Scan(&row.ID, &row.Name); err != nil {
			continue
		}
		result = append(result, row)
	}
	return result, nil
}
