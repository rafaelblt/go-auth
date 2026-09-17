package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/rafaelblt/go-auth/internal/user"
)

type UserRepo struct {
	db DB
}

func NewUserRepo(db DB) (*UserRepo, error) {
	if db == nil {
		return nil, errors.New("db nil")
	}
	return &UserRepo{db: db}, nil
}

// Add guards the insert with ON CONFLICT (username), so of two concurrent
// registrations of the same username only one is applied. The loser inserts
// nothing and gets user.ErrUsernameAlreadyExists. A conflict on any other
// column, the primary key included, is still an error.
//
// See docs/development/decisions/0048-duplicate-username-reported-by-the-writer.md.
func (repo *UserRepo) Add(ctx context.Context, usr *user.User) error {
	model, err := mapUserToModel(usr)
	if err != nil {
		return fmt.Errorf("map user to model failed: %w", err)
	}

	sql := `INSERT INTO users
			(id, username, status, created_at, updated_at)
			VALUES
			(@id, @username, @status, @created_at, @updated_at)
			ON CONFLICT (username) DO NOTHING`
	tag, err := repo.db.Exec(ctx, sql, pgx.StrictStructArgs(model))

	if err != nil {
		return fmt.Errorf("user insert failed: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return user.ErrUsernameAlreadyExists
	}

	return nil
}

func (repo *UserRepo) FindByID(ctx context.Context, id user.ID) (*user.User, error) {
	if id.IsZero() {
		return nil, errors.New("id zero")
	}

	sql := "SELECT * FROM users WHERE id = $1"

	rows, err := repo.db.Query(ctx, sql, id.Value().String())
	if err != nil {
		return nil, fmt.Errorf("db query failed: %w", err)
	}
	defer rows.Close()

	model, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[userModel])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("collect one row failed: %w", err)
	}

	usr, err := mapUserToEntity(model)
	if err != nil {
		return nil, err
	}

	return usr, nil
}

func (repo *UserRepo) FindByUsername(ctx context.Context, username user.Username) (*user.User, error) {
	sql := "SELECT * FROM users WHERE username = $1"

	rows, err := repo.db.Query(ctx, sql, username.String())
	if err != nil {
		return nil, fmt.Errorf("db query failed: %w", err)
	}
	defer rows.Close()

	model, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[userModel])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("collect one row failed: %w", err)
	}

	usr, err := mapUserToEntity(model)
	if err != nil {
		return nil, err
	}

	return usr, nil
}

func (repo *UserRepo) ExistsByUsername(ctx context.Context, username user.Username) (bool, error) {
	var exists bool

	sql := `SELECT EXISTS( SELECT 1 FROM users WHERE username=$1 )`
	err := repo.db.QueryRow(ctx, sql, username.String()).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check username existence: %w", err)
	}

	return exists, nil
}
