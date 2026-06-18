package domain

import "time"

const (
	ReportStatusPending  = 0
	ReportStatusHandled  = 1
	ReportStatusRejected = 2
)

const ReportTimeLayout = "2006-01-02T15:04:05"

type ReportListFilter struct {
	Status     *int
	ReasonType string
	Keyword    string
	Page       int
	PageSize   int
}

// ReportItem 管理端举报列表/详情
type ReportItem struct {
	ID             int64   `json:"id"`
	PostID         int64   `json:"post_id"`
	ReporterID     int64   `json:"reporter_id"`
	ReasonType     string  `json:"reason_type"`
	Description    string  `json:"description"`
	Status         int     `json:"status"`
	HandledAt      *string `json:"handled_at"`
	HandlerID      *int64  `json:"handler_id"`
	HandleNote     string  `json:"handle_note"`
	CreatedAt      string  `json:"created_at"`
	ReporterName   string  `json:"reporter_name"`
	ReporterPhone  string  `json:"reporter_phone"`
	HandlerName    string  `json:"handler_name"`
	PostTitle      string  `json:"post_title"`
	PostContent    string  `json:"post_content"`
	PostAuthorName string  `json:"post_author_name"`
}

type HandleReportInput struct {
	ReportID   int64
	HandlerID  int64
	Status     int
	HandleNote string
	RemovePost bool
}

func FormatReportTime(t time.Time) string {
	return t.Format(ReportTimeLayout)
}
