package validation

type Accumulator struct {
	result []FieldError
}

func NewAccumulator() *Accumulator {
	result := make([]FieldError, 0)
	return &Accumulator{result: result}
}

func (acc *Accumulator) Add(field string, issues Issues) {
	for _, iss := range issues {
		fieldErr := FieldError{field: field, issue: iss}
		acc.result = append(acc.result, fieldErr)
	}
}

func (acc *Accumulator) Err() error {
	if len(acc.result) > 0 {
		return ValidationError{acc.result}
	}
	return nil
}
