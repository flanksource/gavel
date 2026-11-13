package testrunner

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/flanksource/gavel/testrunner/parsers"
	"github.com/flanksource/gavel/testrunner/runners"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
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

var _ = Describe("subprocess timeout argument", func() {
	DescribeTable("arms the in-process deadline only on go test commands",
		func(framework parsers.Framework, budget string, want []string) {
			workDir := GinkgoT().TempDir()
			GinkgoT().Setenv("HOME", GinkgoT().TempDir())
			Expect(os.WriteFile(filepath.Join(workDir, ".gavel.yaml"), []byte("runtime: {timeout: "+budget+"}\n"), 0o600)).To(Succeed())
			if framework == parsers.Ginkgo {
				code := "package sample\nimport \"github.com/onsi/ginkgo/v2\"\nvar _ = ginkgo.Describe(\"timeout\", func() {})\n"
				Expect(os.WriteFile(filepath.Join(workDir, "sample_test.go"), []byte(code), 0o600)).To(Succeed())
			}
			runner, ok := DefaultRegistry(workDir).Get(framework)
			Expect(ok).To(BeTrue())
			o := &TestOrchestrator{RunOptions: RunOptions{WorkDir: workDir}}
			run, err := o.buildCommand(runner, ".", runners.TestOptions{})
			Expect(err).NotTo(HaveOccurred())
			var timeouts []string
			for _, arg := range run.Process.Args {
				if strings.Contains(arg, "timeout") {
					timeouts = append(timeouts, arg)
				}
			}
			Expect(timeouts).To(Equal(want))
		},
		Entry("go test, 5m budget", parsers.GoTest, "5m", []string{"-timeout=4m30s"}),
		Entry("Ginkgo, 5m budget", parsers.Ginkgo, "5m", []string{"-timeout=4m30s"}),
		Entry("go test, budget too short for a margin", parsers.GoTest, "20s", nil),
		Entry("Playwright has no go test deadline", parsers.Playwright, "5m", nil),
	)
})

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
