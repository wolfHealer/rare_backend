package db

import (
	"database/sql"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

var MySQL *sql.DB

func InitMySQL(dsn string) error {
	// dsn 示例: "user:pass@tcp(127.0.0.1:3306)/dbname?parseTime=true"
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return err
	}
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err := db.Ping(); err != nil {
		return err
	}
	MySQL = db
	return nil
}

// 更新示例：更新用户的 email
func UpdateUserEmail(userID int64, email string) (int64, error) {
	res, err := MySQL.Exec("UPDATE users SET email = ? WHERE id = ?", email, userID)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}
