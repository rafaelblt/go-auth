package me_test

import (
	"testing"

	"github.com/rafaelblt/go-auth/internal/domain/user"
	"github.com/rafaelblt/go-auth/internal/testutil/porttest"
	"github.com/rafaelblt/go-auth/internal/testutil/usertest"
	"github.com/rafaelblt/go-auth/internal/usecase/me"
	"github.com/stretchr/testify/require"
)

type TestHelper struct {
	t                        *testing.T
	FakeAccessTokenValidator *porttest.FakeAccessTokenValidator
	FakeUserReader           *porttest.FakeUserReader
}

func NewTestHelper(t *testing.T) TestHelper {
	helper := TestHelper{
		t:                        t,
		FakeAccessTokenValidator: porttest.NewFakeAccessTokenValidator(),
		FakeUserReader:           porttest.NewFakeUserReader(),
	}
	return helper
}

func (helper *TestHelper) UseCase() *me.Me {
	helper.t.Helper()
	uc, err := me.New(me.Config{
		AccessTokenValidator: helper.FakeAccessTokenValidator,
		UserReader:           helper.FakeUserReader,
	})
	require.NoError(helper.t, err)
	return uc
}

// AddUserOfToken stores a user with the ID the validator's claims name, and
// returns it.
func (helper *TestHelper) AddUserOfToken() *user.User {
	helper.t.Helper()
	usr := usertest.NewUser(helper.t, func(p *user.RestoreParams) {
		p.ID = helper.FakeAccessTokenValidator.Claims().UserID
		p.Username = usertest.MustUsername(helper.t, "alice")
	})
	helper.FakeUserReader.InsertUser(usr)
	return usr
}
