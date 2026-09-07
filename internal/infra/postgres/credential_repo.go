package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/rafaelblt/go-auth/internal/password"
	"github.com/rafaelblt/go-auth/internal/user"
)

type CredentialRepo struct {
	db DB
}

func NewCredentialRepo(db DB) (*CredentialRepo, error) {
	if db == nil {
		return nil, errors.New("db nil")
	}
	return &CredentialRepo{db: db}, nil
}

func (repo *CredentialRepo) Add(ctx context.Context, cred *password.Credential) error {
	model, err := mapCredentialToModel(cred)
	if err != nil {
		return fmt.Errorf("map credential to model failed: %w", err)
	}

	sql := `INSERT INTO password_credentials
			(id, user_id, hash, created_at, updated_at)
			VALUES
			(@id, @user_id, @hash, @created_at, @updated_at)`
	_, err = repo.db.Exec(ctx, sql, pgx.StrictStructArgs(model))

	if err != nil {
		return fmt.Errorf("credential insert failed: %w", err)
	}

	return nil
}

func (repo *CredentialRepo) FindByID(ctx context.Context, id password.ID) (*password.Credential, error) {
	sql := "SELECT * FROM password_credentials WHERE id = $1"

	return repo.findOne(ctx, sql, id.Value().String())
}

func (repo *CredentialRepo) FindByUserID(
	ctx context.Context, userID user.ID,
) (*password.Credential, error) {
	sql := "SELECT * FROM password_credentials WHERE user_id = $1"

	return repo.findOne(ctx, sql, userID.Value().String())
}

func (repo *CredentialRepo) findOne(
	ctx context.Context, sql string, args ...any,
) (*password.Credential, error) {
	rows, err := repo.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("db query failed: %w", err)
	}
	defer rows.Close()

	model, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[credentialModel])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("collect one row failed: %w", err)
	}

	cred, err := mapCredentialToEntity(model)
	if err != nil {
		return nil, err
	}

	return cred, nil
}
