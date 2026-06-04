package validation

type TooShort struct {
	min int
}

func IssueTooShort(min int) TooShort {
	ts := TooShort{min: min}
	return ts
}

func (ts TooShort) Error() string {
	return ts.Code()
}

func (ts TooShort) Code() string {
	return "TOO_SHORT"
}

func (ts TooShort) Details() map[string]any {
	return map[string]any{"min": ts.min}
}
