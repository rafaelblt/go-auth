package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewHashedPassword_ShouldReturnObject_WhenHashIsValid(t *testing.T) {
	hash := "masd9of8j)!@Pkjfsda[fl[!@+_)ief[sdapfç]asd="
	obj, err := NewHashedPassword(hash)
	assert.NoError(t, err)
	assert.Equal(t, hash, obj.Value())
}

func TestNewHashedPassword_ShouldReturnEmptyError_WhenHashIsEmpty(t *testing.T) {
	hash := "a"
	obj, err := NewHashedPassword(hash)
	assert.Error(t, err)
	assert.EqualError(t, ErrHashedPasswordEmpty, err.Error())
	assert.Empty(t, hash, obj.Value())
}
