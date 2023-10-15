package entities

import (
	"context"
	"errors"

	"github.com/jmoiron/sqlx"
)

var (
	ErrLoginAlreadyTaken   = errors.New("login already taken")
	ErrOrderAlreadyCreated = errors.New("order already created")
	ErrOrderCreatedByOther = errors.New("order created by other user")
)

type Storage interface {
	CreateUser(ctx context.Context, login string, pwdhash []byte, sessionToken string) (*User, *Session, error)
	FindUser(ctx context.Context, request *UserAuthRequest) (*User, error)
	CreateSession(ctx context.Context, tx *sqlx.Tx, user *User, token string) (*Session, error)
	FindSession(ctx context.Context, token string) (*Session, error)
	CreateOrder(ctx context.Context, userID int, number uint64) (*Order, bool, error)
	FindOrder(ctx context.Context, number uint64) (*Order, error)
	FindUserOrders(ctx context.Context, userID int) ([]Order, error)
	CreateAccrual(ctx context.Context, accrual *Accrual) error
	GetBalance(ctx context.Context, userID int) (*Balance, error)
}
