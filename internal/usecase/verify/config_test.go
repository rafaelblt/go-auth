package verify_test

import (
	"testing"

	"github.com/rafaelblt/go-auth/internal/usecase/verify"
	"github.com/stretchr/testify/assert"
)

func TestNewVerify(t *testing.T) {
	helper := NewTestHelper(t)
	testCases := []struct {
		desc      string
		config    verify.Config
		expectErr bool
	}{
		{
			desc: "valid case",
			config: verify.Config{
				AccessTokenValidator: helper.FakeAccessTokenValidator,
				Clock:                helper.FakeClock,
			},
			expectErr: false,
		},
		{
			desc: "access token validator nil",
			config: verify.Config{
				AccessTokenValidator: nil,
				Clock:                helper.FakeClock,
			},
			expectErr: true,
		},
		{
			desc: "clock nil",
			config: verify.Config{
				AccessTokenValidator: helper.FakeAccessTokenValidator,
				Clock:                nil,
			},
			expectErr: true,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			uc, err := verify.New(tC.config)
			if tC.expectErr {
				assert.Error(t, err)
				assert.Zero(t, uc)
				return
			}
			assert.NoError(t, err)
			assert.NotZero(t, uc)
		})
	}
}
