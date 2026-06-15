package refreshtoken

import (
	"testing"

	"github.com/rafaelblt/go-auth/internal/shared"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerator_ReturnsRefreshTokenGenerated(t *testing.T) {
	g := NewGenerator()
	generated, err := g.Generate()

	assert.NoError(t, err)
	assert.NotZero(t, generated.Raw)
	assert.NotZero(t, generated.Hash)
}

func TestGenerator_ReturnsDifferentResults(t *testing.T) {
	g := NewGenerator()

	raws := shared.NewSet[string]()
	hashes := shared.NewSet[string]()

	total := 100000
	for range total {
		generated, err := g.Generate()
		require.NoError(t, err)
		raws.Add(generated.Raw)
		hashes.Add(generated.Hash.Value())
	}

	assert.Equal(t, total, raws.Len())
	assert.Equal(t, total, hashes.Len())
}
