package service

import (
	"fmt"
	"strconv"
	"strings"
)

const (
	AuthRateLimitPerMinuteSettingKey        = "rate_limit.auth_requests_per_minute"
	GuestUploadRateLimitPerMinuteSettingKey = "rate_limit.guest_upload_requests_per_minute"
	defaultAuthRateLimitPerMinute           = 20
	defaultGuestUploadRateLimitPerMinute    = 60
)

// DefaultAuthRateLimitPerMinute returns the built-in per-IP request cap for
// unauthenticated auth endpoints within one minute.
func DefaultAuthRateLimitPerMinute() string {
	return strconv.Itoa(defaultAuthRateLimitPerMinute)
}

// DefaultGuestUploadRateLimitPerMinute returns the built-in per-IP request cap
// for anonymous uploads within one minute. The default is intentionally higher
// than auth endpoints because the public upload page supports multi-file bursts.
func DefaultGuestUploadRateLimitPerMinute() string {
	return strconv.Itoa(defaultGuestUploadRateLimitPerMinute)
}

// NormalizeRequestsPerMinute validates and normalizes a rate-limit value that
// is stored as a positive integer request count per minute.
func NormalizeRequestsPerMinute(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", fmt.Errorf("cannot be empty")
	}

	value, err := strconv.ParseInt(trimmed, 10, 64)
	if err != nil {
		return "", fmt.Errorf("must be a positive integer")
	}
	if value < 1 {
		return "", fmt.Errorf("must be at least 1")
	}
	if value > 1_000_000 {
		return "", fmt.Errorf("value is too large")
	}

	return strconv.FormatInt(value, 10), nil
}

// ParseRequestsPerMinute parses the stored per-minute rate-limit value into an
// integer after applying the same validation rules as the admin settings UI.
func ParseRequestsPerMinute(raw string) (int, error) {
	normalized, err := NormalizeRequestsPerMinute(raw)
	if err != nil {
		return 0, err
	}
	value, _ := strconv.Atoi(normalized)
	return value, nil
}
