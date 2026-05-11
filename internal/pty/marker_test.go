package pty

import (
	"strconv"
	"testing"
)

func TestMarkerRegexMatchesValidExit(t *testing.T) {
	cases := []struct {
		in         string
		wantExit   int
		wantCmdID  string
	}{
		{"__KITE_END_0_c_abcdefghijkl__", 0, "c_abcdefghijkl"},
		{"__KITE_END_1_c_aaaaaaaaaaaa__", 1, "c_aaaaaaaaaaaa"},
		{"__KITE_END_130_c_zzzzzzzzzzzz__", 130, "c_zzzzzzzzzzzz"},
		{"__KITE_END_-1_c_22222222222a__", -1, "c_22222222222a"},
	}
	for _, c := range cases {
		t.Run(c.in, func(t *testing.T) {
			m := markerRe.FindStringSubmatch(c.in)
			if m == nil {
				t.Fatalf("no match for %q", c.in)
			}
			gotExit, _ := strconv.Atoi(m[1])
			if gotExit != c.wantExit {
				t.Errorf("exit: %d, want %d", gotExit, c.wantExit)
			}
			if m[2] != c.wantCmdID {
				t.Errorf("cmdID: %s, want %s", m[2], c.wantCmdID)
			}
		})
	}
}

func TestMarkerRegexRejectsInvalid(t *testing.T) {
	for _, bad := range []string{
		"__KITE_END_0_c_short__",                 // cmd id too short
		"__KITE_END_0_c_AAAAAAAAAAAA__",          // uppercase not in [a-z2-7]
		"__KITE_END_0_d_abcdefghijkl__",          // wrong prefix
		"__KITE_END__c_abcdefghijkl__",           // missing exit code
		"__KITE_END_0_c_abcdefghijkl_extra__",    // extra suffix
		"prefix__KITE_END_0_c_abcdefghijkl__ ok", // embedded — actually valid! see next test
	} {
		// The "embedded" case is intentionally matched (we look for the marker
		// anywhere in the byte stream). Skip just that one here.
		if bad == "prefix__KITE_END_0_c_abcdefghijkl__ ok" {
			continue
		}
		if m := markerRe.FindStringSubmatch(bad); m != nil {
			t.Errorf("%q unexpectedly matched: %v", bad, m)
		}
	}
}

func TestMarkerRegexFindsInMiddleOfStream(t *testing.T) {
	stream := "some output\n__KITE_END_0_c_abcdefghijkl__\nmore"
	idx := markerRe.FindStringIndex(stream)
	if idx == nil {
		t.Fatal("expected match within stream")
	}
	if stream[:idx[0]] != "some output\n" {
		t.Errorf("pre-marker: %q", stream[:idx[0]])
	}
	if stream[idx[1]:] != "\nmore" {
		t.Errorf("post-marker: %q", stream[idx[1]:])
	}
}
