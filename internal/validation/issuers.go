package validation

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
