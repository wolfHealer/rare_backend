package repo

import (
	"database/sql"
	"time"

	"rare_backend/internal/module/resource/charity/domain"
	"rare_backend/internal/pkg/db"
	"rare_backend/internal/pkg/search"
)

type ChannelRepo struct{}

func NewChannelRepo() *ChannelRepo {
	return &ChannelRepo{}
}

type ChannelRow struct {
	ID                   uint
	ChannelType          string
	Name                 string
	ApplyCondition       string
	ResponseTime         string
	ContactPhone         string
	ContactUrl           string
	HelpLetterTemplate   string
	CrowdfundingTemplate string
	AuditStatus          int
	RejectReason         sql.NullString
	Sort                 int
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

func (r *ChannelRepo) buildListWhere(filter domain.ChannelListFilter) (string, []interface{}) {
	whereClause := "WHERE 1=1"
	args := []interface{}{}

	if filter.AuditStatus >= 0 && filter.AuditStatus <= 2 {
		whereClause += " AND audit_status = ?"
		args = append(args, filter.AuditStatus)
	} else if filter.AuditStatus == -1 {
		whereClause += " AND audit_status = 1"
	}

	if filter.ChannelType != "" {
		whereClause += " AND channel_type = ?"
		args = append(args, filter.ChannelType)
	}
	if filter.Keyword != "" {
		if clause, arg, ok := search.MatchClause("name", filter.Keyword); ok {
			whereClause += clause
			args = append(args, arg)
		}
	}
	return whereClause, args
}

func (r *ChannelRepo) CountList(filter domain.ChannelListFilter) (int64, error) {
	whereClause, args := r.buildListWhere(filter)
	var total int64
	err := db.MySQL.QueryRow("SELECT COUNT(*) FROM help_channel "+whereClause, args...).Scan(&total)
	return total, err
}

func (r *ChannelRepo) List(filter domain.ChannelListFilter) ([]ChannelRow, error) {
	whereClause, args := r.buildListWhere(filter)
	offset := (filter.Page - 1) * filter.PageSize

	listQuery := `
		SELECT id, channel_type, name, apply_condition, response_time, 
		       contact_phone, contact_url, help_letter_template, crowdfunding_template,
		       audit_status, reject_reason, sort, created_at, updated_at
		FROM help_channel 
		` + whereClause + `
		ORDER BY sort DESC, id ASC
		LIMIT ? OFFSET ?
	`
	args = append(args, filter.PageSize, offset)

	rows, err := db.MySQL.Query(listQuery, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []ChannelRow
	for rows.Next() {
		var row ChannelRow
		if err := rows.Scan(
			&row.ID, &row.ChannelType, &row.Name,
			&row.ApplyCondition, &row.ResponseTime, &row.ContactPhone, &row.ContactUrl,
			&row.HelpLetterTemplate, &row.CrowdfundingTemplate,
			&row.AuditStatus, &row.RejectReason, &row.Sort,
			&row.CreatedAt, &row.UpdatedAt,
		); err != nil {
			continue
		}
		list = append(list, row)
	}
	return list, nil
}

func (r *ChannelRepo) GetByID(id uint) (*ChannelRow, error) {
	query := `
		SELECT id, channel_type, name, apply_condition, response_time, 
		       contact_phone, contact_url, help_letter_template, crowdfunding_template,
		       audit_status, reject_reason, sort, created_at, updated_at
		FROM help_channel
		WHERE id = ?
	`
	var row ChannelRow
	err := db.MySQL.QueryRow(query, id).Scan(
		&row.ID, &row.ChannelType, &row.Name,
		&row.ApplyCondition, &row.ResponseTime, &row.ContactPhone, &row.ContactUrl,
		&row.HelpLetterTemplate, &row.CrowdfundingTemplate,
		&row.AuditStatus, &row.RejectReason, &row.Sort,
		&row.CreatedAt, &row.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *ChannelRepo) Create(in domain.CreateChannelInput, auditStatus int) (int64, error) {
	insertQuery := `
		INSERT INTO help_channel 
		(channel_type, name, apply_condition, response_time, contact_phone, contact_url,
		 help_letter_template, crowdfunding_template, audit_status, reject_reason, sort, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, '', ?, NOW(), NOW())
	`
	result, err := db.MySQL.Exec(insertQuery,
		in.ChannelType, in.Name, in.ApplyCondition, in.ResponseTime,
		in.ContactPhone, in.ContactUrl, in.HelpLetterTemplate, in.CrowdfundingTemplate,
		auditStatus, in.Sort)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (r *ChannelRepo) Update(id uint, updateFields []string, args []interface{}) error {
	if len(updateFields) == 0 {
		return nil
	}
	updateFields = append(updateFields, "updated_at = NOW()")
	args = append(args, id)
	updateQuery := `UPDATE help_channel SET ` + joinUpdateFields(updateFields) + ` WHERE id = ?`
	_, err := db.MySQL.Exec(updateQuery, args...)
	return err
}

func (r *ChannelRepo) Delete(id uint) (int64, error) {
	result, err := db.MySQL.Exec(`DELETE FROM help_channel WHERE id = ?`, id)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}
