package logout_test

import (
	"errors"
	"testing"
	"time"

	"github.com/rafaelblt/go-auth/internal/domain/session"
	"github.com/rafaelblt/go-auth/internal/shared"
	"github.com/rafaelblt/go-auth/internal/testutil"
	"github.com/rafaelblt/go-auth/internal/testutil/porttest"
	"github.com/rafaelblt/go-auth/internal/testutil/sessiontest"
	"github.com/rafaelblt/go-auth/internal/usecase"
	"github.com/rafaelblt/go-auth/internal/usecase/logout"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type TestHelper struct {
	t                      *testing.T
	FakeSessionReader      *porttest.FakeSessionReader
	FakeRefreshTokenReader *porttest.FakeRefreshTokenReader
	FakeUnitOfWork         *porttest.FakeUnitOfWork
	FakeClock              *porttest.FakeClock
}

func NewTestHelper(t *testing.T) TestHelper {
	helper := TestHelper{
		t:                      t,
		FakeSessionReader:      porttest.NewFakeSessionReader(),
		FakeRefreshTokenReader: porttest.NewFakeRefreshTokenReader(),
		FakeUnitOfWork:         porttest.NewFakeUnitOfWork(),
		FakeClock:              porttest.NewFakeClock(),
	}
	return helper
}

func (helper *TestHelper) UseCase() *logout.Logout {
	helper.t.Helper()
	uc, err := logout.New(logout.Config{
		SessionReader:      helper.FakeSessionReader,
		RefreshTokenReader: helper.FakeRefreshTokenReader,
		UnitOfWork:         helper.FakeUnitOfWork,
		Clock:              helper.FakeClock,
	})
	require.NoError(helper.t, err)
	return uc
}

func (helper *TestHelper) ValidInput() logout.Input {
	helper.t.Helper()
	return helper.Seed().Input()
}

// Fixture is a session and one of its refresh tokens, stored in the fakes.
type Fixture struct {
	Session *session.Session
	Token   *session.RefreshToken
	Raw     string
}

func (f Fixture) Input() logout.Input {
	return logout.Input{RefreshToken: f.Raw}
}

// Seed stores an active session and an unused, unexpired token of it.
func (helper *TestHelper) Seed() Fixture {
	helper.t.Helper()
	return helper.seed(nil, nil)
}

func (helper *TestHelper) SeedUsedToken() Fixture {
	helper.t.Helper()
	return helper.seed(nil, helper.usedToken)
}

func (helper *TestHelper) SeedExpiredToken() Fixture {
	helper.t.Helper()
	return helper.seed(nil, func(p *session.RefreshTokenRestoreParams) {
		p.ExpiresAt = helper.FakeClock.Now().Add(-time.Minute)
	})
}

func (helper *TestHelper) SeedRevokedSession() Fixture {
	helper.t.Helper()
	return helper.seed(helper.revokedSession, nil)
}

func (helper *TestHelper) SeedUsedTokenOfRevokedSession() Fixture {
	helper.t.Helper()
	return helper.seed(helper.revokedSession, helper.usedToken)
}

// SeedTokenWithoutSession stores a token whose session the reader cannot find.
func (helper *TestHelper) SeedTokenWithoutSession() Fixture {
	helper.t.Helper()

	secret := sessiontest.NewRefreshTokenSecret(helper.t)
	token := sessiontest.NewRefreshToken(helper.t, func(p *session.RefreshTokenRestoreParams) {
		p.Hash = secret.Hash()
		p.ExpiresAt = helper.FakeClock.Now().Add(time.Hour)
	})
	helper.FakeRefreshTokenReader.Insert(token)

	return Fixture{Token: token, Raw: secret.Value()}
}

func (helper *TestHelper) revokedSession(p *session.SessionRestoreParams) {
	p.RevokedAt = shared.Ptr(helper.FakeClock.Now().Add(-time.Minute))
}

func (helper *TestHelper) usedToken(p *session.RefreshTokenRestoreParams) {
	p.UsedAt = shared.Ptr(helper.FakeClock.Now().Add(-time.Minute))
}

func (helper *TestHelper) seed(
	sessOverride func(p *session.SessionRestoreParams),
	tokenOverride func(p *session.RefreshTokenRestoreParams),
) Fixture {
	helper.t.Helper()

	sess := sessiontest.NewSession(helper.t, sessOverride)
	helper.FakeSessionReader.Insert(sess)

	secret := sessiontest.NewRefreshTokenSecret(helper.t)
	token := sessiontest.NewRefreshToken(helper.t, func(p *session.RefreshTokenRestoreParams) {
		p.SessionID = sess.ID()
		p.Hash = secret.Hash()
		p.ExpiresAt = helper.FakeClock.Now().Add(time.Hour)
		if tokenOverride != nil {
			tokenOverride(p)
		}
	})
	helper.FakeRefreshTokenReader.Insert(token)

	return Fixture{Session: sess, Token: token, Raw: secret.Value()}
}

// AssertNoWrites checks that neither sessions nor refresh tokens were written.
func (helper *TestHelper) AssertNoWrites() {
	helper.t.Helper()
	assert.Empty(helper.t, helper.FakeUnitOfWork.FakeSessionWriter.Revoked(), "session revoked")
	helper.AssertNoRefreshTokenWrites()
}

func (helper *TestHelper) AssertNoRefreshTokenWrites() {
	helper.t.Helper()
	assert.Empty(helper.t, helper.FakeUnitOfWork.FakeRefreshTokenWriter.MarkedUsed(), "refresh token marked used")
	assert.Empty(helper.t, helper.FakeUnitOfWork.FakeRefreshTokenWriter.Adds(), "refresh token added")
}

// AssertSessionRevokedNow checks that the session was written once, revoked at
// the clock's time.
func (helper *TestHelper) AssertSessionRevokedNow(sess *session.Session) {
	helper.t.Helper()
	revoked := testutil.Only(helper.t, helper.FakeUnitOfWork.FakeSessionWriter.Revoked())
	assert.Equal(helper.t, sess.ID(), revoked.ID())
	revokedAt, isRevoked := revoked.RevokedAt()
	require.True(helper.t, isRevoked)
	assert.Equal(helper.t, helper.FakeClock.Now(), revokedAt)
}

// AssertUnexpectedError checks that err is not a use case error, so the API
// answers it as an internal error.
func AssertUnexpectedError(t *testing.T, err error) {
	t.Helper()
	require.Error(t, err)
	var uerr usecase.UseCaseError
	assert.False(t, errors.As(err, &uerr), "expected an unexpected error, got use case error %v", err)
}
