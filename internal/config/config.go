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

	flag.StringVar(&cfg.RunAddress, "a", "localhost:8080", "Address and port to run service")
	flag.StringVar(&cfg.DatabaseURI, "d", "", "Database connection URI")
	flag.StringVar(&cfg.AccrualSystemAddress, "r", "", "Accrual system address")
	flag.Parse()

	if err := env.Parse(&cfg); err != nil {
		return nil, fmt.Errorf("failed to parse environment variables: %w", err)
	}

	return &cfg, nil
}
