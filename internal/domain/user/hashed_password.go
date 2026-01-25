package user

import "github.com/rafaelblt/go-auth/internal/domain/shared/errors"

type HashedPassword struct {
	value string
}

var ErrHashedPasswordEmpty = errors.NewDomainError(
	"HASHED_PASSWORD_EMPTY", "The hashed password is empty.",
)

func NewHashedPassword(hash string) (HashedPassword, error) {
	if hash == "" {
		return HashedPassword{}, ErrHashedPasswordEmpty
	}
	return HashedPassword{value: hash}, nil
}

func (h HashedPassword) String() string {
	return h.value
}
