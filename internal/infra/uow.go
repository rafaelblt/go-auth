package infra

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/rafaelblt/go-auth/internal/port"
)

type TxBeginner interface {
	Begin(ctx context.Context) (pgx.Tx, error)
}

type UnitOfWork struct {
	beginner TxBeginner
}

func NewUnitOfWork(beginner TxBeginner) (*UnitOfWork, error) {
	if beginner == nil {
		return nil, errors.New("tx beginner cannot be nil")
	}
	uow := &UnitOfWork{beginner}
	return uow, nil
}

type workFn = func(deps port.UowDeps) error

func (uow *UnitOfWork) Do(ctx context.Context, fn workFn) error {
	tx, err := uow.beginner.Begin(ctx)
	if err != nil {
		return fmt.Errorf("tx begin failed: %w", err)
	}

	defer tx.Rollback(ctx)

	deps, err := buildUowDeps(tx)
	if err != nil {
		return err
	}

	if err = fn(deps); err != nil {
		return err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("tx commit failed: %w", err)
	}

	return nil
}

func buildUowDeps(tx pgx.Tx) (port.UowDeps, error) {
	if tx == nil {
		return port.UowDeps{}, errors.New("cannot build uow deps with nil tx")
	}
	userRepo, err := NewUserRepo(tx)
	if err != nil {
		return port.UowDeps{}, fmt.Errorf("user repo build failed: %w", err)
	}
	credRepo, err := NewCredentialRepo(tx)
	if err != nil {
		return port.UowDeps{}, fmt.Errorf("credential repo build failed: %w", err)
	}
	deps := port.UowDeps{
		UserWriter:       userRepo,
		CredentialWriter: credRepo,
	}
	return deps, nil
}
