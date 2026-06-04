package validation

type TooLong struct {
	max int
}

func IssueTooLong(max int) TooLong {
	iss := TooLong{max: max}
	return iss
}

func (tl TooLong) Error() string {
	return tl.Code()
}

func (tl TooLong) Code() string {
	return "TOO_LONG"
}

func (tl TooLong) Details() map[string]any {
	return map[string]any{"max": tl.max}
}
