package fixtures

import (
	"strings"
	"testing"
	"time"

	"github.com/flanksource/clicky/task"
)

func TestCompactLinePassing(t *testing.T) {
	result := FixtureResult{
		Status:   task.StatusPASS,
		Duration: 1200 * time.Millisecond,
		Command:  "bash -c 'echo hi'",
		Stdout:   "hi",
	}
	result.Test.Name = "Simple Test"

	got := stripFixtureANSI(result.compactLine().String())
	if !strings.Contains(got, "PASS") && !strings.Contains(got, "✓") {
		t.Errorf("compact pass line missing status indicator: %q", got)
	}
	if !strings.Contains(got, "Simple Test") {
		t.Errorf("compact pass line missing name: %q", got)
	}
	// Verbose blocks must NOT appear in the compact output.
	if strings.Contains(got, "$ bash") {
		t.Errorf("compact pass line leaked command block: %q", got)
	}
	if strings.Contains(got, "stdout") {
		t.Errorf("compact pass line leaked stdout block: %q", got)
	}
}

func TestCompactLineFailingWithError(t *testing.T) {
	result := FixtureResult{
		Status:   task.StatusFAIL,
		Duration: 500 * time.Millisecond,
		Error:    "assertion failed: expected 1, got 2",
	}
	result.Test.Name = "Broken Test"

	got := stripFixtureANSI(result.compactLine().String())
	if !strings.Contains(got, "FAIL") && !strings.Contains(got, "✗") {
		t.Errorf("compact fail line missing status indicator: %q", got)
	}
	if !strings.Contains(got, "Broken Test") {
		t.Errorf("compact fail line missing name: %q", got)
	}
	if !strings.Contains(got, "assertion failed") {
		t.Errorf("compact fail line missing error snippet: %q", got)
	}
}

func TestCompactLineFailingPicksStderrFirst(t *testing.T) {
	result := FixtureResult{
		Status: task.StatusFAIL,
		Stdout: "ignored output",
		Stderr: "real-error: connection refused\nstack frame 2",
	}
	result.Test.Name = "Network Test"

	got := stripFixtureANSI(result.compactLine().String())
	if !strings.Contains(got, "real-error: connection refused") {
		t.Errorf("compact fail line should prefer stderr first line: %q", got)
	}
	// Should NOT include the stack frame (only first non-blank line).
	if strings.Contains(got, "stack frame 2") {
		t.Errorf("compact fail line leaked second stderr line: %q", got)
	}
	// Stdout comes after stderr in the priority order, so when stderr exists it shouldn't appear.
	if strings.Contains(got, "ignored output") {
		t.Errorf("compact fail line leaked stdout when stderr is set: %q", got)
	}
}

func TestCompactLineFailingFallsBackThroughSources(t *testing.T) {
	// Only Stdout is set — error and stderr both empty.
	result := FixtureResult{
		Status: task.StatusFAIL,
		Stdout: "the only output line",
	}
	result.Test.Name = "Stdout-only Test"

	got := stripFixtureANSI(result.compactLine().String())
	if !strings.Contains(got, "the only output line") {
		t.Errorf("compact fail line should fall back to stdout: %q", got)
	}
}

func TestCompactLineStripsANSIFromFailureSnippet(t *testing.T) {
	result := FixtureResult{
		Status: task.StatusFAIL,
		Stderr: "\x1b[31merror: boom\x1b[0m",
	}
	result.Test.Name = "ANSI Test"

	got := stripFixtureANSI(result.compactLine().String())
	if strings.Contains(got, "\x1b[") {
		t.Errorf("compact fail line leaked ANSI escapes: %q", got)
	}
	if !strings.Contains(got, "error: boom") {
		t.Errorf("compact fail line missing error message: %q", got)
	}
}

func TestFirstNonBlankFixtureLine(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"empty", "", ""},
		{"only blank", "\n\n  \t\n", ""},
		{"first real line", "\n  \nreal line\nsecond", "real line"},
		{"strips ansi", "\x1b[31merror\x1b[0m", "error"},
		{"trims surrounding ws", "   padded   \nnext", "padded"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := firstNonBlankFixtureLine(tc.in); got != tc.want {
				t.Errorf("firstNonBlankFixtureLine(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestFixtureFailureSnippetPriority(t *testing.T) {
	// Error wins over stderr/stdout/CEL.
	r := FixtureResult{
		Error:         "explicit error",
		Stderr:        "stderr content",
		Stdout:        "stdout content",
		CELExpression: "cel expr",
	}
	if got := fixtureFailureSnippet(r); got != "explicit error" {
		t.Errorf("want explicit error to win, got %q", got)
	}

	// Stderr wins when error is empty.
	r.Error = ""
	if got := fixtureFailureSnippet(r); got != "stderr content" {
		t.Errorf("want stderr to win, got %q", got)
	}

	// Stdout wins when error + stderr both empty.
	r.Stderr = ""
	if got := fixtureFailureSnippet(r); got != "stdout content" {
		t.Errorf("want stdout to win, got %q", got)
	}

	// CEL wins when all others empty.
	r.Stdout = ""
	if got := fixtureFailureSnippet(r); got != "cel expr" {
		t.Errorf("want cel to win, got %q", got)
	}

	// All empty → empty string.
	r.CELExpression = ""
	if got := fixtureFailureSnippet(r); got != "" {
		t.Errorf("want empty string, got %q", got)
	}
}
