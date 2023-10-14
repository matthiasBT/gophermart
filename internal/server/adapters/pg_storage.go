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

type PGStorage struct {
	logger logging.ILogger
	db     *sqlx.DB
}

func NewPGStorage(logger logging.ILogger, db *sqlx.DB) *PGStorage {
	migrations.Migrate(db)
	return &PGStorage{logger: logger, db: db}
}

func (st *PGStorage) CreateUser(
	ctx context.Context, login string, pwdhash []byte, sessionToken string,
) (*entities.User, *entities.Session, error) {
	st.logger.Infof("Creating a new user: %s", login)
	var user = entities.User{}
	tx, err := st.tx(ctx)
	if err != nil {
		return nil, nil, err
	}
	defer tx.Commit()
	query := "insert into users(login, password_hash) values ($1, $2) returning *"
	err = tx.GetContext(ctx, &user, query, login, pwdhash)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			st.logger.Infof("Login is already taken")
			return nil, nil, entities.ErrLoginAlreadyTaken
		}
		st.logger.Errorf("Failed to create a user record: %s", err.Error())
		return nil, nil, err
	}
	st.logger.Infof("User created: %s", login)
	session, err := st.CreateSession(ctx, tx, &user, sessionToken)
	if err != nil {
		tx.Rollback()
		return nil, nil, err
	}
	return &user, session, nil
}

func (st *PGStorage) CreateSession(
	ctx context.Context, tx *sqlx.Tx, user *entities.User, token string,
) (*entities.Session, error) {
	st.logger.Infof("Creating a session for a user: %s", user.Login)
	var session = entities.Session{}
	if tx == nil {
		trx, err := st.tx(ctx)
		if err != nil {
			return nil, err
		}
		tx = trx
		defer tx.Commit()
	}
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

func (st *PGStorage) CreateOrder(ctx context.Context, userID int, number uint64) (*entities.Order, bool, error) {
	order, err := st.FindOrder(ctx, number)
	if err != nil {
		return nil, false, err
	}
	if order != nil {
		return order, true, nil
	}
	var result = entities.Order{}
	query := "insert into orders(user_id, number, status, uploaded_at) values ($1, $2, $3, $4) returning *"
	uploadedAt := time.Now()
	if err := st.db.GetContext(ctx, &result, query, userID, number, "NEW", uploadedAt); err != nil {
		st.logger.Errorf("Failed to create an order: %s", err.Error())
		return nil, false, err
	}
	st.logger.Infof("Order created!")
	return &result, false, nil
}

func (st *PGStorage) FindOrder(ctx context.Context, number uint64) (*entities.Order, error) {
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
	query := "select * from orders where user_id = $1 order by uploaded_at"
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

func (st *PGStorage) tx(ctx context.Context) (*sqlx.Tx, error) {
	tx, err := st.db.BeginTxx(ctx, &txOpt)
	if err != nil {
		st.logger.Errorf("Failed to open a transaction: %s", err.Error())
		return nil, err
	}
	return tx, nil
}
