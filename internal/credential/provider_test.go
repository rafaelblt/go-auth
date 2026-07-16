package credential

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewProvider(t *testing.T) {
	testCases := []struct {
		desc      string
		value     string
		expectErr bool
	}{
		{desc: "empty value", value: "", expectErr: true},
		{desc: "reserved provider", value: "local", expectErr: true},
		{desc: "valid value", value: "github", expectErr: false},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			provider, err := NewProvider(tC.value)
			if tC.expectErr {
				assert.Error(t, err)
				assert.Zero(t, provider)
			} else {
				require.NoError(t, err)
				require.NotZero(t, provider)
				assert.Equal(t, provider.String(), tC.value)
			}
		})
	}
}

func TestParseProvider(t *testing.T) {
	testCases := []struct {
		desc     string
		value    string
		expected string
		wantErr  bool
	}{
		{
			desc:     "custom provider",
			value:    "rafaelblt.com",
			expected: "rafaelblt.com",
			wantErr:  false,
		},
		{
			desc:     "custom provider with uppercase and spaces",
			value:    "   GITHUB.COM    ",
			expected: "github.com",
			wantErr:  false,
		},
		{
			desc:     "local provider",
			value:    ProviderLocal.String(),
			expected: ProviderLocal.String(),
			wantErr:  false,
		},
		{
			desc:     "empty provider",
			value:    "",
			wantErr:  true,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			provider, err := ParseProvider(tC.value)
			if tC.wantErr {
				assert.Error(t, err)
				assert.Zero(t, provider)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tC.expected, provider.String())
		})
	}
}
