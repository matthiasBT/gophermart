package repositories

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/matthiasBT/gophermart/internal/infra/config"
	"github.com/matthiasBT/gophermart/internal/infra/logging"
	"github.com/matthiasBT/gophermart/internal/server/entities"
)

type PGUserRepo struct {
	logger  logging.ILogger
	storage entities.Storage
}

func NewPGUserRepo(logger logging.ILogger, storage entities.Storage) *PGUserRepo {
	return &PGUserRepo{
		logger:  logger,
		storage: storage,
	}
}

func (r *PGUserRepo) CreateUser(
	ctx context.Context, tx entities.Tx, login string, pwdhash []byte,
) (*entities.User, error) {
	r.logger.Infof("Creating a new user: %s", login)
	var user = entities.User{}
	query := "insert into users(login, password_hash) values ($1, $2) returning *"
	if err := tx.GetContext(ctx, &user, query, login, pwdhash); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			r.logger.Infof("Login is already taken")
			return nil, entities.ErrLoginAlreadyTaken
		}
		r.logger.Errorf("Failed to create a user record: %s", err.Error())
		return nil, err
	}
	r.logger.Infof("User created: %s", login)
	return &user, nil
}

func (r *PGUserRepo) CreateSession(
	ctx context.Context, tx entities.Tx, user *entities.User, token string,
) (*entities.Session, error) {
	r.logger.Infof("Creating a session for a user: %s", user.Login)
	var session = entities.Session{}
	query := "insert into sessions(user_id, token, expires_at) values ($1, $2, $3) returning *"
	expiresAt := time.Now().Add(config.SessionTTL)
	if err := tx.GetContext(ctx, &session, query, user.ID, token, expiresAt); err != nil {
		r.logger.Errorf("Failed to create a user session: %s", err.Error())
		return nil, err
	}
	r.logger.Infof("Session created!")
	return &session, nil
}

func (r *PGUserRepo) FindUser(ctx context.Context, request *entities.UserAuthRequest) (*entities.User, error) {
	r.logger.Infof("Searching for a user: %s", request.Login)
	var user = entities.User{}
	query := "select * from users where login = $1"
	if err := r.storage.GetContext(ctx, &user, query, request.Login); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			r.logger.Infoln("User not found")
			return nil, nil
		}
		r.logger.Errorf("Failed to find the user: %s", err.Error())
		return nil, err
	}
	r.logger.Infoln("User found")
	return &user, nil
}

func (r *PGUserRepo) FindSession(ctx context.Context, token string) (*entities.Session, error) {
	r.logger.Infof("Looking for a session")
	var session = entities.Session{}
	query := "select * from sessions where token = $1"
	if err := r.storage.GetContext(ctx, &session, query, token); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			r.logger.Infoln("Session not found")
			return nil, nil
		}
		r.logger.Errorf("Failed to find the session: %s", err.Error())
		return nil, err
	}
	r.logger.Infoln("Session found")
	return &session, nil
}
