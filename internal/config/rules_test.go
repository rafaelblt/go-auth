package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHostPort_AcceptsWhatTheServerCanListenOn(t *testing.T) {
	for _, address := range []string{"localhost:8080", "0.0.0.0:8080", ":8080", "[::1]:8080", "localhost:http"} {
		t.Run(address, func(t *testing.T) {
			assert.NoError(t, hostPort(address))
		})
	}
}

func TestHostPort_ReturnsError(t *testing.T) {
	testCases := []struct {
		address  string
		expected string
	}{
		{"8080", "must be host:port, such as localhost:8080 (address 8080: missing port in address)"},
		{"localhost", "must be host:port, such as localhost:8080 (address localhost: missing port in address)"},
		{"localhost:99999", "must be host:port, such as localhost:8080 (address 99999: invalid port)"},
	}
	for _, tC := range testCases {
		t.Run(tC.address, func(t *testing.T) {
			assert.EqualError(t, hostPort(tC.address), tC.expected)
		})
	}
}

func TestOneOf_ListsTheAllowedValues(t *testing.T) {
	assert.NoError(t, checkLogFormat(LogFormatText))
	assert.EqualError(t, checkLogFormat("xml"), `must be json or text, not "xml"`)
	assert.EqualError(t, checkRateLimitLevel(""), `must be off, relaxed, normal or strict, not ""`)
}
