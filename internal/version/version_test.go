package version

import "testing"

func TestGetFallsBackToDevWhenLdflagsNotSet(t *testing.T) {
	// In the test binary there's no -X injection, so Get() must always return
	// non-empty sentinel strings rather than empty values that would break
	// JSON consumers expecting fixed keys.
	info := Get()
	if info.Version == "" {
		t.Fatal("Version must not be empty even without ldflags")
	}
	if info.Commit == "" {
		t.Fatal("Commit must not be empty even without ldflags")
	}
	if info.Date == "" {
		t.Fatal("Date must not be empty even without ldflags")
	}
	if info.Go == "" {
		t.Fatal("Go runtime version must always be populated from runtime/debug")
	}
}

func TestGetIsCached(t *testing.T) {
	// Consumers use Get in hot paths (e.g. every /health hit). It must be a
	// pure read after the first call — sync.Once guarantees that.
	a := Get()
	b := Get()
	if a != b {
		t.Fatalf("Get() returned different values: %+v vs %+v", a, b)
	}
}
