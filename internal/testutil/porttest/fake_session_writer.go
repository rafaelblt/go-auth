package porttest

import (
	"context"
	"errors"
	"time"

	"github.com/rafaelblt/go-auth/internal/domain/session"
	"github.com/rafaelblt/go-auth/internal/domain/user"
)

type FakeSessionWriter struct {
	adds        []*session.Session
	updates     []*session.Session
	revocations []FakeSessionRevocation
	err         error
}

type FakeSessionRevocation struct {
	UserID    user.ID
	RevokedAt time.Time
}

func NewFakeSessionWriter() *FakeSessionWriter {
	return &FakeSessionWriter{
		adds:        []*session.Session{},
		updates:     []*session.Session{},
		revocations: []FakeSessionRevocation{},
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

func (w *FakeSessionWriter) RevokeAllByUserID(
	ctx context.Context, userID user.ID, revokedAt time.Time,
) error {
	if w.err != nil {
		return w.err
	}
	if userID.IsZero() {
		return errors.New("user id zero")
	}
	w.revocations = append(w.revocations, FakeSessionRevocation{
		UserID:    userID,
		RevokedAt: revokedAt,
	})
	return nil
}

func (w *FakeSessionWriter) Adds() []*session.Session {
	return w.adds
}

func (w *FakeSessionWriter) Updates() []*session.Session {
	return w.updates
}

func (w *FakeSessionWriter) Revocations() []FakeSessionRevocation {
	return w.revocations
}

func (w *FakeSessionWriter) SetError(err error) {
	w.err = err
}
