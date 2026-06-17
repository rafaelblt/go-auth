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

func (repo UserRepo) FindByID(ctx context.Context, id user.ID) (*user.User, error) {
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

	usr, err := repo.mapModel(model)
	if err != nil {
		return nil, err
	}

	return usr, nil
}

func (repo UserRepo) FindByUsername(ctx context.Context, username user.Username) (*user.User, error) {
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

	usr, err := repo.mapModel(model)
	if err != nil {
		return nil, err
	}

	return usr, nil
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

func (repo *UserRepo) mapModel(model userModel) (*user.User, error) {
	id, err := user.ParseID(model.ID)
	if err != nil {
		return nil, fmt.Errorf("parse user id failed: %w", err)
	}
	username, err := user.NewUsername(model.Username)
	if err != nil {
		return nil, fmt.Errorf("username creation failed: %w", err)
	}
	status, err := user.ParseStatus(model.Status)
	if err != nil {
		return nil, fmt.Errorf("parse user status failed: %w", err)
	}

	usr, err := user.RestoreUser(user.RestoreParams{
		ID:        id,
		Username:  username,
		Status:    status,
		CreatedAt: model.CreatedAt,
		UpdatedAt: model.UpdatedAt,
	})
	if err != nil {
		return nil, fmt.Errorf("restore user failed: %w", err)
	}

	return usr, nil
}
