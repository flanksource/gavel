package todos

import "strings"

// ReplaceOrAppendSection replaces the markdown "## <header>" section in content
// with newSection (which must include its own "## <header>" heading line), or
// appends it when the section is absent. A section runs until the next "## "
// heading. When appending and one of insertBefore's headers is present, the new
// section is inserted immediately before the earliest such header; otherwise it
// is appended at the end. It generalizes the original "## Latest Failure"
// section replacement so the criteria and verification sections share one
// implementation.
func ReplaceOrAppendSection(content, header, newSection string, insertBefore ...string) string {
	lines := strings.Split(content, "\n")
	want := "## " + header

	start, end, insertAt := -1, -1, -1
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		switch {
		case trimmed == want || strings.HasPrefix(trimmed, want+" "):
			if start < 0 {
				start = i
			}
		case start >= 0 && end < 0 && strings.HasPrefix(trimmed, "## "):
			end = i
		}
		if insertAt < 0 && matchesAny(trimmed, insertBefore) {
			insertAt = i
		}
	}
	if start >= 0 && end < 0 {
		end = len(lines)
	}

	section := strings.TrimRight(newSection, "\n") + "\n"

	var b strings.Builder
	switch {
	case start >= 0:
		writeLines(&b, lines[:start])
		b.WriteString(section)
		writeTail(&b, lines[end:])
	case insertAt >= 0:
		writeLines(&b, lines[:insertAt])
		b.WriteString(section)
		b.WriteString("\n")
		writeTail(&b, lines[insertAt:])
	default:
		trimmed := strings.TrimRight(content, "\n")
		if trimmed != "" {
			b.WriteString(trimmed)
			b.WriteString("\n\n")
		}
		b.WriteString(section)
	}
	return b.String()
}

func matchesAny(trimmed string, headers []string) bool {
	for _, h := range headers {
		want := "## " + h
		if trimmed == want || strings.HasPrefix(trimmed, want+" ") {
			return true
		}
	}
	return false
}

func writeLines(b *strings.Builder, lines []string) {
	for _, l := range lines {
		b.WriteString(l)
		b.WriteString("\n")
	}
}

func writeTail(b *strings.Builder, lines []string) {
	for i, l := range lines {
		b.WriteString(l)
		if i < len(lines)-1 {
			b.WriteString("\n")
		}
	}
}
