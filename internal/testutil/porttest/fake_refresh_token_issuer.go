package porttest

import (
	"fmt"
	"time"

	"github.com/rafaelblt/go-auth/internal/port"
	"github.com/rafaelblt/go-auth/internal/session"
)

type FakeRefreshTokenIssuer struct {
	payloads []port.RefreshTokenPayload
	issueds  []port.RefreshTokenIssued
	err      error
}

func NewFakeRefreshTokenIssuer() *FakeRefreshTokenIssuer {
	return &FakeRefreshTokenIssuer{}
}

func (iss *FakeRefreshTokenIssuer) Issue(payload port.RefreshTokenPayload) (port.RefreshTokenIssued, error) {
	iss.payloads = append(iss.payloads, payload)

	if iss.err != nil {
		return port.RefreshTokenIssued{}, iss.err
	}

	token, err := session.NewRefreshToken(session.RefreshTokenCreationParams{
		SessionID: payload.SessionID,
		Hash:      []byte("[default-fake-refresh-token-hash]"),
		ParentID:  nil,
		IssuedAt:  time.Now().UTC(),
		ExpiresAt: time.Now().UTC().Add(10000000),
	})

	if err != nil {
		e := fmt.Errorf("default token in fake refresh token issuer is invalid: %w", err)
		return port.RefreshTokenIssued{}, e
	}

	issued := port.RefreshTokenIssued{
		Token:    token,
		RawValue: "[default-fake-refresh-token-raw-value]",
	}
	iss.issueds = append(iss.issueds, issued)

	return issued, nil
}

func (iss *FakeRefreshTokenIssuer) LastPayload() port.RefreshTokenPayload {
	return iss.payloads[len(iss.payloads)-1]
}

func (iss *FakeRefreshTokenIssuer) LastIssued() port.RefreshTokenIssued {
	return iss.issueds[len(iss.issueds)-1]
}

func (iss *FakeRefreshTokenIssuer) SetError(err error) {
	iss.err = err
}
