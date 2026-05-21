package register

import "github.com/rafaelblt/go-auth/internal/usecase"

type Config struct {
	UserExistsChecker usecase.UserExistsChecker
	UnitOfWork        usecase.UnitOfWork
	PasswordHasher    usecase.PasswordHasher
	Clock             usecase.Clock
}
