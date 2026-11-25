// Package patchselect reduces a unified git diff to a chosen set of changed
// lines, so the result can be applied to the index with `git apply --cached`
// (staging) or `git apply --cached --reverse` (unstaging). It follows the
// git-gui / lazygit line-selection algorithm.
package patchselect

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// Side names which file a selected line number refers to: Old for the
// pre-image line number of a '-' line, New for the post-image line number of a
// '+' line.
type Side string

const (
	Old Side = "old"
	New Side = "new"
)

// Line is one selected changed line.
type Line struct {
	Side Side `json:"side"`
	Line int  `json:"line"`
}

// Options select the lines to keep and the direction the result is applied in.
type Options struct {
	Lines []Line
	// Reverse builds a patch for `git apply --reverse`: unselected '+' lines
	// become context and unselected '-' lines are dropped. Forward (staging)
	// drops unselected '+' lines and turns unselected '-' lines into context.
	Reverse bool
}

var (
	ErrBinary      = errors.New("binary diffs cannot be staged by line")
	ErrNoSelection = errors.New("no changed lines selected")
)

var hunkHeader = regexp.MustCompile(`^@@ -(\d+)(?:,(\d+))? \+(\d+)(?:,(\d+))? @@(.*)$`)

type hunk struct {
	oldStart, oldCount, newStart, newCount int
	suffix                                 string
	body                                   []string
}

type section struct {
	header []string
	hunks  []hunk
	binary bool
}

// Filter returns patch reduced to the selected changed lines. Every selected
// line must be a changed line of patch.
func Filter(patch string, opts Options) (string, error) {
	selected, err := selectionSet(opts.Lines)
	if err != nil {
		return "", err
	}
	sections, err := parse(patch)
	if err != nil {
		return "", err
	}
	if len(sections) == 0 {
		return "", errors.New("patch has no hunks")
	}
	matched := map[Line]bool{}
	var out strings.Builder
	for _, sec := range sections {
		if sec.binary {
			return "", ErrBinary
		}
		out.WriteString(filterSection(sec, selectionState{selected: selected, matched: matched, reverse: opts.Reverse}))
	}
	for _, line := range opts.Lines {
		if !matched[line] {
			return "", fmt.Errorf("%s:%d is not a changed line of the diff", line.Side, line.Line)
		}
	}
	if out.Len() == 0 {
		return "", ErrNoSelection
	}
	return out.String(), nil
}

func selectionSet(lines []Line) (map[Line]bool, error) {
	if len(lines) == 0 {
		return nil, ErrNoSelection
	}
	set := make(map[Line]bool, len(lines))
	for _, line := range lines {
		if line.Side != Old && line.Side != New {
			return nil, fmt.Errorf("line %d has unknown side %q (want %q or %q)", line.Line, line.Side, Old, New)
		}
		if line.Line <= 0 {
			return nil, fmt.Errorf("line number %d must be positive", line.Line)
		}
		set[line] = true
	}
	return set, nil
}

func parse(patch string) ([]section, error) {
	var sections []section
	for _, line := range splitLines(patch) {
		if strings.HasPrefix(line, "diff --git ") || len(sections) == 0 {
			sections = append(sections, section{})
		}
		current := &sections[len(sections)-1]
		if strings.HasPrefix(line, "@@") {
			parsed, err := parseHunkHeader(line)
			if err != nil {
				return nil, err
			}
			current.hunks = append(current.hunks, parsed)
			continue
		}
		if len(current.hunks) == 0 {
			current.header = append(current.header, line)
			current.binary = current.binary || strings.HasPrefix(line, "Binary files ") || line == "GIT binary patch"
			continue
		}
		if line == "" || !strings.ContainsRune(" +-\\", rune(line[0])) {
			return nil, fmt.Errorf("unexpected line in hunk: %q", line)
		}
		last := &current.hunks[len(current.hunks)-1]
		last.body = append(last.body, line)
	}
	return sections, nil
}

func splitLines(patch string) []string {
	if patch == "" {
		return nil
	}
	return strings.Split(strings.TrimSuffix(patch, "\n"), "\n")
}

func parseHunkHeader(line string) (hunk, error) {
	match := hunkHeader.FindStringSubmatch(line)
	if match == nil {
		return hunk{}, fmt.Errorf("malformed hunk header %q", line)
	}
	number := func(value string) int {
		if value == "" {
			return 1
		}
		parsed, _ := strconv.Atoi(value) // the regexp only admits digits
		return parsed
	}
	return hunk{
		oldStart: number(match[1]), oldCount: number(match[2]),
		newStart: number(match[3]), newCount: number(match[4]),
		suffix: match[5],
	}, nil
}

type selectionState struct {
	selected map[Line]bool
	matched  map[Line]bool
	reverse  bool
}

