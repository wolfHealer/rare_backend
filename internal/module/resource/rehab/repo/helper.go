package repo

import (
	"database/sql"
	"fmt"
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

func buildInPlaceholders(count int) string {
	if count == 0 {
		return ""
	}
	placeholders := make([]string, count)
	for i := range placeholders {
		placeholders[i] = "?"
	}
	return strings.Join(placeholders, ",")
}

func buildInQuery(base string, ids []uint64) (string, []interface{}) {
	if len(ids) == 0 {
		return "", nil
	}
	args := make([]interface{}, len(ids))
	for i, id := range ids {
		args[i] = id
	}
	return fmt.Sprintf(base, buildInPlaceholders(len(ids))), args
}
