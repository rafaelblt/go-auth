package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var validValues = []string{
	"blatantss",
}

func TestNewStringVO_ShouldReturnStringVO_WhenInputIsValid(t *testing.T) {
	for _, input := range validValues {
		t.Run(input, func(t *testing.T) {
			vo, err := newStringVO(input)
			assert.NoError(t, err)
			assert.NotEmpty(t, vo)
			assert.Equal(t, input, vo.String())
		})
	}
}

func TestNewStringVO_ShouldReturnEmptyError_WhenInputIsEmpty(t *testing.T) {
	input := ""

	vo, err := newStringVO(input)

	assert.Empty(t, vo)
	assert.Error(t, err)
	assert.ErrorIs(t, err, errStringVOEmpty)
}
