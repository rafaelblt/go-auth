package refreshtoken

import (
	"context"
	"testing"

	"github.com/rafaelblt/go-auth/internal/session"
	"github.com/rafaelblt/go-auth/internal/testutil/porttest"
	"github.com/rafaelblt/go-auth/internal/testutil/sessiontest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type ResolverTestHelper struct {
	t                      *testing.T
	FakeRefreshTokenReader *porttest.FakeRefreshTokenReader
}

func NewResolverTestHelper(t *testing.T) *ResolverTestHelper {
	helper := ResolverTestHelper{
		t:                      t,
		FakeRefreshTokenReader: porttest.NewFakeRefreshTokenReader(),
	}
	return &helper
}

func (h *ResolverTestHelper) Resolver() *Resolver {
	resolver, err := NewResolver(ResolverConfig{
		RefreshTokenReader: h.FakeRefreshTokenReader,
	})
	require.NoError(h.t, err)
	return resolver
}

func (h *ResolverTestHelper) GetValidToken() (*session.RefreshToken, string) {
	tokenBytes, err := generateToken()
	require.NoError(h.t, err)
	hash, err := hashToken(tokenBytes)
	require.NoError(h.t, err)

	refreshToken := sessiontest.NewRefreshToken(h.t, func(params *session.RefreshTokenRestoreParams) {
		params.Hash = hash
	})

	h.FakeRefreshTokenReader.Insert(refreshToken)
	return refreshToken, encodeToken(tokenBytes)
}

func TestResolver_ReturnsInvalidTokenError(t *testing.T) {
	helper := NewResolverTestHelper(t)
	resolver := helper.Resolver()

	token, err := resolver.Resolve(context.Background(), "invalid")

	assert.Zero(t, token)
	assert.ErrorIs(t, err, session.ErrTokenInvalid)
}

func TestResolver_ReturnsInvalidToken_WithEmptyRaw(t *testing.T) {
	helper := NewResolverTestHelper(t)
	resolver := helper.Resolver()

	token, err := resolver.Resolve(context.Background(), "")

	assert.Zero(t, token)
	assert.ErrorIs(t, err, session.ErrTokenInvalid)
}

func TestResolver_ReturnsTokenWithValidRaw(t *testing.T) {
	helper := NewResolverTestHelper(t)
	token, raw := helper.GetValidToken()

	resolver := helper.Resolver()
	ctx := context.Background()
	result, err := resolver.Resolve(ctx, raw)

	assert.NoError(t, err)
	assert.Equal(t, token.ID(), result.ID())
}
