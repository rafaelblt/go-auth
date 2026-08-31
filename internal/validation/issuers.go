package validation

const (
	CodeTooLong           = "TOO_LONG"
	CodeTooShort          = "TOO_SHORT"
	CodeInvalidCharacters = "INVALID_CHARACTERS"
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

func IssueInvalidChars() Issue {
	iss := Issue{
		code:    CodeInvalidCharacters,
		details: map[string]any{},
	}
	return iss
}
