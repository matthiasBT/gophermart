package config

import (
	"flag"

	"github.com/caarlos0/env/v9"
)

type Config struct {
	ServerAddr  string `env:"RUN_ADDRESS"`
	DatabaseDSN string `env:"DATABASE_DSN"`
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
	// TODO: add checks for empty values right before the return
	if conf.ServerAddr == "" {
		conf.ServerAddr = *flagServerAddr
	}
	if conf.DatabaseDSN == "" {
		conf.DatabaseDSN = *flagDBDSN
	}
	if conf.AccrualAddr == "" {
		conf.AccrualAddr = *flagAccrualAddr
	}
	return conf, nil
}
