package api

import (
	"net/http"
	"net/http/httptest"
	"net/netip"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestClientAddress(t *testing.T) {
	tenSlashEight := []netip.Prefix{netip.MustParsePrefix("10.0.0.0/8")}
	testCases := []struct {
		desc           string
		remoteAddr     string
		forwardedFor   []string
		trustedProxies []netip.Prefix
		expected       string
	}{
		{
			desc:           "untrusted peer with forwarded for",
			remoteAddr:     "198.51.100.7:4321",
			forwardedFor:   []string{"203.0.113.9"},
			trustedProxies: tenSlashEight,
			expected:       "198.51.100.7",
		},
		{
			desc:           "no trusted proxies",
			remoteAddr:     "10.0.0.1:1234",
			forwardedFor:   []string{"203.0.113.9"},
			trustedProxies: nil,
			expected:       "10.0.0.1",
		},
		{
			desc:           "trusted peer, one entry",
			remoteAddr:     "10.0.0.1:1234",
			forwardedFor:   []string{"203.0.113.9"},
			trustedProxies: tenSlashEight,
			expected:       "203.0.113.9",
		},
		{
			desc:           "trusted peer, spoofed leftmost entry",
			remoteAddr:     "10.0.0.1:1234",
			forwardedFor:   []string{"198.51.100.66, 203.0.113.9, 10.0.0.2"},
			trustedProxies: tenSlashEight,
			expected:       "203.0.113.9",
		},
		{
			desc:           "every entry trusted",
			remoteAddr:     "10.0.0.1:1234",
			forwardedFor:   []string{"10.0.0.3, 10.0.0.2"},
			trustedProxies: tenSlashEight,
			expected:       "10.0.0.3",
		},
		{
			desc:           "trusted peer, no header",
			remoteAddr:     "10.0.0.1:1234",
			forwardedFor:   nil,
			trustedProxies: tenSlashEight,
			expected:       "10.0.0.1",
		},
		{
			desc:           "trusted peer, rightmost entry garbage",
			remoteAddr:     "10.0.0.1:1234",
			forwardedFor:   []string{"203.0.113.9, garbage"},
			trustedProxies: tenSlashEight,
			expected:       "10.0.0.1",
		},
		{
			desc:           "two header lines",
			remoteAddr:     "10.0.0.1:1234",
			forwardedFor:   []string{"198.51.100.1", "10.0.0.3"},
			trustedProxies: tenSlashEight,
			expected:       "198.51.100.1",
		},
		{
			desc:           "spaces around entries",
			remoteAddr:     "10.0.0.1:1234",
			forwardedFor:   []string{"  203.0.113.9 ,  10.0.0.2  "},
			trustedProxies: tenSlashEight,
			expected:       "203.0.113.9",
		},
		{
			desc:           "ipv4-mapped peer matched by ipv4 prefix",
			remoteAddr:     "[::ffff:10.0.0.1]:1234",
			forwardedFor:   []string{"203.0.113.9"},
			trustedProxies: tenSlashEight,
			expected:       "203.0.113.9",
		},
		{
			desc:           "ipv4-mapped untrusted peer",
			remoteAddr:     "[::ffff:198.51.100.7]:1234",
			forwardedFor:   nil,
			trustedProxies: tenSlashEight,
			expected:       "198.51.100.7",
		},
		{
			desc:           "ipv6 client counts as its /64",
			remoteAddr:     "[2001:db8:1:2:aaaa::1]:1234",
			forwardedFor:   nil,
			trustedProxies: nil,
			expected:       "2001:db8:1:2::/64",
		},
		{
			desc:           "remote addr not ip and port",
			remoteAddr:     "pipe",
			forwardedFor:   []string{"203.0.113.9"},
			trustedProxies: tenSlashEight,
			expected:       "pipe",
		},
		{
			desc:           "ipv4-mapped client entry",
			remoteAddr:     "10.0.0.1:1234",
			forwardedFor:   []string{"::ffff:203.0.113.9"},
			trustedProxies: tenSlashEight,
			expected:       "203.0.113.9",
		},
		{
			desc:           "ipv4-mapped trusted hop",
			remoteAddr:     "10.0.0.1:1234",
			forwardedFor:   []string{"203.0.113.9, ::ffff:10.0.0.2"},
			trustedProxies: tenSlashEight,
			expected:       "203.0.113.9",
		},
		{
			desc:           "zoned client entry",
			remoteAddr:     "10.0.0.1:1234",
			forwardedFor:   []string{"fe80::1%eth0"},
			trustedProxies: tenSlashEight,
			expected:       "fe80::/64",
		},
		{
			desc:           "rightmost entry with a port",
			remoteAddr:     "10.0.0.1:1234",
			forwardedFor:   []string{"203.0.113.9:5555"},
			trustedProxies: tenSlashEight,
			expected:       "10.0.0.1",
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/v1/auth/login", nil)
			req.RemoteAddr = tC.remoteAddr
			for _, line := range tC.forwardedFor {
				req.Header.Add("X-Forwarded-For", line)
			}

			client := clientAddress(req, tC.trustedProxies)

			assert.Equal(t, tC.expected, client)
		})
	}
}

func TestRetryAfterSeconds(t *testing.T) {
	testCases := []struct {
		desc     string
		value    time.Duration
		expected string
	}{
		{desc: "fraction rounds up", value: 1500 * time.Millisecond, expected: "2"},
		{desc: "whole seconds", value: 6 * time.Second, expected: "6"},
		{desc: "one nanosecond", value: time.Nanosecond, expected: "1"},
		{desc: "zero", value: 0, expected: "1"},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			assert.Equal(t, tC.expected, retryAfterSeconds(tC.value))
		})
	}
}
