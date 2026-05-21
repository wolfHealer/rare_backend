package repo

import (
	"strings"
	"time"

	"rare_backend/internal/pkg/db"
)

func JoinUpdateFields(fields []string) string {
	return strings.Join(fields, ", ")
}

func ExecDynamicUpdate(table string, id uint, setParts []string, args []interface{}) error {
	if len(setParts) == 0 {
		return nil
	}
	query := "UPDATE " + table + " SET " + JoinUpdateFields(setParts) + " WHERE id = ?"
	_, err := db.MySQL.Exec(query, args...)
	return err
}

func SoftDelete(table string, id uint) error {
	_, err := db.MySQL.Exec(
		"UPDATE "+table+" SET status = 0, updated_at = ? WHERE id = ?",
		time.Now(), id,
	)
	return err
}

func ExistsByID(table string, id uint) (bool, error) {
	var exists uint
	err := db.MySQL.QueryRow("SELECT id FROM "+table+" WHERE id = ?", id).Scan(&exists)
	return err == nil, err
}

func ExistsActiveByID(table string, id uint) (bool, error) {
	var exists uint
	err := db.MySQL.QueryRow("SELECT id FROM "+table+" WHERE id = ? AND status = 1", id).Scan(&exists)
	return err == nil, err
}
