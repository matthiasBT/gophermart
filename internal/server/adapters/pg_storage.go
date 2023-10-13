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
	var user = make([]entities.User, 1)
	tx, err := st.tx(ctx)
	if err != nil {
		return nil, nil, err
	}
	defer tx.Commit()
	query := "INSERT INTO users(login, password_hash) VALUES ($1, $2) RETURNING *"
	err = tx.SelectContext(ctx, &user, query, login, pwdhash)
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
	session, err := st.CreateSession(ctx, tx, &user[0], sessionToken)
	if err != nil {
		tx.Rollback()
		return nil, nil, err
	}
	return &user[0], session, nil
}

func (st *PGStorage) CreateSession(
	ctx context.Context, tx *sqlx.Tx, user *entities.User, token string,
) (*entities.Session, error) {
	st.logger.Infof("Creating a session for a user: %s", user.Login)
	var session = make([]entities.Session, 1)
	if tx == nil {
		trx, err := st.tx(ctx)
		if err != nil {
			return nil, err
		}
		tx = trx
		defer tx.Commit()
	}
	query := "INSERT INTO session(user_id, token, expires_at) VALUES ($1, $2, $3) RETURNING *"
	expiresAt := time.Now().Add(config.SessionTTL)
	if err := tx.SelectContext(ctx, &session, query, user.ID, token, expiresAt); err != nil {
		st.logger.Errorf("Failed to create a user session: %s", err.Error())
		return nil, err
	}
	st.logger.Infof("Session created!")
	return &session[0], nil
}

func (st *PGStorage) FindUser(ctx context.Context, request *entities.UserAuthRequest) (*entities.User, error) {
	st.logger.Infof("Searching for a user: %s", request.Login)
	var user = entities.User{}
	query := "SELECT * FROM users WHERE login = $1"
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

func (st *PGStorage) tx(ctx context.Context) (*sqlx.Tx, error) {
	tx, err := st.db.BeginTxx(ctx, &txOpt)
	if err != nil {
		st.logger.Errorf("Failed to open a transaction: %s", err.Error())
		return nil, err
	}
	return tx, nil
}
