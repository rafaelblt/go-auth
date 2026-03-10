package testutil

import (
	"testing"
	"time"

	"github.com/rafaelblt/go-auth/internal/domain"
	"github.com/stretchr/testify/require"
)

type TestUser struct {
	Entity *domain.User
}

func DefaultUser(t *testing.T) TestUser {
	t.Helper()
	username, err := domain.NewUsername("default_user")
	require.NoError(t, err)
	entity, err := domain.RestoreUser(
		domain.UserRestoreParams{
			ID: domain.NewUserID(),
			Username: username,
			Status: domain.UserStatusActive,
			CreatedAt: time.Date(2026, 3, 10, 16, 0, 0, 0, time.UTC),
			UpdatedAt: time.Date(2026, 3, 10, 16, 0, 0, 0, time.UTC),
		},
	)
	require.NoError(t, err)
	return TestUser{Entity: entity}
}

func OtherUser(t *testing.T) TestUser {
	t.Helper()
	username, err := domain.NewUsername("other_user")
	require.NoError(t, err)
	entity, err := domain.RestoreUser(
		domain.UserRestoreParams{
			ID: domain.NewUserID(),
			Username: username,
			Status: domain.UserStatusActive,
			CreatedAt: time.Date(2000, 2, 20, 0, 0, 0, 0, time.UTC),
			UpdatedAt: time.Date(2000, 2, 20, 0, 0, 0, 0, time.UTC),
		},
	)
	require.NoError(t, err)
	return TestUser{Entity: entity}
}
