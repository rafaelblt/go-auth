package porttest

import (
	"context"
	"errors"

	"github.com/rafaelblt/go-auth/internal/session"
)

type FakeRefreshTokenWriter struct {
	adds       []*session.RefreshToken
	markedUsed []*session.RefreshToken
	err        error
}

func NewFakeRefreshTokenWriter() *FakeRefreshTokenWriter {
	return &FakeRefreshTokenWriter{
		adds:       []*session.RefreshToken{},
		markedUsed: []*session.RefreshToken{},
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

func (w *FakeRefreshTokenWriter) MarkUsed(
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
	if !token.IsUsed() {
		return errors.New("refresh token not used")
	}
	w.markedUsed = append(w.markedUsed, token)
	return nil
}

func (w *FakeRefreshTokenWriter) Adds() []*session.RefreshToken {
	return w.adds
}

func (w *FakeRefreshTokenWriter) MarkedUsed() []*session.RefreshToken {
	return w.markedUsed
}

func (w *FakeRefreshTokenWriter) SetError(err error) {
	w.err = err
}
