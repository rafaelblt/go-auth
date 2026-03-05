package domain

type UserRepository interface {
	Add(*User) error

	GetByID(UserID) (*User, error)
	GetByUsername(Username) (*User, error)

	ExistsByUsername(Username) (bool, error)
}

type UserCredentialsRepository interface {
	Add(*UserCredentials) error

	GetByID(UserID) (*UserCredentials, error)
}

type PasswordHasher interface {
	Hash(PlainPassword) (HashedPassword, error)
	Verify(PlainPassword, HashedPassword) (bool, error)
}
