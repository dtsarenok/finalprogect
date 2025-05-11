package main_test

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"distributed-calculator/internal/auth"
	"distributed-calculator/internal/calculator"
	"distributed-calculator/internal/storage"
)

func SetupServer(t *testing.T) (http.Handler, *sql.DB) {
	t.Helper()

	db, err := storage.NewSQLite(":memory:")
	if err != nil {
		t.Fatalf("failed to create in-memory db: %v", err)
	}

	mux := http.NewServeMux()
	mux.Handle("/api/v1/register", auth.RegisterHandler(db))
	mux.Handle("/api/v1/login", auth.LoginHandler(db))

	jwtMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tokenStr := r.Header.Get("Authorization")
			if tokenStr == "" {
				http.Error(w, "missing token", http.StatusUnauthorized)
				return
			}
			const prefix = "Bearer "
			if len(tokenStr) <= len(prefix) || tokenStr[:len(prefix)] != prefix {
				http.Error(w, "invalid token format", http.StatusUnauthorized)
				return
			}
			tokenStr = tokenStr[len(prefix):]

			claims, err := auth.ParseToken(tokenStr)
			if err != nil {
				http.Error(w, "invalid token", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), auth.UserIDKey, claims.UserID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}

	mux.Handle("/api/v1/calculate", jwtMiddleware(calculator.CalculateHandler(db)))

	return mux, db
}

func TestIntegration_FullFlow(t *testing.T) {
	handler, _ := SetupServer(t)

	registerPayload := `{"login":"user1","password":"pass123"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/register", bytes.NewBufferString(registerPayload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Result().StatusCode != http.StatusOK {
		t.Fatalf("register failed: status %d", w.Result().StatusCode)
	}

	loginPayload := `{"login":"user1","password":"pass123"}`
	req = httptest.NewRequest(http.MethodPost, "/api/v1/login", bytes.NewBufferString(loginPayload))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Result().StatusCode != http.StatusOK {
		t.Fatalf("login failed: status %d", w.Result().StatusCode)
	}

	var loginResp map[string]string
	if err := json.NewDecoder(w.Result().Body).Decode(&loginResp); err != nil {
		t.Fatalf("failed to decode login response: %v", err)
	}
	token, ok := loginResp["token"]
	if !ok || token == "" {
		t.Fatal("token not found in login response")
	}

	calcPayload := `{"expression":"2+2*2"}`
	req = httptest.NewRequest(http.MethodPost, "/api/v1/calculate", bytes.NewBufferString(calcPayload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Result().StatusCode != http.StatusOK {
		t.Fatalf("calculate failed: status %d", w.Result().StatusCode)
	}

	var calcResp struct {
		Result string `json:"result"`
	}
	if err := json.NewDecoder(w.Result().Body).Decode(&calcResp); err != nil {
		t.Fatalf("failed to decode calculate response: %v", err)
	}

	expected := "6"
	if calcResp.Result != expected {
		t.Fatalf("expected result %s, got %s", expected, calcResp.Result)
	}
}

func TestIntegration_UnauthorizedCalculate(t *testing.T) {
	handler, _ := SetupServer(t)

	calcPayload := `{"expression":"2+2"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/calculate", bytes.NewBufferString(calcPayload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Result().StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401 Unauthorized, got %d", w.Result().StatusCode)
	}
}

func TestIntegration_InvalidLogin(t *testing.T) {
	handler, _ := SetupServer(t)

	loginPayload := `{"login":"nonexistent","password":"pass"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/login", bytes.NewBufferString(loginPayload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Result().StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401 Unauthorized for invalid login, got %d", w.Result().StatusCode)
	}
}

func TestIntegration_InvalidRegister(t *testing.T) {
	handler, _ := SetupServer(t)

	registerPayload := `{"login":"","password":"pass"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/register", bytes.NewBufferString(registerPayload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Result().StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400 Bad Request for empty login, got %d", w.Result().StatusCode)
	}
