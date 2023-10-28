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
	GetContext(ctx context.Context, dest any, query string, args ...any) error
	ExecContext(ctx context.Context, query string, args ...any) error
}

type Storage interface {
	Tx(ctx context.Context) (Tx, error)
	SelectContext(ctx context.Context, dest any, query string, args ...any) error
	GetContext(ctx context.Context, dest any, query string, args ...any) error
}

type UserRepo interface {
	CreateUser(ctx context.Context, tx Tx, login string, pwdhash []byte) (*User, error)
	FindUser(ctx context.Context, request *UserAuthRequest) (*User, error)
	CreateSession(ctx context.Context, tx Tx, user *User, token string) (*Session, error)
	FindSession(ctx context.Context, token string) (*Session, error)
}

type OrderRepo interface {
	CreateOrder(ctx context.Context, userID int, number string) (*Order, bool, error)
	FindOrder(ctx context.Context, number string) (*Order, error)
	FindUserOrders(ctx context.Context, userID int) ([]Order, error)
	FetchUnprocessedOrders(ctx context.Context, limit int) ([]Order, error)
	UpdateOrderStatus(ctx context.Context, tx Tx, number string, status string) error
}

type AccrualRepo interface {
	GetBalance(ctx context.Context, userID int) (*Balance, error)
	CreateWithdrawal(ctx context.Context, withdrawal *Accrual) (*Accrual, error)
	FindUserWithdrawals(ctx context.Context, userID int) ([]Accrual, error)
	CreateAccrual(ctx context.Context, tx Tx, userID int, accrual *AccrualResponse) error
}
