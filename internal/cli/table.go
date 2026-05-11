package cli

import (
	"fmt"
	"io"
	"strings"
	"unicode/utf8"
)

// printTable prints headers + rows in a single-line, padded layout.
func printTable(w io.Writer, headers []string, rows [][]string) {
	widths := make([]int, len(headers))
	for i, h := range headers {
		widths[i] = utf8.RuneCountInString(h)
	}
	for _, row := range rows {
		for i, cell := range row {
			if i >= len(widths) {
				continue
			}
			if c := utf8.RuneCountInString(cell); c > widths[i] {
				widths[i] = c
			}
		}
	}

	render := func(cells []string) string {
		parts := make([]string, len(headers))
		for i, w := range widths {
			cell := ""
			if i < len(cells) {
				cell = cells[i]
			}
			parts[i] = cell + strings.Repeat(" ", w-utf8.RuneCountInString(cell))
		}
		return strings.Join(parts, "  ")
	}

	fmt.Fprintln(w, render(headers))
	fmt.Fprintln(w, render(separators(widths)))
	for _, row := range rows {
		fmt.Fprintln(w, render(row))
	}
}

func separators(widths []int) []string {
	out := make([]string, len(widths))
	for i, w := range widths {
		out[i] = strings.Repeat("-", w)
	}
	return out
}
