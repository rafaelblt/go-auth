package domain

import (
	"strings"
)

type Email struct {
	value string
}

var ErrEmailEmpty = NewDomainError("EMAIL_EMPTY", "the email is empty")
var ErrEmailInvalidFormat = NewDomainError("EMAIL_INVALID_FORMAT", "the email is invalid")

func NewEmail(email string) (Email, error) {
	normalized := normalizeEmail(email)

	if normalized == "" {
		return Email{}, ErrEmailEmpty
	}

	if hasValidFormat(normalized) {
		return Email{value: normalized}, nil 
	}

	return Email{}, ErrEmailInvalidFormat
}

func ValidateEmail(email string) []DomainError {
	normalized := normalizeEmail(email)

	var errs []DomainError

	if normalized == "" {
		errs = append(errs, ErrEmailEmpty)
		return errs
	}
	if !hasValidFormat(normalized) {
		errs = append(errs, ErrEmailInvalidFormat)
	}

	return errs
}

func normalizeEmail(raw string) string {
	return strings.TrimSpace(strings.ToLower(raw))
}

func hasValidFormat(email string) bool {
	if email == "" {
		return false // empty
	}
	if strings.Contains(email, " ") {
		return false // email contains spaces
	}
	at := strings.IndexByte(email, '@')
	if at <= 0 || at != strings.LastIndex(email, "@") || at == len(email)-1 {
		return false // invalid '@' placement
	}
	domain := email[at+1:]
	if domain == "" ||
		strings.HasPrefix(domain, ".") ||
		strings.HasSuffix(domain, ".") ||
		!strings.Contains(domain, ".") {
		return false // invalid domain
	}
	return true
}

func (e Email) String() string {
	return e.value
}
