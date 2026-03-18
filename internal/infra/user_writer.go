package infra

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/rafaelblt/go-auth/internal/domain"
)

type PGDB interface {
	Exec(ctx context.Context, sql string, args ...any) (cmdTag pgconn.CommandTag, err error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Commit(ctx context.Context) error
}

type UserWriter struct {
	db PGDB
}

func NewUserWriter(db PGDB) (UserWriter, error) {
	if db == nil {
		return UserWriter{}, errors.New("PGDB cannot be nil")
	}
	return UserWriter{db: db}, nil
}

func (writer UserWriter) Save(ctx context.Context, user *domain.User) error {
	model, err := MapUserDomainToModel(user)
	if err != nil {
		return err
	}
	_, err = writer.db.Exec(ctx,
		`INSERT INTO users (id, username, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)`,
		model.ID, model.Username, model.Status, model.CreatedAt, model.UpdatedAt,
	)
	if err != nil {
		return err
	}
	return nil
}
