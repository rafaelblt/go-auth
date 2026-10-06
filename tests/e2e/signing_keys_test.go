package e2e

import (
	"context"
	"crypto/ed25519"
	"encoding/base64"
	"fmt"
	"net/http"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/rafaelblt/go-auth/internal/config"
	"github.com/rafaelblt/go-auth/internal/shared"
	"github.com/rafaelblt/go-auth/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// keyByKid returns a keyfunc that selects the key by the token's kid, as a
// verifier does.
func keyByKid(jwks JWKSResponseBody) jwt.Keyfunc {
	return func(token *jwt.Token) (any, error) {
		kid, _ := token.Header["kid"].(string)
		for _, key := range jwks.Keys {
			if key.Kid == kid {
				x, err := base64.RawURLEncoding.DecodeString(key.X)
				if err != nil {
					return nil, fmt.Errorf("raw url decode key failed: %w", err)
				}
				return ed25519.PublicKey(x), nil
			}
		}
		return nil, fmt.Errorf("kid %q not in the JWKS", kid)
	}
}

func kids(jwks JWKSResponseBody) []string {
	kids := make([]string, 0, len(jwks.Keys))
	for _, key := range jwks.Keys {
		kids = append(kids, key.Kid)
	}
	return kids
}

func TestSigningKeys_AreSharedBetweenInstances(t *testing.T) {
	env := testApp.NewEnv(t)
	client := startSecondApp(t, config.ConfigParams{})

	usr, pwd := env.Fixtures.CreateUserAndPassword(t)
	resp := client.Post(t, LoginPath, LoginRequestBody{
		Username: usr.Username().String(),
		Password: pwd.Value(),
	})
	require.Equal(t, http.StatusOK, resp.StatusCode)
	accessToken := DecodeBody[LoginResponseBody](t, resp).AccessToken.Value

	sharedJWKS := DecodeBody[JWKSResponseBody](t, env.Client.Get(t, JWKSPath))
	secondJWKS := DecodeBody[JWKSResponseBody](t, client.Get(t, JWKSPath))
	require.NotEmpty(t, sharedJWKS.Keys)
	assert.ElementsMatch(t, kids(sharedJWKS), kids(secondJWKS))

	token, err := jwt.ParseWithClaims(
		accessToken,
		&jwt.RegisteredClaims{},
		keyByKid(sharedJWKS),
		jwt.WithValidMethods([]string{"EdDSA"}),
		jwt.WithExpirationRequired(),
	)
	require.NoError(t, err, "token from the second app does not verify against the shared app's JWKS")
	claims := token.Claims.(*jwt.RegisteredClaims)
	assert.Equal(t, usr.ID().String(), claims.Subject)
}

// The suite's database is migrated before its app starts, so only this test
// starts an app that migrates the database itself, as the quick start does.
func TestSigningKeys_FirstKeyIsAdded_WhenTheAppMigratesAnEmptyDatabase(t *testing.T) {
	db := testutil.NewDatabaseForTest(t, context.Background())

	client := startSecondApp(t, config.ConfigParams{
		DatabaseURL: db.ConnectionString(),
		AutoMigrate: shared.Ptr(true),
	})

	resp := client.Get(t, JWKSPath)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	jwks := DecodeBody[JWKSResponseBody](t, resp)
	assert.Len(t, jwks.Keys, 1)
}
