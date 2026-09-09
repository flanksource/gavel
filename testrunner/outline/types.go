// Package outline enumerates supported tests without executing test bodies and
// annotates each with location, size, complexity, duplication, and run history.
package outline

import (
	"context"
	"fmt"
	"strings"

	"github.com/flanksource/gavel/testrunner/history"
	"github.com/flanksource/gavel/testrunner/parsers"
)

type Options struct {
	WorkDir      string
	Paths        []string            // positional path filters, relative to WorkDir
	Frameworks   []parsers.Framework // empty = every supported outline framework
	FixtureFiles []string            // fixture globs; empty = config or **/*.fixture.md
	AISummary    bool
	Duplication  bool
	History      bool
	Context      context.Context
}

// SupportedFrameworks returns the stable set accepted by test outline. Fixture
// is intentionally not part of parsers.AllFrameworks because it is executed by
// the fixture runner rather than a testrunner framework runner.
func SupportedFrameworks() []parsers.Framework {
	return []parsers.Framework{
		parsers.GoTest,
		parsers.Ginkgo,
		parsers.Jest,
		parsers.Vitest,
		parsers.Playwright,
		parsers.Fixture,
	}
}

// ParseFramework resolves CLI names accepted by test outline.
func ParseFramework(name string) (parsers.Framework, error) {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "fixture", "fixtures":
		return parsers.Fixture, nil
	}
	framework, err := parsers.ParseFramework(name)
	if err == nil {
		return framework, nil
	}
	known := make([]string, 0, len(SupportedFrameworks()))
	for _, supported := range SupportedFrameworks() {
		known = append(known, supported.String())
	}
	return "", fmt.Errorf("unknown framework %q; known: %s", name, strings.Join(known, ", "))
}

// Entry is one test (or ginkgo container) in the outline.
type Entry struct {
	Framework      parsers.Framework `json:"framework"`
	File           string            `json:"file"`
	Line           int               `json:"line,omitempty"`
	EndLine        int               `json:"end_line,omitempty"`
	Name           string            `json:"name"`
	Suite          []string          `json:"suite,omitempty"`
	Container      bool              `json:"container,omitempty"`
	Dynamic        bool              `json:"dynamic,omitempty"`
	Pending        bool              `json:"pending,omitempty"`
	Focused        bool              `json:"focused,omitempty"`
	Labels         []string          `json:"labels,omitempty"`
	SizeLines      int               `json:"size_lines,omitempty"`
	Complexity     int               `json:"complexity,omitempty"`
	DuplicationPct float64           `json:"duplication_pct,omitempty"`
	Description    string            `json:"description,omitempty"`
	AISummary      string            `json:"ai_summary,omitempty"`
	Error          string            `json:"error,omitempty"` // collection failed for this package/file
	History        *history.Entry    `json:"history,omitempty"`
	Children       []*Entry          `json:"children,omitempty"`

	calls []string // called identifiers in the test body, for static descriptions
}

type Report struct {
	WorkDir  string   `json:"work_dir"`
	RunCount int      `json:"run_count,omitempty"`
	Entries  []*Entry `json:"entries"`
}

// Leaves returns every executable (non-container) entry in the report,
// depth-first. Error rows (a package whose tests could not be collected) are
// not tests and are excluded, so they never count toward totals or get history,
// duplication, or AI annotations; they still render in the tree.
func (r *Report) Leaves() []*Entry {
	var leaves []*Entry
	var walk func(*Entry)
	walk = func(e *Entry) {
		if !e.Container && e.Error == "" {
			leaves = append(leaves, e)
		}
		for _, child := range e.Children {
			walk(child)
		}
	}
	for _, e := range r.Entries {
		walk(e)
	}
	return leaves
}
