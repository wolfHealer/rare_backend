package repo

import (
	"database/sql"
	"strings"
)

func nullString(ns sql.NullString) string {
	if ns.Valid {
		return ns.String
	}
	return ""
}

func joinUpdateFields(fields []string) string {
	return strings.Join(fields, ", ")
}
