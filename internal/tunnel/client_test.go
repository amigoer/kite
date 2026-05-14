package tunnel

import (
	"testing"
	"time"
)

func TestBuildConnectURL(t *testing.T) {
	cases := []struct {
		in      string
		want    string
		wantErr bool
	}{
		{"https://panel.example.com", "wss://panel.example.com" + ConnectPath, false},
		{"http://panel.example.com:8080", "ws://panel.example.com:8080" + ConnectPath, false},
		{"wss://panel.example.com", "wss://panel.example.com" + ConnectPath, false},
		{"ws://127.0.0.1:9090", "ws://127.0.0.1:9090" + ConnectPath, false},
		// Bare host, no scheme → default to wss (production-safe).
		{"panel.example.com", "wss://panel.example.com" + ConnectPath, false},
		// Strip stale path / query / fragment, we own the path.
		{"https://panel.example.com/v2/api?x=1#frag", "wss://panel.example.com" + ConnectPath, false},
		// Unsupported scheme.
		{"ftp://panel.example.com", "", true},
	}
	for _, tc := range cases {
		got, err := buildConnectURL(tc.in)
		if tc.wantErr {
			if err == nil {
				t.Errorf("buildConnectURL(%q) succeeded, want error", tc.in)
			}
			continue
		}
		if err != nil {
			t.Errorf("buildConnectURL(%q) unexpected error: %v", tc.in, err)
			continue
		}
		if got != tc.want {
			t.Errorf("buildConnectURL(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestNextBackoff(t *testing.T) {
	if got := nextBackoff(ReconnectMin); got != 2*time.Second {
		t.Errorf("first doubling = %v, want 2s", got)
	}
	// Cap at ReconnectMax.
	if got := nextBackoff(ReconnectMax); got != ReconnectMax {
		t.Errorf("at max = %v, want %v", got, ReconnectMax)
	}
	if got := nextBackoff(ReconnectMax - time.Second); got != ReconnectMax {
		t.Errorf("near-max doubles past cap = %v, want %v", got, ReconnectMax)
	}
}

func TestJitterStaysInBounds(t *testing.T) {
	base := 4 * time.Second
	for i := 0; i < 500; i++ {
		got := jitter(base)
		// Allow a tiny floating-point slack (1ms).
		lo := time.Duration(float64(base) * (1 - ReconnectJitterFr))
		hi := time.Duration(float64(base) * (1 + ReconnectJitterFr))
		if got < lo-time.Millisecond || got > hi+time.Millisecond {
			t.Errorf("jitter(%v) = %v, outside [%v, %v]", base, got, lo, hi)
		}
	}
}
