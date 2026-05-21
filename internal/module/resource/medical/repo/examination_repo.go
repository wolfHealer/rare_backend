package repo

import (
	"database/sql"
	"strconv"
	"time"

	"rare_backend/internal/module/resource/medical/domain"
	"rare_backend/internal/pkg/db"
)

type ExaminationRepo struct {
	rel *DiseaseRelRepo
}

func NewExaminationRepo(rel *DiseaseRelRepo) *ExaminationRepo {
	return &ExaminationRepo{rel: rel}
}

type examinationListRow struct {
	ID          uint64
	ExamName    string
	ExamType    string
	ExamPurpose string
	SampleNotes string
	Institution string
	Sort        int
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type examinationDetailRow struct {
	ID                uint64
	ExamName          string
	ExamType          string
	ExamPurpose       string
	ReferenceValue    sql.NullString
	AbnormalInterpret sql.NullString
	SampleNotes       sql.NullString
	Institution       sql.NullString
	TemplateExcel     sql.NullString
	TemplateWord      sql.NullString
	CompareTemplate   sql.NullString
	AuditStatus       int8
	RejectReason      sql.NullString
	Sort              int
	CreatedAt         time.Time
}

func (r *ExaminationRepo) buildListWhere(filter domain.ExaminationListFilter) (string, []interface{}, string) {
	whereClause := "WHERE 1=1"
	args := []interface{}{}

	if filter.AuditStatus != "" {
		if auditStatus, err := strconv.Atoi(filter.AuditStatus); err == nil {
			whereClause += " AND em.audit_status = ?"
			args = append(args, auditStatus)
		}
	}

	if filter.Keyword != "" {
		whereClause += " AND (em.exam_name LIKE ? OR em.exam_purpose LIKE ?)"
		args = append(args, "%"+filter.Keyword+"%", "%"+filter.Keyword+"%")
	}

	if filter.ExamType != "" {
		whereClause += " AND em.exam_type = ?"
		args = append(args, filter.ExamType)
	}

	joinClause := ""
	if filter.DiseaseID > 0 {
		joinClause = "JOIN exam_manual_disease_rel emdr ON em.id = emdr.exam_manual_id"
		whereClause += " AND emdr.disease_id = ?"
		args = append(args, filter.DiseaseID)
	}

	return whereClause, args, joinClause
}

func (r *ExaminationRepo) CountList(filter domain.ExaminationListFilter) (int64, error) {
	whereClause, args, joinClause := r.buildListWhere(filter)
	var total int64
	countQuery := "SELECT COUNT(*) FROM examination_manual em " + joinClause + " " + whereClause
	err := db.MySQL.QueryRow(countQuery, args...).Scan(&total)
	return total, err
}

func (r *ExaminationRepo) List(filter domain.ExaminationListFilter) ([]examinationListRow, error) {
	whereClause, args, joinClause := r.buildListWhere(filter)
	offset := (filter.Page - 1) * filter.PageSize

	listQuery := `
		SELECT em.id, em.exam_name, em.exam_type, em.exam_purpose, em.sample_notes, em.institution, em.sort, em.created_at, em.updated_at
		FROM examination_manual em
		` + joinClause + `
		` + whereClause + `
		ORDER BY em.sort ASC, em.created_at DESC
		LIMIT ? OFFSET ?
	`
	listArgs := append([]interface{}{}, args...)
	listArgs = append(listArgs, filter.PageSize, offset)

	rows, err := db.MySQL.Query(listQuery, listArgs...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []examinationListRow
	for rows.Next() {
		var item examinationListRow
		if err := rows.Scan(&item.ID, &item.ExamName, &item.ExamType, &item.ExamPurpose, &item.SampleNotes, &item.Institution, &item.Sort, &item.CreatedAt, &item.UpdatedAt); err != nil {
			continue
		}
		list = append(list, item)
	}
	return list, nil
}

func (r *ExaminationRepo) GetByID(id uint64) (*examinationDetailRow, error) {
	query := `
		SELECT id, exam_name, exam_type, exam_purpose, reference_value, abnormal_interpret, sample_notes, institution,
		       template_excel, template_word, compare_template, audit_status, reject_reason, sort, created_at
		FROM examination_manual
		WHERE id = ?
	`
	var item examinationDetailRow
	if err := db.MySQL.QueryRow(query, id).Scan(
		&item.ID, &item.ExamName, &item.ExamType, &item.ExamPurpose,
		&item.ReferenceValue, &item.AbnormalInterpret, &item.SampleNotes, &item.Institution,
		&item.TemplateExcel, &item.TemplateWord, &item.CompareTemplate,
		&item.AuditStatus, &item.RejectReason, &item.Sort, &item.CreatedAt,
	); err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *ExaminationRepo) Create(in domain.CreateExaminationInput) (int64, error) {
	tx, err := db.MySQL.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	templateExcel := ""
	templateWord := ""
	compareTemplate := ""

	if in.Templates != nil {
		templateExcel = in.Templates.Excel
		templateWord = in.Templates.Word
		compareTemplate = in.Templates.Compare
	}

	initAuditStatus := int8(0)
	if in.AuditStatus != nil {
		initAuditStatus = *in.AuditStatus
	}

	insertQuery := `
		INSERT INTO examination_manual (
			exam_type, exam_name, exam_purpose, reference_value, abnormal_interpret, 
			sample_notes, institution, template_excel, template_word, compare_template, 
			audit_status, sort, created_at, updated_at
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW())
	`

	res, err := tx.Exec(insertQuery,
		in.ExamType,
		in.ExamName,
		in.ExamPurpose,
		in.ReferenceValue,
		in.AbnormalInterpret,
		in.SampleNotes,
		in.Institution,
		templateExcel,
		templateWord,
		compareTemplate,
		initAuditStatus,
		in.Sort,
	)
	if err != nil {
		return 0, err
	}

	manualID, _ := res.LastInsertId()

	if len(in.DiseaseIDs) > 0 {
		if err := r.rel.InsertExamManualDiseaseRel(tx, uint64(manualID), in.DiseaseIDs); err != nil {
			return 0, err
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return manualID, nil
}

func (r *ExaminationRepo) Update(id uint64, in domain.UpdateExaminationInput) error {
	tx, err := db.MySQL.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	fields := []string{}
	values := []interface{}{}

	if in.ExamName != nil {
		fields = append(fields, "exam_name=?")
		values = append(values, *in.ExamName)
	}
	if in.ExamType != nil {
		fields = append(fields, "exam_type=?")
		values = append(values, *in.ExamType)
	}
	if in.ExamPurpose != nil {
		fields = append(fields, "exam_purpose=?")
		values = append(values, *in.ExamPurpose)
	}
	if in.ReferenceValue != nil {
		fields = append(fields, "reference_value=?")
		values = append(values, *in.ReferenceValue)
	}
	if in.AbnormalInterpret != nil {
		fields = append(fields, "abnormal_interpret=?")
		values = append(values, *in.AbnormalInterpret)
	}
	if in.SampleNotes != nil {
		fields = append(fields, "sample_notes=?")
		values = append(values, *in.SampleNotes)
	}
	if in.Institution != nil {
		fields = append(fields, "institution=?")
		values = append(values, *in.Institution)
	}

	if in.Templates != nil {
		fields = append(fields, "template_excel=?")
		values = append(values, in.Templates.Excel)
		fields = append(fields, "template_word=?")
		values = append(values, in.Templates.Word)
		fields = append(fields, "compare_template=?")
		values = append(values, in.Templates.Compare)
	}

	if in.Sort != nil {
		fields = append(fields, "sort=?")
		values = append(values, *in.Sort)
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
		sqlStr := "UPDATE examination_manual SET " + joinUpdateFields(fields) + " WHERE id=?"
		if _, err := tx.Exec(sqlStr, values...); err != nil {
			return err
		}
	}

	if in.HasDiseaseIDs {
		if _, err := tx.Exec("DELETE FROM exam_manual_disease_rel WHERE exam_manual_id = ?", id); err != nil {
			return err
		}
		if len(in.DiseaseIDs) > 0 {
			if err := r.rel.InsertExamManualDiseaseRel(tx, id, in.DiseaseIDs); err != nil {
				return err
			}
		}
	}

	return tx.Commit()
}

func (r *ExaminationRepo) Delete(id uint64) error {
	tx, err := db.MySQL.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec("DELETE FROM exam_manual_disease_rel WHERE exam_manual_id = ?", id); err != nil {
		return err
	}

	result, err := tx.Exec("DELETE FROM examination_manual WHERE id = ?", id)
	if err != nil {
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return tx.Commit()
}
