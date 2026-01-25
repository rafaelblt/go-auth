package user

type Status string

const (
	StatusActive Status = "active"
)

func (s Status) String() string { return string(s) }

func (s Status) IsActive() bool { return s == StatusActive }
