package database

import (
	"database/sql"

	_ "github.com/go-sql-driver/mysql"
)

var DB *sql.DB

func Connect(dsn string) error {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return err
	}
	if err = db.Ping(); err != nil {
		return err
	}
	DB = db
	return nil
}

// Migrate menambah kolom yang dibutuhkan (error diabaikan kalau kolom sudah ada).
func Migrate() {
	DB.Exec("ALTER TABLE users ADD COLUMN menus TEXT NULL")
	DB.Exec("ALTER TABLE melk_susu ADD COLUMN ref_id BIGINT NULL")
}
