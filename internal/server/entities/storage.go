package entities

import (
	"context"
	"errors"
)

var (
	ErrLoginAlreadyTaken = errors.New("unknown metric")
)

type Storage interface {
	CreateUser(ctx context.Context, request *UserCreateRequest) (*User, error)
}
