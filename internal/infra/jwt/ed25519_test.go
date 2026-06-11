package jwt

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/rafaelblt/go-auth/internal/testutil/porttest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func generatePrivateKey(t *testing.T) ed25519.PrivateKey {
	_, priv, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)
	return priv
}

func edSigner(t *testing.T) *Ed25519Signer {
	key := generatePrivateKey(t)
	clock := porttest.NewFakeClock()
	s, err := NewEd25519(Ed25519Config{
		PrivateKey: key,
		Clock:      clock,
	})
	require.NoError(t, err)
	return s
}

func validClaims() claims {
	return claims{
		Subject:   "subject",
		ExpiresAt: time.Now().Add(time.Hour).Truncate(time.Second),
	}
}

func TestNewEd25519_ReturnsSigner(t *testing.T) {
	priv := generatePrivateKey(t)
	clock := porttest.NewFakeClock()

	signer, err := NewEd25519(Ed25519Config{
		PrivateKey: priv,
		Clock:      clock,
	})

	require.NoError(t, err)
	require.NotNil(t, signer)
	assert.Equal(t, clock, signer.clock)
	assert.Equal(t, priv, signer.privateKey)
	expectedPubKey := priv.Public().(ed25519.PublicKey)
	assert.Equal(t, expectedPubKey, signer.publicKey)
}

func TestNewEd25519_ReturnsError_WhenPrivateKeyIsInvalid(t *testing.T) {
	signer, err := NewEd25519(Ed25519Config{
		PrivateKey: ed25519.PrivateKey{},
	})

	require.Error(t, err)
	require.Nil(t, signer)
}

func TestEd25519_Sign_ReturnsToken(t *testing.T) {
	signer := edSigner(t)
	expectedClaims := validClaims()

	token, err := signer.sign(expectedClaims)

	require.NoError(t, err)
	require.NotZero(t, token)
}

func TestEd25519_Sign_ReturnsError_WhenSubjectIsEmpty(t *testing.T) {
	signer := edSigner(t)

	token, err := signer.sign(claims{
		Subject:   "",
		ExpiresAt: time.Now(),
	})

	assert.Error(t, err)
	assert.Zero(t, token)
}

func TestEd25519_Sign_ReturnsError_WhenExpiresAtIsZero(t *testing.T) {
	signer := edSigner(t)

	token, err := signer.sign(claims{
		Subject:   "aaaa",
		ExpiresAt: time.Time{},
	})

	assert.Error(t, err)
	assert.Zero(t, token)
}

func TestEd25519_Parse_ReturnsError_ForRandomToken(t *testing.T) {
	signer := edSigner(t)

	claims, err := signer.parse("random")

	assert.ErrorIs(t, err, errTokenInvalid)
	assert.Zero(t, claims)
}

func TestEd25519_Parse_ReturnsClaims_ForValidToken(t *testing.T) {
	signer := edSigner(t)
	expectedClaims := validClaims()

	valid, err := signer.sign(expectedClaims)
	require.NoError(t, err)

	actualClaims, err := signer.parse(valid)

	require.NoError(t, err)
	require.NotZero(t, actualClaims)
	assert.Equal(t, expectedClaims, actualClaims)
}

func TestEd25519_Parse_ReturnsError_ForTokenWithNoClaims(t *testing.T) {
	signer := edSigner(t)
	token, err := jwt.New(jwt.SigningMethodEdDSA).SignedString(signer.privateKey)
	require.NoError(t, err)

	claims, err := signer.parse(token)

	assert.ErrorIs(t, err, errTokenInvalid)
	assert.Zero(t, claims)
}

func TestEd25519_Parse_ReturnsError_ForTokenWithDifferentKey(t *testing.T) {
	signer1 := edSigner(t)
	signer2 := edSigner(t)

	token, err := signer1.sign(validClaims())
	require.NoError(t, err)

	claims, err := signer2.parse(token)

	assert.ErrorIs(t, err, errTokenInvalid)
	assert.Zero(t, claims)
}

func TestEd25519_Parse_ReturnsError_ForTokenWithDifferentMethod(t *testing.T) {
	signer := edSigner(t)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{Subject: "sub"})
	raw, err := token.SignedString([]byte(signer.publicKey))
	require.NoError(t, err)

	_, err = signer.parse(raw)

	assert.ErrorIs(t, err, errTokenInvalid)
}

func TestEd25519_Parse_ReturnsError_ForTokenWithAlgNone(t *testing.T) {
	signer := edSigner(t)

	token := jwt.NewWithClaims(jwt.SigningMethodNone, jwt.RegisteredClaims{Subject: "sub"})
	raw, err := token.SignedString(jwt.UnsafeAllowNoneSignatureType)
	require.NoError(t, err)

	_, err = signer.parse(raw)

	assert.ErrorIs(t, err, errTokenInvalid)
}

func TestEd25519_Parse_ReturnsError_ForTokenExpired(t *testing.T) {
	fakeClock := porttest.NewFakeClock()
	signer := edSigner(t)
	signer.clock = fakeClock

	claims := validClaims()
	claims.ExpiresAt = fakeClock.Now().Add(-1)
	token, err := signer.sign(claims)
	require.NoError(t, err)

	_, err = signer.parse(token)

	assert.ErrorIs(t, err, errTokenExpired)
}

func TestEd25519_Parse_ReturnsError_ForTokenWithExpiresAtEqualsNow(t *testing.T) {
	fakeClock := porttest.NewFakeClock()
	signer := edSigner(t)
	signer.clock = fakeClock

	claims := validClaims()
	claims.ExpiresAt = fakeClock.Now()
	token, err := signer.sign(claims)
	require.NoError(t, err)

	_, err = signer.parse(token)

	assert.ErrorIs(t, err, errTokenExpired)
}

func TestEd25519_Parse_ReturnsError_ForTokenWithPayloadModified(t *testing.T) {
	// Arrange
	signer := edSigner(t)

	validToken, err := signer.sign(claims{
		Subject:   "sub",
		ExpiresAt: time.Now(),
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

	// Act
	_, err = signer.parse(modifiedToken)

	// Assert
	assert.ErrorIs(t, err, errTokenInvalid)
}
