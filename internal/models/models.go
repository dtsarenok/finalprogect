package models

// User представляет пользователя системы
type User struct {
	ID           int64  // Уникальный идентификатор пользователя
	Login        string // Логин пользователя
	PasswordHash string // Хеш пароля пользователя
}

// Calculation представляет вычисление пользователя
type Calculation struct {
	ID         int64  // Уникальный идентификатор вычисления
	UserID     int64  // Идентификатор пользователя, которому принадлежит вычисление
	Expression string // Выражение для вычисления
	Result     string // Результат вычисления
	CreatedAt  string // Время создания (формат можно уточнить)
}
