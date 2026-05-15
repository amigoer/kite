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
			// Submatch layout: [1]=full END token, [2]=exit, [3]=cmdID,
			// [4]=full PROMPT token. END matches must leave [4] empty.
			if m[4] != "" {
				t.Errorf("expected end-marker match, got prompt: %q", m[4])
			}
			gotExit, _ := strconv.Atoi(m[2])
			if gotExit != c.wantExit {
				t.Errorf("exit: %d, want %d", gotExit, c.wantExit)
			}
			if m[3] != c.wantCmdID {
				t.Errorf("cmdID: %s, want %s", m[3], c.wantCmdID)
			}
		})
	}
}

func TestMarkerRegexMatchesPromptSentinel(t *testing.T) {
	m := markerRe.FindStringSubmatch("__KITE_PROMPT__")
	if m == nil {
		t.Fatal("expected prompt sentinel to match")
	}
	if m[4] != "__KITE_PROMPT__" {
		t.Errorf("group [4] should be the prompt sentinel, got %q", m[4])
	}
	if m[1] != "" || m[2] != "" || m[3] != "" {
		t.Errorf("end-marker groups should be empty on prompt match: %v", m[1:4])
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

func TestStripScrollbackClear(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"no_clear", "hello world", "hello world"},
		{"just_clear_scrollback", "\x1b[3J", ""},
		{"full_clear", "\x1b[H\x1b[2J\x1b[3J", "\x1b[H\x1b[2J"},
		{"clear_amid_output", "before\x1b[3Jafter", "beforeafter"},
		{"multiple_clears", "a\x1b[3Jb\x1b[3Jc", "abc"},
		// We must NOT touch the "erase screen" sequence — only the
		// scrollback extension. Otherwise `clear` wouldn't visibly clear.
		{"keep_erase_screen", "\x1b[2J", "\x1b[2J"},
		// Don't get fooled by lookalikes.
		{"keep_other_csi", "\x1b[3K", "\x1b[3K"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := string(stripScrollbackClear([]byte(tc.in)))
			if got != tc.want {
				t.Errorf("stripScrollbackClear(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}
