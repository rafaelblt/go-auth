package bcrypt

import (
	"errors"
	"fmt"

	"github.com/rafaelblt/go-auth/internal/credential"
	"golang.org/x/crypto/bcrypt"
)

type Hasher struct {
	cost int
}

type Config struct {
	Cost int
}

func NewHasher(cfg Config) (*Hasher, error) {
	if cfg.Cost < bcrypt.MinCost || cfg.Cost > bcrypt.MaxCost {
		e := fmt.Errorf("bcrypt cost %d is outside allowed inclusive range %d-%d",
			cfg.Cost, bcrypt.MinCost, bcrypt.MaxCost)
		return nil, e
	}

	hasher := Hasher{cost: cfg.Cost}
	return &hasher, nil
}

func (h Hasher) Hash(plain credential.PlainPassword) (credential.Secret, error) {
	if plain.IsZero() {
		return credential.Secret{},
			errors.New("plain password cannot be zero in bcrypt hashing")
	}

	bytes, err := bcrypt.GenerateFromPassword(
		[]byte(plain.Value()),
		h.cost,
	)
	if err != nil {
		return credential.Secret{},
			fmt.Errorf("bcrypt hash generation failed: %w", err)
	}

	secret, err := credential.NewSecret(string(bytes))
	if err != nil {
		return credential.Secret{},
			fmt.Errorf("new credential secret failed: %w", err)
	}

	return secret, err
}

func (h Hasher) Verify(plain credential.PlainPassword, hash credential.Secret) (bool, error) {
	if plain.IsZero() {
		return false,
			errors.New("plain password cannot be zero in bcrypt verify")
	}
	if hash.IsZero() {
		return false,
			errors.New("credential secret cannot be zero in bcrypt verify")
	}

	err := bcrypt.CompareHashAndPassword(
		[]byte(hash.Value()),
		[]byte(plain.Value()),
	)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
		return false, nil
	}
	return false, fmt.Errorf("bcrypt compare hash failed: %w", err)
}
