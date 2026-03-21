package domain_test

import (
	"testing"
	"time"

	"github.com/rafaelblt/go-auth/internal/domain"
	"github.com/rafaelblt/go-auth/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewUser(t *testing.T) {
	validUsername, err := domain.NewUsername("username")
	require.NoError(t, err)
	testCases := []struct {
		desc      string
		params    domain.NewUserParams
		expectErr bool
	}{
		{
			desc: "valid case",
			params: domain.NewUserParams{
				Username:  validUsername,
				CreatedAt: time.Now().UTC(),
			},
			expectErr: false,
		},
		{
			desc: "username zero",
			params: domain.NewUserParams{
				Username:  domain.Username{},
				CreatedAt: time.Now().UTC(),
			},
			expectErr: true,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			user, err := domain.NewUser(tC.params)
			if tC.expectErr {
				assert.Error(t, err)
				assert.Nil(t, user)
			} else {
				require.NoError(t, err)
				require.NotNil(t, user)
				assert.NotZero(t, user.ID())
				assert.Equal(t, tC.params.Username, user.Username())
				assert.Equal(t, domain.UserStatusActive, user.Status())
				assert.Equal(t, tC.params.CreatedAt, user.CreatedAt())
				assert.Equal(t, tC.params.CreatedAt, user.UpdatedAt())
			}
		})
	}
}

func TestRestoreUser(t *testing.T) {
	validUsername, err := domain.NewUsername("username")
	require.NoError(t, err)
	testCases := []struct {
		desc      string
		params    domain.UserRestoreParams
		expectErr bool
	}{
		{
			desc: "valid case",
			params: domain.UserRestoreParams{
				ID:        domain.NewUserID(),
				Username:  validUsername,
				Status:    domain.UserStatusActive,
				CreatedAt: time.Now().UTC(),
				UpdatedAt: time.Now().UTC(),
			},
			expectErr: false,
		},
		{
			desc: "user id zero",
			params: domain.UserRestoreParams{
				ID:        domain.UserID{},
				Username:  validUsername,
				Status:    domain.UserStatusActive,
				CreatedAt: time.Now().UTC(),
				UpdatedAt: time.Now().UTC(),
			},
			expectErr: true,
		},
		{
			desc: "username zero",
			params: domain.UserRestoreParams{
				ID:        domain.NewUserID(),
				Username:  domain.Username{},
				Status:    domain.UserStatusActive,
				CreatedAt: time.Now().UTC(),
				UpdatedAt: time.Now().UTC(),
			},
			expectErr: true,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			user, err := domain.RestoreUser(tC.params)
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

func TestChangeUsername(t *testing.T) {
	testCases := []struct {
		desc      string
		user      *domain.User
		username  domain.Username
		updatedAt time.Time
		expectErr bool
	}{
		{
			desc:      "valid update",
			user:      testutil.DefaultUser(t).Entity,
			username:  testutil.OtherUser(t).Entity.Username(),
			updatedAt: time.Now().UTC(),
			expectErr: false,
		},
		{
			desc:      "username zero",
			user:      testutil.DefaultUser(t).Entity,
			username:  domain.Username{},
			updatedAt: time.Now().UTC(),
			expectErr: true,
		},
		{
			desc:      "updated at before created at",
			user:      testutil.DefaultUser(t).Entity,
			username:  testutil.OtherUser(t).Entity.Username(),
			updatedAt: testutil.DefaultUser(t).Entity.CreatedAt().Add(-1),
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

func TestIsZero(t *testing.T) {
	testCases := []struct {
		desc   string
		user   *domain.User
		isZero bool
	}{
		{
			desc:   "user zero",
			user:   &domain.User{},
			isZero: true,
		},
		{
			desc:   "valid user",
			user:   testutil.DefaultUser(t).Entity,
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
