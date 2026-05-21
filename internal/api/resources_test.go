package api

import (
	"testing"

	"github.com/rafaelblt/go-auth/internal/domain"
	"github.com/rafaelblt/go-auth/internal/testutil/domaintest"
	"github.com/rafaelblt/go-auth/internal/usecase"
	"github.com/stretchr/testify/assert"
)

func mapUserToDTO(t *testing.T, user *domain.User) usecase.UserDTO {
	t.Helper()
	dto := usecase.MapUserToDTO(user)
	return dto
}

func TestMapUserDTOToResource(t *testing.T) {
	testCases := []struct {
		desc      string
		dto       usecase.UserDTO
		expectErr bool
	}{
		{desc: "user dto zero", dto: usecase.UserDTO{}, expectErr: true},
		{desc: "default user", dto: mapUserToDTO(t, domaintest.NewUser(t, nil))},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			resource, err := mapUserDTOToResource(tC.dto)
			if tC.expectErr {
				assert.Error(t, err)
				assert.Zero(t, resource)
			} else {
				assert.NoError(t, err)
				assert.NotZero(t, resource)
				assert.Equal(t, tC.dto.ID(), resource.ID)
				assert.Equal(t, tC.dto.Username(), resource.Username)
				assert.Equal(t, tC.dto.Status(), resource.Status)
				assert.Equal(t, tC.dto.CreatedAt(), resource.CreatedAt)
				assert.Equal(t, tC.dto.UpdatedAt(), resource.UpdatedAt)
			}
		})
	}
}
