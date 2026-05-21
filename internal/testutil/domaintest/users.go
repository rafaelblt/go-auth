package domaintest

import (
	"testing"
	"time"

	"github.com/rafaelblt/go-auth/internal/domain"
	"github.com/stretchr/testify/require"
)

func NewUser(t *testing.T, override func(*domain.UserRestoreParams)) *domain.User {
	t.Helper()

    params := domain.UserRestoreParams{
        ID:        domain.NewUserID(),
        Username:  MustUsername(t, "Default User"),
        Status:    domain.UserStatusActive,
        CreatedAt: time.Date(2007, 8, 9, 20, 45, 0, 0, time.UTC),
        UpdatedAt: time.Date(2026, 5, 21, 16, 0, 7, 0, time.UTC),
    }

    if override != nil {
        override(&params)
    }

    entity, err := domain.RestoreUser(params)
    require.NoError(t, err)
    return entity
}

func MustUsername(t *testing.T, value string) domain.Username {
	t.Helper()
	username, err := domain.NewUsername(value)
	require.NoError(t, err)
	return username
}
