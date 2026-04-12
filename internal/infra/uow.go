package infra

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/rafaelblt/go-auth/internal/usecase"
)

type UnitOfWork struct {
	tx   pgx.Tx
	deps usecase.UowDeps
}

func NewUnitOfWork(tx pgx.Tx) (*UnitOfWork, error) {
	deps, err := buildUowDeps(tx)
	if err != nil {
		return nil, err
	}
	uow := &UnitOfWork{tx, deps}
	return uow, nil
}

func buildUowDeps(tx pgx.Tx) (usecase.UowDeps, error) {
	if tx == nil {
		return usecase.UowDeps{}, errors.New("cannot build uow deps with nil tx")
	}
	userRepo, err := NewUserRepo(tx)
	if err != nil {
		return usecase.UowDeps{}, fmt.Errorf("user repo build failed: %w", err)
	}
	credRepo, err := NewCredentialRepo(tx)
	if err != nil {
		return usecase.UowDeps{}, fmt.Errorf("credential repo build failed: %w", err)
	}
	deps := usecase.UowDeps{
		UserWriter:       userRepo,
		CredentialWriter: credRepo,
	}
	return deps, nil
}

type execFn = func(deps usecase.UowDeps) error
func (uow *UnitOfWork) Do(ctx context.Context, fn execFn) error {
	defer uow.tx.Rollback(ctx)

	err := fn(uow.deps)
	if err != nil {
		return err
	}

	err = uow.tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("uow commit failed: %w", err)
	}

	return nil
}
