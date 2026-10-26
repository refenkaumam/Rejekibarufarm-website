package database

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/go-sql-driver/mysql"
)

var DB *sql.DB

func Connect(dsn string) error {
	cfg, err := mysql.ParseDSN(dsn)
	if err != nil {
		return fmt.Errorf("format DB_DSN tidak valid")
	}
	if cfg.Timeout == 0 {
		cfg.Timeout = 5 * time.Second
	}
	if cfg.ReadTimeout == 0 {
		cfg.ReadTimeout = 30 * time.Second
	}
	if cfg.WriteTimeout == 0 {
		cfg.WriteTimeout = 30 * time.Second
	}
	db, err := sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		return err
	}
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(30 * time.Minute)
	if err = db.Ping(); err != nil {
		return err
	}
	DB = db
	return nil
}

func columnExists(table, col string) (bool, error) {
	var n int
	err := DB.QueryRow(`SELECT COUNT(*) FROM information_schema.COLUMNS
		WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME=? AND COLUMN_NAME=?`, table, col).Scan(&n)
	return n > 0, err
}

// Migrate hanya menambah kolom/tabel baru (aditif), tidak pernah menghapus atau mengubah data.
func Migrate() error {
	cols := []struct{ table, col, ddl string }{
		{"users", "menus", "ALTER TABLE users ADD COLUMN menus TEXT NULL"},
		{"melk_susu", "ref_id", "ALTER TABLE melk_susu ADD COLUMN ref_id BIGINT NULL"},
		{"users", "token_version", "ALTER TABLE users ADD COLUMN token_version INT NOT NULL DEFAULT 0"},
	}
	for _, m := range cols {
		ok, err := columnExists(m.table, m.col)
		if err != nil {
			return err
		}
		if !ok {
			if _, err := DB.Exec(m.ddl); err != nil {
				return fmt.Errorf("%s.%s: %w (sudahkah migrations/001_init_peternakan.sql diimpor?)", m.table, m.col, err)
			}
		}
	}
	_, err := DB.Exec(`CREATE TABLE IF NOT EXISTS audit_log (
		id BIGINT AUTO_INCREMENT PRIMARY KEY,
		at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		user_id BIGINT NULL,
		username VARCHAR(50) NULL,
		ip VARCHAR(45) NULL,
		aksi VARCHAR(30) NOT NULL,
		tabel VARCHAR(30) NULL,
		row_id BIGINT NULL,
		sebelum MEDIUMTEXT NULL,
		sesudah MEDIUMTEXT NULL,
		KEY idx_audit_at (at),
		KEY idx_audit_row (tabel, row_id)
	)`)
	return err
}
