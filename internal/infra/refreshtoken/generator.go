package refreshtoken

import "github.com/rafaelblt/go-auth/internal/port"

type Generator struct{}

func NewGenerator() *Generator {
	return &Generator{}
}

type generated = port.RefreshTokenGenerated

func (g *Generator) Generate() (generated, error) {
	token, err := generateToken()
	if err != nil {
		return generated{}, nil
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
