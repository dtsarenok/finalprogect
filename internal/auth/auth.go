package auth

import (
	"distributed-calculator/internal/models"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var jwtKey = []byte("your_secret_key_here")

var db *gorm.DB

func InitDB(database *gorm.DB) {
	db = database
}

type Claims struct {
	UserID uint `json:"user_id"`
	jwt.RegisteredClaims
}

func RegisterUser(login, password string) error {
	var user models.User
	if err := db.Where("login = ?", login).First(&user).Error; err == nil {
		return errors.New("user already exists")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user = models.User{
		Login:        login,
		PasswordHash: string(hashedPassword), // Используем PasswordHash, а не Password
	}

	return db.Create(&user).Error
}

func AuthenticateUser(login, password string) (string, error) {
	var user models.User
	if err := db.Where("login = ?", login).First(&user).Error; err != nil {
		return "", errors.New("invalid login or password")
	}

	// Используем PasswordHash для сравнения
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", errors.New("invalid login or password")
	}

	expirationTime := time.Now().Add(24 * time.Hour)

	claims := &Claims{
		UserID: uint(user.ID), // Явное приведение типа int64 -> uint
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func ParseToken(tokenStr string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil || !token.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}
