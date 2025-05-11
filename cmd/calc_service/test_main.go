package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func setupTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}

	_, err = db.Exec(`CREATE TABLE users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		login TEXT UNIQUE NOT NULL,
		password_hash TEXT NOT NULL
	)`)
	if err != nil {
		t.Fatalf("failed to create table: %v", err)
	}

	return db
}

func TestRegisterAndLogin(t *testing.T) {
	db := setupTestDB(t)
	handler := SetupRouter(db)

	registerBody := `{"login":"testuser","password":"secret123"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/register", bytes.NewBufferString(registerBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)
	if w.Result().StatusCode != http.StatusOK {
		t.Fatalf("expected status 200 on register, got %d", w.Result().StatusCode)
	}

	loginBody := `{"login":"testuser","password":"secret123"}`
	req = httptest.NewRequest(http.MethodPost, "/api/v1/login", bytes.NewBufferString(loginBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()

	handler.ServeHTTP(w, req)
	if w.Result().StatusCode != http.StatusOK {
		t.Fatalf("expected status 200 on login, got %d", w.Result().StatusCode)
	}

	var resp map[string]string
	if err := json.NewDecoder(w.Result().Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode login response: %v", err)
	}
	token, ok := resp["token"]
	if !ok || token == "" {
		t.Fatal("expected token in login response")
	}
}
