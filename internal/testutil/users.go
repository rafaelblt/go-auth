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

func mustUsername(t *testing.T, value string) domain.Username {
	t.Helper()
	username, err := domain.NewUsername(value)
	require.NoError(t, err)
	return username
}

func DefaultUser(t *testing.T) TestUser {
	t.Helper()
	entity, err := domain.RestoreUser(
		domain.UserRestoreParams{
			ID: domain.NewUserID(),
			Username: mustUsername(t, "default_user"),
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
	entity, err := domain.RestoreUser(
		domain.UserRestoreParams{
			ID: domain.NewUserID(),
			Username: mustUsername(t, "other_user"),
			Status: domain.UserStatusActive,
			CreatedAt: time.Date(2000, 2, 20, 0, 0, 0, 0, time.UTC),
			UpdatedAt: time.Date(2000, 2, 20, 0, 0, 0, 0, time.UTC),
		},
	)
	require.NoError(t, err)
	return TestUser{Entity: entity}
}
