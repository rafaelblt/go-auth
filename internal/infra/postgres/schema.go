package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

var ErrSchemaNotInitialized = errors.New("schema not initialized")
var ErrSchemaDirty = errors.New("schema in dirty state")

const undefinedTableCode = "42P01"

func SchemaVersion(ctx context.Context, db DB) (version uint, err error) {
	if db == nil {
		return 0, errors.New("db nil")
	}

	var current int64
	var dirty bool

	sql := "SELECT version, dirty FROM schema_migrations LIMIT 1"
	err = db.QueryRow(ctx, sql).Scan(&current, &dirty)

	switch {
	case isUndefinedTable(err):
		return 0, fmt.Errorf("%w: schema_migrations table not found", ErrSchemaNotInitialized)
	case errors.Is(err, pgx.ErrNoRows):
		return 0, fmt.Errorf("%w: no version in schema_migrations", ErrSchemaNotInitialized)
	case err != nil:
		return 0, fmt.Errorf("schema version query failed: %w", err)
	}

	if dirty {
		return 0, fmt.Errorf("%w: version %d", ErrSchemaDirty, current)
	}
	if current < 0 {
		return 0, fmt.Errorf("%w: negative version %d", ErrSchemaNotInitialized, current)
	}

	return uint(current), nil
}

func isUndefinedTable(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == undefinedTableCode
}
