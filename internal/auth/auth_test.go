package auth

import (
	"errors"
	"testing"
	"time"

	"distributed-calculator/internal/models"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	err = db.AutoMigrate(&models.User{})
	if err != nil {
		t.Fatalf("failed to migrate test db: %v", err)
	}
	return db
}

func TestRegisterUser(t *testing.T) {
	db := setupTestDB(t)
	InitDB(db)

	// Регистрируем нового пользователя
	err := RegisterUser("testuser", "password123")
	if err != nil {
		t.Fatalf("unexpected error on register: %v", err)
	}

	// Попытка зарегистрировать того же пользователя должна вернуть ошибку
	err = RegisterUser("testuser", "password123")
	if !errors.Is(err, errors.New("user already exists")) && err == nil {
		t.Fatalf("expected 'user already exists' error, got %v", err)
	}
}

func TestAuthenticateUser(t *testing.T) {
	db := setupTestDB(t)
	InitDB(db)

	// Сначала регистрируем пользователя
	err := RegisterUser("authuser", "secret")
	if err != nil {
		t.Fatalf("failed to register user: %v", err)
	}

	// Правильный логин и пароль
	token, err := AuthenticateUser("authuser", "secret")
	if err != nil {
		t.Fatalf("authentication failed: %v", err)
	}
	if token == "" {
		t.Fatalf("expected token, got empty string")
	}

	// Неправильный пароль
	_, err = AuthenticateUser("authuser", "wrongpassword")
	if err == nil {
		t.Fatalf("expected error on wrong password, got nil")
	}

	// Неправильный логин
	_, err = AuthenticateUser("wronguser", "secret")
	if err == nil {
		t.Fatalf("expected error on wrong login, got nil")
	}
}

func TestParseToken(t *testing.T) {
	db := setupTestDB(t)
	InitDB(db)

	err := RegisterUser("tokenuser", "pass")
	if err != nil {
		t.Fatalf("failed to register user: %v", err)
	}

	token, err := AuthenticateUser("tokenuser", "pass")
	if err != nil {
		t.Fatalf("failed to authenticate user: %v", err)
	}

	claims, err := ParseToken(token)
	if err != nil {
		t.Fatalf("failed to parse token: %v", err)
	}

	if claims.UserID == 0 {
		t.Fatalf("expected valid user ID in claims, got 0")
	}

	// Проверка истечения срока действия токена
	if claims.ExpiresAt.Time.Before(time.Now()) {
		t.Fatalf("token already expired")
	}
}
