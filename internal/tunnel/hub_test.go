package tunnel

import (
	"context"
	"errors"
	"log/slog"
	"testing"
)

func TestExtractBearer(t *testing.T) {
	cases := map[string]string{
		"":                   "",
		"Bearer ":            "",
		"Bearer  abc  ":      "abc",
		"Bearer abc":         "abc",
		"Bearer    xyz":      "xyz",
		"bearer abc":         "", // case-sensitive on scheme — daemons we control
		"Token abc":          "",
		"Basic dXNlcjpwdw==": "",
	}
	for in, want := range cases {
		if got := extractBearer(in); got != want {
			t.Errorf("extractBearer(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestHubAuthorize(t *testing.T) {
	h := NewHub(map[string]string{
		"laptop":      "tok-laptop",
		"home-server": "tok-server",
	}, slog.Default())

	cases := []struct {
		name, token string
		want        error
	}{
		{"", "", ErrUnknownName},
		{"laptop", "", ErrBadToken},
		{"laptop", "wrong", ErrBadToken},
		{"laptop", "tok-laptop", nil},
		{"home-server", "tok-server", nil},
		{"home-server", "tok-laptop", ErrBadToken}, // right shape, wrong daemon's token
		{"unknown", "tok-laptop", ErrUnknownName},
	}
	for _, tc := range cases {
		got := h.authorize(tc.name, tc.token)
		if !errors.Is(got, tc.want) {
			t.Errorf("authorize(%q,%q) = %v, want %v", tc.name, tc.token, got, tc.want)
		}
	}
}

func TestHubAuthorizeRejectsAllWhenEmpty(t *testing.T) {
	h := NewHub(nil, slog.Default())
	if err := h.authorize("any", "any"); !errors.Is(err, ErrUnknownName) {
		t.Errorf("empty allow-list authorize = %v, want ErrUnknownName", err)
	}
}

func TestHubDialOfflineReturnsSentinel(t *testing.T) {
	h := NewHub(map[string]string{"laptop": "tok"}, slog.Default())
	_, err := h.Dial(context.Background(), "laptop")
	if !errors.Is(err, ErrDaemonOffline) {
		t.Errorf("Dial with no session = %v, want ErrDaemonOffline", err)
	}
	// Unknown name behaves identically to offline at Dial time — we
	// don't leak whether a name is configured to browser callers.
	_, err = h.Dial(context.Background(), "unknown")
	if !errors.Is(err, ErrDaemonOffline) {
		t.Errorf("Dial for unknown name = %v, want ErrDaemonOffline", err)
	}
}

func TestHubSnapshotIncludesOfflineConfiguredDaemons(t *testing.T) {
	h := NewHub(map[string]string{
		"laptop":      "tok1",
		"home-server": "tok2",
	}, slog.Default())
	snap := h.Snapshot()
	if len(snap) != 2 {
		t.Fatalf("snapshot len = %d, want 2", len(snap))
	}
	for _, d := range snap {
		if d.Connected {
			t.Errorf("daemon %q reported connected with no live session", d.Name)
		}
	}
}

func TestConstantTimeEq(t *testing.T) {
	if !constantTimeEq("abc", "abc") {
		t.Error("equal strings should match")
	}
	if constantTimeEq("abc", "abd") {
		t.Error("differing strings should not match")
	}
	if constantTimeEq("abc", "abcd") {
		t.Error("different lengths should not match")
	}
	if !constantTimeEq("", "") {
		t.Error("two empty strings should match")
	}
}
