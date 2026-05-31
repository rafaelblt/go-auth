package postgrestest

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

func ResetDB(ctx context.Context, pool *pgxpool.Pool) error {
	_, err := pool.Exec(ctx, `
        TRUNCATE TABLE users, credentials
        RESTART IDENTITY CASCADE
    `)
	if err != nil {
		return fmt.Errorf("truncate tables failed: %w", err)
	}
	return nil
}
