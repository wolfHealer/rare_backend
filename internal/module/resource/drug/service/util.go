package service

import "database/sql"

func nullString(ns sql.NullString) string {
	if ns.Valid {
		return ns.String
	}
	return ""
}

// convertDrugType 转换药品类型枚举
func convertDrugType(drugType string) string {
	typeMap := map[string]string{
		"进口":  "imported",
		"国产":  "domestic",
		"仿制药": "generic",
	}
	if val, ok := typeMap[drugType]; ok {
		return val
	}
	return drugType
}

func mapYesNo(val int8) string {
	if val == 1 {
		return "是"
	}
	return "否"
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
