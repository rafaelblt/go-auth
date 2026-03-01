package domain

type HashedPassword struct {
	value string
}

var ErrHashedPasswordEmpty = NewDomainError(
	"HASHED_PASSWORD_EMPTY", "the hashed password is empty",
)

func NewHashedPassword(hash string) (HashedPassword, error) {
	if hash == "" {
		return HashedPassword{}, ErrHashedPasswordEmpty
	}
	return HashedPassword{value: hash}, nil
}

func (h HashedPassword) Value() string {
	return h.value
}
