package verify_test

import (
	"testing"

	"github.com/rafaelblt/go-auth/internal/testutil/porttest"
	"github.com/rafaelblt/go-auth/internal/usecase/verify"
	"github.com/stretchr/testify/require"
)

type TestHelper struct {
	t                        *testing.T
	FakeAccessTokenValidator *porttest.FakeAccessTokenValidator
	FakeClock                *porttest.FakeClock
}

func NewTestHelper(t *testing.T) TestHelper {
	helper := TestHelper{
		t:                        t,
		FakeAccessTokenValidator: porttest.NewFakeAccessTokenValidator(),
		FakeClock:                porttest.NewFakeClock(),
	}
	return helper
}

func (helper *TestHelper) UseCase() *verify.Verify {
	helper.t.Helper()
	uc, err := verify.New(verify.Config{
		AccessTokenValidator: helper.FakeAccessTokenValidator,
		Clock:                helper.FakeClock,
	})
	require.NoError(helper.t, err)
	return uc
}
