package bootstrap

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"
)

const EnvDatabaseURL = "DATABASE_URL"

type Config struct {
	DatabaseURL     string
	BcryptCost      int
	RefreshTokenTTL time.Duration
}

func LoadConfig() (Config, error) {
	err := godotenv.Load()
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return Config{}, fmt.Errorf("godotenv load failed: %w", err)
	}

	cfg := Config{}

	dbURL, ok := os.LookupEnv(EnvDatabaseURL)
	if !ok {
		return Config{}, fmt.Errorf("the env var '%s' is required", EnvDatabaseURL)
	}
	cfg.DatabaseURL = dbURL

	cfg.BcryptCost = 8
	cfg.RefreshTokenTTL = time.Hour * 24 * 7

	return cfg, nil
}
