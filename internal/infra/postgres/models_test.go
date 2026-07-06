package postgres

import (
	"testing"
	"time"

	"github.com/rafaelblt/go-auth/internal/credential"
	"github.com/rafaelblt/go-auth/internal/session"
	"github.com/rafaelblt/go-auth/internal/shared"
	"github.com/rafaelblt/go-auth/internal/testutil/credentialtest"
	"github.com/rafaelblt/go-auth/internal/testutil/sessiontest"
	"github.com/rafaelblt/go-auth/internal/testutil/usertest"
	"github.com/rafaelblt/go-auth/internal/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMapUserToModel(t *testing.T) {
	testCases := []struct {
		desc      string
		entity    *user.User
		expectErr bool
	}{
		{
			desc:      "user nil",
			entity:    nil,
			expectErr: true,
		},
		{
			desc:      "user zero",
			entity:    &user.User{},
			expectErr: true,
		},
		{
			desc:   "default user",
			entity: usertest.NewUser(t, nil),
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			model, err := mapUserToModel(tC.entity)
			if tC.expectErr {
				assert.Error(t, err)
				assert.Zero(t, model)
				return
			}
			require.NotZero(t, model)
			assert.Equal(t, tC.entity.ID().String(), model.ID)
			assert.Equal(t, tC.entity.Username().String(), model.Username)
			assert.Equal(t, tC.entity.Status().String(), model.Status)
			assert.Equal(t, tC.entity.CreatedAt(), model.CreatedAt)
			assert.Equal(t, tC.entity.UpdatedAt(), model.UpdatedAt)
		})
	}
}

func TestMapUserToEntity(t *testing.T) {
	testCases := []struct {
		desc      string
		model     userModel
		expectErr bool
	}{
		{
			desc:      "model zero",
			model:     userModel{},
			expectErr: true,
		},
		{
			desc: "valid model",
			model: userModel{
				ID:        shared.NewEntityID().String(),
				Username:  "username",
				Status:    user.StatusActive.String(),
				CreatedAt: time.Now().UTC(),
				UpdatedAt: time.Now().UTC(),
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			entity, err := mapUserToEntity(tC.model)
			if tC.expectErr {
				assert.Error(t, err)
				assert.Nil(t, entity)
				return
			}
			require.False(t, entity.IsZero())
			assert.Equal(t, tC.model.ID, entity.ID().String())
			assert.Equal(t, tC.model.Username, entity.Username().String())
			assert.Equal(t, tC.model.Status, entity.Status().String())
			assert.Equal(t, tC.model.CreatedAt, entity.CreatedAt())
			assert.Equal(t, tC.model.UpdatedAt, entity.UpdatedAt())
		})
	}
}

func TestMapCredentialToModel(t *testing.T) {
	testCases := []struct {
		desc      string
		entity    *credential.Credential
		expectErr bool
	}{
		{
			desc:      "credential nil",
			entity:    nil,
			expectErr: true,
		},
		{
			desc:      "credential zero",
			entity:    &credential.Credential{},
			expectErr: true,
		},
		{
			desc:   "default credential",
			entity: credentialtest.NewCredential(t, nil),
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			model, err := mapCredentialToModel(tC.entity)
			if tC.expectErr {
				assert.Error(t, err)
				assert.Zero(t, model)
				return
			}
			require.NotZero(t, model)
			assert.Equal(t, tC.entity.ID().String(), model.ID)
			assert.Equal(t, tC.entity.UserID().String(), model.UserID)
			assert.Equal(t, tC.entity.Kind().String(), model.Kind)
			assert.Equal(t, tC.entity.Provider().String(), model.Provider)
			assert.Equal(t, tC.entity.Secret().Value(), model.Secret)
			assert.Equal(t, tC.entity.CreatedAt(), model.CreatedAt)
			assert.Equal(t, tC.entity.UpdatedAt(), model.UpdatedAt)
		})
	}
}

