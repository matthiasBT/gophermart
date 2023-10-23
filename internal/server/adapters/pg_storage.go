package adapters

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jmoiron/sqlx"
	"github.com/matthiasBT/gophermart/internal/infra/config"
	"github.com/matthiasBT/gophermart/internal/infra/logging"
	"github.com/matthiasBT/gophermart/internal/infra/migrations"
	"github.com/matthiasBT/gophermart/internal/server/entities"
)

var txOpt = sql.TxOptions{
	Isolation: sql.LevelReadCommitted,
	ReadOnly:  false,
}

type PGTx struct {
	tx *sqlx.Tx
}

func (pgtx *PGTx) Commit() error {
	return pgtx.tx.Commit()
}

func (pgtx *PGTx) Rollback() error {
	return pgtx.tx.Rollback()
}

func (pgtx *PGTx) GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	return pgtx.tx.GetContext(ctx, dest, query, args...)
}

func (pgtx *PGTx) ExecContext(ctx context.Context, query string, args ...any) error {
	_, err := pgtx.tx.ExecContext(ctx, query, args...)
	return err
}

type PGStorage struct {
	logger logging.ILogger
	db     *sqlx.DB
}

func NewPGStorage(logger logging.ILogger, dsn string) *PGStorage {
	db := sqlx.MustOpen("pgx", dsn)
	migrations.Migrate(db)
	return &PGStorage{logger: logger, db: db}
}

func (st *PGStorage) Shutdown() {
	if err := st.db.Close(); err != nil {
		st.logger.Errorf("Failed to cleanup the DB resources: %v", err)
	}
}

func (st *PGStorage) CreateUser(
	ctx context.Context, tx entities.Tx, login string, pwdhash []byte,
) (*entities.User, error) {
	st.logger.Infof("Creating a new user: %s", login)
	var user = entities.User{}
	query := "insert into users(login, password_hash) values ($1, $2) returning *"
	if err := tx.GetContext(ctx, &user, query, login, pwdhash); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			st.logger.Infof("Login is already taken")
			return nil, entities.ErrLoginAlreadyTaken
		}
		st.logger.Errorf("Failed to create a user record: %s", err.Error())
		return nil, err
	}
	st.logger.Infof("User created: %s", login)
	return &user, nil
}

func (st *PGStorage) CreateSession(
	ctx context.Context, tx entities.Tx, user *entities.User, token string,
) (*entities.Session, error) {
	st.logger.Infof("Creating a session for a user: %s", user.Login)
	var session = entities.Session{}
	query := "insert into sessions(user_id, token, expires_at) values ($1, $2, $3) returning *"
	expiresAt := time.Now().Add(config.SessionTTL)
	if err := tx.GetContext(ctx, &session, query, user.ID, token, expiresAt); err != nil {
		st.logger.Errorf("Failed to create a user session: %s", err.Error())
		return nil, err
	}
	st.logger.Infof("Session created!")
	return &session, nil
}

func (st *PGStorage) FindUser(ctx context.Context, request *entities.UserAuthRequest) (*entities.User, error) {
	st.logger.Infof("Searching for a user: %s", request.Login)
	var user = entities.User{}
	query := "select * from users where login = $1"
	if err := st.db.GetContext(ctx, &user, query, request.Login); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			st.logger.Infoln("User not found")
			return nil, nil
		}
		st.logger.Errorf("Failed to find the user: %s", err.Error())
		return nil, err
	}
	st.logger.Infoln("User found")
	return &user, nil
}

func (st *PGStorage) FindSession(ctx context.Context, token string) (*entities.Session, error) {
	st.logger.Infof("Looking for a session")
	var session = entities.Session{}
	query := "select * from sessions where token = $1"
	if err := st.db.GetContext(ctx, &session, query, token); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			st.logger.Infoln("Session not found")
			return nil, nil
		}
		st.logger.Errorf("Failed to find the session: %s", err.Error())
		return nil, err
	}
	st.logger.Infoln("Session found")
	return &session, nil
}

func (st *PGStorage) CreateOrder(ctx context.Context, userID int, number string) (*entities.Order, bool, error) {
	st.logger.Infof("Creating order %s for user %d", number, userID)
	order, err := st.FindOrder(ctx, number)
	if err != nil {
		return nil, false, err
	}
	if order != nil {
		return order, true, nil
	}
	var result = entities.Order{}
	query := "insert into orders(user_id, number, status, uploaded_at) values ($1, $2, $3, $4) returning *"
	if err := st.db.GetContext(ctx, &result, query, userID, number, "NEW", time.Now()); err != nil {
		st.logger.Errorf("Failed to create an order: %s", err.Error())
		return nil, false, err
	}
	st.logger.Infof("Order created!")
	return &result, false, nil
}

func (st *PGStorage) FindOrder(ctx context.Context, number string) (*entities.Order, error) {
	st.logger.Infof("Searching for an order: %d", number)
	var order = entities.Order{}
	query := "select * from orders where number = $1"
	if err := st.db.GetContext(ctx, &order, query, number); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			st.logger.Infoln("Order not found")
			return nil, nil
		}
		st.logger.Errorf("Failed to find the order: %s", err.Error())
		return nil, err
	}
	st.logger.Infoln("Order found")
	return &order, nil
}

