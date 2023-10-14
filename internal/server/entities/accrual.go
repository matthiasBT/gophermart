package entities

import "context"

type IAccrualClient interface {
	GetAccrual(ctx context.Context, orderNumber uint64) (*AccrualResponse, error)
}
