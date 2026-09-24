package jwt

import (
	"errors"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/rafaelblt/go-auth/internal/domain/session"
	"github.com/rafaelblt/go-auth/internal/domain/user"
	"github.com/rafaelblt/go-auth/internal/port"
	"github.com/rafaelblt/go-auth/internal/testutil"
	"github.com/rafaelblt/go-auth/internal/testutil/porttest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TEST HELPER

type ServiceTestHelper struct {
	t          *testing.T
	FakeClock  *porttest.FakeClock
	FakeSigner *FakeSigner
	Expiration time.Duration
}

func NewServiceTestHelper(t *testing.T) *ServiceTestHelper {
	helper := ServiceTestHelper{
		t:          t,
		FakeClock:  porttest.NewFakeClock(),
		FakeSigner: NewFakeSigner(),
		Expiration: time.Minute,
	}
	return &helper
}

func (sth *ServiceTestHelper) Service() *AccessTokenService {
	return &AccessTokenService{
		signer:     sth.FakeSigner,
		clock:      sth.FakeClock,
		expiration: sth.Expiration,
	}
}

func (sth *ServiceTestHelper) ValidPayload() port.AccessTokenPayload {
	return port.AccessTokenPayload{UserID: user.NewID()}
}

// FAKE SIGNER

type FakeSigner struct {
	defaultToken  string
	defaultClaims jwt.RegisteredClaims
	signedClaims  []jwt.RegisteredClaims
	parsedTokens  []string
	signErr       error
	parseErr      error
}

func NewFakeSigner() *FakeSigner {
	s := FakeSigner{
		defaultToken:  "default fake signed token",
		defaultClaims: jwt.RegisteredClaims{Subject: user.NewID().String()},
		signedClaims:  make([]jwt.RegisteredClaims, 0),
		parsedTokens:  make([]string, 0),
	}
	return &s
}

func (fs *FakeSigner) Sign(claims jwt.RegisteredClaims) (string, error) {
	fs.signedClaims = append(fs.signedClaims, claims)
	if fs.signErr != nil {
		return "", fs.signErr
	}
	return fs.defaultToken, nil
}

func (fs *FakeSigner) Parse(token string) (jwt.RegisteredClaims, error) {
	fs.parsedTokens = append(fs.parsedTokens, token)
	if fs.parseErr != nil {
		return jwt.RegisteredClaims{}, fs.parseErr
	}
	return fs.defaultClaims, nil
}

func (fs *FakeSigner) DefaultSignedToken() string {
	return fs.defaultToken
}

func (fs *FakeSigner) DefaultParsedClaims() jwt.RegisteredClaims {
	return fs.defaultClaims
}

func (fs *FakeSigner) SignedClaims() []jwt.RegisteredClaims {
	return fs.signedClaims
}

func (fs *FakeSigner) ParsedTokens() []string {
	return fs.parsedTokens
}

func (fs *FakeSigner) SetSignError(err error) {
	fs.signErr = err
}

func (fs *FakeSigner) SetParseError(err error) {
	fs.parseErr = err
}

// TESTS

func TestNewAccessTokenService(t *testing.T) {
	cfg := AccessTokenServiceConfig{
		Signer:     NewFakeSigner(),
		Clock:      porttest.NewFakeClock(),
		Expiration: time.Hour,
	}

	service, err := NewAccessTokenService(cfg)

	require.NoError(t, err)
	require.NotZero(t, service)
	assert.Equal(t, cfg.Signer, service.signer)
	assert.Equal(t, cfg.Clock, service.clock)
	assert.Equal(t, cfg.Expiration, service.expiration)
}

func TestNewAccessTokenService_ReturnsError_WhenConfigIsInvalid(t *testing.T) {
	testCases := []struct {
		desc string
		cfg  AccessTokenServiceConfig
	}{
		{
			desc: "signer nil",
			cfg: AccessTokenServiceConfig{
				Signer:     nil,
				Clock:      porttest.NewFakeClock(),
				Expiration: time.Minute,
			},
		},
		{
			desc: "clock nil",
			cfg: AccessTokenServiceConfig{
				Signer:     NewFakeSigner(),
				Clock:      nil,
				Expiration: time.Minute,
			},
		},
		{
			desc: "expiration too short",
			cfg: AccessTokenServiceConfig{
				Signer:     NewFakeSigner(),
				Clock:      porttest.NewFakeClock(),
				Expiration: minExpiration - 1,
			},
		},
		{
			desc: "expiration too long",
			cfg: AccessTokenServiceConfig{
				Signer:     NewFakeSigner(),
				Clock:      porttest.NewFakeClock(),
				Expiration: maxExpiration + 1,
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			service, err := NewAccessTokenService(tC.cfg)
			assert.Zero(t, service)
			assert.Error(t, err)
		})
	}
}

func TestAccessTokenService_Issue_ReturnsSignedTokenAndItsTimes(t *testing.T) {
	helper := NewServiceTestHelper(t)

	issued, err := helper.Service().Issue(helper.ValidPayload())

	require.NoError(t, err)
	require.NotZero(t, issued)
	assert.Equal(t, helper.FakeSigner.DefaultSignedToken(), issued.Token.Value())
	assert.Equal(t, helper.FakeClock.Now(), issued.IssuedAt)
	assert.Equal(t,
		helper.FakeClock.Now().Add(helper.Expiration).Truncate(time.Second),
		issued.ExpiresAt)
}

func TestAccessTokenService_Issue_SignsTheTokenWithClaims(t *testing.T) {
	helper := NewServiceTestHelper(t)
	payload := helper.ValidPayload()

	helper.Service().Issue(payload)

	claims := testutil.Only(t, helper.FakeSigner.SignedClaims())
	assert.Equal(t, payload.UserID.String(), claims.Subject)
	assert.Equal(t,
		helper.FakeClock.Now().Add(helper.Expiration).Truncate(time.Second),
		claims.ExpiresAt.Time)
}

func TestAccessTokenService_Issue_ReturnsError_WhenUserIDIsZero(t *testing.T) {
	helper := NewServiceTestHelper(t)
	payload := port.AccessTokenPayload{UserID: user.ID{}}

	issued, err := helper.Service().Issue(payload)

	assert.Zero(t, issued)
	assert.Error(t, err)
}

func TestAccessTokenService_Issue_ReturnsError_WhenSignerFails(t *testing.T) {
	helper := NewServiceTestHelper(t)
	expectedErr := errors.New("unexpected error for test")
	helper.FakeSigner.SetSignError(expectedErr)

	issued, err := helper.Service().Issue(helper.ValidPayload())

	assert.Zero(t, issued)
	assert.ErrorIs(t, err, expectedErr)
}

func TestAccessTokenService_Validate_ReturnsClaims(t *testing.T) {
	helper := NewServiceTestHelper(t)
	expected := helper.FakeSigner.DefaultParsedClaims()

	claims, err := helper.Service().Validate("token")

	require.NoError(t, err)
	require.NotZero(t, claims)
	assert.Equal(t, expected.Subject, claims.UserID.String())
}

func TestAccessTokenService_Validate_ReturnsErrTokenInvalid(t *testing.T) {
	helper := NewServiceTestHelper(t)
	helper.FakeSigner.SetParseError(ErrTokenInvalid)

	claims, err := helper.Service().Validate("invalid")

	assert.Zero(t, claims)
	assert.ErrorIs(t, err, session.ErrTokenInvalid)
}

func TestAccessTokenService_Validate_ReturnsErrTokenExpired(t *testing.T) {
	helper := NewServiceTestHelper(t)
	helper.FakeSigner.SetParseError(ErrTokenExpired)

	claims, err := helper.Service().Validate("expired")

	assert.Zero(t, claims)
	assert.ErrorIs(t, err, session.ErrTokenExpired)
}

func TestAccessTokenService_Validate_ReturnsError_WhenSignerFails(t *testing.T) {
	helper := NewServiceTestHelper(t)
	expectedErr := errors.New("unexpected error for test")
	helper.FakeSigner.SetParseError(expectedErr)

	claims, err := helper.Service().Validate("token")

	assert.Zero(t, claims)
	assert.ErrorIs(t, err, expectedErr)
}
