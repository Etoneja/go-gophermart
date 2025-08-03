package config

import (
	"flag"
	"fmt"
	"os"
	"time"
)

type Config struct {
	Debug                bool
	ServerAddress        string
	DatabaseURL          string
	JWTSecret            string
	AccrualSystemAddress string
	WorkerPoolSize       int
	WorkerInterval       time.Duration
}

func LoadConfig() (*Config, error) {
	cfg := &Config{}

	flag.BoolVar(&cfg.Debug, "debug", false, "Enable debug mode")
	flag.StringVar(&cfg.ServerAddress, "a", ":8080", "Server address to listen on")
	flag.StringVar(&cfg.DatabaseURL, "d", "", "Database connection URL")
	flag.StringVar(&cfg.JWTSecret, "jwt-secret", "default-secret", "JWT secret key")
	flag.StringVar(&cfg.AccrualSystemAddress, "r", "", "Accrual System API base URL")
	flag.IntVar(&cfg.WorkerPoolSize, "worker-pool-size", 5, "Worker pool size")
	flag.DurationVar(&cfg.WorkerInterval, "worker-interval", 5*time.Second, "Worker processing interval")
	flag.Parse()

	if envServerAddress := os.Getenv("RUN_ADDRESS"); envServerAddress != "" {
		cfg.ServerAddress = envServerAddress
	}
	if envDatabaseURL := os.Getenv("DATABASE_URI"); envDatabaseURL != "" {
		cfg.DatabaseURL = envDatabaseURL
	}
	if envJWTSecret := os.Getenv("JWT_SECRET"); envJWTSecret != "" {
		cfg.JWTSecret = envJWTSecret
	}
	if envAccrualSystemAddress := os.Getenv("ACCRUAL_SYSTEM_ADDRESS"); envAccrualSystemAddress != "" {
		cfg.AccrualSystemAddress = envAccrualSystemAddress
	}

	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("database URL is required")
	}

	if cfg.AccrualSystemAddress == "" {
		return nil, fmt.Errorf("accrual system address is required")
	}

	return cfg, nil
}
