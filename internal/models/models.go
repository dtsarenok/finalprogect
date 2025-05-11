package models


type User struct {
	ID           int64  
	Login        string
	PasswordHash string
}

type Calculation struct {
	ID         int64 
	UserID     int64
	Expression string
	Result     string
	CreatedAt  string
}
