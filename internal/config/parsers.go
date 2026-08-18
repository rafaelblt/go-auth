package config

import (
	"fmt"
	"strconv"
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
