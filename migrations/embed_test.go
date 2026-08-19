package migrations

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLatest(t *testing.T) {
	latest, err := Latest()
	assert.NoError(t, err)
	assert.Equal(t, uint(4), latest)
}
