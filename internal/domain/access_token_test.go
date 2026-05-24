package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewAccessToken(t *testing.T) {
	testCases := []struct {
		desc       string
		value      string
		expectErr  bool
	}{
		{
			desc:      "empty value",
			value:     "",
			expectErr: true,
		},
		{
			desc:       "valid value",
			value:      "abc",
			expectErr:  false,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			token, err := NewAccessToken(tC.value)
			if tC.expectErr {
				assert.Error(t, err)
				assert.Zero(t, token)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tC.value, token.Value())
		})
	}
}
