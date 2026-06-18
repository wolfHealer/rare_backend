package domain

// 举报类型（与前端 reason_type 对齐）
const (
	ReportReasonSpam    = "spam"
	ReportReasonAbuse   = "abuse"
	ReportReasonIllegal = "illegal"
	ReportReasonMisinfo = "misinfo"
	ReportReasonOther   = "other"
)

var ValidReportReasons = map[string]struct{}{
	ReportReasonSpam:    {},
	ReportReasonAbuse:   {},
	ReportReasonIllegal: {},
	ReportReasonMisinfo: {},
	ReportReasonOther:   {},
}

type ReportPostInput struct {
	PostID      int64
	ReporterID  int64
	ReasonType  string
	Description string
}
