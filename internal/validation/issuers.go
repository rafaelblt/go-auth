package validation

import "strings"

func IssueTooLong(max int, unit LengthUnit) Issue {
	iss := Issue{
		code: CodeTooLong,
		details: map[string]any{
			KeyMaxLength:  max,
			KeyUnitLength: unit,
		},
	}
	return iss
}

func IssueTooShort(min int, unit LengthUnit) Issue {
	iss := Issue{
		code: CodeTooShort,
		details: map[string]any{
			KeyMinLength:  min,
			KeyUnitLength: unit,
		},
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

func IssueRequired() Issue {
	iss := Issue{
		code:    CodeRequired,
		details: map[string]any{},
	}
	return iss
}

func IssueNotPositive() Issue {
	iss := Issue{
		code:    CodeNotPositive,
		details: map[string]any{},
	}
	return iss
}

func IssueNotAllowed(allowed ...string) Issue {
	iss := Issue{
		code: CodeNotAllowed,
		details: map[string]any{
			KeyAllowed: strings.Join(allowed, ", "),
		},
	}
	return iss
}
