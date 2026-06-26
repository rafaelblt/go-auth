package porttest

import (
	"context"
	"errors"

	"github.com/rafaelblt/go-auth/internal/session"
)

type FakeSessionWriter struct {
	data []*session.Session
	err  error
}

func NewFakeSessionWriter() *FakeSessionWriter {
	return &FakeSessionWriter{data: []*session.Session{}}
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
	w.data = append(w.data, sess)
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
	var idx int
	var found *session.Session
	for i, s := range w.data {
		if s.ID() == sess.ID() {
			idx = i
			found = s
		}
	}
	if found == nil {
		return errors.New("session not found")
	}
	w.data[idx] = found
	return nil
}

func (w *FakeSessionWriter) Data() []*session.Session {
	return w.data
}

func (w *FakeSessionWriter) SetError(err error) { w.err = err }
