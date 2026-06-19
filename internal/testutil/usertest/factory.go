package usertest

import (
	"testing"
	"time"

	"github.com/rafaelblt/go-auth/internal/user"
	"github.com/stretchr/testify/require"
)

func NewUser(t *testing.T, override func(p *user.RestoreParams)) *user.User {
	t.Helper()

    params := user.RestoreParams{
        ID:        user.NewID(),
        Username:  MustUsername(t, "Default User"),
        Status:    user.StatusActive,
        CreatedAt: time.Date(2007, 8, 9, 20, 45, 0, 0, time.UTC),
        UpdatedAt: time.Date(2026, 5, 21, 16, 0, 7, 0, time.UTC),
    }

    if override != nil {
        override(&params)
    }

    entity, err := user.RestoreUser(params)
    require.NoError(t, err)
    return entity
}
