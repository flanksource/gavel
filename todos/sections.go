package todos

import (
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/text"
)

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
	headings := topLevelSectionHeadings(content)

	start, end, insertAt := -1, -1, -1
	for i := range lines {
		heading := headings[i]
		switch {
		case heading == want || strings.HasPrefix(heading, want+" "):
			if start < 0 {
				start = i
			}
		case start >= 0 && end < 0 && heading != "":
			end = i
		}
		if insertAt < 0 && matchesAny(heading, insertBefore) {
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

func topLevelSectionHeadings(content string) map[int]string {
	source := []byte(content)
	document := goldmark.New().Parser().Parse(text.NewReader(source))
	headings := make(map[int]string)
	for _, heading := range documentHeadings(document) {
		if heading.Level != 2 {
			continue
		}
		line := strings.Count(content[:headingLineStart(source, heading)], "\n")
		headings[line] = "## " + strings.TrimSpace(markdownNodeText(heading, source))
	}
	return headings
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
