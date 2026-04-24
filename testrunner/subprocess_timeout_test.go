package testrunner

import (
	"testing"
	"time"

	"github.com/flanksource/gavel/testrunner/parsers"
)

func TestSubprocessTimeoutFor(t *testing.T) {
	tests := []struct {
		name    string
		budget  time.Duration
		wantSub time.Duration
	}{
		// 10% margin dominates for budgets > 200s.
		{"5m budget -> 4m30s", 5 * time.Minute, 4*time.Minute + 30*time.Second},
		{"1h budget -> 54m", 60 * time.Minute, 54 * time.Minute},

		// 20s floor kicks in once budget/10 would fall below 20s.
		{"1m budget -> 40s", 1 * time.Minute, 40 * time.Second},
		{"30s budget -> 10s", 30 * time.Second, 10 * time.Second},

		// At or below the floor, there's no room for a margin: skip the flag.
		{"20s budget -> 0 (skip flag)", 20 * time.Second, 0},
		{"10s budget -> 0 (skip flag)", 10 * time.Second, 0},
		{"zero budget -> 0 (skip flag)", 0, 0},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := subprocessTimeoutFor(tc.budget)
			if got != tc.wantSub {
				t.Errorf("subprocessTimeoutFor(%s) = %s, want %s", tc.budget, got, tc.wantSub)
			}
		})
	}
}

func TestSubprocessTimeoutArgsGoTest(t *testing.T) {
	args := subprocessTimeoutArgs(parsers.GoTest, 4*time.Minute+30*time.Second)
	if len(args) != 1 {
		t.Fatalf("want 1 arg, got %v", args)
	}
	if args[0] != "-timeout=4m30s" {
		t.Errorf("want -timeout=4m30s, got %q", args[0])
	}
}

func TestSubprocessTimeoutArgsGinkgo(t *testing.T) {
	args := subprocessTimeoutArgs(parsers.Ginkgo, 4*time.Minute+30*time.Second)
	want := []string{"--timeout=4m30s", "--poll-progress-after=4m20s"}
	if len(args) != len(want) {
		t.Fatalf("want %d args, got %v", len(want), args)
	}
	for i, w := range want {
		if args[i] != w {
			t.Errorf("arg %d: want %q, got %q", i, w, args[i])
		}
	}
}

func TestSubprocessTimeoutArgsGinkgoShortBudget(t *testing.T) {
	// When subTimeout - 10s would go non-positive, poll-after falls back
	// to half the budget so we still arm a progress report.
	args := subprocessTimeoutArgs(parsers.Ginkgo, 8*time.Second)
	if len(args) != 2 {
		t.Fatalf("want 2 args, got %v", args)
	}
	if args[0] != "--timeout=8s" {
		t.Errorf("want --timeout=8s, got %q", args[0])
	}
	if args[1] != "--poll-progress-after=4s" {
		t.Errorf("expected half-budget fallback --poll-progress-after=4s, got %q", args[1])
	}
}

func TestSubprocessTimeoutArgsUnknownFramework(t *testing.T) {
	// Playwright, Jest, Vitest etc. don't use go/ginkgo timeout flags.
	if got := subprocessTimeoutArgs(parsers.Playwright, time.Minute); got != nil {
		t.Errorf("want nil for non-go frameworks, got %v", got)
	}
}

func TestMarkSubtreeTimedOutFlagsNonTerminalLeaves(t *testing.T) {
	// Subtree: one already-failed leaf, one that was still running when the
	// package got killed. Only the second should flip to TimedOut/Failed.
	tree := parsers.Test{
		Name: "Pkg",
		Children: []parsers.Test{
			{Name: "already_failed", Failed: true, Message: "assert mismatch"},
			{Name: "still_running"}, // no terminal flag
		},
	}
	marked := markSubtreeTimedOut(&tree, "killed after 3s")
	if !marked {
		t.Fatal("expected at least one leaf to be marked")
	}
	if !tree.TimedOut {
		t.Error("parent should be flagged TimedOut when a descendant timed out")
	}
	if tree.Children[0].TimedOut {
		t.Error("already-failed child should not be marked timed out")
	}
	if tree.Children[0].Message != "assert mismatch" {
		t.Error("already-failed child's message should be preserved")
	}
	if !tree.Children[1].TimedOut {
		t.Error("non-terminal child should be marked timed out")
	}
	if !tree.Children[1].Failed {
		t.Error("timed out child should also be marked failed")
	}
	if tree.Children[1].Message != "killed after 3s" {
		t.Errorf("non-terminal child should inherit the marker message, got %q", tree.Children[1].Message)
	}
}

func TestExtractStackTraceFromPanicOutput(t *testing.T) {
	stdout := "some preamble\npanic: test timed out after 30s\n\tgoroutine 1 [running]:\nmain.foo()"
	got := extractStackTrace(stdout, "")
	if got == "" {
		t.Fatal("expected stack trace to be extracted")
	}
	if got[:5] != "panic" {
		t.Errorf("expected stack to start at 'panic:', got %q", got[:5])
	}
}

func TestExtractStackTraceReturnsEmptyWhenNoMarker(t *testing.T) {
	if got := extractStackTrace("boring log\n", "also boring\n"); got != "" {
		t.Errorf("expected empty extraction, got %q", got)
	}
}

func TestMarkSubtreeTimedOutReturnsFalseWhenAllTerminal(t *testing.T) {
	tree := parsers.Test{
		Name: "Pkg",
		Children: []parsers.Test{
			{Name: "ok", Passed: true},
			{Name: "bad", Failed: true},
		},
	}
	if markSubtreeTimedOut(&tree, "killed") {
		t.Error("expected no marking when every leaf is terminal")
	}
	if tree.TimedOut {
		t.Error("parent should not be flagged when no descendant timed out")
	}
}
