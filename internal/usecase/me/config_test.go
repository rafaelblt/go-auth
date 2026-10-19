package me_test

import (
	"testing"

	"github.com/rafaelblt/go-auth/internal/usecase/me"
	"github.com/stretchr/testify/assert"
)

func TestNewMe(t *testing.T) {
	helper := NewTestHelper(t)
	testCases := []struct {
		desc      string
		config    me.Config
		expectErr bool
	}{
		{
			desc: "valid case",
			config: me.Config{
				AccessTokenValidator: helper.FakeAccessTokenValidator,
				UserReader:           helper.FakeUserReader,
			},
			expectErr: false,
		},
		{
			desc: "access token validator nil",
			config: me.Config{
				AccessTokenValidator: nil,
				UserReader:           helper.FakeUserReader,
			},
			expectErr: true,
		},
		{
			desc: "user reader nil",
			config: me.Config{
				AccessTokenValidator: helper.FakeAccessTokenValidator,
				UserReader:           nil,
			},
			expectErr: true,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			uc, err := me.New(tC.config)
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
