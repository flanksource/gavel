// Package outline statically enumerates tests (go test functions, ginkgo
// specs, vitest tests) without executing them and annotates each with
// location, size, complexity, duplication, and run history so test quality
// can be evaluated at a glance.
package outline

import (
	"context"

	"github.com/flanksource/gavel/testrunner/history"
	"github.com/flanksource/gavel/testrunner/parsers"
)

type Options struct {
	WorkDir     string
	Paths       []string            // positional path filters, relative to WorkDir
	Frameworks  []parsers.Framework // empty = gotest + ginkgo + vitest
	AISummary   bool
	Duplication bool
	History     bool
	Context     context.Context
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
// depth-first.
func (r *Report) Leaves() []*Entry {
	var leaves []*Entry
	var walk func(*Entry)
	walk = func(e *Entry) {
		if !e.Container {
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
