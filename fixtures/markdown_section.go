package fixtures

import (
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

// markdownSection returns the body under the first heading whose text equals
// title (case-insensitively), up to the next heading at the same level or
// higher. A document with no such heading yields the empty string — the caller
// asked for a section that is not there, and inventing the whole document in its
// place is how a scoped read silently becomes an unscoped one.
func markdownSection(content, title string) string {
	source := []byte(content)
	document := goldmark.New().Parser().Parse(text.NewReader(source))

	var target *ast.Heading
	for node := document.FirstChild(); node != nil; node = node.NextSibling() {
		heading, ok := node.(*ast.Heading)
		if !ok {
			continue
		}
		if target == nil {
			if strings.EqualFold(strings.TrimSpace(nodeText(heading, source)), strings.TrimSpace(title)) {
				target = heading
			}
			continue
		}
		if heading.Level <= target.Level {
			return strings.TrimSpace(content[sectionContentStart(source, target):sectionLineStart(source, heading)])
		}
	}
	if target == nil {
		return ""
	}
	return strings.TrimSpace(content[sectionContentStart(source, target):])
}

// sectionLineStart is the offset of the first byte of a heading's own line.
func sectionLineStart(source []byte, heading *ast.Heading) int {
	if heading.Lines().Len() == 0 {
		return 0
	}
	start := heading.Lines().At(0).Start
	for start > 0 && source[start-1] != '\n' {
		start--
	}
	return start
}

// sectionContentStart is the offset just past a heading's line — where its body
// begins. A setext heading is underlined, so its body starts a line later.
func sectionContentStart(source []byte, heading *ast.Heading) int {
	start := sectionLineStart(source, heading)
	end := sectionLineEnd(source, start)
	if strings.HasPrefix(strings.TrimLeft(string(source[start:end]), " \t"), "#") {
		return end
	}
	return sectionLineEnd(source, end)
}

func sectionLineEnd(source []byte, start int) int {
	for start < len(source) {
		start++
		if source[start-1] == '\n' {
			break
		}
	}
	return start
}

func nodeText(node ast.Node, source []byte) string {
	var builder strings.Builder
	_ = ast.Walk(node, func(child ast.Node, entering bool) (ast.WalkStatus, error) {
		if entering {
			if textNode, ok := child.(*ast.Text); ok {
				builder.Write(textNode.Segment.Value(source))
			}
		}
		return ast.WalkContinue, nil
	})
	return builder.String()
}
