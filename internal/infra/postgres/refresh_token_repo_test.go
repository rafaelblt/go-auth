package postgres_test

import (
	"context"
	"testing"
	"time"

	"github.com/rafaelblt/go-auth/internal/infra/postgres"
	"github.com/rafaelblt/go-auth/internal/session"
	"github.com/rafaelblt/go-auth/internal/testutil/sessiontest"
	"github.com/rafaelblt/go-auth/internal/testutil/usertest"
	"github.com/rafaelblt/go-auth/internal/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// HELPER

type RefreshTokenRepoTestHelper struct {
	t  *testing.T
	db postgres.DB
}

func NewRefreshTokenRepoTestHelper(t *testing.T) *RefreshTokenRepoTestHelper {
	db := poolFactory.Acquire(t)
	return &RefreshTokenRepoTestHelper{t, db}
}

func (helper *RefreshTokenRepoTestHelper) Repo() *postgres.RefreshTokenRepo {
	helper.t.Helper()

	repo, err := postgres.NewRefreshTokenRepo(helper.db)
	require.NoError(helper.t, err)

	return repo
}

func (helper *RefreshTokenRepoTestHelper) SaveSession(sess *session.Session) {
	helper.t.Helper()

	userRepo, err := postgres.NewUserRepo(helper.db)
	require.NoError(helper.t, err)

	sessRepo, err := postgres.NewSessionRepo(helper.db)
	require.NoError(helper.t, err)

	usr := usertest.NewUser(helper.t, func(p *user.RestoreParams) {
		p.ID = sess.UserID()
	})
	require.NoError(helper.t, userRepo.Add(context.Background(), usr))

	require.NoError(helper.t, sessRepo.Add(context.Background(), sess))
}

func (helper *RefreshTokenRepoTestHelper) CheckRefreshTokenIsSaved(token *session.RefreshToken) bool {
	var result bool

	var parentIDPtr *string
	parentID, ok := token.ParentID()
	if ok {
		idString := parentID.String()
		parentIDPtr = &idString
	}

	var usedAtPtr *time.Time
	usedAt, ok := token.UsedAt()
	if ok {
		usedAtPtr = &usedAt
	}

	err := helper.db.QueryRow(context.Background(),
		`SELECT EXISTS(
			SELECT 1 FROM refresh_tokens WHERE
			id=$1 AND
			session_id=$2 AND
			parent_id IS NOT DISTINCT FROM $3 AND
			hash=$4 AND
			created_at=$5 AND
			expires_at=$6 AND
			used_at IS NOT DISTINCT FROM $7
		)`,
		token.ID().Value(),
		token.SessionID().Value(),
		parentIDPtr,
		token.Hash().Value(),
		token.CreatedAt(),
		token.ExpiresAt(),
		usedAtPtr,
	).Scan(&result)
	require.NoError(helper.t, err)
	return result
}

// TESTS

func TestNewRefreshTokenRepo_WithDBNil(t *testing.T) {
	uow, err := postgres.NewRefreshTokenRepo(nil)
	assert.Error(t, err)
	assert.Zero(t, uow)
}

func TestNewRefreshTokenRepo_WithValidDB(t *testing.T) {
	db := poolFactory.Acquire(t)

	uow, err := postgres.NewRefreshTokenRepo(db)

	assert.NoError(t, err)
	assert.NotNil(t, uow)
}

func TestRefreshTokenRepo_Add(t *testing.T) {
	helper := NewRefreshTokenRepoTestHelper(t)

	sess := sessiontest.NewSession(t, nil)
	helper.SaveSession(sess)

	token := sessiontest.NewRefreshToken(t, func(p *session.RefreshTokenRestoreParams) {
		p.SessionID = sess.ID()
	})

	repo := helper.Repo()
	err := repo.Add(context.Background(), token)

	require.NoError(t, err)
	require.True(t, helper.CheckRefreshTokenIsSaved(token))
}

// TODO: Update
// TODO: FindByHash
