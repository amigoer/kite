package service

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
)

const (
	DefaultQuotaSettingKey  = "default_quota"
	defaultStorageQuotaByte = int64(10 * 1024 * 1024 * 1024)
	unlimitedStorageQuota   = int64(-1)
)

var storageQuotaPattern = regexp.MustCompile(`(?i)^([0-9]+(?:\.[0-9]+)?)\s*(b|kb|mb|gb|tb)$`)

// DefaultStorageQuotaBytes returns the compiled fallback quota for new regular
// users when the runtime settings table does not override it.
func DefaultStorageQuotaBytes() int64 {
	return defaultStorageQuotaByte
}

// DefaultQuotaSettingValue returns the persisted string form used by runtime
// settings and defaults. The value is stored in bytes so it can be written
// directly into users.storage_limit without another conversion step.
func DefaultQuotaSettingValue() string {
	return strconv.FormatInt(DefaultStorageQuotaBytes(), 10)
}

// UnlimitedStorageQuotaBytes returns the sentinel value used throughout the
// application to represent "no quota limit".
func UnlimitedStorageQuotaBytes() int64 {
	return unlimitedStorageQuota
}

// NormalizeDefaultQuota validates and normalizes the admin-facing quota input.
// The persisted form is always an int64 byte count encoded as a decimal string.
func NormalizeDefaultQuota(raw string) (string, error) {
	bytes, err := ParseStorageQuotaBytes(raw)
	if err != nil {
		return "", err
	}
	return strconv.FormatInt(bytes, 10), nil
}

// ParseStorageQuotaBytes parses either a raw byte count (for example
// "10737418240") or a human-friendly size (for example "10 GB") into bytes.
func ParseStorageQuotaBytes(raw string) (int64, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return 0, fmt.Errorf("cannot be empty")
	}
	if trimmed == strconv.FormatInt(UnlimitedStorageQuotaBytes(), 10) {
		return UnlimitedStorageQuotaBytes(), nil
	}

	if bytes, err := strconv.ParseInt(trimmed, 10, 64); err == nil {
		if bytes < 1 {
			return 0, fmt.Errorf("must be greater than 0")
		}
		return bytes, nil
	}

	matches := storageQuotaPattern.FindStringSubmatch(trimmed)
	if len(matches) != 3 {
		return 0, fmt.Errorf("must be a positive size like 10 GB or 512 MB")
	}

	value, err := strconv.ParseFloat(matches[1], 64)
	if err != nil || value <= 0 {
		return 0, fmt.Errorf("must be a positive size like 10 GB or 512 MB")
	}

	factor := storageQuotaUnitFactor(matches[2])
	if factor == 0 {
		return 0, fmt.Errorf("unsupported size unit")
	}

	bytesFloat := value * float64(factor)
	if bytesFloat > float64(math.MaxInt64) {
		return 0, fmt.Errorf("value is too large")
	}

	bytes := int64(math.Round(bytesFloat))
	if bytes < 1 {
		return 0, fmt.Errorf("must be greater than 0")
	}

	return bytes, nil
}

func storageQuotaUnitFactor(unit string) int64 {
	switch strings.ToLower(strings.TrimSpace(unit)) {
	case "b":
		return 1
	case "kb":
		return 1024
	case "mb":
		return 1024 * 1024
	case "gb":
		return 1024 * 1024 * 1024
	case "tb":
		return 1024 * 1024 * 1024 * 1024
	default:
		return 0
	}
}
