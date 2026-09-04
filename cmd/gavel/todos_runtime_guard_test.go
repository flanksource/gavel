package main

import (
	"errors"
	"strings"
	"testing"
)

// A retired `todos run` flag is answered with its replacement, not cobra's bare
// "unknown flag"; a flag that was never retired keeps cobra's error, and a
// non-flag error passes through untouched.
func TestRetiredTodoRunFlagsNameTheirReplacement(t *testing.T) {
	for flag, replacement := range retiredTodoRunFlags {
		err := retiredFlagError(todosRunCmd, errors.New("unknown flag: --"+flag))
		if err == nil || !strings.Contains(err.Error(), replacement) {
			t.Errorf("--%s: got %v, want the replacement %q", flag, err, replacement)
		}
		if !strings.Contains(err.Error(), "--"+flag+" was retired") {
			t.Errorf("--%s: got %v, want it named as retired", flag, err)
		}
	}
	unknown := errors.New("unknown flag: --bogus")
	if err := retiredFlagError(todosRunCmd, unknown); err != unknown {
		t.Errorf("unretired flag: got %v, want cobra's error passed through", err)
	}
	other := errors.New("invalid argument \"x\" for \"--max-turns\"")
	if err := retiredFlagError(todosRunCmd, other); err != other {
		t.Errorf("non unknown-flag error: got %v, want it passed through", err)
	}
	for _, flag := range []string{"driver", "mode", "prompt", "group-by", "check"} {
		if _, ok := retiredTodoRunFlags[flag]; !ok {
			t.Errorf("--%s is documented as retired in MANUAL.md but missing from retiredTodoRunFlags", flag)
		}
	}
}

func TestLegacyFileBackedTODOSyncCLISurfacesAreRemoved(t *testing.T) {
	for _, command := range prCmd.Commands() {
		if command.Name() == "fix" {
			t.Fatal("retired gavel pr fix command is still registered")
		}
	}

	for _, test := range []struct {
		path        []string
		flags       []string
		replacement string
	}{
		{path: []string{"pr", "status"}, flags: []string{"sync-todos"}},
		{path: []string{"lint"}, flags: []string{"sync-todos", "group-by"}},
		{path: []string{"test"}, flags: []string{"sync-todos", "todos-dir", "todo-template"}},
		// The driver vocabulary is gone: the execution mechanism is the compact
		// model's own mode segment, so there is no second axis to select it with.
		{path: []string{"todos", "run"}, flags: []string{"driver"}, replacement: "--model cli:opus:high"},
		// The run/plan mode, the prompt name and grouping are gone too: a run is
		// one lifecycle step for one todo, named with --step.
		{path: []string{"todos", "run"}, flags: []string{"mode", "prompt", "group-by"}, replacement: "--step"},
		// The checks suite is part of the todo's definition of done, rendered from
		// configuration; no run spec is consulted, so a per-run force flag had
		// nothing to force.
		{path: []string{"todos", "run"}, flags: []string{"check"}, replacement: ".gavel.yaml checks.enabled (loop budget: todos.run.workflow.verify.maxIterations)"},
	} {
		command, _, err := rootCmd.Find(test.path)
		if err != nil {
			t.Fatalf("find %v: %v", test.path, err)
		}
		for _, flag := range test.flags {
			if command.Flags().Lookup(flag) == nil {
				continue
			}
			if test.replacement != "" {
				t.Errorf("%v still exposes retired --%s; it is replaced by %s", test.path, flag, test.replacement)
				continue
			}
			t.Errorf("%v still exposes retired --%s", test.path, flag)
		}
	}
}
