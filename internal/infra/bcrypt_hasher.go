package infra

import (
	"errors"
	"fmt"

	"github.com/rafaelblt/go-auth/internal/credential"
	"golang.org/x/crypto/bcrypt"
)

type BcryptHasher struct {
	cost int
}

type BcryptConfig struct {
	Cost int
}

func NewBcryptHasher(cfg BcryptConfig) (BcryptHasher, error) {
	if cfg.Cost < bcrypt.MinCost || cfg.Cost > bcrypt.MaxCost {
		return BcryptHasher{},
		fmt.Errorf("bcrypt cost %d is outside allowed inclusive range %d-%d",
			cfg.Cost, bcrypt.MinCost, bcrypt.MaxCost)
	}

	hasher := BcryptHasher{cost: cfg.Cost}
	return hasher, nil
}

func (hasher BcryptHasher) Hash(plain credential.PlainPassword) (credential.Secret, error) {
	if plain.IsZero() {
		return credential.Secret{},
		errors.New("plain password cannot be zero in bcrypt hashing")
	}

	bytes, err := bcrypt.GenerateFromPassword(
		[]byte(plain.Value()),
		hasher.cost,
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

func (hasher BcryptHasher) Verify(plain credential.PlainPassword, hash credential.Secret) (bool, error) {
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
