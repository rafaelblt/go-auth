package domain_test

import (
	"testing"

	"github.com/rafaelblt/go-auth/internal/domain"
	"github.com/stretchr/testify/assert"
)

func TestNewDomainError(t *testing.T) {
	code, msg := "code", "message"

	err := domain.NewDomainError(code, msg)

	assert.Equal(t, err.Code(), code)
	assert.Equal(t, err.Message(), msg)
}
