// Package postgres is the PostgreSQL adapter: the connection pool, one
// repository per table, the unit of work, and the schema version check. It is
// the only code in the service that writes SQL.
//
// See docs/architecture/persistence.md.
package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPool(ctx context.Context, conn string) (*pgxpool.Pool, error) {
	cfg, err := parseConnString(conn)
	if err != nil {
		return nil, fmt.Errorf("connection string %w", err)
	}

	cfg.ConnConfig.RuntimeParams["timezone"] = "UTC"
	cfg.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
		conn.TypeMap().RegisterType(&pgtype.Type{
			Name:  "timestamptz",
			OID:   pgtype.TimestamptzOID,
			Codec: &pgtype.TimestamptzCodec{ScanLocation: time.UTC},
		})
		return nil
	}

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("new with config: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping: %w", err)
	}

	return pool, nil
}

// CheckConnString is the rule for the connection string, which DATABASE_URL
// sets: what NewPool parses before it connects. Its error never holds the
// connection string.
func CheckConnString(conn string) error {
	_, err := parseConnString(conn)
	return err
}

// parseConnString refuses a blank string, which pgx would not: it would fall
// back to the libpq environment variables.
//
// See docs/configuration.md#database_url.
func parseConnString(conn string) (*pgxpool.Config, error) {
	if strings.TrimSpace(conn) == "" {
		return nil, errors.New("must not be blank")
	}

	cfg, err := pgxpool.ParseConfig(conn)
	if err != nil {
		return nil, fmt.Errorf("must be a PostgreSQL connection string: %s", parseFailure(err))
	}
	return cfg, nil
}

// parseFailure is what pgx says is wrong, without the connection string it
// quotes, which masks the password but not the host, user or database.
func parseFailure(err error) string {
	var parseErr *pgconn.ParseConfigError
	if !errors.As(err, &parseErr) {
		return "pgx could not parse it"
	}
	unquoted := *parseErr
	unquoted.ConnString = ""
	return strings.TrimPrefix(unquoted.Error(), "cannot parse ``: ")
}
