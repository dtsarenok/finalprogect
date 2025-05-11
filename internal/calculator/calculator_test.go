package calculator

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"distributed-calculator/internal/auth"

	_ "github.com/mattn/go-sqlite3"
)

func setupTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}

	_, err = db.Exec(`
	CREATE TABLE users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		login TEXT UNIQUE NOT NULL,
		password_hash TEXT NOT NULL
	);
	CREATE TABLE calculations (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		expression TEXT NOT NULL,
		result TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY(user_id) REFERENCES users(id)
	);
	`)
	if err != nil {
		t.Fatalf("failed to create tables: %v", err)
	}

	return db
}

func contextWithUserID(userID int64) context.Context {
	return context.WithValue(context.Background(), auth.UserIDKey, userID)
}

func TestCalculateHandler_Success(t *testing.T) {
	db := setupTestDB(t)

	res, err := db.Exec("INSERT INTO users (login, password_hash) VALUES (?, ?)", "testuser", "hash")
	if err != nil {
		t.Fatalf("failed to insert user: %v", err)
	}
	userID, err := res.LastInsertId()
	if err != nil {
		t.Fatalf("failed to get user id: %v", err)
	}

	handler := CalculateHandler(db)

	reqBody := `{"expression": "2+3*4"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/calculate", bytes.NewBufferString(reqBody))
	req = req.WithContext(contextWithUserID(userID))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}

	var calcResp CalculateResponse
	if err := json.NewDecoder(resp.Body).Decode(&calcResp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	expectedResult := "14"
	if calcResp.Result != expectedResult {
		t.Fatalf("expected result %s, got %s", expectedResult, calcResp.Result)
	}

	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM calculations WHERE user_id = ? AND expression = ?", userID, "2+3*4").Scan(&count)
	if err != nil {
		t.Fatalf("failed to query calculations: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 calculation saved, got %d", count)
	}
}

func TestCalculateHandler_InvalidExpression(t *testing.T) {
	db := setupTestDB(t)

	handler := CalculateHandler(db)

	reqBody := `{"expression": "2++2"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/calculate", bytes.NewBufferString(reqBody))
	req = req.WithContext(contextWithUserID(1))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler(w, req)

	if w.Result().StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid expression, got %d", w.Result().StatusCode)
	}
}

func TestCalculateHandler_Unauthorized(t *testing.T) {
	db := setupTestDB(t)

	handler := CalculateHandler(db)

	reqBody := `{"expression": "2+2"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/calculate", bytes.NewBufferString(reqBody))
	// Без контекста с userID
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler(w, req)

	if w.Result().StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401 for unauthorized, got %d", w.Result().StatusCode)
	}
}

func TestCalculateHandler_InvalidJSON(t *testing.T) {
	db := setupTestDB(t)

	handler := CalculateHandler(db)

	reqBody := `{invalid json}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/calculate", bytes.NewBufferString(reqBody))
	req = req.WithContext(contextWithUserID(1))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler(w, req)

	if w.Result().StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid json, got %d", w.Result().StatusCode)
	}
}
