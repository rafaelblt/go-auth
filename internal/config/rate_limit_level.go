package config

type RateLimitLevel string

const (
	RateLimitOff     RateLimitLevel = "off"
	RateLimitRelaxed RateLimitLevel = "relaxed"
	RateLimitNormal  RateLimitLevel = "normal"
	RateLimitStrict  RateLimitLevel = "strict"
)

var checkRateLimitLevel = oneOf(RateLimitOff, RateLimitRelaxed, RateLimitNormal, RateLimitStrict)
