package store

import (
	"database/sql"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

const dbPath = "store/groceries.db"

var db *sql.DB

func Init() error {
	if err := os.MkdirAll(filepath.Dir(dbPath), 0o755); err != nil {
		return err
	}

	dsn := dbPath + "?_pragma=journal_mode(WAL)&_pragma=foreign_keys(on)&_pragma=busy_timeout(5000)"
	conn, err := sql.Open("sqlite", dsn)
	if err != nil {
		return err
	}

	db = conn
	return migrate()
}

func migrate() error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS groceries (
			uuid TEXT PRIMARY KEY,
			name TEXT NOT NULL
		);
		CREATE TABLE IF NOT EXISTS needed_groceries (
			uuid TEXT PRIMARY KEY,
			FOREIGN KEY (uuid) REFERENCES groceries(uuid) ON DELETE CASCADE
		);
	`)
	return err
}

func Close() error {
	if db == nil {
		return nil
	}
	return db.Close()
}
