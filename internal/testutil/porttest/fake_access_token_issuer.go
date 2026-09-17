package porttest

import (
	"fmt"
	"time"

	"github.com/rafaelblt/go-auth/internal/domain/session"
	"github.com/rafaelblt/go-auth/internal/port"
)

type FakeAccessTokenIssuer struct {
	payloads []port.AccessTokenPayload
	issueds  []port.AccessTokenIssued
	err      error
}

func NewFakeAccessTokenIssuer() *FakeAccessTokenIssuer {
	return &FakeAccessTokenIssuer{}
}

func (iss *FakeAccessTokenIssuer) Issue(payload port.AccessTokenPayload) (port.AccessTokenIssued, error) {
	iss.payloads = append(iss.payloads, payload)

	if iss.err != nil {
		return port.AccessTokenIssued{}, iss.err
	}

	token, err := session.NewAccessToken("default")
	if err != nil {
		e := fmt.Errorf("the default access token in fake issuer is invalid: %w", err)
		return port.AccessTokenIssued{}, e
	}

	issued := port.AccessTokenIssued{
		Token:     token,
		ExpiresAt: time.Now().UTC().AddDate(1, 0, 0),
	}
	iss.issueds = append(iss.issueds, issued)

	return issued, nil
}

func (iss *FakeAccessTokenIssuer) SetError(err error) {
	iss.err = err
}

func (iss *FakeAccessTokenIssuer) Payloads() []port.AccessTokenPayload {
	return iss.payloads
}

func (iss *FakeAccessTokenIssuer) Issueds() []port.AccessTokenIssued {
	return iss.issueds
}
