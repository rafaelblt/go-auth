package infra

import (
	"context"
	"errors"

	"github.com/rafaelblt/go-auth/internal/domain"
)

type CredentialRepo struct {
	db PGDB
}

func NewCredentialRepo(db PGDB) (CredentialRepo, error) {
	if db == nil {
		return CredentialRepo{}, errors.New("PGDB cannot be nil")
	}
	return CredentialRepo{db: db}, nil
}

func (repo CredentialRepo) Save(ctx context.Context, cred *domain.Credential) error {
	model, err := MapCredentialDomainToModel(cred)
	if err != nil {
		return err
	}
	_, err = repo.db.Exec(ctx,
		`INSERT INTO credentials
			(id, user_id, kind, provider, secret, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		model.ID,
		model.UserID,
		model.Kind,
		model.Provider,
		model.Secret,
		model.CreatedAt,
		model.UpdatedAt,
	)
	if err != nil {
		return err
	}
	return nil
}
