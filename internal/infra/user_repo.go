package infra

import (
	"context"
	"errors"
	"fmt"

	"github.com/rafaelblt/go-auth/internal/user"
)

type UserRepo struct {
	db PGDB
}

func NewUserRepo(db PGDB) (UserRepo, error) {
	if db == nil {
		return UserRepo{}, errors.New("PGDB cannot be nil")
	}
	return UserRepo{db: db}, nil
}

func (repo UserRepo) Save(ctx context.Context, usr *user.User) error {
	model, err := MapUserToModel(usr)
	if err != nil {
		return fmt.Errorf("failed to map user domain to model: %w", err)
	}
	_, err = repo.db.Exec(ctx,
		`INSERT INTO users (id, username, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)`,
		model.ID, model.Username, model.Status, model.CreatedAt, model.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("new user insert failed: %w", err)
	}
	return nil
}

func (repo UserRepo) ExistsByUsername(ctx context.Context, username user.Username) (bool, error) {
	var exists bool

	query := `SELECT EXISTS( SELECT 1 FROM users WHERE username=$1 )`
	err := repo.db.QueryRow(ctx, query, username.String()).Scan(&exists)

	if err != nil {
		return false, fmt.Errorf("failed to check username existence: %w", err)
	}

	return exists, nil
}
