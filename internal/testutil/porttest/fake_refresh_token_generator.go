package porttest

import (
	"github.com/rafaelblt/go-auth/internal/port"
)

type FakeRefreshTokenGenerator struct {
	gens []port.RefreshTokenGenerated
	err  error
}

func NewFakeRefreshTokenGenerator() *FakeRefreshTokenGenerator {
	return &FakeRefreshTokenGenerator{}
}

func (rtg *FakeRefreshTokenGenerator) Generate() (port.RefreshTokenGenerated, error) {
	generated := port.RefreshTokenGenerated{
		Raw:  "[default-fake-refresh-token-raw-value]",
		Hash: []byte("[default-fake-refresh-token-hash-value]"),
	}
	rtg.gens = append(rtg.gens, generated)

	return generated, nil
}

func (rtg *FakeRefreshTokenGenerator) Generated() []port.RefreshTokenGenerated {
	return rtg.gens
}

func (rtg *FakeRefreshTokenGenerator) SetError(err error) {
	rtg.err = err
}
