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
// Unknown escapes are dropped. This is deliberately minimal: clicky's
// renderer only uses the sequences above, and settling a full VT100 is out
// of scope.
func settleANSI(raw string) string {
	var rows [][]byte
	rows = append(rows, nil)
	row := 0
	col := 0

	ensureCol := func(r, c int) {
		for len(rows[r]) < c {
			rows[r] = append(rows[r], ' ')
		}
	}

	writeRune := func(b byte) {
		ensureCol(row, col)
		if col < len(rows[row]) {
			rows[row][col] = b
		} else {
			rows[row] = append(rows[row], b)
		}
		col++
	}

	i := 0
	for i < len(raw) {
		b := raw[i]
		switch {
		case b == 0x1b && i+1 < len(raw) && raw[i+1] == '[':
			// CSI sequence: ESC [ params intermediates final
			j := i + 2
			paramsStart := j
			for j < len(raw) && (raw[j] == ';' || raw[j] == '?' || raw[j] == ' ' || (raw[j] >= '0' && raw[j] <= '9')) {
				j++
			}
			if j >= len(raw) {
				i = len(raw)
				break
			}
			params := raw[paramsStart:j]
			final := raw[j]
			i = j + 1

			n := parseFirstParam(params, 1)
			switch final {
			case 'A':
				row -= n
				if row < 0 {
					row = 0
				}
			case 'B':
				row += n
				for row >= len(rows) {
					rows = append(rows, nil)
				}
			case 'E':
				row += n
				col = 0
				for row >= len(rows) {
					rows = append(rows, nil)
				}
			case 'F':
				row -= n
				if row < 0 {
					row = 0
				}
				col = 0
			case 'G':
				col = n - 1
				if col < 0 {
					col = 0
				}
			case 'H', 'f':
				// cursor position: default 1;1
				r, c := parseTwoParams(params, 1, 1)
				row = r - 1
				col = c - 1
				if row < 0 {
					row = 0
				}
				if col < 0 {
					col = 0
				}
				for row >= len(rows) {
					rows = append(rows, nil)
				}
			case 'K':
				// erase in line
				mode := parseFirstParam(params, 0)
				switch mode {
				case 0:
					if col < len(rows[row]) {
						rows[row] = rows[row][:col]
					}
				case 1:
					ensureCol(row, col)
					for k := 0; k < col && k < len(rows[row]); k++ {
						rows[row][k] = ' '
					}
				case 2:
					rows[row] = nil
				}
			case 'J':
				mode := parseFirstParam(params, 0)
				if mode == 2 {
					rows = [][]byte{nil}
					row = 0
					col = 0
				}
			default:
				// SGR (m), cursor hide/show (?25l/h), etc. — no visual effect
			}
		case b == 0x1b && i+1 < len(raw) && raw[i+1] == ']':
			// OSC: ESC ] ... BEL or ST (ESC \)
			j := i + 2
			for j < len(raw) {
				if raw[j] == 0x07 {
					j++
					break
				}
				if raw[j] == 0x1b && j+1 < len(raw) && raw[j+1] == '\\' {
					j += 2
					break
				}
				j++
			}
			i = j
		case b == 0x1b:
			// Lone ESC followed by an unknown introducer; skip it and the next byte.
			i += 2
			if i > len(raw) {
				i = len(raw)
			}
		case b == '\r':
			col = 0
			i++
		case b == '\n':
			row++
			col = 0
			for row >= len(rows) {
				rows = append(rows, nil)
			}
			i++
		case b == '\b':
			if col > 0 {
				col--
			}
			i++
		case b == '\t':
			next := (col/8 + 1) * 8
			for col < next {
				writeRune(' ')
			}
			i++
		case b < 0x20:
			// drop other control chars
			i++
		default:
			writeRune(b)
			i++
		}
	}

	var b strings.Builder
	for idx, r := range rows {
		b.Write(r)
		if idx < len(rows)-1 {
			b.WriteByte('\n')
		}
	}
	return b.String()
}

func parseFirstParam(params string, def int) int {
	p := strings.TrimPrefix(params, "?")
	if p == "" {
		return def
	}
	if i := strings.IndexByte(p, ';'); i >= 0 {
		p = p[:i]
	}
	n := 0
	if p == "" {
		return def
	}
	for _, c := range p {
		if c < '0' || c > '9' {
			return def
		}
		n = n*10 + int(c-'0')
	}
	return n
}

func parseTwoParams(params string, d1, d2 int) (int, int) {
	p := strings.TrimPrefix(params, "?")
	if p == "" {
		return d1, d2
	}
	parts := strings.SplitN(p, ";", 2)
	n1 := d1
	n2 := d2
	if parts[0] != "" {
		n := 0
		ok := true
		for _, c := range parts[0] {
			if c < '0' || c > '9' {
				ok = false
				break
			}
			n = n*10 + int(c-'0')
		}
		if ok {
			n1 = n
		}
	}
	if len(parts) == 2 && parts[1] != "" {
		n := 0
		ok := true
		for _, c := range parts[1] {
			if c < '0' || c > '9' {
				ok = false
				break
			}
			n = n*10 + int(c-'0')
		}
		if ok {
			n2 = n
		}
	}
	return n1, n2
}

// duplicateLines returns every non-empty line that appears more than once in
// the ANSI-settled view of raw. Leading/trailing whitespace is trimmed for
// the comparison so spinner frames like " ⠋ task" and "  ⠋ task" don't
// spuriously differ.
func duplicateLines(raw string) []DupLine {
	settled := settleANSI(raw)
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

// hasDuplicateLines is the CEL-friendly boolean companion to duplicateLines.
func hasDuplicateLines(raw string) bool {
	return len(duplicateLines(raw)) > 0
}

// finalText exposes settleANSI under a name that reads well in CEL.
func finalText(raw string) string {
	return settleANSI(raw)
}

// Debug_SettleANSI is an exported wrapper for the hack/ analysis scripts.
// Not part of the public API — named with underscore to discourage use.
func Debug_SettleANSI(raw string) string { return settleANSI(raw) }

// Debug_DuplicateLines is an exported wrapper for hack/ analysis scripts.
func Debug_DuplicateLines(raw string) []DupLine { return duplicateLines(raw) }
