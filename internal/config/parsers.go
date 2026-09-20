package config

import (
	"fmt"
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
		return 0, fmt.Errorf("invalid int value %q: %w", val, err)
	}
	return n, nil
}

func durationEnvParser(val string) (time.Duration, error) {
	d, err := time.ParseDuration(val)
	if err != nil {
		return 0, fmt.Errorf("invalid duration value %q: %w", val, err)
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
		return false, fmt.Errorf("invalid bool value %q", val)
	}
}

// logFormatEnvParser only converts. Whether the format is one the app knows
// is decided by NewConfig, which can also report the accepted values.
func logFormatEnvParser(val string) (LogFormat, error) {
	return LogFormat(strings.ToLower(val)), nil
}
