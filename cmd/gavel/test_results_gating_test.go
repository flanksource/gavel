package main

import (
	"io"
	"os"
	"strings"
	"testing"

	"github.com/flanksource/clicky"
	"github.com/flanksource/gavel/testrunner"
	"github.com/flanksource/gavel/testrunner/parsers"
)

// withFormat sets clicky's resolved output format for the duration of fn and
// restores it after, so format-gated behaviour can be exercised without a real
// CLI invocation. Format is a process-global, hence the restore.
func withFormat(format string, fn func()) {
	prev := clicky.Flags.Format
	clicky.Flags.Format = format
	defer func() { clicky.Flags.Format = prev }()
	fn()
}

// captureStdout runs fn with os.Stdout redirected to a pipe and returns what it
// wrote, so the gating test can assert on whether the section breakdown printed.
func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	orig := os.Stdout
	os.Stdout = w
	fn()
	_ = w.Close()
	os.Stdout = orig
	out, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	return string(out)
}

func passingTestTree() []parsers.Test {
	return []parsers.Test{{
		Name:     "pkg",
		Passed:   true,
		Children: []parsers.Test{{Name: "TestThing", Passed: true}},
	}}
}

func TestPrintTestRunResults_SkipsBreakdownForSerializedFormat(t *testing.T) {
	opts := testrunner.RunOptions{ShowPassed: true}
	out := captureStdout(t, func() {
		withFormat("markdown", func() {
			printTestRunResults(passingTestTree(), opts, parsers.Tests(passingTestTree()).Sum(), nil)
		})
	})
	if out != "" {
		t.Errorf("serialized format must not print the terminal breakdown (it would float above the tree), got:\n%s", out)
	}
}

func TestPrintTestRunResults_PrintsBreakdownForPretty(t *testing.T) {
	opts := testrunner.RunOptions{ShowPassed: true}
	out := captureStdout(t, func() {
		withFormat("pretty", func() {
			printTestRunResults(passingTestTree(), opts, parsers.Tests(passingTestTree()).Sum(), nil)
		})
	})
	if !strings.Contains(out, "Passed tests") {
		t.Errorf("pretty format must print the passed-tests breakdown, got:\n%s", out)
	}
	if !strings.Contains(out, "Test summary") {
		t.Errorf("pretty format must print the summary, got:\n%s", out)
	}
}
