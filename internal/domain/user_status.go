package domain

type UserStatus string

const (
	UserStatusActive UserStatus = "active"
)

func (s UserStatus) String() string { return string(s) }
func (s UserStatus) IsActive() bool { return s == UserStatusActive }
