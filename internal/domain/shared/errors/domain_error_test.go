package errors

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewDomainError(t *testing.T) {
	code, msg := "code", "message"

	err := NewDomainError(code, msg)

	assert.Equal(t, err.Code(), code)
	assert.Equal(t, err.Message(), msg)
}
