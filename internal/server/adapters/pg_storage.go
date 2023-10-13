package adapters

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jmoiron/sqlx"
	"github.com/matthiasBT/gophermart/internal/infra/logging"
	"github.com/matthiasBT/gophermart/internal/infra/migrations"
	"github.com/matthiasBT/gophermart/internal/server/entities"
)

type PGStorage struct {
	logger logging.ILogger
	db     *sqlx.DB
	crypto entities.ICryptoProvider
}

func NewPGStorage(logger logging.ILogger, db *sqlx.DB, crypto entities.ICryptoProvider) *PGStorage {
	migrations.Migrate(db)
	return &PGStorage{logger: logger, db: db, crypto: crypto}
}

func (st *PGStorage) CreateUser(ctx context.Context, userReq *entities.UserCreateRequest) (*entities.User, error) {
	pwdhash, err := st.crypto.HashPassword(userReq.Password)
	if err != nil {
		return nil, err
	}
	var user = make([]entities.User, 1)
	txOpt := sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
		ReadOnly:  false,
	}
	tx, err := st.db.BeginTxx(ctx, &txOpt)
	if err != nil {
		st.logger.Errorf("Failed to open a transaction: %s", err.Error())
		return nil, err
	}
	defer tx.Commit()
	err = tx.SelectContext(
		ctx, &user, "INSERT INTO users(login, password_hash) VALUES ($1, $2) RETURNING *", userReq.Login, pwdhash,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return nil, entities.ErrLoginAlreadyTaken
		}
		st.logger.Errorf("Failed to create a user record: %s", err.Error())
		return nil, err
	}
	return &user[0], nil
}
