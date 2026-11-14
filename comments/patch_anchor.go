package comments

import (
	"fmt"
	"slices"
	"strings"
)

type PatchAnchor struct {
	Version   int      `json:"version"`
	StartLine int      `json:"startLine"`
	EndLine   int      `json:"endLine"`
	Before    []string `json:"before"`
	Lines     []string `json:"lines"`
	After     []string `json:"after"`
}
type PatchLocation struct {
	State     string `json:"state"`
	StartLine int    `json:"startLine,omitempty"`
	EndLine   int    `json:"endLine,omitempty"`
}

func CapturePatch(source string, start, end int) (PatchAnchor, error) {
	lines := strings.Split(strings.ReplaceAll(source, "\r\n", "\n"), "\n")
	if start < 1 || end < start || end > len(lines) {
		return PatchAnchor{}, fmt.Errorf("%w: source range outside document", ErrInvalid)
	}
	return PatchAnchor{Version: 1, StartLine: start, EndLine: end, Before: slices.Clone(lines[max(0, start-4) : start-1]), Lines: slices.Clone(lines[start-1 : end]), After: slices.Clone(lines[end:min(len(lines), end+3)])}, nil
}
func ResolvePatch(source string, anchor PatchAnchor) PatchLocation {
	if anchor.Version != 1 || len(anchor.Lines) == 0 || anchor.StartLine < 1 || anchor.EndLine-anchor.StartLine+1 != len(anchor.Lines) || len(anchor.Before) > 3 || len(anchor.After) > 3 {
		return PatchLocation{State: "outdated"}
	}
	lines := strings.Split(strings.ReplaceAll(source, "\r\n", "\n"), "\n")
	hunk := append(append(slices.Clone(anchor.Before), anchor.Lines...), anchor.After...)
	expected := anchor.StartLine - 1 - len(anchor.Before)
	matches := func(i int) bool {
		return i >= 0 && i+len(hunk) <= len(lines) && slices.Equal(lines[i:i+len(hunk)], hunk)
	}
	location := func(i int) PatchLocation {
		start := i + len(anchor.Before) + 1
		return PatchLocation{State: "matched", StartLine: start, EndLine: start + len(anchor.Lines) - 1}
	}
	if matches(expected) {
		return location(expected)
	}
	found := -1
	for i := 0; i+len(hunk) <= len(lines); i++ {
		if matches(i) {
			if found >= 0 {
				return PatchLocation{State: "ambiguous"}
			}
			found = i
		}
	}
	if found >= 0 {
		return location(found)
	}
	return PatchLocation{State: "outdated"}
}
