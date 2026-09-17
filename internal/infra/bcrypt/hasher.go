package bcrypt

import (
	"errors"
	"fmt"

	"github.com/rafaelblt/go-auth/internal/domain/password"
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

func (h Hasher) Hash(plain password.Plain) (password.Hashed, error) {
	if plain.IsZero() {
		return password.Hashed{},
			errors.New("plain password cannot be zero in bcrypt hashing")
	}

	bytes, err := bcrypt.GenerateFromPassword(
		[]byte(plain.Value()),
		h.cost,
	)
	if err != nil {
		return password.Hashed{},
			fmt.Errorf("bcrypt hash generation failed: %w", err)
	}

	hashed, err := password.NewHashed(string(bytes))
	if err != nil {
		return password.Hashed{},
			fmt.Errorf("new hashed password failed: %w", err)
	}

	return hashed, err
}

func (h Hasher) Verify(plain password.Plain, hash password.Hashed) (bool, error) {
	if plain.IsZero() {
		return false,
			errors.New("plain password cannot be zero in bcrypt verify")
	}
	if hash.IsZero() {
		return false,
			errors.New("hashed password cannot be zero in bcrypt verify")
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
