package repo

import (
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"

	"rare_backend/internal/module/community/domain"
	"rare_backend/internal/pkg/db"
)

// ReportRepo 帖子举报（依赖 post_report 表，需在 RDS 中手动建表）
type ReportRepo struct{}

func NewReportRepo() *ReportRepo {
	return &ReportRepo{}
}

const reportSelectSQL = `
	SELECT
		r.id, r.post_id, r.reporter_id, r.reason_type, r.description,
		r.status, r.handle_note, r.handler_id, r.created_at, r.handled_at,
		COALESCE(p.title, ''), COALESCE(p.content, ''),
		COALESCE(reporter.display_name, ''), COALESCE(reporter.phone, ''),
		COALESCE(author.display_name, ''), COALESCE(handler.display_name, '')
	FROM post_report r
	INNER JOIN post p ON p.id = r.post_id
	LEFT JOIN user reporter ON reporter.id = r.reporter_id
	LEFT JOIN user author ON author.id = p.user_id
	LEFT JOIN user handler ON handler.id = r.handler_id
`

func (r *ReportRepo) Create(in domain.ReportPostInput) (int64, error) {
	res, err := db.MySQL.Exec(`
		INSERT INTO post_report (post_id, reporter_id, reason_type, description, status)
		VALUES (?, ?, ?, ?, 0)
	`, in.PostID, in.ReporterID, in.ReasonType, nullIfEmpty(in.Description))
	if err != nil {
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
			return 0, domain.ErrReportDuplicate
		}
		return 0, err
	}
	return res.LastInsertId()
}

func (r *ReportRepo) List(filter domain.ReportListFilter) ([]domain.ReportItem, int64, error) {
	where, args := buildReportListWhere(filter)

	countQuery := "SELECT COUNT(*) FROM post_report r " +
		"INNER JOIN post p ON p.id = r.post_id " +
		"LEFT JOIN user reporter ON reporter.id = r.reporter_id " + where
	var total int64
	if err := db.MySQL.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	offset := (filter.Page - 1) * filter.PageSize
	listArgs := append(append([]interface{}{}, args...), filter.PageSize, offset)
	query := reportSelectSQL + where + `
		ORDER BY r.status ASC, r.created_at DESC
		LIMIT ? OFFSET ?`

	rows, err := db.MySQL.Query(query, listArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	list := make([]domain.ReportItem, 0)
	for rows.Next() {
		item, err := scanReportItem(rows)
		if err != nil {
			return nil, 0, err
		}
		list = append(list, item)
	}
	return list, total, rows.Err()
}

func (r *ReportRepo) GetDetailByID(reportID int64) (*domain.ReportItem, error) {
	row := db.MySQL.QueryRow(reportSelectSQL+` WHERE r.id = ?`, reportID)
	item, err := scanReportItem(row)
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *ReportRepo) GetByID(reportID int64) (status int, postID int64, err error) {
	err = db.MySQL.QueryRow(
		`SELECT status, post_id FROM post_report WHERE id = ?`, reportID,
	).Scan(&status, &postID)
	return status, postID, err
}

func (r *ReportRepo) Handle(in domain.HandleReportInput) error {
	res, err := db.MySQL.Exec(`
		UPDATE post_report
		SET status = ?, handle_note = ?, handler_id = ?, handled_at = ?
		WHERE id = ? AND status = 0
	`, in.Status, nullIfEmpty(in.HandleNote), in.HandlerID, time.Now(), in.ReportID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		var cur int
		err := db.MySQL.QueryRow(`SELECT status FROM post_report WHERE id = ?`, in.ReportID).Scan(&cur)
		if err == sql.ErrNoRows {
			return domain.ErrReportNotFound
		}
		if err == nil {
			return domain.ErrReportAlreadyDone
		}
		return err
	}
	return nil
}

func buildReportListWhere(filter domain.ReportListFilter) (string, []interface{}) {
	where := "WHERE 1=1"
	args := []interface{}{}
	if filter.Status != nil {
		where += " AND r.status = ?"
		args = append(args, *filter.Status)
	}
	if filter.ReasonType != "" {
		where += " AND r.reason_type = ?"
		args = append(args, filter.ReasonType)
	}
	if kw := strings.TrimSpace(filter.Keyword); kw != "" {
		like := "%" + kw + "%"
		where += " AND (p.title LIKE ? OR p.content LIKE ? OR reporter.display_name LIKE ? OR reporter.phone LIKE ? OR r.description LIKE ?)"
		args = append(args, like, like, like, like, like)
	}
	return where, args
}

type reportScanner interface {
	Scan(dest ...any) error
}

func scanReportItem(s reportScanner) (domain.ReportItem, error) {
	var item domain.ReportItem
	var desc, note sql.NullString
	var handlerID sql.NullInt64
	var handledAt sql.NullTime
	var createdAt time.Time
	var reporterPhone sql.NullString

	err := s.Scan(
		&item.ID, &item.PostID, &item.ReporterID, &item.ReasonType, &desc,
		&item.Status, &note, &handlerID, &createdAt, &handledAt,
		&item.PostTitle, &item.PostContent,
		&item.ReporterName, &reporterPhone,
		&item.PostAuthorName, &item.HandlerName,
	)
	if err != nil {
		return item, err
	}
	if desc.Valid {
		item.Description = desc.String
	}
	if note.Valid {
		item.HandleNote = note.String
	}
	if handlerID.Valid {
		id := handlerID.Int64
		item.HandlerID = &id
	}
	if handledAt.Valid {
		formatted := domain.FormatReportTime(handledAt.Time)
		item.HandledAt = &formatted
	}
	item.CreatedAt = domain.FormatReportTime(createdAt)
	if reporterPhone.Valid {
		item.ReporterPhone = maskPhone(reporterPhone.String)
	}
	return item, nil
}

func maskPhone(phone string) string {
	phone = strings.TrimSpace(phone)
	if phone == "" || strings.HasPrefix(phone, "deleted_") {
		return ""
	}
	if len(phone) >= 3 {
		return phone[:3] + "****"
	}
	return "****"
}

func nullIfEmpty(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}
