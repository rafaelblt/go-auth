package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/rafaelblt/go-auth/internal/user"
)

type UserRepo struct {
	db DB
}

func NewUserRepo(db DB) (UserRepo, error) {
	if db == nil {
		return UserRepo{}, errors.New("DB nil")
	}
	return UserRepo{db: db}, nil
}

func (repo UserRepo) Save(ctx context.Context, usr *user.User) error {
	if usr == nil {
		return errors.New("cannot save a nil user in database")
	}
	if usr.IsZero() {
		return errors.New("cannot save a zero user in database")
	}

	sql := `INSERT INTO users
			(id, username, status, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5)`

	_, err := repo.db.Exec(ctx, sql,
		usr.ID().Value(),
		usr.Username().String(),
		usr.Status().String(),
		usr.CreatedAt(),
		usr.UpdatedAt(),
	)

	if err != nil {
		return fmt.Errorf("user insert failed: %w", err)
	}

	return nil
}

func (repo UserRepo) ExistsByUsername(ctx context.Context, username user.Username) (bool, error) {
	var exists bool

	sql := `SELECT EXISTS( SELECT 1 FROM users WHERE username=$1 )`
	err := repo.db.QueryRow(ctx, sql, username.String()).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check username existence: %w", err)
	}

	return exists, nil
}
