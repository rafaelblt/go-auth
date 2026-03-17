package domain

import (
	"fmt"
	"strings"
)

type UserStatus string

const (
	UserStatusActive UserStatus = "active"
)

func NewUserStatus(value string) (UserStatus, error) {
	switch strings.TrimSpace(strings.ToLower(value)) {
	case "active":
		return UserStatusActive, nil
	}
	return "", fmt.Errorf("invalid user status: %s", value)
}

func (s UserStatus) String() string { return string(s) }
func (s UserStatus) IsActive() bool { return s == UserStatusActive }
