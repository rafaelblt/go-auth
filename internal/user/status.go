package user

type Status struct {
	value string
}

var (
	StatusActive = Status{"active"}
)

func (s Status) String() string { return s.value }
func (s Status) IsActive() bool { return s == StatusActive }
func (s Status) IsZero() bool   { return s.value == "" }
