package todos

import (
	"strings"
	"testing"

	"github.com/flanksource/gavel/todos/types"
)

// checkTaskName labels the rows a bulk check renders. Filename() falls back to
// the title for DB-backed TODOs, which two TODOs can share, so the short id has
// to disambiguate them or a triage sweep renders a column of identical rows.
func TestCheckTaskNameDisambiguatesSameTitledTODOs(t *testing.T) {
	first := &types.TODO{
		ID: "1111", ShortID: "ab12cd", Provider: "db",
		TODOFrontmatter: types.TODOFrontmatter{Title: "Fix the parser"},
	}
	second := &types.TODO{
		ID: "2222", ShortID: "ef34gh", Provider: "db",
		TODOFrontmatter: types.TODOFrontmatter{Title: "Fix the parser"},
	}

	if first.Filename() != second.Filename() {
		t.Skip("Filename() already disambiguates these; the guard below is then redundant")
	}
	if checkTaskName(first) == checkTaskName(second) {
		t.Fatalf("two same-titled TODOs share the task label %q", checkTaskName(first))
	}
	for _, todo := range []*types.TODO{first, second} {
		name := checkTaskName(todo)
		if !strings.Contains(name, todo.ShortID) {
			t.Errorf("label %q omits short id %q", name, todo.ShortID)
		}
		if !strings.Contains(name, "Fix the parser") {
			t.Errorf("label %q dropped the title", name)
		}
	}
}

func TestCheckTaskNameWithoutShortID(t *testing.T) {
	todo := &types.TODO{FilePath: "todo.md"}
	if got := checkTaskName(todo); got != todo.Filename() {
		t.Fatalf("checkTaskName() = %q, want the filename %q", got, todo.Filename())
	}
}

// Each check runs the TODO's fixture — a real test suite — so an unbounded
// fan-out over a large triage selection thrashes the machine rather than
// finishing sooner.
func TestDefaultCheckConcurrencyIsBounded(t *testing.T) {
	if DefaultCheckConcurrency <= 0 {
		t.Fatalf("DefaultCheckConcurrency = %d, want a positive bound", DefaultCheckConcurrency)
	}
	if DefaultCheckConcurrency > 16 {
		t.Errorf("DefaultCheckConcurrency = %d, which is high enough to thrash on a laptop", DefaultCheckConcurrency)
	}
}
