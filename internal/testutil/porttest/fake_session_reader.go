package porttest

import (
	"context"

	"github.com/rafaelblt/go-auth/internal/session"
	"github.com/rafaelblt/go-auth/internal/shared"
)

type FakeSessionReader struct {
	data map[session.SessionID]*session.Session
	err  error
}

func NewFakeSessionReader() *FakeSessionReader {
	reader := FakeSessionReader{
		data: make(map[session.SessionID]*session.Session, 0),
		err:  nil,
	}
	return &reader
}

func (r *FakeSessionReader) FindByID(
	ctx context.Context, id session.SessionID,
) (*session.Session, error) {
	if r.err != nil {
		return nil, r.err
	}
	for _, sess := range r.data {
		if sess.ID() == id {
			return shared.ClonePtr(sess), nil
		}
	}
	return nil, nil
}

func (r *FakeSessionReader) Insert(sess *session.Session) {
	r.data[sess.ID()] = shared.ClonePtr(sess)
}

func (r *FakeSessionReader) SetError(err error) {
	r.err = err
}
