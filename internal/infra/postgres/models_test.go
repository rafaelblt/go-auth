package postgres

import (
	"testing"
	"time"

	"github.com/rafaelblt/go-auth/internal/domain/password"
	"github.com/rafaelblt/go-auth/internal/domain/session"
	"github.com/rafaelblt/go-auth/internal/domain/user"
	"github.com/rafaelblt/go-auth/internal/port"
	"github.com/rafaelblt/go-auth/internal/shared"
	"github.com/rafaelblt/go-auth/internal/testutil/passwordtest"
	"github.com/rafaelblt/go-auth/internal/testutil/sessiontest"
	"github.com/rafaelblt/go-auth/internal/testutil/usertest"
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

func TestMapPasswordToModel(t *testing.T) {
	testCases := []struct {
		desc      string
		entity    *password.Password
		expectErr bool
	}{
		{
			desc:      "password nil",
			entity:    nil,
			expectErr: true,
		},
		{
			desc:      "password zero",
			entity:    &password.Password{},
			expectErr: true,
		},
		{
			desc:   "default password",
			entity: passwordtest.NewPassword(t, nil),
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			model, err := mapPasswordToModel(tC.entity)
			if tC.expectErr {
				assert.Error(t, err)
				assert.Zero(t, model)
				return
			}
			require.NotZero(t, model)
			assert.Equal(t, tC.entity.ID().String(), model.ID)
			assert.Equal(t, tC.entity.UserID().String(), model.UserID)
			assert.Equal(t, tC.entity.Hash().Value(), model.Hash)
			assert.Equal(t, tC.entity.CreatedAt(), model.CreatedAt)
			assert.Equal(t, tC.entity.UpdatedAt(), model.UpdatedAt)
		})
	}
}

func TestMapPasswordToEntity(t *testing.T) {
	testCases := []struct {
		desc      string
		model     passwordModel
		expectErr bool
	}{
		{
			desc:      "model zero",
			model:     passwordModel{},
			expectErr: true,
		},
		{
			desc: "valid model",
			model: passwordModel{
				ID:        shared.NewEntityID().String(),
				UserID:    shared.NewEntityID().String(),
				Hash:      "hash",
				CreatedAt: time.Now().UTC(),
				UpdatedAt: time.Now().UTC(),
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			entity, err := mapPasswordToEntity(tC.model)
			if tC.expectErr {
				assert.Error(t, err)
				assert.Nil(t, entity)
				return
			}
			require.False(t, entity.IsZero())
			assert.Equal(t, tC.model.ID, entity.ID().String())
			assert.Equal(t, tC.model.UserID, entity.UserID().String())
			assert.Equal(t, tC.model.Hash, entity.Hash().Value())
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
			assert.Equal(t, tC.entity.ExpiresAt(), model.ExpiresAt)
			assert.Equal(t, shared.PtrFromOk(tC.entity.UsedAt()), model.UsedAt)
			assert.Equal(t, tC.entity.CreatedAt(), model.CreatedAt)
			assert.Equal(t, tC.entity.UpdatedAt(), model.UpdatedAt)
			parentID, hasParent := tC.entity.ParentID()
			if hasParent {
				assert.Equal(t, parentID.String(), *model.ParentID)
			} else {
				assert.Nil(t, model.ParentID)
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
				Hash:      sessiontest.NewRefreshTokenHash(t, "model").Value(),
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
				Hash:      sessiontest.NewRefreshTokenHash(t, "model").Value(),
				UsedAt:    shared.Ptr(time.Now().UTC()),
				CreatedAt: time.Now().UTC(),
				ExpiresAt: time.Now().UTC(),
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
			require.False(t, entity.IsZero())
			assert.Equal(t, tC.model.ID, entity.ID().String())
			assert.Equal(t, tC.model.SessionID, entity.SessionID().String())
			assert.Equal(t, tC.model.Hash, entity.Hash().Value())
			assert.Equal(t, tC.model.ExpiresAt, entity.ExpiresAt())
			assert.Equal(t, tC.model.UsedAt, shared.PtrFromOk(entity.UsedAt()))
			assert.Equal(t, tC.model.CreatedAt, entity.CreatedAt())
			assert.Equal(t, tC.model.UpdatedAt, entity.UpdatedAt())
			parentID, hasParent := entity.ParentID()
			if tC.model.ParentID == nil {
				assert.False(t, hasParent)
			} else {
				assert.Equal(t, *tC.model.ParentID, parentID.String())
			}
		})
	}
}

func TestMapSigningKeyToModel(t *testing.T) {
	key := port.StoredSigningKey{
		Generation: 7,
		Seed:       []byte("0123456789abcdef0123456789abcdef"),
		ActiveAt:   time.Now().UTC().Add(time.Hour),
		CreatedAt:  time.Now().UTC(),
	}

	model := mapSigningKeyToModel(key)

	assert.Equal(t, key.Generation, model.Generation)
	assert.Equal(t, key.Seed, model.Seed)
	assert.Equal(t, key.ActiveAt, model.ActiveAt)
	assert.Equal(t, key.CreatedAt, model.CreatedAt)
}

func TestMapSigningKeyToRecord(t *testing.T) {
	model := signingKeyModel{
		Generation: 7,
		Seed:       []byte("0123456789abcdef0123456789abcdef"),
		ActiveAt:   time.Now().UTC().Add(time.Hour),
		CreatedAt:  time.Now().UTC(),
	}

	key := mapSigningKeyToRecord(model)

	assert.Equal(t, model.Generation, key.Generation)
	assert.Equal(t, model.Seed, key.Seed)
	assert.Equal(t, model.ActiveAt, key.ActiveAt)
	assert.Equal(t, model.CreatedAt, key.CreatedAt)
}
