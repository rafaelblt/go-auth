package domain_test

import (
	"testing"
	"time"

	"github.com/rafaelblt/go-auth/internal/domain"
	"github.com/rafaelblt/go-auth/internal/testutil"
	"github.com/stretchr/testify/assert"
)

func hashedPassword() domain.HashedPassword {
	return testutil.Must(domain.NewHashedPassword("hash"))
}

func TestNewPasswordCredential_ShouldReturnPasswordCredential(t *testing.T) {
	hash := hashedPassword()
	createdAt := time.Now()

	credential, err := domain.NewPasswordCredential(hash, createdAt)

	assert.NoError(t, err)
	assert.NotZero(t, credential.ID())
	assert.Equal(t, hash, credential.Hashed())
	assert.Equal(t, createdAt, credential.CreatedAt())
}

func TestNewPasswordCredential_ShouldReturnError_WhenHashedPasswordIsZero(t *testing.T) {
	hash := domain.HashedPassword{}
	createdAt := time.Now()

	credential, err := domain.NewPasswordCredential(hash, createdAt)

	assert.Zero(t, credential)
	assert.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrHashedPasswordZero)
}
