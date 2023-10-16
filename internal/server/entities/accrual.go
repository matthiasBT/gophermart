package entities

import "context"

type IAccrualClient interface {
	GetAccrual(ctx context.Context, orderNumber string) (*AccrualResponse, error)
}
