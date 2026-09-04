package fixtures

import (
	"strings"
)

// DupLine describes a single line that appears more than once in the
// settled-text view of a PTY capture.
type DupLine struct {
	Text  string `json:"text"`
	Count int    `json:"count"`
}

// settleANSI interprets the ANSI byte stream the way a real terminal would,
// producing the text the user actually sees at end of stream. It honors:
//
//   - CSI n A                cursor up n rows (default 1)
//   - CSI n B                cursor down n rows (default 1)
//   - CSI n E / CSI n F      cursor next/prev line (start of line)
//   - CSI 2K / CSI K / CSI 0K / CSI 1K  erase in line
//   - CSI 2J                 erase display (used rarely by clicky)
//   - CSI ?25l / ?25h        hide/show cursor (ignored — no visual effect)
//   - CR (\r)                column 0
//   - LF (\n)                new row (auto-scroll append)
//
// When width > 0 the grid auto-wraps printable output at that column (DEC
// pending-wrap semantics), so a line longer than the terminal occupies multiple
// *physical* rows and cursor moves operate on those physical rows. This is what
// surfaces wrap-induced redraw bugs: a renderer that emits a cursor-up sized by
// logical line count lands mid-content when the content wrapped, leaving the
// smear behind. width <= 0 disables wrapping (rows grow unbounded — the legacy
// behavior for callers that don't know the terminal width).
//
// Unknown escapes are dropped. This is deliberately minimal: clicky's
// renderer only uses the sequences above, and settling a full VT100 is out
// of scope.
func settleANSI(raw string, width int) string {
	screen := newANSIScreen(width, 0)
	screen.Write([]byte(raw))
	return screen.String()
}

// duplicateLines returns every non-empty line that appears more than once in
// the ANSI-settled view of raw. Leading/trailing whitespace is trimmed for
// the comparison so spinner frames like " ⠋ task" and "  ⠋ task" don't
// spuriously differ.
func duplicateLines(raw string, width int) []DupLine {
	return duplicateSettledLines(settleANSI(raw, width))
}

func duplicateSettledLines(settled string) []DupLine {
	counts := make(map[string]int)
	order := []string{}
	for _, line := range strings.Split(settled, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		if _, seen := counts[trimmed]; !seen {
			order = append(order, trimmed)
		}
		counts[trimmed]++
	}
	var dups []DupLine
	for _, line := range order {
		if counts[line] > 1 {
			dups = append(dups, DupLine{Text: line, Count: counts[line]})
		}
	}
	return dups
}

// finalText exposes settleANSI under a name that reads well in CEL.
func finalText(raw string, width int) string {
	return settleANSI(raw, width)
}

// Debug_SettleANSI is an exported wrapper for the hack/ analysis scripts.
// Not part of the public API — named with underscore to discourage use.
func Debug_SettleANSI(raw string, width int) string { return settleANSI(raw, width) }

// Debug_DuplicateLines is an exported wrapper for hack/ analysis scripts.
func Debug_DuplicateLines(raw string, width int) []DupLine { return duplicateLines(raw, width) }
