package fixtures

import (
	"fmt"
	"slices"
	"strings"

	"github.com/goccy/go-yaml/ast"
	"github.com/goccy/go-yaml/parser"
)

// lineSpan is a 1-based, inclusive range of source lines.
type lineSpan struct{ start, end int }

// stepSpans locates every step of every fixture document in data. A step ends
// on the line before the next step, top-level key or document, or at the end
// of the file, without the blank lines, comments or separators in between.
func stepSpans(data []byte) ([][]lineSpan, error) {
	file, err := parser.ParseBytes(data, 0)
	if err != nil {
		return nil, err
	}
	var starts [][]int
	var boundaries []int
	for i, doc := range file.Docs {
		if doc.Body == nil {
			continue
		}
		body, ok := doc.Body.(*ast.MappingNode)
		if !ok {
			return nil, fmt.Errorf("document %d: want a mapping, got %s", i+1, doc.Body.Type())
		}
		var docStarts []int
		for _, field := range body.Values {
			boundaries = append(boundaries, field.Key.GetToken().Position.Line)
			if field.Key.GetToken().Value != "steps" {
				continue
			}
			value := field.Value
			if anchor, ok := value.(*ast.AnchorNode); ok {
				value = anchor.Value
			}
			steps, ok := value.(*ast.SequenceNode)
			if !ok {
				return nil, fmt.Errorf("document %d: steps: want a sequence, got %s", i+1, value.Type())
			}
			for _, step := range steps.Values {
				docStarts = append(docStarts, step.GetToken().Position.Line)
			}
		}
		boundaries = append(boundaries, docStarts...)
		starts = append(starts, docStarts)
	}
	slices.Sort(boundaries)
	lines := strings.Split(string(data), "\n")
	spans := make([][]lineSpan, len(starts))
	for i, docStarts := range starts {
		for _, start := range docStarts {
			end := len(lines)
			if next, _ := slices.BinarySearch(boundaries, start+1); next < len(boundaries) {
				end = boundaries[next] - 1
			}
			spans[i] = append(spans[i], lineSpan{start: start, end: trimTrailingFiller(lines, start, end)})
		}
	}
	return spans, nil
}

// setMarkdownEndLines ends each test node of a markdown tree on the line
// before the next node, or at the end of content; a table row is one line.
func setMarkdownEndLines(tree *FixtureNode, content string) {
	var boundaries []int
	tree.Walk(func(n *FixtureNode) {
		if n.Origin != nil && n.Origin.Line > 0 {
			boundaries = append(boundaries, n.Origin.Line)
		}
	})
	slices.Sort(boundaries)
	lines := strings.Split(content, "\n")
	tree.Walk(func(n *FixtureNode) {
		if n.Test == nil || n.Origin == nil || n.Origin.Line == 0 {
			return
		}
		if n.Origin.Kind == "table-row" {
			n.Origin.EndLine = n.Origin.Line
			return
		}
		end := len(lines)
		if next, _ := slices.BinarySearch(boundaries, n.Origin.Line+1); next < len(boundaries) {
			end = boundaries[next] - 1
		}
		n.Origin.EndLine = trimTrailingFiller(lines, n.Origin.Line, end)
	})
}

// trimTrailingFiller moves end back over blank lines, comments and document
// separators, never past start.
func trimTrailingFiller(lines []string, start, end int) int {
	for end > start {
		line := strings.TrimSpace(lines[end-1])
		if line != "" && !strings.HasPrefix(line, "#") && line != "---" && line != "..." {
			break
		}
		end--
	}
	return end
}
