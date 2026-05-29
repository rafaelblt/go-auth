package porttest

import (
	"context"

	"github.com/rafaelblt/go-auth/internal/port"
)

type FakeUnitOfWork struct {
	FakeUserWriter         *FakeUserWriter
	FakeCredentialWriter   *FakeCredentialWriter
	FakeSessionWriter      *FakeSessionWriter
	FakeRefreshTokenWriter *FakeRefreshTokenWriter
}

func NewFakeUnitOfWork() *FakeUnitOfWork {
	uow := FakeUnitOfWork{
		FakeUserWriter:         NewFakeUserWriter(),
		FakeCredentialWriter:   NewFakeCredentialWriter(),
		FakeSessionWriter:      NewFakeSessionWriter(),
		FakeRefreshTokenWriter: NewFakeRefreshTokenWriter(),
	}
	return &uow
}

func (uow FakeUnitOfWork) Do(ctx context.Context, fn func(deps port.UowDeps) error) error {
	return fn(port.UowDeps{
		UserWriter:         uow.FakeUserWriter,
		CredentialWriter:   uow.FakeCredentialWriter,
		SessionWriter:      uow.FakeSessionWriter,
		RefreshTokenWriter: uow.FakeRefreshTokenWriter,
	})
}
