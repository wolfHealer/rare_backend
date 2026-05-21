package repo

import (
	"database/sql"
	"strings"
)

func nullStringPtr(ns sql.NullString) *string {
	if ns.Valid {
		return &ns.String
	}
	return nil
}

func joinUpdateFields(fields []string) string {
	return strings.Join(fields, ", ")
}
