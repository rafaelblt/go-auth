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
	Address         string
	DatabaseURL     string
	BcryptCost      int
	AccessTokenTTL  time.Duration
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

func (c Config) Validate() error {
	if c.Address == "" {
		return errors.New("address empty")
	}
	if c.DatabaseURL == "" {
		return errors.New("database url empty")
	}
	if c.BcryptCost <= 0 {
		return errors.New("bcrypt cost zero or negative")
	}
	if c.AccessTokenTTL <= 0 {
		return errors.New("access token ttl zero or negative")
	}
	if c.RefreshTokenTTL <= 0 {
		return errors.New("refresh token ttl zero or negative")
	}
	return nil
}
