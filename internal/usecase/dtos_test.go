package usecase_test

import (
	"testing"

	"github.com/rafaelblt/go-auth/internal/domain"
	"github.com/rafaelblt/go-auth/internal/testutil/domaintest"
	"github.com/rafaelblt/go-auth/internal/usecase"
	"github.com/stretchr/testify/assert"
)

func TestMapUserToDTO(t *testing.T) {
	testCases := []struct {
		desc  string
		user  *domain.User
		panic bool
	}{
		{
			desc:  "default user",
			user:  domaintest.DefaultUser(t),
			panic: false,
		},
		{
			desc:  "user zero",
			user:  &domain.User{},
			panic: true,
		},
		{
			desc:  "user nil",
			user:  nil,
			panic: true,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			if tC.panic {
				assert.Panics(t, func() {
					usecase.MapUserToDTO(tC.user)
				})
			} else {
				dto := usecase.MapUserToDTO(tC.user)
				assert.Equal(t, tC.user.ID().Value().String(), dto.ID())
				assert.Equal(t, tC.user.Username().String(), dto.Username())
				assert.Equal(t, tC.user.Status().String(), dto.Status())
				assert.Equal(t, tC.user.CreatedAt(), dto.CreatedAt())
				assert.Equal(t, tC.user.UpdatedAt(), dto.UpdatedAt())
			}
		})
	}
}
