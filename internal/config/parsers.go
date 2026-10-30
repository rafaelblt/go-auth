package config

import (
	"encoding/base64"
	"errors"
	"fmt"
	"net/netip"
	"strconv"
	"strings"
	"time"
)

func stringEnvParser(val string) (string, error) {
	return val, nil
}

func intEnvParser(val string) (int, error) {
	n, err := strconv.Atoi(val)
	if err != nil {
		return 0, fmt.Errorf("must be a whole number, not %q", val)
	}
	return n, nil
}

func durationEnvParser(val string) (time.Duration, error) {
	d, err := time.ParseDuration(val)
	if err != nil {
		return 0, fmt.Errorf("must be a duration such as 30m, 1h30m or 168h, not %q", val)
	}
	return d, nil
}

func boolEnvParser(val string) (bool, error) {
	switch strings.ToLower(val) {
	case "true":
		return true, nil
	case "false":
		return false, nil
	default:
		return false, fmt.Errorf("must be true or false, not %q", val)
	}
}

// logFormatEnvParser only converts. Whether the format is one the app knows
// is decided by NewConfig, which also reports the accepted values.
func logFormatEnvParser(val string) (LogFormat, error) {
	return LogFormat(strings.ToLower(val)), nil
}

// rateLimitLevelEnvParser only converts, like logFormatEnvParser.
func rateLimitLevelEnvParser(val string) (RateLimitLevel, error) {
	return RateLimitLevel(strings.ToLower(val)), nil
}

// trustedProxiesEnvParser stops at any entry it cannot read, since a typo in
// this list would silently change whose X-Forwarded-For is believed. An
// address written as IPv4-mapped IPv6 becomes IPv4, which is how the client
// address is compared.
func trustedProxiesEnvParser(val string) ([]netip.Prefix, error) {
	if strings.TrimSpace(val) == "" {
		return nil, nil
	}

	entries := strings.Split(val, ",")
	prefixes := make([]netip.Prefix, 0, len(entries))
	for _, entry := range entries {
		entry = strings.TrimSpace(entry)
		prefix, err := parseTrustedProxy(entry)
		if err != nil {
			return nil, fmt.Errorf("must list IP addresses and CIDR prefixes, separated by commas: %q is not one (%w)", entry, err)
		}
		prefixes = append(prefixes, prefix)
	}
	return prefixes, nil
}

// sessionRetentionEnvParser reads off in any case, as RATE_LIMIT does, and
// anything else as a duration. Whether the duration is long enough is decided
// by NewConfig.
func sessionRetentionEnvParser(val string) (SessionRetention, error) {
	if strings.EqualFold(val, "off") {
		return SessionRetentionOff(), nil
	}

	d, err := time.ParseDuration(val)
	if err != nil {
		return SessionRetention{}, fmt.Errorf("must be off or a duration such as 720h, not %q", val)
	}
	return SessionRetentionOf(d), nil
}

// encryptionKeyEnvParser never puts the value in its errors: it is a secret.
// Whether the key has the right size is decided by NewConfig.
func encryptionKeyEnvParser(val string) ([]byte, error) {
	val = strings.TrimSpace(val)
	if val == "" {
		return nil, nil
	}

	key, err := base64.StdEncoding.DecodeString(val)
	if err != nil {
		return nil, fmt.Errorf("must be standard base64 (%w)", err)
	}
	return key, nil
}

func parseTrustedProxy(entry string) (netip.Prefix, error) {
	if entry == "" {
		return netip.Prefix{}, errors.New("empty entry")
	}

	if strings.Contains(entry, "/") {
		prefix, err := netip.ParsePrefix(entry)
		if err != nil {
			return netip.Prefix{}, err
		}
		if !prefix.Addr().Is4In6() {
			return prefix, nil
		}
		if prefix.Bits() < 96 {
			return netip.Prefix{}, errors.New("IPv4-mapped prefix shorter than /96")
		}
		return netip.PrefixFrom(prefix.Addr().Unmap(), prefix.Bits()-96), nil
	}

	addr, err := netip.ParseAddr(entry)
	if err != nil {
		return netip.Prefix{}, err
	}
	if addr.Zone() != "" {
		return netip.Prefix{}, errors.New("zone not allowed")
	}
	addr = addr.Unmap()
	return netip.PrefixFrom(addr, addr.BitLen()), nil
}
