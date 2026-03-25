package infra

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type PGDB interface {
	Exec(ctx context.Context, sql string, args ...any) (cmdTag pgconn.CommandTag, err error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Commit(ctx context.Context) error
}
