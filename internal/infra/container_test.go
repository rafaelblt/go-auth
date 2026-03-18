package infra_test

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/rafaelblt/go-auth/migrations"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

type TestDB struct {
	container *postgres.PostgresContainer
	dbURL     string
}

func (db TestDB) ConnectionString() string {
	return db.dbURL
}

var testDB TestDB

func TestMain(m *testing.M) {
	ctx := context.Background()

	container, err := createContainer(ctx)
	if err != nil {
		log.Fatalf("could not start container: %v", err)
	}

	testDB, err = createTestDB(ctx, container)
	if err != nil {
		log.Fatalf("db setup failed: %v", err)
	}

	code := m.Run()

	if err := container.Terminate(ctx); err != nil {
		log.Printf("failed to terminate container: %v", err)
	}

	os.Exit(code)
}

func createContainer(ctx context.Context) (*postgres.PostgresContainer, error) {
	container, err := postgres.Run(
		ctx,
		"postgres:16",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("pguser"),
		postgres.WithPassword("pgpass"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30 * time.Second),
		),
	)
	return container, err
}

func createTestDB(ctx context.Context, container *postgres.PostgresContainer) (TestDB, error) {
	dbURL, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		return TestDB{}, fmt.Errorf("could not get connection string: %w", err)
	}

	err = runMigrations(dbURL)
	if err != nil {
		return TestDB{}, fmt.Errorf("migration failed: %w", err)
	}

	db :=  TestDB{
		container: container,
		dbURL: dbURL,
	}
	return db, err
}

func runMigrations(dbURL string) error {
	driver, err := iofs.New(migrations.FS, ".")
	mgrt, err := migrate.NewWithSourceInstance("iofs", driver, dbURL)
	if err != nil {
		return err
	}

	err = mgrt.Up()

	if err != nil && err != migrate.ErrNoChange {
		return err
	}

	return nil
}
