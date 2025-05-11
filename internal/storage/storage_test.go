package storage

import (
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func TestNewSQLite(t *testing.T) {
	db, err := NewSQLite(":memory:")
	if err != nil {
		t.Fatalf("failed to create new SQLite DB: %v", err)
	}
	defer db.Close()

	if !tableExists(t, db, "users") {
		t.Fatal("table 'users' does not exist after migration")
	}

	if !tableExists(t, db, "calculations") {
		t.Fatal("table 'calculations' does not exist after migration")
	}
}

func tableExists(t *testing.T, db *sql.DB, tableName string) bool {
	t.Helper()
	query := `SELECT name FROM sqlite_master WHERE type='table' AND name=?;`
	var name string
	err := db.QueryRow(query, tableName).Scan(&name)
	if err != nil {
		if err == sql.ErrNoRows {
			return false
		}
		t.Fatalf("error checking table existence: %v", err)
	}
	return name == tableName
}
