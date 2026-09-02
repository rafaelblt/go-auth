package user_test

import (
	"testing"
	"time"

	"github.com/rafaelblt/go-auth/internal/testutil/usertest"
	"github.com/rafaelblt/go-auth/internal/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewUser(t *testing.T) {
	username := usertest.MustUsername(t, "username")
	testCases := []struct {
		desc      string
		params    user.CreationParams
		expectErr bool
	}{
		{
			desc: "valid case",
			params: user.CreationParams{
				Username:  username,
				CreatedAt: time.Now().UTC(),
			},
			expectErr: false,
		},
		{
			desc: "username zero",
			params: user.CreationParams{
				Username:  user.Username{},
				CreatedAt: time.Now().UTC(),
			},
			expectErr: true,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			usr, err := user.NewUser(tC.params)
			if tC.expectErr {
				assert.Error(t, err)
				assert.Nil(t, usr)
			} else {
				require.NoError(t, err)
				require.NotNil(t, usr)
				assert.NotZero(t, usr.ID())
				assert.Equal(t, tC.params.Username, usr.Username())
				assert.Equal(t, user.StatusActive, usr.Status())
				assert.Equal(t, tC.params.CreatedAt, usr.CreatedAt())
				assert.Equal(t, tC.params.CreatedAt, usr.UpdatedAt())
			}
		})
	}
}

func TestRestoreUser(t *testing.T) {
	username := usertest.MustUsername(t, "username")
	testCases := []struct {
		desc      string
		params    user.RestoreParams
		expectErr bool
	}{
		{
			desc: "valid case",
			params: user.RestoreParams{
				ID:        user.NewID(),
				Username:  username,
				Status:    user.StatusActive,
				CreatedAt: time.Now().UTC(),
				UpdatedAt: time.Now().UTC(),
			},
			expectErr: false,
		},
		{
			desc: "user id zero",
			params: user.RestoreParams{
				ID:        user.ID{},
				Username:  username,
				Status:    user.StatusActive,
				CreatedAt: time.Now().UTC(),
				UpdatedAt: time.Now().UTC(),
			},
			expectErr: true,
		},
		{
			desc: "username zero",
			params: user.RestoreParams{
				ID:        user.NewID(),
				Username:  user.Username{},
				Status:    user.StatusActive,
				CreatedAt: time.Now().UTC(),
				UpdatedAt: time.Now().UTC(),
			},
			expectErr: true,
		},
		{
			desc: "status zero",
			params: user.RestoreParams{
				ID:        user.NewID(),
				Username:  username,
				Status:    user.Status{},
				CreatedAt: time.Now().UTC(),
				UpdatedAt: time.Now().UTC(),
			},
			expectErr: true,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			user, err := user.RestoreUser(tC.params)
			if tC.expectErr {
				assert.Error(t, err)
				assert.Nil(t, user)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tC.params.ID, user.ID())
				assert.Equal(t, tC.params.Username, user.Username())
				assert.Equal(t, tC.params.Status, user.Status())
				assert.Equal(t, tC.params.CreatedAt, user.CreatedAt())
				assert.Equal(t, tC.params.UpdatedAt, user.UpdatedAt())
			}
		})
	}
}

func TestUser_ChangeUsername(t *testing.T) {
	testCases := []struct {
		desc      string
		user      *user.User
		username  user.Username
		updatedAt time.Time
		expectErr bool
	}{
		{
			desc:      "valid change",
			user:      usertest.NewUser(t, nil),
			username:  usertest.MustUsername(t, "other_username"),
			updatedAt: time.Now().UTC(),
			expectErr: false,
		},
		{
			desc:      "username zero",
			user:      usertest.NewUser(t, nil),
			username:  user.Username{},
			updatedAt: time.Now().UTC(),
			expectErr: true,
		},
		{
			desc:      "updated at before created at",
			user:      usertest.NewUser(t, nil),
			username:  usertest.MustUsername(t, "other_username"),
			updatedAt: time.Date(1, 1, 1, 1, 1, 0, 0, time.UTC),
			expectErr: true,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			err := tC.user.ChangeUsername(tC.username, tC.updatedAt)
			if tC.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tC.username, tC.user.Username())
				assert.Equal(t, tC.updatedAt, tC.user.UpdatedAt())
			}
		})
	}
}

func TestUser_IsZero(t *testing.T) {
	testCases := []struct {
		desc   string
		user   *user.User
		isZero bool
	}{
		{
			desc:   "user zero",
			user:   &user.User{},
			isZero: true,
		},
		{
			desc:   "valid user",
			user:   usertest.NewUser(t, nil),
			isZero: false,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			result := tC.user.IsZero()
			assert.Equal(t, tC.isZero, result)
		})
	}
}
