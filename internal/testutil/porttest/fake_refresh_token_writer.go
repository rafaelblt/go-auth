package porttest

import (
	"context"

	"github.com/rafaelblt/go-auth/internal/session"
)

type FakeRefreshTokenWriter struct {
	saved []*session.RefreshToken
	err   error
}

func NewFakeRefreshTokenWriter() *FakeRefreshTokenWriter {
	return &FakeRefreshTokenWriter{saved: []*session.RefreshToken{}}
}

func (w *FakeRefreshTokenWriter) Save(
	ctx context.Context, session *session.RefreshToken,
) error {
	if w.err == nil {
		w.saved = append(w.saved, session)
		return nil
	}
	return w.err
}

func (w *FakeRefreshTokenWriter) SavedTokens() []*session.RefreshToken {
	return w.saved
}

func (w *FakeRefreshTokenWriter) SetError(err error) { w.err = err }
