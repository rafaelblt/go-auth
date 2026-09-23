// Package migrate applies the embedded SQL migrations with golang-migrate.
//
// See docs/architecture/persistence/migrations.md.
package migrate

import (
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/rafaelblt/go-auth/migrations"
)

func RunMigrations(dbURL string) error {
	driver, err := iofs.New(migrations.FS, ".")
	if err != nil {
		return fmt.Errorf("iofs new driver failed: %w", err)
	}

	mgrt, err := migrate.NewWithSourceInstance("iofs", driver, dbURL)
	if err != nil {
		return fmt.Errorf("new migrate instance failed: %w", err)
	}

	err = mgrt.Up()

	if err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("migration up failed: %w", err)
	}

	return nil
}
