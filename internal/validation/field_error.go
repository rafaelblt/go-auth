package validation

type FieldError struct {
	field string
	err   error
}

func NewFieldError(field string, err error) FieldError {
	return FieldError{field, err}
}

func (fe FieldError) Field() string {
	return fe.field
}

func (fe FieldError) Err() error {
	return fe.err
}