func TestMapCredentialToEntity(t *testing.T) {
	testCases := []struct {
		desc      string
		model     credentialModel
		expectErr bool
	}{
		{
			desc:      "model zero",
			model:     credentialModel{},
			expectErr: true,
		},
		{
			desc: "valid model",
			model: credentialModel{
				ID:        shared.NewEntityID().String(),
				UserID:    shared.NewEntityID().String(),
				Kind:      credential.KindPassword.String(),
				Provider:  credential.ProviderLocal.String(),
				Secret:    "secret",
				CreatedAt: time.Now().UTC(),
				UpdatedAt: time.Now().UTC(),
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			entity, err := mapCredentialToEntity(tC.model)
			if tC.expectErr {
				assert.Error(t, err)
				assert.Nil(t, entity)
				return
			}
			require.False(t, entity.IsZero())
			assert.Equal(t, tC.model.ID, entity.ID().String())
			assert.Equal(t, tC.model.UserID, entity.UserID().String())
			assert.Equal(t, tC.model.Kind, entity.Kind().String())
			assert.Equal(t, tC.model.Provider, entity.Provider().String())
			assert.Equal(t, tC.model.Secret, entity.Secret().Value())
			assert.Equal(t, tC.model.CreatedAt, entity.CreatedAt())
			assert.Equal(t, tC.model.UpdatedAt, entity.UpdatedAt())
		})
	}
}

func TestMapSessionToModel(t *testing.T) {
	testCases := []struct {
		desc      string
		entity    *session.Session
		expectErr bool
	}{
		{
			desc:      "session nil",
			entity:    nil,
			expectErr: true,
		},
		{
			desc:      "session zero",
			entity:    &session.Session{},
			expectErr: true,
		},
		{
			desc: "session without revoked at",
			entity: sessiontest.NewSession(t, func(p *session.SessionRestoreParams) {
				p.RevokedAt = nil
			}),
		},
		{
			desc: "session with revoked at",
			entity: sessiontest.NewSession(t, func(p *session.SessionRestoreParams) {
				p.RevokedAt = shared.Ptr(time.Now())
			}),
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			model, err := mapSessionToModel(tC.entity)
			if tC.expectErr {
				assert.Error(t, err)
				assert.Zero(t, model)
				return
			}
			require.NotZero(t, model)
			assert.Equal(t, tC.entity.ID().String(), model.ID)
			assert.Equal(t, tC.entity.UserID().String(), model.UserID)
			assert.Equal(t, tC.entity.CreatedAt(), model.CreatedAt)
			assert.Equal(t, tC.entity.UpdatedAt(), model.UpdatedAt)
			revokedAt, isRevoked := tC.entity.RevokedAt()
			if isRevoked {
				assert.Equal(t, revokedAt, *model.RevokedAt)
			} else {
				assert.Nil(t, model.RevokedAt)
			}
		})
	}
}

func TestMapSessionToEntity(t *testing.T) {
	testCases := []struct {
		desc      string
		model     sessionModel
		expectErr bool
	}{
		{
			desc:      "model zero",
			model:     sessionModel{},
			expectErr: true,
		},
		{
			desc: "model with only required fields",
			model: sessionModel{
				ID:        shared.NewEntityID().String(),
				UserID:    shared.NewEntityID().String(),
				CreatedAt: time.Now().UTC(),
				UpdatedAt: time.Now().UTC(),
			},
		},
		{
			desc: "model with all fields",
			model: sessionModel{
				ID:        shared.NewEntityID().String(),
				UserID:    shared.NewEntityID().String(),
				CreatedAt: time.Now().UTC(),
				UpdatedAt: time.Now().UTC(),
				RevokedAt: shared.Ptr(time.Now().UTC()),
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			entity, err := mapSessionToEntity(tC.model)
			if tC.expectErr {
				assert.Error(t, err)
				assert.Nil(t, entity)
				return
			}
			require.False(t, entity.IsZero(), "refresh token is zero")
			assert.Equal(t, tC.model.ID, entity.ID().String())
			assert.Equal(t, tC.model.UserID, entity.UserID().String())
			assert.Equal(t, tC.model.CreatedAt, entity.CreatedAt())
			assert.Equal(t, tC.model.UpdatedAt, entity.UpdatedAt())
			revokedAt, isRevoked := entity.RevokedAt()
			if tC.model.RevokedAt == nil {
				assert.False(t, isRevoked)
			} else {
				assert.Equal(t, *tC.model.RevokedAt, revokedAt)
			}
		})
	}
}

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
			desc: "refresh token without parent id and used at",
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
