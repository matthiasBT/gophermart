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
	UploadedAt time.Time `db:"uploaded_at"`
	Accrual    float32   `db:"accrual"`
}

type Accrual struct {
	ID      int     `db:"id"`
	UserID  int     `db:"user_id"`
	OrderID int     `db:"order_id"`
	Amount  float32 `db:"amount"`
}

type AccrualResponse struct {
	OrderNumber string  `json:"order"`
	Status      string  `json:"status"`
	Amount      float32 `json:"accrual"`
}
