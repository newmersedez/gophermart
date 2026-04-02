package config

import (
	"flag"
	"fmt"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	RunAddress           string `env:"RUN_ADDRESS"`
	DatabaseURI          string `env:"DATABASE_URI"`
	AccrualSystemAddress string `env:"ACCRUAL_SYSTEM_ADDRESS"`
}

func NewConfig() (*Config, error) {
	cfg := Config{}

	if err := env.Parse(&cfg); err != nil {
		return nil, fmt.Errorf("failed to parse environment variables: %w", err)
	}

	var runAddress string
	var databaseURI string
	var accrualSystemAddress string

	flag.StringVar(&runAddress, "a", "", "Address and port to run service")
	flag.StringVar(&databaseURI, "d", "", "Database connection URI")
	flag.StringVar(&accrualSystemAddress, "r", "", "Accrual system address")
	flag.Parse()

	if runAddress != "" {
		cfg.RunAddress = runAddress
	}
	if databaseURI != "" {
		cfg.DatabaseURI = databaseURI
	}
	if accrualSystemAddress != "" {
		cfg.AccrualSystemAddress = accrualSystemAddress
	}

	if cfg.RunAddress == "" {
		cfg.RunAddress = "localhost:8080"
	}

	return &cfg, nil
}
