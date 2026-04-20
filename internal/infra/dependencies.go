package infra

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rafaelblt/go-auth/internal/usecase"
)

type DependencyContainer struct {
	pool  *pgxpool.Pool
	clock usecase.Clock
}

type DependenciesConfig struct {
	DatabaseConnection string
}

func NewDependencyContainer(
	ctx context.Context, cfg DependenciesConfig,
) (*DependencyContainer, error) {
	pool, err := NewPool(ctx, cfg.DatabaseConnection)
	if err != nil {
		return nil, fmt.Errorf("pool creation failed: %w", err)
	}
	container := DependencyContainer{
		pool:  pool,
		clock: NewSystemClock(),
	}
	return &container, nil
}

func (ctr *DependencyContainer) BuildRegister() (usecase.Register, error) {
	userExistsChecker, err := NewUserRepo(ctr.pool)
	if err != nil {
		return usecase.Register{},
			fmt.Errorf("user repo as user exists checker creation failed: %w", err)
	}
	uow, err := ctr.buildUow()
	if err != nil {
		return usecase.Register{}, err
	}
	hasher, err := ctr.buildPwdHasher()
	if err != nil {
		return usecase.Register{}, err
	}
	uc, err := usecase.NewRegister(usecase.RegisterConfig{
		UserExistsChecker: userExistsChecker,
		UnitOfWork:        uow,
		PasswordHasher:    hasher,
		Clock:             ctr.clock,
	})
	return uc, nil
}

func (ctr *DependencyContainer) buildUow() (usecase.UnitOfWork, error) {
	uow, err := NewUnitOfWork(ctr.pool)
	if err != nil {
		return nil, fmt.Errorf("unit of work creation failed: %w", err)
	}
	return uow, nil
}

func (ctr *DependencyContainer) buildPwdHasher() (usecase.PasswordHasher, error) {
	hasher, err := NewBcryptHasher(BcryptConfig{
		Cost: 8,
	})
	if err != nil {
		return nil, fmt.Errorf("bcrypt hasher creation failed: %w", err)
	}
	return hasher, nil
}

func (ctr *DependencyContainer) Close() {
	ctr.pool.Close()
}
