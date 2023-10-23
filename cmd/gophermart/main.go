package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
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
	r.Mount("/api", controller.Route())
	return r
}

func gracefulShutdown(srv *http.Server, done chan struct{}, logger logging.ILogger) {
	quitChannel := make(chan os.Signal, 1)
	signal.Notify(quitChannel, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quitChannel
	logger.Infof("Received signal: %v\n", sig)
	done <- struct{}{}
	time.Sleep(2 * time.Second)

	if err := srv.Shutdown(context.Background()); err != nil {
		log.Fatalf("Server shutdown failed: %v\n", err.Error())
	}
}

func main() {
	logger := logging.SetupLogger()
	conf, err := config.Read()
	if err != nil {
		logger.Fatal(err)
	}
	storage := adapters.NewPGStorage(logger, conf.DatabaseDSN)
	defer storage.Shutdown()
	crypto := adapters.CryptoProvider{Logger: logger}
	controller := usecases.NewBaseController(logger, storage, &crypto)
	r := setupServer(logger, storage, controller)
	srv := http.Server{Addr: conf.ServerAddr, Handler: r}

	done := make(chan struct{}, 1)
	jobs := make(chan entities.Job, config.WorkerJobsCapacity)
	supplier := adapters.NewSupplier(
		storage, logger, jobs, done, time.NewTicker(config.WorkerInterval).C, config.WorkerJobsCapacity,
	)
	ctx := context.Background()
	go supplier.Run(ctx)
	accrualDriver := adapters.NewAccrualClient(
		logger, conf.AccrualAddr, config.DefaultAccrualRequestTimeoutSec, config.MaxAccrualRequestAttempts,
	)
	for i := 0; i < config.WorkerPoolSize; i++ {
		worker := adapters.NewCollector(
			fmt.Sprintf("worker-%d", i),
			accrualDriver,
			storage,
			logger,
			jobs,
			done,
		)
		go worker.Run(ctx)
	}

	go func() {
		logger.Infof("Launching the server at %s\n", conf.ServerAddr)
		if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			logger.Fatal(err)
		}
	}()

	gracefulShutdown(&srv, done, logger)
}
