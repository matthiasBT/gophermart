package main

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/matthiasBT/gophermart/internal/infra/auth"
	"github.com/matthiasBT/gophermart/internal/infra/config"
	"github.com/matthiasBT/gophermart/internal/infra/logging"
	"github.com/matthiasBT/gophermart/internal/server/adapters"
	"github.com/matthiasBT/gophermart/internal/server/entities"
	"github.com/matthiasBT/gophermart/internal/server/usecases"
)

func setupServer(logger logging.ILogger, storage entities.Storage, controller *usecases.BaseController) *chi.Mux {
	r := chi.NewRouter()
	r.Use(logging.Middleware(logger))
	r.Use(auth.Middleware(logger, storage))
	r.Mount("/", controller.Route())
	return r
}

// TODO: посмотреть вебинар или почитать статьи по graceful shutdown, см. каналы Пачки
func main() {
	logger := logging.SetupLogger()
	conf, err := config.Read()
	if err != nil {
		logger.Fatal(err)
	}
	logger.Infof(
		"Config. Server address: %s. Database DSN: %s. Accrual system URL: %s",
		conf.ServerAddr,
		conf.DatabaseDSN,
		conf.AccrualAddr,
	)
	db := sqlx.MustOpen("pgx", conf.DatabaseDSN)
	defer db.Close()
	storage := adapters.NewPGStorage(logger, db)
	crypto := adapters.CryptoProvider{Logger: logger}
	accrualDriver := adapters.NewAccrualClient(
		logger, conf.AccrualAddr, config.DefaultAccrualRequestTimeoutSec, config.MaxAccrualRequestAttempts,
	)
	controller := usecases.NewBaseController(logger, storage, &crypto, accrualDriver)
	r := setupServer(logger, storage, controller)

	srv := http.Server{Addr: conf.ServerAddr, Handler: r}
	logger.Infof("Launching the server at %s\n", conf.ServerAddr)
	if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		logger.Fatal(err)
	}
}
