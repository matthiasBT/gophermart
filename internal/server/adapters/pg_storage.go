package adapters

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
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

func (st *PGStorage) Tx(ctx context.Context) (entities.Tx, error) {
	tx, err := st.db.BeginTxx(ctx, &txOpt)
	if err != nil {
		st.logger.Errorf("Failed to open a transaction: %s", err.Error())
		return nil, err
	}
	trans := PGTx{tx: tx}
	return &trans, nil
}

func (st *PGStorage) SelectContext(ctx context.Context, dest any, query string, args ...any) error {
	return st.db.SelectContext(ctx, dest, query, args...)
}

func (st *PGStorage) GetContext(ctx context.Context, dest any, query string, args ...any) error {
	return st.db.GetContext(ctx, dest, query, args...)
}
