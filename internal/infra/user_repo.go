package infra

import (
	"context"
	"errors"
	"fmt"

	"github.com/rafaelblt/go-auth/internal/domain"
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

func (repo UserRepo) Save(ctx context.Context, user *domain.User) error {
	model, err := MapUserDomainToModel(user)
	if err != nil {
		return err
	}
	_, err = repo.db.Exec(ctx,
		`INSERT INTO users (id, username, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)`,
		model.ID, model.Username, model.Status, model.CreatedAt, model.UpdatedAt,
	)
	if err != nil {
		return err
	}
	return nil
}

func (repo UserRepo) ExistsByUsername(ctx context.Context, username domain.Username) (bool, error) {
	var exists bool

	query := `SELECT EXISTS( SELECT 1 FROM users WHERE username=$1 )`
	err := repo.db.QueryRow(ctx, query, username.String()).Scan(&exists)

	if err != nil {
		return false, fmt.Errorf("failed to check username existence: %w", err)
	}

	return exists, nil
}
