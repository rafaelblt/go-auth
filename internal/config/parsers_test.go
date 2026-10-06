package config

import (
	"bytes"
	"encoding/base64"
	"net/netip"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStringEnvParser(t *testing.T) {
	value, err := stringEnvParser("hello")

	require.NoError(t, err)
	assert.Equal(t, "hello", value)
}

func TestIntEnvParser(t *testing.T) {
	testCases := []struct {
		desc     string
		value    string
		expected int
	}{
		{
			desc:     "positive",
			value:    "42",
			expected: 42,
		},
		{
			desc:     "negative",
			value:    "-7",
			expected: -7,
		},
		{
			desc:     "zero",
			value:    "0",
			expected: 0,
		},
	}

	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			value, err := intEnvParser(tC.value)

			require.NoError(t, err)
			assert.Equal(t, tC.expected, value)
		})
	}
}

func TestIntEnvParser_ReturnsError(t *testing.T) {
	value, err := intEnvParser("invalid")

	assert.Error(t, err)
	assert.Zero(t, value)
}

func TestDurationEnvParser(t *testing.T) {
	testCases := []struct {
		desc     string
		value    string
		expected time.Duration
	}{
		{
			desc:     "milliseconds",
			value:    "300ms",
			expected: 300 * time.Millisecond,
		},
		{
			desc:     "seconds",
			value:    "30s",
			expected: 30 * time.Second,
		},
		{
			desc:     "minutes",
			value:    "15m",
			expected: 15 * time.Minute,
		},
		{
			desc:     "hours",
			value:    "1h",
			expected: 1 * time.Hour,
		},
		{
			desc:     "minutes and seconds",
			value:    "12m30s",
			expected: (12 * time.Minute) + (30 * time.Second),
		},
		{
			desc:     "hours and minutes",
			value:    "2h45m",
			expected: (2 * time.Hour) + (45 * time.Minute),
		},
	}

	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			value, err := durationEnvParser(tC.value)

			require.NoError(t, err)
			assert.Equal(t, tC.expected, value)
		})
	}
}

func TestDurationEnvParser_ReturnsError(t *testing.T) {
	value, err := durationEnvParser("invalid")

	assert.Error(t, err)
	assert.Zero(t, value)
}

func TestBoolEnvParser(t *testing.T) {
	testCases := []struct {
		desc     string
		value    string
		expected bool
	}{
		{
			desc:     "lowercase true",
			value:    "true",
			expected: true,
		},
		{
			desc:     "lowercase false",
			value:    "false",
			expected: false,
		},
		{
			desc:     "uppercase true",
			value:    "TRUE",
			expected: true,
		},
		{
			desc:     "uppercase false",
			value:    "FALSE",
			expected: false,
		},
		{
			desc:     "mixed case true",
			value:    "truE",
			expected: true,
		},
		{
			desc:     "mixed case false",
			value:    "falSe",
			expected: false,
		},
	}

	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			value, err := boolEnvParser(tC.value)

			require.NoError(t, err)
			assert.Equal(t, tC.expected, value)
		})
	}
}

func TestBoolEnvParser_ReturnsError(t *testing.T) {
	value, err := boolEnvParser("invalid")

	assert.Error(t, err)
	assert.Zero(t, value)
}

func TestLogFormatEnvParser(t *testing.T) {
	testCases := []struct {
		desc     string
		value    string
		expected LogFormat
	}{
		{
			desc:     "lowercase json",
			value:    "json",
			expected: LogFormatJSON,
		},
		{
			desc:     "lowercase text",
			value:    "text",
			expected: LogFormatText,
		},
		{
			desc:     "uppercase json",
			value:    "JSON",
			expected: LogFormatJSON,
		},
		{
			desc:     "mixed case text",
			value:    "Text",
			expected: LogFormatText,
		},
	}

	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			value, err := logFormatEnvParser(tC.value)

			require.NoError(t, err)
			assert.Equal(t, tC.expected, value)
		})
	}
}

func TestRateLimitLevelEnvParser(t *testing.T) {
	testCases := []struct {
		desc     string
		value    string
		expected RateLimitLevel
	}{
		{
			desc:     "lowercase off",
			value:    "off",
			expected: RateLimitOff,
		},
		{
			desc:     "uppercase normal",
			value:    "NORMAL",
			expected: RateLimitNormal,
		},
		{
			desc:     "mixed case strict",
			value:    "Strict",
			expected: RateLimitStrict,
		},
		{
			desc:     "lowercase relaxed",
			value:    "relaxed",
			expected: RateLimitRelaxed,
		},
	}

	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			value, err := rateLimitLevelEnvParser(tC.value)

			require.NoError(t, err)
			assert.Equal(t, tC.expected, value)
		})
	}
}

