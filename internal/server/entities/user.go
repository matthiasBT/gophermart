package entities

import "time"

type UserCreateRequest struct {
	Login    string
	Password string
}

type User struct {
	ID           int    `db:"id"`
	Login        string `db:"login"`
	PasswordHash string `db:"password_hash"`
}

type Session struct {
	ID        int       `db:"id"`
	UserID    int       `db:"user_id"`
	Token     string    `db:"token"`
	ExpiresAt time.Time `db:"expires_at"`
}
