package porttest

import (
	"context"
	"errors"

	"github.com/rafaelblt/go-auth/internal/session"
)

type FakeRefreshTokenWriter struct {
	data []*session.RefreshToken
	err  error
}

func NewFakeRefreshTokenWriter() *FakeRefreshTokenWriter {
	return &FakeRefreshTokenWriter{data: []*session.RefreshToken{}}
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
	w.data = append(w.data, token)
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
	var idx int
	var found *session.RefreshToken
	for i, t := range w.data {
		if t.ID() == token.ID() {
			idx = i
			found = t
		}
	}
	if found == nil {
		return errors.New("refresh token not found")
	}
	w.data[idx] = found
	return nil
}

func (w *FakeRefreshTokenWriter) Data() []*session.RefreshToken {
	return w.data
}

func (w *FakeRefreshTokenWriter) SetError(err error) { w.err = err }
