package entities

import (
	"context"
	"errors"
)

var (
	ErrLoginAlreadyTaken = errors.New("login already taken")
)

type Tx interface {
	Commit() error
	Rollback() error
	GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
}

type Storage interface {
	Tx(ctx context.Context) (Tx, error)

	CreateUser(ctx context.Context, tx Tx, login string, pwdhash []byte) (*User, error)
	FindUser(ctx context.Context, request *UserAuthRequest) (*User, error)
	CreateSession(ctx context.Context, tx Tx, user *User, token string) (*Session, error)
	FindSession(ctx context.Context, token string) (*Session, error)

	CreateOrder(ctx context.Context, userID int, number string) (*Order, bool, error)
	FindOrder(ctx context.Context, number string) (*Order, error)
	FindUserOrders(ctx context.Context, userID int) ([]Order, error)

	CreateAccrual(ctx context.Context, accrual *Accrual) error
	GetBalance(ctx context.Context, userID int) (*Balance, error)
	CreateWithdrawal(ctx context.Context, withdrawal *Withdrawal) (*Withdrawal, error)
	FindUserWithdrawals(ctx context.Context, userID int) ([]Withdrawal, error)
}
