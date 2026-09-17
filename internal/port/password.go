package port

import "github.com/rafaelblt/go-auth/internal/domain/password"

type PasswordHasher interface {
	Hash(password.Plain) (password.Hashed, error)
}

type PasswordChecker interface {
	Verify(password.Plain, password.Hashed) (bool, error)
}
