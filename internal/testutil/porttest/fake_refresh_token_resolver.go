package porttest

import (
	"context"

	"github.com/rafaelblt/go-auth/internal/session"
)

type FakeRefreshTokenResolver struct {
	data map[string]*session.RefreshToken
	err  error
}

func NewFakeRefreshTokenResolver() *FakeRefreshTokenResolver {
	return &FakeRefreshTokenResolver{
		data: make(map[string]*session.RefreshToken),
		err:  nil,
	}
}

func (r *FakeRefreshTokenResolver) Resolve(
	ctx context.Context, raw string,
) (*session.RefreshToken, error) {
	if r.err != nil {
		return nil, r.err
	}
	token, ok := r.data[raw]
	if !ok {
		return nil, session.ErrTokenInvalid
	}
	return token, nil
}

func (r *FakeRefreshTokenResolver) Insert(raw string, token *session.RefreshToken) {
	r.data[raw] = token
}

func (r *FakeRefreshTokenResolver) SetError(err error) {
	r.err = err
}
