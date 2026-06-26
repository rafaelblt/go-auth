package porttest

import (
	"context"

	"github.com/rafaelblt/go-auth/internal/session"
)

type FakeSessionWriter struct {
	saved []*session.Session
	err   error
}

func NewFakeSessionWriter() *FakeSessionWriter {
	return &FakeSessionWriter{saved: []*session.Session{}}
}

func (w *FakeSessionWriter) Add(
	ctx context.Context, session *session.Session,
) error {
	if w.err == nil {
		w.saved = append(w.saved, session)
		return nil
	}
	return w.err
}

func (w *FakeSessionWriter) SavedSessions() []*session.Session {
	return w.saved
}

func (w *FakeSessionWriter) SetError(err error) { w.err = err }
