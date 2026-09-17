package porttest

import (
	"context"
	"errors"

	"github.com/rafaelblt/go-auth/internal/domain/session"
)

type FakeSessionWriter struct {
	adds    []*session.Session
	updates []*session.Session
	err     error
}

func NewFakeSessionWriter() *FakeSessionWriter {
	return &FakeSessionWriter{
		adds:    []*session.Session{},
		updates: []*session.Session{},
	}
}

func (w *FakeSessionWriter) Add(
	ctx context.Context, sess *session.Session,
) error {
	if w.err != nil {
		return w.err
	}
	if sess == nil {
		return errors.New("session nil")
	}
	if sess.IsZero() {
		return errors.New("session zero")
	}
	w.adds = append(w.adds, sess)
	return w.err
}

func (w *FakeSessionWriter) Update(
	ctx context.Context, sess *session.Session,
) error {
	if w.err != nil {
		return w.err
	}
	if sess == nil {
		return errors.New("session nil")
	}
	if sess.IsZero() {
		return errors.New("session zero")
	}
	w.updates = append(w.updates, sess)
	return nil
}

func (w *FakeSessionWriter) Adds() []*session.Session {
	return w.adds
}

func (w *FakeSessionWriter) Updates() []*session.Session {
	return w.updates
}

func (w *FakeSessionWriter) SetError(err error) {
	w.err = err
}
