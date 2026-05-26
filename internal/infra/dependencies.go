package infra

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rafaelblt/go-auth/internal/port"
	"github.com/rafaelblt/go-auth/internal/usecase/register"
)

type DependencyContainer struct {
	pool  *pgxpool.Pool
	clock port.Clock
}

type DependenciesConfig struct {
	DatabasePool *pgxpool.Pool
}

func NewDependencyContainer(
	ctx context.Context, cfg DependenciesConfig,
) (*DependencyContainer, error) {
	if cfg.DatabasePool == nil {
		return nil, errors.New("database pool of dependencies config cannot be nil")
	}
	container := DependencyContainer{
		pool:  cfg.DatabasePool,
		clock: NewSystemClock(),
	}
	return &container, nil
}

func (ctr *DependencyContainer) BuildRegister() (register.Register, error) {
	userExistsChecker, err := NewUserRepo(ctr.pool)
	if err != nil {
		return register.Register{},
			fmt.Errorf("user repo as user exists checker creation failed: %w", err)
	}
	uow, err := ctr.buildUow()
	if err != nil {
		return register.Register{}, err
	}
	hasher, err := ctr.buildPwdHasher()
	if err != nil {
		return register.Register{}, err
	}
	uc, err := register.New(register.Config{
		UserExistsChecker: userExistsChecker,
		UnitOfWork:        uow,
		PasswordHasher:    hasher,
		Clock:             ctr.clock,
	})
	return uc, nil
}

func (ctr *DependencyContainer) buildUow() (port.UnitOfWork, error) {
	uow, err := NewUnitOfWork(ctr.pool)
	if err != nil {
		return nil, fmt.Errorf("unit of work creation failed: %w", err)
	}
	return uow, nil
}

func (ctr *DependencyContainer) buildPwdHasher() (port.PasswordHasher, error) {
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
