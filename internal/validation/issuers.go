package validation

const (
	CodeTooLong  = "TOO_LONG"
	CodeTooShort = "TOO_SHORT"
)

func IssueTooLong(max int) Issue {
	iss := Issue{
		code:    CodeTooLong,
		details: map[string]any{"max": max},
	}
	return iss
}

func IssueTooShort(min int) Issue {
	iss := Issue{
		code:    CodeTooShort,
		details: map[string]any{"min": min},
	}
	return iss
}