func TestTrustedProxiesEnvParser(t *testing.T) {
	testCases := []struct {
		desc     string
		value    string
		expected []netip.Prefix
	}{
		{
			desc:     "empty",
			value:    "",
			expected: nil,
		},
		{
			desc:     "spaces only",
			value:    "   ",
			expected: nil,
		},
		{
			desc:     "one cidr",
			value:    "10.0.0.0/8",
			expected: []netip.Prefix{netip.MustParsePrefix("10.0.0.0/8")},
		},
		{
			desc:     "bare ipv4",
			value:    "192.0.2.1",
			expected: []netip.Prefix{netip.MustParsePrefix("192.0.2.1/32")},
		},
		{
			desc:     "bare ipv6",
			value:    "2001:db8::1",
			expected: []netip.Prefix{netip.MustParsePrefix("2001:db8::1/128")},
		},
		{
			desc:     "bare ipv4-mapped ipv6",
			value:    "::ffff:10.0.0.1",
			expected: []netip.Prefix{netip.MustParsePrefix("10.0.0.1/32")},
		},
		{
			desc:     "ipv4-mapped ipv6 cidr",
			value:    "::ffff:10.0.0.0/104",
			expected: []netip.Prefix{netip.MustParsePrefix("10.0.0.0/8")},
		},
		{
			desc:  "several with spaces",
			value: " 10.0.0.0/8 ,192.0.2.1,  2001:db8::/32 ",
			expected: []netip.Prefix{
				netip.MustParsePrefix("10.0.0.0/8"),
				netip.MustParsePrefix("192.0.2.1/32"),
				netip.MustParsePrefix("2001:db8::/32"),
			},
		},
	}

	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			value, err := trustedProxiesEnvParser(tC.value)

			require.NoError(t, err)
			assert.Equal(t, tC.expected, value)
		})
	}
}

func TestTrustedProxiesEnvParser_ReturnsError(t *testing.T) {
	testCases := []struct {
		desc  string
		value string
	}{
		{
			desc:  "not an address",
			value: "proxy.local",
		},
		{
			desc:  "empty entry between commas",
			value: "10.0.0.1,,10.0.0.2",
		},
		{
			desc:  "trailing comma",
			value: "10.0.0.1,",
		},
		{
			desc:  "prefix length out of range",
			value: "10.0.0.0/33",
		},
		{
			desc:  "ipv6 zone",
			value: "fe80::1%eth0",
		},
		{
			desc:  "ipv4-mapped prefix shorter than /96",
			value: "::ffff:0:0/80",
		},
	}

	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			value, err := trustedProxiesEnvParser(tC.value)

			require.Error(t, err)
			assert.Contains(t, err.Error(), "invalid trusted proxy")
			assert.Nil(t, value)
		})
	}
}

// encryptionKeyForTest is a fixed key of the 32 bytes SIGNING_KEY_ENCRYPTION_KEY
// takes.
var encryptionKeyForTest = bytes.Repeat([]byte{0x2a}, 32)

func TestEncryptionKeyEnvParser(t *testing.T) {
	encoded := base64.StdEncoding.EncodeToString(encryptionKeyForTest)

	testCases := []struct {
		desc     string
		value    string
		expected []byte
	}{
		{
			desc:     "standard base64 of 32 bytes",
			value:    encoded,
			expected: encryptionKeyForTest,
		},
		{
			desc:     "surrounding spaces",
			value:    "  " + encoded + " ",
			expected: encryptionKeyForTest,
		},
		{
			desc:     "empty",
			value:    "",
			expected: nil,
		},
		{
			desc:     "spaces only",
			value:    "   ",
			expected: nil,
		},
	}

	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			value, err := encryptionKeyEnvParser(tC.value)

			require.NoError(t, err)
			assert.Equal(t, tC.expected, value)
		})
	}
}

func TestEncryptionKeyEnvParser_ReturnsError(t *testing.T) {
	testCases := []struct {
		desc  string
		value string
	}{
		{
			desc:  "not base64",
			value: "not base64!",
		},
		{
			desc:  "base64 of 16 bytes",
			value: base64.StdEncoding.EncodeToString(bytes.Repeat([]byte{0x2a}, 16)),
		},
		{
			desc:  "base64 of 33 bytes",
			value: base64.StdEncoding.EncodeToString(bytes.Repeat([]byte{0x2a}, 33)),
		},
		{
			desc:  "url-safe base64 of 32 bytes",
			value: base64.RawURLEncoding.EncodeToString(bytes.Repeat([]byte{0xfb}, 32)),
		},
	}

	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			value, err := encryptionKeyEnvParser(tC.value)

			require.Error(t, err)
			assert.Nil(t, value)
			assert.NotContains(t, err.Error(), tC.value)
		})
	}
}
