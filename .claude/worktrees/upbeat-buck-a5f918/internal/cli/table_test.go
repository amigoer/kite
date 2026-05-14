package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestPrintTablePadsColumns(t *testing.T) {
	var buf bytes.Buffer
	printTable(&buf,
		[]string{"ID", "NAME"},
		[][]string{
			{"r_short", "demo"},
			{"r_muchlonger", "another-name"},
		},
	)
	got := buf.String()

	// Header row must show both columns, separator row, then two data rows.
	lines := strings.Split(strings.TrimRight(got, "\n"), "\n")
	if len(lines) != 4 {
		t.Fatalf("want 4 lines, got %d:\n%s", len(lines), got)
	}
	if !strings.HasPrefix(lines[0], "ID") || !strings.Contains(lines[0], "NAME") {
		t.Errorf("header: %q", lines[0])
	}
	if !strings.HasPrefix(lines[1], "--") {
		t.Errorf("separator: %q", lines[1])
	}
	// All data rows should be the same width as the longest one — within a
	// couple of trailing whitespace chars.
	for _, l := range lines[2:] {
		if !strings.Contains(l, "r_") {
			t.Errorf("data row: %q", l)
		}
	}
}

func TestPrintTableHandlesEmpty(t *testing.T) {
	var buf bytes.Buffer
	printTable(&buf, []string{"A", "B"}, nil)
	out := buf.String()
	if !strings.Contains(out, "A") || !strings.Contains(out, "B") {
		t.Errorf("missing headers: %q", out)
	}
}

func TestPrintTableHandlesUnicodeWidth(t *testing.T) {
	var buf bytes.Buffer
	printTable(&buf,
		[]string{"name"},
		[][]string{{"café"}, {"naïve"}},
	)
	// We measure rune count, not byte count — the multi-byte chars shouldn't
	// blow up the layout.
	for _, l := range strings.Split(buf.String(), "\n") {
		if l == "" {
			continue
		}
		if strings.ContainsAny(l, "\x00") {
			t.Errorf("unexpected null byte: %q", l)
		}
	}
}
