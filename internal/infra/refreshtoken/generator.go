package refreshtoken

import "github.com/rafaelblt/go-auth/internal/port"

type Generator struct{}

func NewGenerator() *Generator {
	return &Generator{}
}

type generated = port.RefreshTokenGenerated

func (g *Generator) Generate() (generated, error) {
	// Unreachable since Go 1.24: crypto/rand.Read never returns an error and
	// crashes the program instead. Kept so a failure is never swallowed.
	token, err := generateToken()
	if err != nil {
		return generated{}, err
	}
	hash, err := hashToken(token)
	if err != nil {
		return generated{}, err
	}
	generated := generated{
		Raw:  encodeToken(token),
		Hash: hash,
	}
	return generated, nil
}
