package repo

import (
	"database/sql"
	"fmt"
	"strings"

	"rare_backend/internal/module/resource/medical/domain"
	"rare_backend/internal/pkg/db"
)

type DiseaseRelRepo struct{}

func NewDiseaseRelRepo() *DiseaseRelRepo {
	return &DiseaseRelRepo{}
}

func (r *DiseaseRelRepo) GetDiseasesByHospital(hospitalID uint64) ([]domain.DiseaseSimpleInfo, error) {
	query := `
		SELECT d.id, d.name, d.alias 
		FROM hospital_disease_rel hdr
		JOIN disease d ON hdr.disease_id = d.id
		WHERE hdr.hospital_id = ?
	`
	rows, err := db.MySQL.Query(query, hospitalID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var diseases []domain.DiseaseSimpleInfo
	for rows.Next() {
		var d domain.DiseaseSimpleInfo
		var alias sql.NullString
		if err := rows.Scan(&d.ID, &d.Name, &alias); err != nil {
			continue
		}
		if alias.Valid {
			d.Alias = alias.String
		}
		diseases = append(diseases, d)
	}
	return diseases, nil
}

func (r *DiseaseRelRepo) GetDiseaseIDsByDoctor(doctorID uint64) ([]uint64, error) {
	rows, err := db.MySQL.Query("SELECT disease_id FROM doctor_disease_rel WHERE doctor_id = ?", doctorID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ids []uint64
	for rows.Next() {
		var id uint64
		if err := rows.Scan(&id); err != nil {
			continue
		}
		ids = append(ids, id)
	}
	return ids, nil
}

func (r *DiseaseRelRepo) ListDiseaseIDsByDoctorIDs(doctorIDs []uint64) (map[uint64][]uint64, error) {
	result := make(map[uint64][]uint64)
	if len(doctorIDs) == 0 {
		return result, nil
	}

	placeholders, args := buildInPlaceholders(doctorIDs)
	query := fmt.Sprintf(
		"SELECT doctor_id, disease_id FROM doctor_disease_rel WHERE doctor_id IN (%s)",
		strings.Join(placeholders, ","),
	)
	rows, err := db.MySQL.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var doctorID, diseaseID uint64
		if err := rows.Scan(&doctorID, &diseaseID); err != nil {
			continue
		}
		result[doctorID] = append(result[doctorID], diseaseID)
	}
	return result, rows.Err()
}

func (r *DiseaseRelRepo) GetDiseaseDetailsByDoctor(doctorID uint64) ([]domain.DiseaseSimpleInfo, error) {
	query := `
		SELECT d.id, d.name, d.alias 
		FROM doctor_disease_rel ddr
		JOIN disease d ON ddr.disease_id = d.id
		WHERE ddr.doctor_id = ?
	`
	rows, err := db.MySQL.Query(query, doctorID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var diseases []domain.DiseaseSimpleInfo
	for rows.Next() {
		var d domain.DiseaseSimpleInfo
		var alias sql.NullString
		if err := rows.Scan(&d.ID, &d.Name, &alias); err != nil {
			continue
		}
		if alias.Valid {
			d.Alias = alias.String
		}
		diseases = append(diseases, d)
	}
	return diseases, nil
}

func (r *DiseaseRelRepo) GetDiseaseIDsByExamManual(manualID uint64) ([]uint64, error) {
	rows, err := db.MySQL.Query("SELECT disease_id FROM exam_manual_disease_rel WHERE exam_manual_id = ?", manualID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ids []uint64
	for rows.Next() {
		var id uint64
		if err := rows.Scan(&id); err != nil {
			continue
		}
		ids = append(ids, id)
	}
	return ids, nil
}

func (r *DiseaseRelRepo) ListDiseaseIDsByExamManualIDs(manualIDs []uint64) (map[uint64][]uint64, error) {
	result := make(map[uint64][]uint64)
	if len(manualIDs) == 0 {
		return result, nil
	}

	placeholders, args := buildInPlaceholders(manualIDs)
	query := fmt.Sprintf(
		"SELECT exam_manual_id, disease_id FROM exam_manual_disease_rel WHERE exam_manual_id IN (%s)",
		strings.Join(placeholders, ","),
	)
	rows, err := db.MySQL.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var manualID, diseaseID uint64
		if err := rows.Scan(&manualID, &diseaseID); err != nil {
			continue
		}
		result[manualID] = append(result[manualID], diseaseID)
	}
	return result, rows.Err()
}

func (r *DiseaseRelRepo) InsertHospitalDiseaseRel(tx *sql.Tx, hospitalID uint64, diseaseIDs []uint64) error {
	stmt, err := tx.Prepare("INSERT INTO hospital_disease_rel (hospital_id, disease_id) VALUES (?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()
	for _, did := range diseaseIDs {
		if _, err := stmt.Exec(hospitalID, did); err != nil {
			return err
		}
	}
	return nil
}

func (r *DiseaseRelRepo) InsertDoctorDiseaseRel(tx *sql.Tx, doctorID uint64, diseaseIDs []uint64) error {
	stmt, err := tx.Prepare("INSERT INTO doctor_disease_rel (doctor_id, disease_id) VALUES (?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()
	for _, did := range diseaseIDs {
		if _, err := stmt.Exec(doctorID, did); err != nil {
			return err
		}
	}
	return nil
}

func (r *DiseaseRelRepo) InsertExamManualDiseaseRel(tx *sql.Tx, manualID uint64, diseaseIDs []uint64) error {
	stmt, err := tx.Prepare("INSERT INTO exam_manual_disease_rel (exam_manual_id, disease_id) VALUES (?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()
	for _, did := range diseaseIDs {
		if _, err := stmt.Exec(manualID, did); err != nil {
			return err
		}
	}
	return nil
}

func (r *DiseaseRelRepo) DeleteHospitalDiseaseRel(tx *sql.Tx, hospitalID uint64) error {
	_, err := tx.Exec("DELETE FROM hospital_disease_rel WHERE hospital_id = ?", hospitalID)
	return err
}

func (r *DiseaseRelRepo) DeleteDoctorDiseaseRel(tx *sql.Tx, doctorID uint64) error {
	_, err := tx.Exec("DELETE FROM doctor_disease_rel WHERE doctor_id = ?", doctorID)
	return err
}

func (r *DiseaseRelRepo) DeleteExamManualDiseaseRel(tx *sql.Tx, manualID uint64) error {
	_, err := tx.Exec("DELETE FROM exam_manual_disease_rel WHERE exam_manual_id = ?", manualID)
	return err
}

func buildInPlaceholders(ids []uint64) ([]string, []interface{}) {
	placeholders := make([]string, len(ids))
	args := make([]interface{}, len(ids))
	for i, id := range ids {
		placeholders[i] = "?"
		args[i] = id
	}
	return placeholders, args
}
