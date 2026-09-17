package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/rafaelblt/go-auth/internal/domain/password"
	"github.com/rafaelblt/go-auth/internal/domain/user"
)

type PasswordRepo struct {
	db DB
}

func NewPasswordRepo(db DB) (*PasswordRepo, error) {
	if db == nil {
		return nil, errors.New("db nil")
	}
	return &PasswordRepo{db: db}, nil
}

func (repo *PasswordRepo) Add(ctx context.Context, pwd *password.Password) error {
	model, err := mapPasswordToModel(pwd)
	if err != nil {
		return fmt.Errorf("map password to model failed: %w", err)
	}

	sql := `INSERT INTO passwords
			(id, user_id, hash, created_at, updated_at)
			VALUES
			(@id, @user_id, @hash, @created_at, @updated_at)`
	_, err = repo.db.Exec(ctx, sql, pgx.StrictStructArgs(model))

	if err != nil {
		return fmt.Errorf("password insert failed: %w", err)
	}

	return nil
}

func (repo *PasswordRepo) FindByID(ctx context.Context, id password.ID) (*password.Password, error) {
	sql := "SELECT * FROM passwords WHERE id = $1"

	return repo.findOne(ctx, sql, id.Value().String())
}

func (repo *PasswordRepo) FindByUserID(
	ctx context.Context, userID user.ID,
) (*password.Password, error) {
	sql := "SELECT * FROM passwords WHERE user_id = $1"

	return repo.findOne(ctx, sql, userID.Value().String())
}

func (repo *PasswordRepo) findOne(
	ctx context.Context, sql string, args ...any,
) (*password.Password, error) {
	rows, err := repo.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("db query failed: %w", err)
	}
	defer rows.Close()

	model, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[passwordModel])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("collect one row failed: %w", err)
	}

	pwd, err := mapPasswordToEntity(model)
	if err != nil {
		return nil, err
	}

	return pwd, nil
}
