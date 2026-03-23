package domain

type UserRepository interface {
	Add(*User) error

	GetByID(UserID) (*User, error)
	GetByUsername(Username) (*User, error)

	ExistsByUsername(Username) (bool, error)
}

type PasswordHasher interface {
	Hash(PlainPassword) (HashedPassword, error)
	Verify(PlainPassword, HashedPassword) (bool, error)
}
