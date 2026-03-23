package domain

type UserRepository interface {
	Add(*User) error

	GetByID(UserID) (*User, error)
	GetByUsername(Username) (*User, error)

	ExistsByUsername(Username) (bool, error)
}

