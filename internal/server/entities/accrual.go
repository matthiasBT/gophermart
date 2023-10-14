package entities

import "context"

type IAccrualClient interface {
	GetAccrual(ctx context.Context, orderID int) (*AccrualResponse, error)
}
