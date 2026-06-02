package validation

func IssueMaxLen(max int) Issue {
	params := map[string]any{"max": max}
	iss := Issue{
		code:   CodeTooLong,
		params: params,
	}
	return iss
}

func IssueMinLen(min int) Issue {
	params := map[string]any{"min": min}
	iss := Issue{
		code:   CodeTooShort,
		params: params,
	}
	return iss
}
