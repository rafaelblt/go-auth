package ed25519

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/rafaelblt/go-auth/internal/testutil/porttest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func signerForTest(t *testing.T) *Signer {
	signer, _ := signerAndFakeClockForTest(t)
	return signer
}

func signerAndFakeClockForTest(t *testing.T) (*Signer, *porttest.FakeClock) {
	clock := porttest.NewFakeClock()
	keyring := keyringForTest(t)
	signer, err := NewSigner(SignerConfig{
		Keyring: keyring,
		Clock:   clock,
	})
	require.NoError(t, err)
	return signer, clock
}

func validClaims() jwt.RegisteredClaims {
	return jwt.RegisteredClaims{
		Subject:   "subject exxample for test",
		Issuer:    "issuer example for test",
		ExpiresAt: jwt.NewNumericDate(time.Date(6000, 6, 6, 6, 6, 0, 0, time.UTC)),
		IssuedAt:  jwt.NewNumericDate(time.Date(1000, 1, 1, 1, 1, 0, 0, time.UTC)),
	}
}

func tokenWithHeader(t *testing.T, signer *Signer, header string) string {
	t.Helper()
	token, err := signer.Sign(validClaims())
	require.NoError(t, err)
	parts := strings.Split(token, ".")
	parts[0] = base64.RawURLEncoding.EncodeToString([]byte(header))
	return strings.Join(parts, ".")
}

func assertNumericDate(t *testing.T, expected, actual *jwt.NumericDate) {
	t.Helper()
	if expected == nil {
		assert.Nilf(t, actual, "expected nil, but is %v", actual)
		return
	}
	require.NotNil(t, actual, "expected not nil, but actual is nil")
	assert.Equal(t, expected.UTC(), actual.UTC())
}

func TestNewSigner_ReturnsSigner_WithKeyringAndClock(t *testing.T) {
	clock := porttest.NewFakeClock()
	keyring := keyringForTest(t)

	signer, err := NewSigner(SignerConfig{
		Keyring: keyring,
		Clock:   clock,
	})

	require.NoError(t, err)
	require.NotNil(t, signer)
	assert.Equal(t, keyring, signer.keyring)
	assert.Equal(t, clock, signer.clock)
}

func TestNewSigner_ReturnsError(t *testing.T) {
	testCases := []struct {
		desc string
		cfg  SignerConfig
	}{
		{
			desc: "clock nil",
			cfg:  SignerConfig{Keyring: keyringForTest(t), Clock: nil},
		},
		{
			desc: "keyring nil",
			cfg:  SignerConfig{Keyring: nil, Clock: porttest.NewFakeClock()},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			signer, err := NewSigner(tC.cfg)
			assert.Zero(t, signer)
			assert.Error(t, err)
		})
	}
}

func TestSigner_Sign_ReturnsError_WhenExpiresAtIsNotDefined(t *testing.T) {
	signer := signerForTest(t)

	token, err := signer.Sign(jwt.RegisteredClaims{
		Subject: "something",
	})

	assert.Zero(t, token)
	assert.Error(t, err)
}

func TestSigner_Sign_ReturnsTokenWithClaims(t *testing.T) {
	testCases := []struct {
		claims jwt.RegisteredClaims
	}{
		{
			claims: validClaims(),
		},
		{
			claims: jwt.RegisteredClaims{
				Subject:   "subject",
				ExpiresAt: jwt.NewNumericDate(time.Now().UTC()),
			},
		},
		{
			claims: jwt.RegisteredClaims{
				Subject:   "other subject",
				Issuer:    "issuer",
				ExpiresAt: jwt.NewNumericDate(time.Date(3000, 3, 3, 3, 3, 0, 0, time.UTC)),
			},
		},
		{
			claims: jwt.RegisteredClaims{
				Subject:   "another subject",
				ExpiresAt: jwt.NewNumericDate(time.Date(6000, 6, 6, 6, 6, 0, 0, time.UTC)),
				NotBefore: jwt.NewNumericDate(time.Date(7000, 7, 7, 7, 7, 0, 0, time.UTC)),
				IssuedAt:  jwt.NewNumericDate(time.Date(5000, 5, 5, 5, 5, 0, 0, time.UTC)),
			},
		},
		{
			claims: jwt.RegisteredClaims{
				ID:        "imagine there's a random uuid here",
				Subject:   "subject subject _ subscriber - subscrition = subprime2",
				ExpiresAt: jwt.NewNumericDate(time.Date(9000, 9, 9, 9, 9, 0, 0, time.UTC)),
				Audience:  jwt.ClaimStrings{"aud", "auraman", "egoman", "chaos"},
			},
		},
	}
	for i, tC := range testCases {
		t.Run(fmt.Sprintf("case %d", i), func(t *testing.T) {
			signer := signerForTest(t)

			token, err := signer.Sign(tC.claims)

			tokenParsed, _, err := jwt.NewParser().ParseUnverified(token, &jwt.RegisteredClaims{})
			require.NoError(t, err)
			claims := tokenParsed.Claims.(*jwt.RegisteredClaims)
			assert.Equal(t, tC.claims.Subject, claims.Subject)
			assert.Equal(t, tC.claims.Audience, claims.Audience)
			assert.Equal(t, tC.claims.Issuer, claims.Issuer)
			assert.Equal(t, tC.claims.ID, claims.ID)
			assertNumericDate(t, tC.claims.ExpiresAt, claims.ExpiresAt)
			assertNumericDate(t, tC.claims.IssuedAt, claims.IssuedAt)
			assertNumericDate(t, tC.claims.NotBefore, claims.NotBefore)
		})
	}
}

