package user

import (
	"fmt"
	"strings"
)

type Status struct {
	value string
}

var (
	StatusActive = Status{"active"}
)

func ParseStatus(value string) (Status, error) {
	var s Status
	switch strings.TrimSpace(strings.ToLower(value)) {
	case "active":
		s = StatusActive
	default:
		return Status{}, fmt.Errorf("invalid status value: %s", value)
	}
	return s, nil
}

func (s Status) String() string { return s.value }
func (s Status) IsActive() bool { return s == StatusActive }
func (s Status) IsZero() bool   { return s.value == "" }
