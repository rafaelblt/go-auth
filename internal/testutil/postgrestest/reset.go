package postgrestest

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const migrationsTable = "schema_migrations"

func ResetDB(ctx context.Context, pool *pgxpool.Pool) error {
	_, err := pool.Exec(ctx, `
        DROP SCHEMA public CASCADE;
        CREATE SCHEMA public;
    `)
	if err != nil {
		return fmt.Errorf("reset schema failed: %w", err)
	}

	pool.Reset()

	return nil
}

func TruncateTables(ctx context.Context, pool *pgxpool.Pool) error {
	rows, err := pool.Query(ctx, `
        SELECT quote_ident(tablename)
        FROM pg_tables
        WHERE schemaname = 'public' AND tablename <> $1
    `, migrationsTable)
	if err != nil {
		return fmt.Errorf("list tables failed: %w", err)
	}

	tables, err := pgx.CollectRows(rows, pgx.RowTo[string])
	if err != nil {
		return fmt.Errorf("collect tables failed: %w", err)
	}

	if len(tables) == 0 {
		return nil
	}

	sql := "TRUNCATE TABLE " + strings.Join(tables, ", ") + " RESTART IDENTITY"
	if _, err := pool.Exec(ctx, sql); err != nil {
		return fmt.Errorf("truncate tables failed: %w", err)
	}

	return nil
}
