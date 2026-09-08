package porttest

import (
	"context"

	"github.com/rafaelblt/go-auth/internal/port"
)

type FakeUnitOfWork struct {
	FakeUserWriter         *FakeUserWriter
	FakePasswordWriter     *FakePasswordWriter
	FakeSessionWriter      *FakeSessionWriter
	FakeRefreshTokenWriter *FakeRefreshTokenWriter
	err                    error
}

func NewFakeUnitOfWork() *FakeUnitOfWork {
	uow := FakeUnitOfWork{
		FakeUserWriter:         NewFakeUserWriter(),
		FakePasswordWriter:     NewFakePasswordWriter(),
		FakeSessionWriter:      NewFakeSessionWriter(),
		FakeRefreshTokenWriter: NewFakeRefreshTokenWriter(),
	}
	return &uow
}

func (uow *FakeUnitOfWork) Do(ctx context.Context, fn func(deps port.UowDeps) error) error {
	if uow.err != nil {
		return uow.err
	}
	return fn(port.UowDeps{
		UserWriter:         uow.FakeUserWriter,
		PasswordWriter:     uow.FakePasswordWriter,
		SessionWriter:      uow.FakeSessionWriter,
		RefreshTokenWriter: uow.FakeRefreshTokenWriter,
	})
}

// SetError makes Do fail without running the callback, simulating a failure of
// the transaction itself instead of a failure of one of the writers.
func (uow *FakeUnitOfWork) SetError(err error) {
	uow.err = err
}
