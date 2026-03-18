package render

import (
	"fmt"
	"strings"
)

// Table prints a simple ASCII table.
func (r *Renderer) Table(headers []string, rows [][]string) {
	// Calculate column widths.
	widths := make([]int, len(headers))
	for i, h := range headers {
		widths[i] = len(h)
	}
	for _, row := range rows {
		for i, cell := range row {
			if i < len(widths) && len(cell) > widths[i] {
				widths[i] = len(cell)
			}
		}
	}

	sep := "+"
	for _, w := range widths {
		sep += strings.Repeat("-", w+2) + "+"
	}

	fmt.Fprintln(r.w, sep)
	header := "|"
	for i, h := range headers {
		header += fmt.Sprintf(" %-*s |", widths[i], h)
	}
	fmt.Fprintln(r.w, header)
	fmt.Fprintln(r.w, sep)

	for _, row := range rows {
		line := "|"
		for i := range headers {
			cell := ""
			if i < len(row) {
				cell = row[i]
			}
			line += fmt.Sprintf(" %-*s |", widths[i], cell)
		}
		fmt.Fprintln(r.w, line)
	}
	fmt.Fprintln(r.w, sep)
}
