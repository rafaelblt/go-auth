package password

import "errors"

type Hashed struct {
	value string
}

func NewHashed(value string) (Hashed, error) {
	if value == "" {
		return Hashed{}, errors.New("hashed password value cannot be empty")
	}
	return Hashed{value}, nil
}

func (h Hashed) Value() string { return h.value }
func (h Hashed) IsZero() bool  { return h.value == "" }
