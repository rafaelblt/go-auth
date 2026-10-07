package porttest

import (
	"time"

	"github.com/rafaelblt/go-auth/internal/domain/user"
	"github.com/rafaelblt/go-auth/internal/port"
)

type FakeAccessTokenValidator struct {
	raws   []string
	claims port.AccessTokenClaims
	err    error
}

func NewFakeAccessTokenValidator() *FakeAccessTokenValidator {
	validator := FakeAccessTokenValidator{
		claims: port.AccessTokenClaims{
			UserID:    user.NewID(),
			ExpiresAt: time.Now().UTC().Add(time.Hour),
		},
	}
	return &validator
}

func (v *FakeAccessTokenValidator) Validate(raw string) (port.AccessTokenClaims, error) {
	v.raws = append(v.raws, raw)

	if v.err != nil {
		return port.AccessTokenClaims{}, v.err
	}
	return v.claims, nil
}

func (v *FakeAccessTokenValidator) SetClaims(claims port.AccessTokenClaims) {
	v.claims = claims
}

func (v *FakeAccessTokenValidator) SetError(err error) {
	v.err = err
}

func (v *FakeAccessTokenValidator) Claims() port.AccessTokenClaims {
	return v.claims
}

func (v *FakeAccessTokenValidator) Raws() []string {
	return v.raws
}
