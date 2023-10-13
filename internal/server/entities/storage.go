package entities

import (
	"context"
	"errors"

	"github.com/jmoiron/sqlx"
)

var (
	ErrLoginAlreadyTaken = errors.New("unknown metric")
)

type Storage interface {
	CreateUser(ctx context.Context, request *UserCreateRequest, sessionToken string) (*User, *Session, error)
	CreateSession(ctx context.Context, tx *sqlx.Tx, user *User, token string) (*Session, error)
}
