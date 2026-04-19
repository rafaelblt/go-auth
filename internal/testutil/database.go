package testutil

import (
	"context"
	"fmt"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rafaelblt/go-auth/internal/infra"
	"github.com/rafaelblt/go-auth/migrations"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

type Database struct {
	container *postgres.PostgresContainer
	pool      *pgxpool.Pool
	conn      string
}

func NewDatabase(ctx context.Context) (*Database, error) {
	testcontainer, err := createContainer(ctx)
	if err != nil {
		return nil, err
	}

	conn, err := testcontainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		return nil, fmt.Errorf("could not get connection string: %w", err)
	}

	err = runMigrations(conn)
	if err != nil {
		return nil, err
	}

	pool, err := infra.NewPool(ctx, conn)
	if err != nil {
		return nil, fmt.Errorf("pool creation failed: %w", err)
	}

	db := Database{
		container: testcontainer,
		pool:      pool,
		conn:      conn,
	}

	return &db, nil
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
				WithStartupTimeout(30*time.Second),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("could not start postgres testcontainer: %w", err)
	}
	return container, nil
}

func runMigrations(dbURL string) error {
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

func (db *Database) ConnectionString() string { return db.conn }
func (db *Database) Pool() *pgxpool.Pool      { return db.pool }

func (db *Database) Close(ctx context.Context) error {
	if err := db.container.Terminate(ctx); err != nil {
		return fmt.Errorf("test container terminate failed: %w", err)
	}
	db.pool.Close()

	return nil
}
