package register

import "github.com/rafaelblt/go-auth/internal/port"

type Config struct {
	UserExistsChecker port.UserExistsChecker
	UnitOfWork        port.UnitOfWork
	PasswordHasher    port.PasswordHasher
	Clock             port.Clock
}
