package port

import "github.com/rafaelblt/go-auth/internal/credential"

type PasswordHasher interface {
	Hash(credential.PlainPassword) (credential.Secret, error)	
}

type PasswordChecker interface {
	Verify(credential.PlainPassword, credential.Secret) (bool, error)
}
