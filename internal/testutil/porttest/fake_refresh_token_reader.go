package porttest

import (
	"context"

	"github.com/rafaelblt/go-auth/internal/session"
)

type refreshTokenID = session.RefreshTokenID
type refreshToken = session.RefreshToken

type FakeRefreshTokenReader struct {
 	data map[refreshTokenID]*refreshToken
	err  error
}

func NewFakeRefreshTokenReader() *FakeRefreshTokenReader {
	return &FakeRefreshTokenReader{
		data: make(map[refreshTokenID]*refreshToken),
	}
}

func (r *FakeRefreshTokenReader) FindByHash(ctx context.Context, hash session.RefreshTokenHash) (*refreshToken, error) {
	if r.err != nil {
		return nil, r.err
	}
	for _, token := range r.data {
		if token.Hash().Equal(hash) {
			return token, nil
		}
	}
	return nil, nil
}

func (r *FakeRefreshTokenReader) Insert(token *refreshToken) {
	r.data[token.ID()] = token
}

func (r *FakeRefreshTokenReader) SetError(err error) {
	r.err = err
}
