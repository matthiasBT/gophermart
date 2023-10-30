package config

import (
	"flag"
	"time"

	"github.com/caarlos0/env/v9"
)

const SessionTTL = 1 * time.Hour
const MaxAccrualRequestAttempts = 5
const DefaultAccrualRequestTimeoutSec = 10

const WorkerPoolSize = 3
const WorkerJobsCapacity = 100
const WorkerInterval = 15 * time.Second

type Config struct {
	ServerAddr  string `env:"RUN_ADDRESS"`
	DatabaseDSN string `env:"DATABASE_URI"`
	AccrualAddr string `env:"ACCRUAL_SYSTEM_ADDRESS"`
}

func Read() (*Config, error) {
	conf := new(Config)
	err := env.Parse(conf)
	if err != nil {
		return nil, err
	}
	flagServerAddr := flag.String("a", "", "Server address. Usage: -a=host:port")
	flagDBDSN := flag.String("d", "", "PostgreSQL database DSN")
	flagAccrualAddr := flag.String("r", "", "Accrual system URL")
	flag.Parse()
	if conf.ServerAddr == "" {
		conf.ServerAddr = *flagServerAddr
	}
	if conf.DatabaseDSN == "" {
		conf.DatabaseDSN = *flagDBDSN
	}
	if conf.AccrualAddr == "" {
		conf.AccrualAddr = *flagAccrualAddr
	}
	if conf.ServerAddr == "" || conf.DatabaseDSN == "" || conf.AccrualAddr == "" {
		panic("Invalid server configuration")
	}
	return conf, nil
}
