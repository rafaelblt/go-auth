package passwordtest

import (
	"testing"
	"time"

	"github.com/rafaelblt/go-auth/internal/domain/password"
	"github.com/rafaelblt/go-auth/internal/domain/user"
	"github.com/stretchr/testify/require"
)

func NewPassword(t *testing.T, override func(p *password.RestoreParams)) *password.Password {
	t.Helper()

	params := password.RestoreParams{
		ID:        password.NewID(),
		UserID:    user.NewID(),
		Hash:      MustHashed(t, "default hash"),
		CreatedAt: time.Date(2007, 8, 9, 20, 45, 0, 0, time.UTC),
		UpdatedAt: time.Date(2026, 5, 21, 16, 0, 7, 0, time.UTC),
	}

	if override != nil {
		override(&params)
	}

	entity, err := password.RestorePassword(params)
	require.NoError(t, err, "password restore failed")
	return entity
}
