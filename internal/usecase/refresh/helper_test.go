package refresh_test

import (
	"errors"
	"testing"
	"time"

	"github.com/rafaelblt/go-auth/internal/domain/session"
	"github.com/rafaelblt/go-auth/internal/shared"
	"github.com/rafaelblt/go-auth/internal/testutil/porttest"
	"github.com/rafaelblt/go-auth/internal/testutil/sessiontest"
	"github.com/rafaelblt/go-auth/internal/usecase"
	"github.com/rafaelblt/go-auth/internal/usecase/refresh"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type TestHelper struct {
	t                      *testing.T
	FakeSessionReader      *porttest.FakeSessionReader
	FakeAccessTokenIssuer  *porttest.FakeAccessTokenIssuer
	FakeRefreshTokenReader *porttest.FakeRefreshTokenReader
	FakeUnitOfWork         *porttest.FakeUnitOfWork
	FakeClock              *porttest.FakeClock
	RefreshTokenTTL        time.Duration
}

func NewTestHelper(t *testing.T) TestHelper {
	helper := TestHelper{
		t:                      t,
		FakeSessionReader:      porttest.NewFakeSessionReader(),
		FakeAccessTokenIssuer:  porttest.NewFakeAccessTokenIssuer(),
		FakeRefreshTokenReader: porttest.NewFakeRefreshTokenReader(),
		FakeUnitOfWork:         porttest.NewFakeUnitOfWork(),
		FakeClock:              porttest.NewFakeClock(),
		RefreshTokenTTL:        24 * time.Hour,
	}
	return helper
}

func (helper *TestHelper) UseCase() *refresh.Refresh {
	helper.t.Helper()
	uc, err := refresh.New(refresh.Config{
		SessionReader:      helper.FakeSessionReader,
		AccessTokenIssuer:  helper.FakeAccessTokenIssuer,
		RefreshTokenReader: helper.FakeRefreshTokenReader,
		UnitOfWork:         helper.FakeUnitOfWork,
		Clock:              helper.FakeClock,
		RefreshTokenTTL:    helper.RefreshTokenTTL,
	})
	require.NoError(helper.t, err)
	return uc
}

func (helper *TestHelper) ValidInput() refresh.Input {
	helper.t.Helper()
	return helper.Seed().Input()
}

// Fixture is a session and one of its refresh tokens, stored in the fakes.
type Fixture struct {
	Session *session.Session
	Token   *session.RefreshToken
	Raw     string
}

func (f Fixture) Input() refresh.Input {
	return refresh.Input{RefreshToken: f.Raw}
}

// Seed stores an active session and an unused, unexpired token of it.
func (helper *TestHelper) Seed() Fixture {
	helper.t.Helper()
	return helper.seed(nil, nil)
}

func (helper *TestHelper) SeedUsedToken() Fixture {
	helper.t.Helper()
	return helper.seed(nil, func(p *session.RefreshTokenRestoreParams) {
		p.UsedAt = shared.Ptr(helper.FakeClock.Now().Add(-time.Minute))
	})
}

func (helper *TestHelper) SeedExpiredToken() Fixture {
	helper.t.Helper()
	return helper.seed(nil, func(p *session.RefreshTokenRestoreParams) {
		p.ExpiresAt = helper.FakeClock.Now().Add(-time.Minute)
	})
}

func (helper *TestHelper) SeedRevokedSession() Fixture {
	helper.t.Helper()
	return helper.seed(func(p *session.SessionRestoreParams) {
		p.RevokedAt = shared.Ptr(helper.FakeClock.Now().Add(-time.Minute))
	}, nil)
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

func (helper *TestHelper) AssertNoAccessTokenIssued() {
	helper.t.Helper()
	assert.Empty(helper.t, helper.FakeAccessTokenIssuer.Issueds(), "access token issued")
}

func (helper *TestHelper) AssertNoRefreshTokenWrites() {
	helper.t.Helper()
	assert.Empty(helper.t, helper.FakeUnitOfWork.FakeRefreshTokenWriter.MarkedUsed(), "refresh token marked used")
	assert.Empty(helper.t, helper.FakeUnitOfWork.FakeRefreshTokenWriter.Adds(), "refresh token added")
}

// AssertNoWrites checks that neither refresh tokens nor sessions were written.
func (helper *TestHelper) AssertNoWrites() {
	helper.t.Helper()
	helper.AssertNoRefreshTokenWrites()
	assert.Empty(helper.t, helper.FakeUnitOfWork.FakeSessionWriter.Updates(), "session updated")
}

// AssertUnexpectedError checks that err is not a use case error, so the API
// answers it as an internal error.
func AssertUnexpectedError(t *testing.T, err error) {
	t.Helper()
	require.Error(t, err)
	var uerr usecase.UseCaseError
	assert.False(t, errors.As(err, &uerr), "expected an unexpected error, got use case error %v", err)
}
