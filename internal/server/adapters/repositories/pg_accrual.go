package repositories

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/matthiasBT/gophermart/internal/infra/logging"
	"github.com/matthiasBT/gophermart/internal/server/entities"
)

type PGAccrualRepo struct {
	logger  logging.ILogger
	storage entities.Storage
}

func NewPGAccrualRepo(logger logging.ILogger, storage entities.Storage) *PGAccrualRepo {
	return &PGAccrualRepo{
		logger:  logger,
		storage: storage,
	}
}

func (a *PGAccrualRepo) CreateAccrual(ctx context.Context, tx entities.Tx, userID int, accrual *entities.AccrualResponse) error {
	a.logger.Infof(
		"Creating accrual. User: %d, order: %d, amount: %f", userID, accrual.OrderNumber, accrual.Amount,
	)
	query := `
		insert into accruals(user_id, order_number, amount)
		values ($1, $2, $3)
		on conflict (user_id, order_number) where processed_at is null
		do update set amount = EXCLUDED.amount
	`
	if err := tx.ExecContext(ctx, query, userID, accrual.OrderNumber, accrual.Amount); err != nil {
		a.logger.Errorf("Failed to create accrual: %v", err)
		return err
	}
	a.logger.Infof("Accrual created!")
	return nil
}

func (a *PGAccrualRepo) GetBalance(ctx context.Context, userID int) (*entities.Balance, error) {
	a.logger.Infof("Calculating user balance: %d", userID)
	var balance = entities.Balance{}
	query := `
		with user_accr as (
			select a.amount
			from accruals a
			where a.user_id = $1
		)
		select
			(select coalesce(sum(amount), 0.0) from user_accr) current,
			(select -1 * coalesce(sum(amount), 0.0) from user_accr where amount < 0) withdrawn
	`
	if err := a.storage.GetContext(ctx, &balance, query, userID); err != nil {
		a.logger.Errorf("Failed to calculate balance: %s", err.Error())
		return nil, err
	}
	a.logger.Infoln("Balance calculated")
	return &balance, nil
}

func (a *PGAccrualRepo) CreateWithdrawal(
	ctx context.Context, withdrawal *entities.Accrual,
) (*entities.Accrual, error) {
	a.logger.Infof(
		"Creating withdrawal for user: %d, order: %d, amount: %f",
		withdrawal.UserID,
		withdrawal.OrderNumber,
		withdrawal.Amount,
	)
	query := "insert into accruals(user_id, order_number, amount, processed_at) values ($1, $2, $3, $4) returning *"
	var res = entities.Accrual{}
	if err := a.storage.GetContext(
		ctx, &res, query, withdrawal.UserID, withdrawal.OrderNumber, -withdrawal.Amount, time.Now(),
	); err != nil {
		a.logger.Errorf("Failed to create withdrawal: %s", err.Error())
		return nil, err
	}
	a.logger.Infof("Withdrawal created!")
	return &res, nil
}

func (a *PGAccrualRepo) FindUserWithdrawals(ctx context.Context, userID int) ([]entities.Accrual, error) {
	a.logger.Infof("Getting user withdrawals: %d", userID)
	var withdrawals []entities.Accrual
	query := `
		select id, user_id, order_number, processed_at, -1 * amount as amount
		from accruals
		where user_id = $1
		and amount < 0
		order by processed_at
	`
	if err := a.storage.SelectContext(ctx, &withdrawals, query, userID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			a.logger.Infoln("Withdrawals not found")
			return nil, nil
		}
		a.logger.Errorf("Failed to find the withdrawals: %s", err.Error())
		return nil, err
	}
	a.logger.Infoln("Withdrawals found")
	return withdrawals, nil
}
