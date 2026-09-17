package bootstrap

import (
	"context"
	"fmt"
	"time"

	"github.com/rafaelblt/go-auth/internal/infra/migrate"
	"github.com/rafaelblt/go-auth/internal/infra/postgres"
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

	switch {
	case current < latest:
		return fmt.Errorf(
			"the database schema version (%d) is behind the latest migration (%d): apply the pending migrations",
			current, latest)
	case current > latest:
		return fmt.Errorf(
			"the database schema version (%d) is ahead of the latest migration (%d): the binary is older than the database",
			current, latest)
	}
	return nil
}
