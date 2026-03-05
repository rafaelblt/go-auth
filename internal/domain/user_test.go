package domain_test

import (
	"testing"
	"time"

	"github.com/rafaelblt/go-auth/internal/domain"
	"github.com/rafaelblt/go-auth/internal/testutil"
	"github.com/stretchr/testify/assert"
)

func username() domain.Username {
	return testutil.Must(domain.NewUsername("valid"))
}

func TestNewUserWithPassword_ShouldReturnUserAndCredentials(t *testing.T) {
	username := username()
	hashed := hashedPassword()
	created := time.Now().UTC()

	user, creds, err := domain.NewUserWithPassword(username, hashed, created)

	assert.NoError(t, err)
	// user
	assert.NotZero(t, user.ID())
	assert.Equal(t, username, user.Username())
	assert.Equal(t, domain.UserStatusActive, user.Status())
	assert.Equal(t, created, user.CreatedAt())
	assert.Equal(t, created, user.UpdatedAt())
	// credentials
	assert.Equal(t, user.ID(), creds.UserID())
	assert.Equal(t, hashed, creds.Password().Hashed())
	assert.Equal(t, created, creds.Password().CreatedAt())
}

func TestNewUserWithPassword_ShouldReturnError_WhenUsernameZero(t *testing.T) {
	username := domain.Username{}
	hashed := hashedPassword()
	created := time.Now().UTC()

	user, creds, err := domain.NewUserWithPassword(username, hashed, created)

	assert.ErrorIs(t, err, domain.ErrUsernameZero)
	assert.Zero(t, user)
	assert.Zero(t, creds)
}

func TestNewUserWithPassword_ShouldReturnError_WhenHashedPasswordZero(t *testing.T) {
	username := username()
	hashed := domain.HashedPassword{}
	created := time.Now().UTC()

	user, creds, err := domain.NewUserWithPassword(username, hashed, created)

	assert.ErrorIs(t, err, domain.ErrHashedPasswordZero)
	assert.Zero(t, user)
	assert.Zero(t, creds)
}
