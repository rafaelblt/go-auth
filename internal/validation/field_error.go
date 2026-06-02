package validation

import "fmt"

type FieldError struct {
	field string
	issue Issue
}

func NewFieldError(field string, issue Issue) FieldError {
	return FieldError{field, issue}
}

func (fe FieldError) Error() string {
	return fmt.Sprintf("[%s]: %s", fe.field, fe.issue)
}

func (fe FieldError) Issue() Issue {
	return fe.issue
}

func (fe FieldError) Field() string {
	return fe.field
}
