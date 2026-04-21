package service

import "testing"

func TestParseStorageQuotaBytesAcceptsHumanReadableSize(t *testing.T) {
	got, err := ParseStorageQuotaBytes("1.5 GB")
	if err != nil {
		t.Fatalf("parse human-readable quota: %v", err)
	}

	const want = int64(1610612736)
	if got != want {
		t.Fatalf("quota bytes = %d, want %d", got, want)
	}
}

func TestNormalizeDefaultQuotaStoresBytes(t *testing.T) {
	got, err := NormalizeDefaultQuota("10 GB")
	if err != nil {
		t.Fatalf("normalize default quota: %v", err)
	}

	if got != DefaultQuotaSettingValue() {
		t.Fatalf("normalized default quota = %q, want %q", got, DefaultQuotaSettingValue())
	}
}

func TestParseStorageQuotaBytesAcceptsRawBytes(t *testing.T) {
	got, err := ParseStorageQuotaBytes("10737418240")
	if err != nil {
		t.Fatalf("parse raw bytes quota: %v", err)
	}
	if got != DefaultStorageQuotaBytes() {
		t.Fatalf("quota bytes = %d, want %d", got, DefaultStorageQuotaBytes())
	}
}

func TestParseStorageQuotaBytesAcceptsUnlimitedSentinel(t *testing.T) {
	got, err := ParseStorageQuotaBytes("-1")
	if err != nil {
		t.Fatalf("parse unlimited quota: %v", err)
	}
	if got != UnlimitedStorageQuotaBytes() {
		t.Fatalf("quota bytes = %d, want %d", got, UnlimitedStorageQuotaBytes())
	}
}

func TestNormalizeDefaultQuotaRejectsInvalidValue(t *testing.T) {
	if _, err := NormalizeDefaultQuota("abc"); err == nil {
		t.Fatal("expected invalid quota error")
	}
}
