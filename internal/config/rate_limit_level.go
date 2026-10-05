package config

import (
	"github.com/rafaelblt/go-auth/internal/shared"
	"github.com/rafaelblt/go-auth/internal/validation"
)

type RateLimitLevel string

const (
	RateLimitOff     RateLimitLevel = "off"
	RateLimitRelaxed RateLimitLevel = "relaxed"
	RateLimitNormal  RateLimitLevel = "normal"
	RateLimitStrict  RateLimitLevel = "strict"
)

func (l RateLimitLevel) valid() bool {
	switch l {
	case RateLimitOff, RateLimitRelaxed, RateLimitNormal, RateLimitStrict:
		return true
	default:
		return false
	}
}

var rateLimitLevelValidators = []validation.Validator[RateLimitLevel]{
	allowedRateLimitLevel(),
}

func allowedRateLimitLevel() validation.Validator[RateLimitLevel] {
	return func(value RateLimitLevel) *validation.Issue {
		if !value.valid() {
			iss := validation.IssueNotAllowed(
				string(RateLimitOff),
				string(RateLimitRelaxed),
				string(RateLimitNormal),
				string(RateLimitStrict),
			)
			return shared.Ptr(iss)
		}
		return nil
	}
}
