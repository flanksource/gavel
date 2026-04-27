package testui

import (
	"testing"

	"github.com/flanksource/gavel/testrunner/parsers"
)

func TestMergeRerunTestsAppendsAttemptsToMatchingLeaves(t *testing.T) {
	existing := []parsers.Test{{
		Framework:   parsers.GoTest,
		PackagePath: "./pkg/a",
		Name:        "TestFoo",
		Failed:      true,
		Message:     "assertion failed",
		Attempts: []parsers.TestAttempt{{
			Sequence: 1,
			RunKind:  "initial",
			Failed:   true,
			Message:  "assertion failed",
		}},
	}}
	incoming := []parsers.Test{{
		Framework:   parsers.GoTest,
		PackagePath: "./pkg/a",
		Name:        "TestFoo",
		Passed:      true,
		Attempts: []parsers.TestAttempt{{
			Sequence: 1,
			RunKind:  "rerun",
			Passed:   true,
		}},
	}}

	merged := mergeRerunTests(existing, incoming)
	if len(merged) != 1 {
		t.Fatalf("want 1 merged test, got %d", len(merged))
	}
	got := merged[0]
	if !got.Passed || got.Failed {
		t.Error("top-level flags should reflect latest attempt (passed)")
	}
	if len(got.Attempts) != 2 {
		t.Fatalf("want 2 attempts after rerun merge, got %d", len(got.Attempts))
	}
	if got.Attempts[0].RunKind != "initial" || got.Attempts[1].RunKind != "rerun" {
		t.Errorf("attempt order wrong: got %+v", got.Attempts)
	}
	if got.Attempts[0].Sequence != 1 || got.Attempts[1].Sequence != 2 {
		t.Errorf("attempt sequence should be renumbered globally, got %d and %d",
			got.Attempts[0].Sequence, got.Attempts[1].Sequence)
	}
}

func TestMergeRerunTestsRenumbersAcrossMultipleReruns(t *testing.T) {
	// Simulate a 3-attempt history (initial + 2 reruns). Each rerun ships
	// an attempt with Sequence=1 because the runner stamps locally; the
	// merge must hand back a 1/2/3 monotonic sequence.
	history := []parsers.Test{{
		Framework: parsers.GoTest, PackagePath: "./pkg/a", Name: "Test",
		Passed:   false,
		Failed:   true,
		Attempts: []parsers.TestAttempt{{Sequence: 1, RunKind: "initial", Failed: true}},
	}}
	rerunOne := []parsers.Test{{
		Framework: parsers.GoTest, PackagePath: "./pkg/a", Name: "Test",
		Failed:   true,
		Attempts: []parsers.TestAttempt{{Sequence: 1, RunKind: "rerun", Failed: true}},
	}}
	rerunTwo := []parsers.Test{{
		Framework: parsers.GoTest, PackagePath: "./pkg/a", Name: "Test",
		Passed:   true,
		Attempts: []parsers.TestAttempt{{Sequence: 1, RunKind: "rerun", Passed: true}},
	}}

	afterOne := mergeRerunTests(history, rerunOne)
	afterTwo := mergeRerunTests(afterOne, rerunTwo)

	atts := afterTwo[0].Attempts
	if len(atts) != 3 {
		t.Fatalf("want 3 attempts after 2 reruns, got %d", len(atts))
	}
	for i, a := range atts {
		if a.Sequence != i+1 {
			t.Errorf("attempt %d: want Sequence %d, got %d", i, i+1, a.Sequence)
		}
	}
}

func TestMergeRerunTestsAppendsUnknownTests(t *testing.T) {
	existing := []parsers.Test{{Framework: parsers.GoTest, PackagePath: "./pkg/a", Name: "Keep", Passed: true}}
	incoming := []parsers.Test{{Framework: parsers.GoTest, PackagePath: "./pkg/a", Name: "Brand new", Failed: true}}

	merged := mergeRerunTests(existing, incoming)
	if len(merged) != 2 {
		t.Fatalf("want 2 tests after merge (1 kept + 1 appended), got %d", len(merged))
	}
	if merged[0].Name != "Keep" || merged[1].Name != "Brand new" {
		t.Errorf("order wrong: %v, %v", merged[0].Name, merged[1].Name)
	}
}

func TestMergeRerunTestsRecursesIntoChildren(t *testing.T) {
	existing := []parsers.Test{{
		Framework:   parsers.Ginkgo,
		PackagePath: "./pkg/b",
		Name:        "Suite",
		Children: []parsers.Test{
			{Framework: parsers.Ginkgo, PackagePath: "./pkg/b", Name: "spec1", Failed: true,
				Attempts: []parsers.TestAttempt{{Sequence: 1, Failed: true}}},
		},
	}}
	incoming := []parsers.Test{{
		Framework:   parsers.Ginkgo,
		PackagePath: "./pkg/b",
		Name:        "Suite",
		Children: []parsers.Test{
			{Framework: parsers.Ginkgo, PackagePath: "./pkg/b", Name: "spec1", Passed: true,
				Attempts: []parsers.TestAttempt{{Sequence: 1, Passed: true}}},
		},
	}}

	merged := mergeRerunTests(existing, incoming)
	if len(merged) != 1 || len(merged[0].Children) != 1 {
		t.Fatalf("expected one suite with one spec, got %+v", merged)
	}
	spec := merged[0].Children[0]
	if !spec.Passed || spec.Failed {
		t.Error("child spec should reflect latest attempt (passed)")
	}
	if len(spec.Attempts) != 2 {
		t.Errorf("want 2 attempts on child spec, got %d", len(spec.Attempts))
	}
}
