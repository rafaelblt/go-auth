package porttest

import (
	"bytes"
	"crypto/sha256"
	"fmt"

	"github.com/rafaelblt/go-auth/internal/domain/session"
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
	if rtg.err != nil {
		return port.RefreshTokenGenerated{}, rtg.err
	}
	hash, err := session.NewRefreshTokenHash(bytes.Repeat([]byte{2}, sha256.Size))
	if err != nil {
		e := fmt.Errorf("refresh token hash creation failed: %w", err)
		return port.RefreshTokenGenerated{}, e
	}
	generated := port.RefreshTokenGenerated{
		Raw:  "[default-fake-refresh-token-raw-value]",
		Hash: hash,
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
