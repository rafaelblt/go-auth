package refresh_test

import (
	"testing"
	"time"

	"github.com/rafaelblt/go-auth/internal/session"
	"github.com/rafaelblt/go-auth/internal/shared"
	"github.com/rafaelblt/go-auth/internal/testutil/porttest"
	"github.com/rafaelblt/go-auth/internal/testutil/sessiontest"
	"github.com/rafaelblt/go-auth/internal/usecase/refresh"
	"github.com/stretchr/testify/require"
)

type TestHelper struct {
	t                         *testing.T
	FakeSessionReader         *porttest.FakeSessionReader
	FakeAccessTokenIssuer     *porttest.FakeAccessTokenIssuer
	FakeRefreshTokenResolver  *porttest.FakeRefreshTokenResolver
	FakeRefreshTokenGenerator *porttest.FakeRefreshTokenGenerator
	FakeUnitOfWork            *porttest.FakeUnitOfWork
	FakeClock                 *porttest.FakeClock
	RefreshTokenTTL           time.Duration
}

func NewTestHelper(t *testing.T) TestHelper {
	helper := TestHelper{
		t:                         t,
		FakeSessionReader:         porttest.NewFakeSessionReader(),
		FakeAccessTokenIssuer:     porttest.NewFakeAccessTokenIssuer(),
		FakeRefreshTokenGenerator: porttest.NewFakeRefreshTokenGenerator(),
		FakeRefreshTokenResolver:  porttest.NewFakeRefreshTokenResolver(),
		FakeUnitOfWork:            porttest.NewFakeUnitOfWork(),
		FakeClock:                 porttest.NewFakeClock(),
		RefreshTokenTTL:           24 * time.Hour,
	}
	return helper
}

func (helper *TestHelper) UseCase() *refresh.Refresh {
	helper.t.Helper()
	uc, err := refresh.New(refresh.Config{
		SessionReader:         helper.FakeSessionReader,
		AccessTokenIssuer:     helper.FakeAccessTokenIssuer,
		RefreshTokenResolver:  helper.FakeRefreshTokenResolver,
		RefreshTokenGenerator: helper.FakeRefreshTokenGenerator,
		UnitOfWork:            helper.FakeUnitOfWork,
		Clock:                 helper.FakeClock,
		RefreshTokenTTL:       helper.RefreshTokenTTL,
	})
	require.NoError(helper.t, err)
	return uc
}

func (helper *TestHelper) ValidInput() refresh.Input {
	helper.t.Helper()
	_, raw := helper.GetRefreshTokenAndRaw()
	return refresh.Input{RefreshToken: raw}
}

func (helper *TestHelper) GetRefreshTokenAndRaw() (*session.RefreshToken, string) {
	helper.t.Helper()

	sess := sessiontest.NewSession(helper.t, nil)
	helper.FakeSessionReader.Insert(sess)

	token := sessiontest.NewRefreshToken(helper.t, func(p *session.RefreshTokenRestoreParams) {
		p.SessionID = sess.ID()
		p.ExpiresAt = helper.FakeClock.Now().Add(1)
	})
	raw := "default_raw"
	helper.FakeRefreshTokenResolver.Insert(raw, token)

	return token, raw
}

func (helper *TestHelper) GetSessionAndTokenRaw() (*session.Session, string) {
	helper.t.Helper()

	sess := sessiontest.NewSession(helper.t, nil)
	helper.FakeSessionReader.Insert(sess)

	token := sessiontest.NewRefreshToken(helper.t, func(p *session.RefreshTokenRestoreParams) {
		p.SessionID = sess.ID()
		p.ExpiresAt = helper.FakeClock.Now().Add(1)
	})
	raw := "default_raw"
	helper.FakeRefreshTokenResolver.Insert(raw, token)

	return sess, raw
}

func (helper *TestHelper) GetTokenAlreadyUsed() string {
	helper.t.Helper()

	sess := sessiontest.NewSession(helper.t, nil)
	helper.FakeSessionReader.Insert(sess)

	token := sessiontest.NewRefreshToken(helper.t, func(p *session.RefreshTokenRestoreParams) {
		p.SessionID = sess.ID()
		p.ExpiresAt = helper.FakeClock.Now().Add(1)
		p.UsedAt = shared.Ptr(time.Now().UTC())
	})
	raw := "default_raw"
	helper.FakeRefreshTokenResolver.Insert(raw, token)

	return raw
}

func (helper *TestHelper) GetTokenExpired() string {
	helper.t.Helper()

	sess := sessiontest.NewSession(helper.t, nil)
	helper.FakeSessionReader.Insert(sess)

	token := sessiontest.NewRefreshToken(helper.t, func(p *session.RefreshTokenRestoreParams) {
		p.SessionID = sess.ID()
		p.ExpiresAt = helper.FakeClock.Now().Add(-1)
	})
	raw := "default_raw"
	helper.FakeRefreshTokenResolver.Insert(raw, token)

	return raw
}

func (helper *TestHelper) GetTokenWithSessionRevoked() string {
	helper.t.Helper()

	sess := sessiontest.NewSession(helper.t, func(p *session.SessionRestoreParams) {
		p.RevokedAt = shared.Ptr(time.Now().UTC())
	})
	helper.FakeSessionReader.Insert(sess)

	token := sessiontest.NewRefreshToken(helper.t, func(p *session.RefreshTokenRestoreParams) {
		p.SessionID = sess.ID()
		p.ExpiresAt = helper.FakeClock.Now().Add(1)
	})
	raw := "default_raw"
	helper.FakeRefreshTokenResolver.Insert(raw, token)

	return raw
}
