package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/rafaelblt/go-auth/internal/credential"
)

type CredentialRepo struct {
	db DB
}

func NewCredentialRepo(db DB) (CredentialRepo, error) {
	if db == nil {
		return CredentialRepo{}, errors.New("DB nil")
	}
	return CredentialRepo{db: db}, nil
}

func (repo CredentialRepo) Save(ctx context.Context, cred *credential.Credential) error {
	if cred == nil {
		return errors.New("cannot save a nil credential in database")
	}
	if cred.IsZero() {
		return errors.New("cannot save a zero credential in database")
	}

	sql := `INSERT INTO credentials
			(id, user_id, kind, provider, secret, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err := repo.db.Exec(ctx, sql,
		cred.ID().Value(),
		cred.UserID().Value(),
		cred.Kind().String(),
		cred.Provider().String(),
		cred.Secret().Value(),
		cred.CreatedAt(),
		cred.UpdatedAt(),
	)

	if err != nil {
		return fmt.Errorf("credential insert failed: %w", err)
	}

	return nil
}
