package domain

type UserCredentials struct {
	userID   UserID
	password PasswordCredential
}

func NewUserCredentialsWithPassword(
	userID UserID, password PasswordCredential,
) (*UserCredentials, error) {
	if userID.IsZero() {
		return nil, ErrUserIDZero
	}
	if password.IsZero() {
		return nil, ErrPasswordCredentialZero
	}
	return &UserCredentials{
		userID:   userID,
		password: password,
	}, nil
}

func (c UserCredentials) UserID() UserID               { return c.userID }
func (c UserCredentials) Password() PasswordCredential { return c.password }
