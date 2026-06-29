package porttest

import (
	"context"
	"errors"

	"github.com/rafaelblt/go-auth/internal/session"
)

type FakeRefreshTokenWriter struct {
	adds    []*session.RefreshToken
	updates []*session.RefreshToken
	err     error
}

func NewFakeRefreshTokenWriter() *FakeRefreshTokenWriter {
	return &FakeRefreshTokenWriter{
		adds:    []*session.RefreshToken{},
		updates: []*session.RefreshToken{},
	}
}

func (w *FakeRefreshTokenWriter) Add(
	ctx context.Context, token *session.RefreshToken,
) error {
	if w.err != nil {
		return w.err
	}
	if token == nil {
		return errors.New("refresh token nil")
	}
	if token.IsZero() {
		return errors.New("refresh token zero")
	}
	w.adds = append(w.adds, token)
	return w.err
}

func (w *FakeRefreshTokenWriter) Update(
	ctx context.Context, token *session.RefreshToken,
) error {
	if w.err != nil {
		return w.err
	}
	if token == nil {
		return errors.New("refresh token nil")
	}
	if token.IsZero() {
		return errors.New("refresh token zero")
	}
	w.updates = append(w.updates, token)
	return nil
}

func (w *FakeRefreshTokenWriter) Adds() []*session.RefreshToken {
	return w.adds
}

func (w *FakeRefreshTokenWriter) Updates() []*session.RefreshToken {
	return w.updates
}

func (w *FakeRefreshTokenWriter) SetError(err error) {
	w.err = err
}
