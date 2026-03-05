package domain_test

import (
	"testing"

	"github.com/rafaelblt/go-auth/internal/domain"
	"github.com/stretchr/testify/assert"
)

func TestNewHashedPassword_ShouldReturnObject_WhenHashIsValid(t *testing.T) {
	hash := "masd9of8j)!@Pkjfsda[fl[!@+_)ief[sdapfç]asd="
	obj, err := domain.NewHashedPassword(hash)
	assert.NoError(t, err)
	assert.Equal(t, hash, obj.Value())
}

func TestNewHashedPassword_ShouldReturnEmptyError_WhenHashIsEmpty(t *testing.T) {
	hash := ""
	obj, err := domain.NewHashedPassword(hash)
	assert.Error(t, err)
	assert.EqualError(t, domain.ErrHashedPasswordEmpty, err.Error())
	assert.Empty(t, hash, obj.Value())
}

func TestHashedPasswordIsZero_ShouldReturnTrue_WhenIsZero(t *testing.T) {
	zeroHash := domain.HashedPassword{}
	assert.True(t, zeroHash.IsZero())
}

func TestHashedPasswordIsZero_ShouldReturnFalse_WhenIsNotZero(t *testing.T) {
	hash, _ := domain.NewHashedPassword("hash")
	assert.False(t, hash.IsZero())
}
