package config

import (
	"github.com/rafaelblt/go-auth/internal/shared"
	"github.com/rafaelblt/go-auth/internal/validation"
)

type LogFormat string

const (
	LogFormatJSON LogFormat = "json"
	LogFormatText LogFormat = "text"
)

func (f LogFormat) valid() bool {
	return f == LogFormatJSON || f == LogFormatText
}

var logFormatValidators = []validation.Validator[LogFormat]{
	allowedLogFormat(),
}

func allowedLogFormat() validation.Validator[LogFormat] {
	return func(value LogFormat) *validation.Issue {
		if !value.valid() {
			iss := validation.IssueNotAllowed(
				string(LogFormatJSON),
				string(LogFormatText),
			)
			return shared.Ptr(iss)
		}
		return nil
	}
}
