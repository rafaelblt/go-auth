package testutil

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/docker/docker/client"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

type Database struct {
	container *postgres.PostgresContainer
	conn      string
}

func NewDatabase(ctx context.Context) (*Database, error) {
	if err := checkDockerAvailable(ctx); err != nil {
		return nil, err
	}

	testcontainer, err := createContainer(ctx)
	if err != nil {
		return nil, err
	}

	conn, err := testcontainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		return nil, fmt.Errorf("could not get connection string: %w", err)
	}

	db := Database{
		container: testcontainer,
		conn:      conn,
	}

	return &db, nil
}

func NewDatabaseForTest(t *testing.T, ctx context.Context) (*Database) {
	db, err := NewDatabase(ctx)
	require.NoError(t, err)

	t.Cleanup(func() {
		db.Close(ctx)
	})

	return db
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

func (db *Database) ConnectionString() string { return db.conn }

func (db *Database) Close(ctx context.Context) error {
	if err := db.container.Terminate(ctx); err != nil {
		return fmt.Errorf("test container terminate failed: %w", err)
	}

	return nil
}

func checkDockerAvailable(ctx context.Context) error {
    client, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
    if err != nil {
        return fmt.Errorf("Docker not found: %w", err)
    }
    defer client.Close()

    if _, err := client.Ping(ctx); err != nil {
        return fmt.Errorf("Docker is unavailable (is it running?): %w", err)
    }

    return nil
}
