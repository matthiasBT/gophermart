package entities

import "time"

type ContextKey struct {
	Key string
}

type UserAuthRequest struct {
	Login    string
	Password string
}

type User struct {
	ID           int    `db:"id"`
	Login        string `db:"login"`
	PasswordHash []byte `db:"password_hash"`
}

type Session struct {
	ID        int       `db:"id"`
	UserID    int       `db:"user_id"`
	Token     string    `db:"token"`
	ExpiresAt time.Time `db:"expires_at"`
}

type Order struct {
	ID         int       `db:"id"`
	UserID     int       `db:"user_id"`
	Number     uint64    `db:"number"`
	Status     string    `db:"status"`
	Accrual    float32   `db:"accrual"` // todo: need to join another table to get it
	UploadedAt time.Time `db:"uploaded_at"`
}

type Accrual struct {
	OrderID int     `db:"order_id"`
	Status  string  `db:"status"`
	Accrual float32 `db:"accrual"`
}

type AccrualResponse struct {
	OrderNumber string  `json:"order"`
	Status      string  `json:"status"`
	Accrual     float32 `json:"accrual"`
}
