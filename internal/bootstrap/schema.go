package bootstrap

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/rafaelblt/go-auth/internal/infra/postgres"
	"github.com/rafaelblt/go-auth/internal/migrate"
	"github.com/rafaelblt/go-auth/migrations"
)

func autoMigrate(dbURL string) error {
	err := migrate.RunMigrations(dbURL)
	if err != nil {
		return fmt.Errorf("run migration failed: %w", err)
	}
	return nil
}

func verifySchema(ctx context.Context, db postgres.DB) error {
	latest, err := migrations.Latest()
	if err != nil {
		return fmt.Errorf("get latest migration version failed: %w", err)
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	current, err := postgres.SchemaVersion(ctx, db)
	if err != nil {
		return fmt.Errorf("get current schema version failed: %w", err)
	}

	if current != latest {
		e := fmt.Sprintf(
			"the database scehama version (%d) is different from the latest migration (%d)",
			current, latest)
		return errors.New(e)
	}
	return nil
}
