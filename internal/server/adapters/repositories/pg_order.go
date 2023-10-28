package repositories

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/matthiasBT/gophermart/internal/infra/logging"
	"github.com/matthiasBT/gophermart/internal/server/entities"
)

type PGOrderRepo struct {
	logger  logging.ILogger
	storage entities.Storage
}

func NewPGOrderRepo(logger logging.ILogger, storage entities.Storage) *PGOrderRepo {
	return &PGOrderRepo{
		logger:  logger,
		storage: storage,
	}
}

func (o *PGOrderRepo) CreateOrder(ctx context.Context, userID int, number string) (*entities.Order, bool, error) {
	o.logger.Infof("Creating order %s for user %d", number, userID)
	order, err := o.FindOrder(ctx, number)
	if err != nil {
		return nil, false, err
	}
	if order != nil {
		return order, true, nil
	}
	var result = entities.Order{}
	query := "insert into orders(user_id, number, status, uploaded_at) values ($1, $2, $3, $4) returning *"
	if err := o.storage.GetContext(ctx, &result, query, userID, number, "NEW", time.Now()); err != nil {
		o.logger.Errorf("Failed to create an order: %s", err.Error())
		return nil, false, err
	}
	o.logger.Infof("Order created!")
	return &result, false, nil
}

func (o *PGOrderRepo) FindOrder(ctx context.Context, number string) (*entities.Order, error) {
	o.logger.Infof("Searching for an order: %d", number)
	var order = entities.Order{}
	query := "select * from orders where number = $1"
	if err := o.storage.GetContext(ctx, &order, query, number); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			o.logger.Infoln("Order not found")
			return nil, nil
		}
		o.logger.Errorf("Failed to find the order: %s", err.Error())
		return nil, err
	}
	o.logger.Infoln("Order found")
	return &order, nil
}

func (o *PGOrderRepo) FindUserOrders(ctx context.Context, userID int) ([]entities.Order, error) {
	o.logger.Infof("Searching for user's orders: %d", userID)
	var orders []entities.Order
	query := `
		select o.*, coalesce(a.amount, 0) as "accrual"
		from orders o
		left join accruals a on o.number = a.order_number
		where o.user_id = $1
		order by uploaded_at
	`
	if err := o.storage.SelectContext(ctx, &orders, query, userID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			o.logger.Infoln("Orders not found")
			return nil, nil
		}
		o.logger.Errorf("Failed to find the orders: %s", err.Error())
		return nil, err
	}
	o.logger.Infoln("Orders found")
	return orders, nil
}

func (o *PGOrderRepo) FetchUnprocessedOrders(ctx context.Context, limit int) ([]entities.Order, error) {
	o.logger.Infof("Getting %d unprocessed orders", limit)
	var orders []entities.Order
	query := "select * from orders where status not in ('INVALID', 'PROCESSED') order by id limit $1"
	if err := o.storage.SelectContext(ctx, &orders, query, limit); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			o.logger.Infoln("Orders not found")
			return nil, nil
		}
		o.logger.Errorf("Failed to find the orders: %v", err)
		return nil, err
	}
	o.logger.Infoln("Unprocessed orders found")
	return orders, nil
}

func (o *PGOrderRepo) UpdateOrderStatus(ctx context.Context, tx entities.Tx, number string, status string) error {
	o.logger.Infof("Updating order %s status: %s", number, status)
	query := "update orders set status = $1 where number = $2"
	if err := tx.ExecContext(ctx, query, status, number); err != nil {
		o.logger.Errorf("Failed to update order: %v", err)
		return err
	}
	o.logger.Infof("Order status updated!")
	return nil
}