func (st *PGStorage) FindUserOrders(ctx context.Context, userID int) ([]entities.Order, error) {
	st.logger.Infof("Searching for user's orders: %d", userID)
	var orders []entities.Order
	query := `
		select o.*, coalesce(a.amount, 0) as "accrual"
		from orders o
		left join accruals a on o.number = a.order_number
		where o.user_id = $1
		order by uploaded_at
	`
	if err := st.db.SelectContext(ctx, &orders, query, userID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			st.logger.Infoln("Orders not found")
			return nil, nil
		}
		st.logger.Errorf("Failed to find the orders: %s", err.Error())
		return nil, err
	}
	st.logger.Infoln("Orders found")
	return orders, nil
}

func (st *PGStorage) CreateAccrual(ctx context.Context, tx entities.Tx, userID int, accrual *entities.AccrualResponse) error {
	st.logger.Infof(
		"Creating accrual. User: %d, order: %d, amount: %f", userID, accrual.OrderNumber, accrual.Amount,
	)
	query := `
		insert into accruals(user_id, order_number, amount)
		values ($1, $2, $3)
		on conflict (user_id, order_number) where processed_at is null
		do update set amount = EXCLUDED.amount
	`
	if err := tx.ExecContext(ctx, query, userID, accrual.OrderNumber, accrual.Amount); err != nil {
		st.logger.Errorf("Failed to create accrual: %v", err)
		return err
	}
	st.logger.Infof("Accrual created!")
	return nil
}

func (st *PGStorage) GetBalance(ctx context.Context, userID int) (*entities.Balance, error) {
	st.logger.Infof("Calculating user balance: %d", userID)
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
	if err := st.db.GetContext(ctx, &balance, query, userID); err != nil {
		st.logger.Errorf("Failed to calculate balance: %s", err.Error())
		return nil, err
	}
	st.logger.Infoln("Balance calculated")
	return &balance, nil
}

func (st *PGStorage) CreateWithdrawal(
	ctx context.Context, withdrawal *entities.Accrual,
) (*entities.Accrual, error) {
	st.logger.Infof(
		"Creating withdrawal for user: %d, order: %d, amount: %f",
		withdrawal.UserID,
		withdrawal.OrderNumber,
		withdrawal.Amount,
	)
	query := "insert into accruals(user_id, order_number, amount, processed_at) values ($1, $2, $3, $4) returning *"
	var res = entities.Accrual{}
	if err := st.db.GetContext(
		ctx, &res, query, withdrawal.UserID, withdrawal.OrderNumber, -withdrawal.Amount, time.Now(),
	); err != nil {
		st.logger.Errorf("Failed to create withdrawal: %s", err.Error())
		return nil, err
	}
	st.logger.Infof("Withdrawal created!")
	return &res, nil
}

func (st *PGStorage) FindUserWithdrawals(ctx context.Context, userID int) ([]entities.Accrual, error) {
	st.logger.Infof("Getting user withdrawals: %d", userID)
	var withdrawals []entities.Accrual
	query := `
		select id, user_id, order_number, processed_at, -1 * amount as amount
		from accruals
		where user_id = $1
		and amount < 0
		order by processed_at
	`
	if err := st.db.SelectContext(ctx, &withdrawals, query, userID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			st.logger.Infoln("Withdrawals not found")
			return nil, nil
		}
		st.logger.Errorf("Failed to find the withdrawals: %s", err.Error())
		return nil, err
	}
	st.logger.Infoln("Withdrawals found")
	return withdrawals, nil
}

func (st *PGStorage) Tx(ctx context.Context) (entities.Tx, error) {
	tx, err := st.db.BeginTxx(ctx, &txOpt)
	if err != nil {
		st.logger.Errorf("Failed to open a transaction: %s", err.Error())
		return nil, err
	}
	trans := PGTx{tx: tx}
	return &trans, nil
}

func (st *PGStorage) FetchUnprocessedOrders(ctx context.Context, limit int) ([]entities.Order, error) {
	st.logger.Infof("Getting %d unprocessed orders", limit)
	var orders []entities.Order
	query := `
		select *
		from orders
		where status not in ('INVALID', 'PROCESSED')
		order by id
		limit $1
	`
	if err := st.db.SelectContext(ctx, &orders, query, limit); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			st.logger.Infoln("Orders not found")
			return nil, nil
		}
		st.logger.Errorf("Failed to find the orders: %v", err)
		return nil, err
	}
	st.logger.Infoln("Unprocessed orders found")
	return orders, nil
}

func (st *PGStorage) UpdateOrderStatus(ctx context.Context, tx entities.Tx, number string, status string) error {
	st.logger.Infof("Updating order %s status: %s", number, status)
	query := "update orders set status = $1 where number = $2"
	if err := tx.ExecContext(ctx, query, status, number); err != nil {
		st.logger.Errorf("Failed to update order: %v", err)
		return err
	}
	st.logger.Infof("Order status updated!")
	return nil
}