// filterSection renders one file's kept hunks, or "" when none keeps a line.
func filterSection(sec section, state selectionState) string {
	var hunks strings.Builder
	offset, contextualized := 0, false
	for _, h := range sec.hunks {
		filtered, ok := filterHunk(h, state)
		if !ok {
			continue
		}
		contextualized = contextualized || filtered.contextualized
		oldStart, newStart := filtered.starts(h, offset, state.reverse)
		offset += filtered.newCount - filtered.oldCount
		fmt.Fprintf(&hunks, "@@ -%d,%d +%d,%d @@%s\n", oldStart, filtered.oldCount, newStart, filtered.newCount, h.suffix)
		for _, line := range filtered.body {
			hunks.WriteString(line + "\n")
		}
	}
	if hunks.Len() == 0 {
		return ""
	}
	header := sec.header
	if contextualized && createsOrDeletes(header) {
		header = rewriteCreationHeader(header)
	}
	if len(header) == 0 {
		return hunks.String()
	}
	return strings.Join(header, "\n") + "\n" + hunks.String()
}

type filteredHunk struct {
	oldCount, newCount int
	body               []string
	// contextualized is set once an unselected change line became context, so
	// a creation or deletion no longer creates or deletes the file.
	contextualized bool
}

func filterHunk(h hunk, state selectionState) (filteredHunk, bool) {
	var out filteredHunk
	oldLine, newLine := h.oldStart, h.newStart
	lastKept, anySelected := false, false
	for _, line := range h.body {
		kind, text := line[0], line[1:]
		if kind == '\\' {
			if lastKept {
				out.body = append(out.body, line)
			}
			continue
		}
		key := Line{Side: New, Line: newLine}
		if kind == '-' {
			key = Line{Side: Old, Line: oldLine}
		}
		lastKept = true
		switch {
		case kind == ' ':
			out.keep(line, true, true)
		case state.selected[key]:
			state.matched[key], anySelected = true, true
			out.keep(line, kind == '-', kind == '+')
		case (kind == '-') != state.reverse:
			out.keep(" "+text, true, true)
			out.contextualized = true
		default:
			lastKept = false
		}
		oldLine, newLine = advance(kind, oldLine, newLine)
	}
	return out, anySelected
}

func (f *filteredHunk) keep(line string, oldSide, newSide bool) {
	f.body = append(f.body, line)
	if oldSide {
		f.oldCount++
	}
	if newSide {
		f.newCount++
	}
}

func advance(kind byte, oldLine, newLine int) (int, int) {
	switch kind {
	case ' ':
		return oldLine + 1, newLine + 1
	case '-':
		return oldLine + 1, newLine
	default:
		return oldLine, newLine + 1
	}
}

// starts recomputes the hunk header starts. Forward application anchors on
// the pre-image, so the old start stays and the new start follows the kept
// hunks' net line delta; reverse anchors on the post-image instead. A side
// with zero lines names the line before the hunk, per the unified format.
func (f filteredHunk) starts(h hunk, offset int, reverse bool) (int, int) {
	first := func(start, count int) int {
		if count == 0 {
			return start + 1
		}
		return start
	}
	start := func(first, count int) int {
		if count == 0 {
			return first - 1
		}
		return first
	}
	if reverse {
		return start(first(h.newStart, h.newCount)-offset, f.oldCount), h.newStart
	}
	return h.oldStart, start(first(h.oldStart, h.oldCount)+offset, f.newCount)
}

func createsOrDeletes(header []string) bool {
	for _, line := range header {
		if strings.HasPrefix(line, "new file mode ") || strings.HasPrefix(line, "deleted file mode ") {
			return true
		}
	}
	return false
}

// rewriteCreationHeader turns a file creation or deletion header into a
// modification header: a partial selection leaves the file in place on both
// sides.
func rewriteCreationHeader(header []string) []string {
	var minus, plus string
	for _, line := range header {
		if strings.HasPrefix(line, "--- ") {
			minus = strings.TrimPrefix(line, "--- ")
		}
		if strings.HasPrefix(line, "+++ ") {
			plus = strings.TrimPrefix(line, "+++ ")
		}
	}
	rewritten := make([]string, 0, len(header))
	for _, line := range header {
		switch {
		case strings.HasPrefix(line, "new file mode "), strings.HasPrefix(line, "deleted file mode "), strings.HasPrefix(line, "index "):
			continue
		case line == "--- /dev/null":
			line = "--- a/" + strings.TrimPrefix(plus, "b/")
		case line == "+++ /dev/null":
			line = "+++ b/" + strings.TrimPrefix(minus, "a/")
		}
		rewritten = append(rewritten, line)
	}
	return rewritten
}