func TestSigner_Parse_ReturnsErrTokenInvalid(t *testing.T) {
	testCases := []struct {
		desc     string
		tokenFor func(t *testing.T, signer *Signer) string
	}{
		{
			desc: "malformed token",
			tokenFor: func(*testing.T, *Signer) string {
				return "not-a-jwt"
			},
		},
		{
			desc: "token without claims",
			tokenFor: func(t *testing.T, signer *Signer) string {
				raw, err := jwt.New(jwt.SigningMethodEdDSA).
					SignedString(signer.keyring.SigningKey().private)
				require.NoError(t, err)
				return raw
			},
		},
		{
			desc: "token signed with another key",
			tokenFor: func(t *testing.T, _ *Signer) string {
				otherSigner := signerForTest(t)
				token, err := otherSigner.Sign(jwt.RegisteredClaims{
					Subject:   "another subject",
					ExpiresAt: jwt.NewNumericDate(time.Now().UTC()),
				})
				require.NoError(t, err)
				return token
			},
		},
		{
			desc: "token signed with HS256 using the public key",
			tokenFor: func(t *testing.T, signer *Signer) string {
				pub := signer.keyring.SigningKey().public
				raw, err := jwt.NewWithClaims(jwt.SigningMethodHS256, validClaims()).
					SignedString([]byte(pub))
				require.NoError(t, err)
				return raw
			},
		},
		{
			desc: "token signed with alg none",
			tokenFor: func(t *testing.T, _ *Signer) string {
				raw, err := jwt.NewWithClaims(jwt.SigningMethodNone, validClaims()).
					SignedString(jwt.UnsafeAllowNoneSignatureType)
				require.NoError(t, err)
				return raw
			},
		},
		{
			desc: "token with an unknown alg",
			tokenFor: func(t *testing.T, signer *Signer) string {
				kid := signer.keyring.SigningKey().id
				return tokenWithHeader(t, signer, `{"alg":"ES999","kid":"`+kid+`","typ":"JWT"}`)
			},
		},
		{
			desc: "token without alg",
			tokenFor: func(t *testing.T, signer *Signer) string {
				kid := signer.keyring.SigningKey().id
				return tokenWithHeader(t, signer, `{"kid":"`+kid+`","typ":"JWT"}`)
			},
		},
		{
			desc: "token with payload modified",
			tokenFor: func(t *testing.T, signer *Signer) string {
				validToken, err := signer.Sign(jwt.RegisteredClaims{
					Subject:   "sub",
					ExpiresAt: jwt.NewNumericDate(time.Now()),
				})
				require.NoError(t, err)

				parts := strings.Split(validToken, ".")
				payload, err := base64.RawURLEncoding.DecodeString(parts[1])
				require.NoError(t, err)

				var claims map[string]any
				require.NoError(t, json.Unmarshal(payload, &claims))
				claims["sub"] = "another-sub"

				modifiedPayload, err := json.Marshal(claims)
				require.NoError(t, err)

				parts[1] = base64.RawURLEncoding.EncodeToString(modifiedPayload)
				modifiedToken := strings.Join(parts, ".")

				return modifiedToken
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			signer := signerForTest(t)

			claims, err := signer.Parse(tC.tokenFor(t, signer))

			assert.Zero(t, claims)
			assert.ErrorIs(t, err, ErrTokenInvalid)
		})
	}
}

func TestSigner_Parse_ReturnsErrTokenExpired(t *testing.T) {
	signer, fakeClock := signerAndFakeClockForTest(t)

	claims := validClaims()
	claims.ExpiresAt = jwt.NewNumericDate(fakeClock.Now().Add(-time.Second))
	token, err := signer.Sign(claims)
	require.NoError(t, err)

	retrievedClaims, err := signer.Parse(token)

	assert.Zero(t, retrievedClaims)
	assert.ErrorIs(t, err, ErrTokenExpired)
}
