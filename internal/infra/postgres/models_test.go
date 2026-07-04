package postgres

import (
	"testing"
	"time"

	"github.com/rafaelblt/go-auth/internal/session"
	"github.com/rafaelblt/go-auth/internal/shared"
	"github.com/rafaelblt/go-auth/internal/testutil/sessiontest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMapRefreshTokenToModel(t *testing.T) {
	testCases := []struct {
		desc      string
		entity    *session.RefreshToken
		expectErr bool
	}{
		{
			desc:      "refresh token nil",
			entity:    nil,
			expectErr: true,
		},
		{
			desc:      "refresh token zero",
			entity:    &session.RefreshToken{},
			expectErr: true,
		},
		{
			desc:   "refresh token without parent id and used at",
			entity: sessiontest.NewRefreshToken(t, func(p *session.RefreshTokenRestoreParams) {
				p.ParentID = nil
				p.UsedAt = nil				
			}),
		},
		{
			desc: "refresh token with parent id",
			entity: sessiontest.NewRefreshToken(t, func(p *session.RefreshTokenRestoreParams) {
				p.ParentID = shared.Ptr(session.NewRefreshTokenID())
			}),
		},
		{
			desc: "refresh token with used at",
			entity: sessiontest.NewRefreshToken(t, func(p *session.RefreshTokenRestoreParams) {
				p.UsedAt = shared.Ptr(time.Now().UTC())
			}),
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			model, err := mapRefreshTokenToModel(tC.entity)
			if tC.expectErr {
				assert.Error(t, err)
				assert.Zero(t, model)
				return
			}
			require.NotZero(t, model)
			assert.Equal(t, tC.entity.ID().String(), model.ID)
			assert.Equal(t, tC.entity.SessionID().String(), model.SessionID)
			assert.Equal(t, tC.entity.Hash().Value(), model.Hash)
			assert.Equal(t, tC.entity.CreatedAt(), model.CreatedAt)
			assert.Equal(t, tC.entity.ExpiresAt(), model.ExpiresAt)
			parentID, hasParent := tC.entity.ParentID()
			if hasParent {
				assert.Equal(t, parentID.String(), *model.ParentID)
			} else {
				assert.Nil(t, model.ParentID)
			}
			usedAt, isUsed := tC.entity.UsedAt()
			if isUsed {
				assert.Equal(t, usedAt, *model.UsedAt)
			} else {
				assert.Nil(t, model.UsedAt)
			}
		})
	}
}

func TestMapRefreshTokenToEntity(t *testing.T) {
	testCases := []struct {
		desc      string
		model     refreshTokenModel
		expectErr bool
	}{
		{
			desc:      "model zero",
			model:     refreshTokenModel{},
			expectErr: true,
		},
		{
			desc: "model with only required fields",
			model: refreshTokenModel{
				ID:        shared.NewEntityID().String(),
				SessionID: shared.NewEntityID().String(),
				Hash:      []byte{2, 0, 3, 0},
				CreatedAt: time.Now().UTC(),
				ExpiresAt: time.Now().UTC(),
			},
		},
		{
			desc: "model with all fields",
			model: refreshTokenModel{
				ID:        shared.NewEntityID().String(),
				SessionID: shared.NewEntityID().String(),
				ParentID:  shared.Ptr(shared.NewEntityID().String()),
				Hash:      []byte{2, 0, 3, 0},
				CreatedAt: time.Now().UTC(),
				ExpiresAt: time.Now().UTC(),
				UsedAt:    shared.Ptr(time.Now().UTC()),
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			entity, err := mapRefreshTokenToEntity(tC.model)
			if tC.expectErr {
				assert.Error(t, err)
				assert.Nil(t, entity)
				return
			}
			require.False(t, entity.IsZero(), "refresh token is zero")
			assert.Equal(t, tC.model.ID, entity.ID().String())
			assert.Equal(t, tC.model.SessionID, entity.SessionID().String())
			assert.Equal(t, tC.model.Hash, entity.Hash().Value())
			assert.Equal(t, tC.model.CreatedAt, entity.CreatedAt())
			assert.Equal(t, tC.model.ExpiresAt, entity.ExpiresAt())
			parentID, hasParent := entity.ParentID()
			if tC.model.ParentID == nil {
				assert.False(t, hasParent)
			} else {
				assert.Equal(t, *tC.model.ParentID, parentID.String())
			}
			usedAt, isUsed := entity.UsedAt()
			if tC.model.UsedAt == nil {
				assert.False(t, isUsed)
			} else {
				assert.Equal(t, *tC.model.UsedAt, usedAt)
			}
		})
	}
}
