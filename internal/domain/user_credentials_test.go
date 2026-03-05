package domain_test

import (
	"testing"
	"time"

	"github.com/rafaelblt/go-auth/internal/domain"
	"github.com/rafaelblt/go-auth/internal/testutil"
	"github.com/stretchr/testify/assert"
)

func passwordCredential() domain.PasswordCredential {
	return testutil.Must(domain.NewPasswordCredential(hashedPassword(), time.Now().UTC()))
}

func TestNewWithPassword_ShouldReturnUserCredentials(t *testing.T) {
	userID := domain.NewUserID()
	password := passwordCredential()

	credentials, err := domain.NewUserCredentialsWithPassword(userID, password)

	assert.NoError(t, err)
	assert.Equal(t, userID, credentials.UserID())
	assert.Equal(t, password, credentials.Password())
}

func TestNewWithPassword_ShouldReturnError_WhenUserIDZero(t *testing.T) {
	userID := domain.UserID{}
	password := passwordCredential()

	credentials, err := domain.NewUserCredentialsWithPassword(userID, password)

	assert.ErrorIs(t, err, domain.ErrUserIDZero)
	assert.Zero(t, credentials)
}

func TestNewWithPassword_ShouldReturnError_WhenPasswordCredentialZero(t *testing.T) {
	userID := domain.NewUserID()
	password := domain.PasswordCredential{}

	credentials, err := domain.NewUserCredentialsWithPassword(userID, password)
	
	assert.ErrorIs(t, err, domain.ErrPasswordCredentialZero)
	assert.Zero(t, credentials)
}
