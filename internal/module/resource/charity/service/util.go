package service

import (
	"database/sql"
	"regexp"
	"strconv"
	"strings"
)

func auditStatusToStatusText(auditStatus int) string {
	switch auditStatus {
	case 1:
		return "open"
	case 0:
		return "pending"
	case 2:
		return "rejected"
	default:
		return "closed"
	}
}

func truncateSummary(text string, maxLen int) string {
	if len(text) <= maxLen {
		return text
	}
	runes := []rune(text)
	if len(runes) <= maxLen {
		return text
	}
	return string(runes[:maxLen]) + "..."
}

func convertChannelType(channelType string) string {
	typeMap := map[string]string{
		"emergency_help":     "urgent",
		"crowdfunding":       "crowdfunding",
		"charity_consulting": "consulting",
		"founding_support":   "foundation",
		"紧急求助":               "urgent",
		"众筹求助":               "crowdfunding",
		"医疗救助":               "medical",
		"生活补助":               "living",
	}
	if t, ok := typeMap[channelType]; ok {
		return t
	}
	return "other"
}

func getFileSize(url string) string {
	if strings.Contains(url, ".pdf") {
		return "1.2MB"
	}
	return "未知"
}

func extractAmountValue(amountStr string) int {
	re := regexp.MustCompile(`\d+`)
	matches := re.FindAllString(amountStr, -1)
	if len(matches) > 0 {
		value, _ := strconv.Atoi(matches[0])
		if strings.Contains(amountStr, "万") {
			value *= 10000
		}
		return value
	}
	return 0
}

func trimApplyDeadline(deadlineStr string) string {
	if strings.Contains(deadlineStr, "T") {
		parts := strings.Split(deadlineStr, "T")
		return parts[0]
	}
	return deadlineStr
}

func parseAuditStatusParam(auditStatusStr string, defaultWhenMissing int) int {
	if auditStatusStr == "" {
		return defaultWhenMissing
	}
	status, err := strconv.Atoi(auditStatusStr)
	if err == nil && (status == 0 || status == 1 || status == 2) {
		return status
	}
	return -2
}

func normalizePage(page, pageSize, defaultPageSize int) (int, int) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = defaultPageSize
	}
	return page, pageSize
}

func nullStringPtr(ns sql.NullString) *string {
	if ns.Valid {
		return &ns.String
	}
	return nil
}
