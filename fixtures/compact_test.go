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

func TestDisplayOptionsForVerbosity(t *testing.T) {
	base := DisplayOptions{
		ShowStdout: OutputOnFailure,
		ShowStderr: OutputOnFailure,
	}

	v1 := DisplayOptionsForVerbosity(1, base, false, false)
	if !v1.ShowPassed {
		t.Fatal("-v should show passed fixture rows")
	}
	if v1.ShowCommand {
		t.Fatal("-v should not show command details yet")
	}

	v2 := DisplayOptionsForVerbosity(2, base, false, false)
	if !v2.ShowCommand {
		t.Fatal("-vv should show command details")
	}

	v3 := DisplayOptionsForVerbosity(3, base, false, false)
	if v3.ShowStdout != OutputAlways || v3.ShowStderr != OutputAlways || !v3.ShowCELVars {
		t.Fatalf("-vvv should show streams and CEL variables, got %+v", v3)
	}

	explicit := DisplayOptionsForVerbosity(3, base, true, false)
	if explicit.ShowStdout != OutputOnFailure {
		t.Fatalf("explicit --show-stdout should override -vvv, got %s", explicit.ShowStdout)
	}
}

func TestApplyDisplayOptionsHidesPassedByDefault(t *testing.T) {
	root := &FixtureNode{
		Name: "root",
		Type: SectionNode,
		Children: []*FixtureNode{{
			Name: "file.md",
			Type: FileNode,
			Children: []*FixtureNode{{
				Name: "pass",
				Type: TestNode,
				Results: &FixtureResult{
					Status: task.StatusPASS,
					Test:   FixtureTest{Name: "pass"},
				},
			}, {
				Name: "fail",
				Type: TestNode,
				Results: &FixtureResult{
					Status: task.StatusFAIL,
					Test:   FixtureTest{Name: "fail"},
					Stdout: "failed output",
				},
			}},
		}},
	}

	root.ApplyDisplayOptions(DisplayOptions{
		ShowStdout: OutputOnFailure,
		ShowStderr: OutputOnFailure,
	})

	file := root.Children[0]
	if len(file.Children) != 1 {
		t.Fatalf("expected only failing child to remain, got %d", len(file.Children))
	}
	if file.Children[0].Name != "fail" {
		t.Fatalf("expected failing child to remain, got %s", file.Children[0].Name)
	}
	if file.Children[0].Results.Display == nil {
		t.Fatal("visible result should carry display options")
	}
}

func TestFixtureResultPrettyHonorsDisplayOptions(t *testing.T) {
	hiddenPass := FixtureResult{
		Status:  task.StatusPASS,
		Test:    FixtureTest{Name: "hidden pass"},
		Display: &DisplayOptions{ShowStdout: OutputOnFailure, ShowStderr: OutputOnFailure},
	}
	if got := strings.TrimSpace(stripFixtureANSI(hiddenPass.Pretty().String())); got != "" {
		t.Fatalf("hidden passing result should render empty text, got %q", got)
	}

	result := FixtureResult{
		Status:  task.StatusPASS,
		Test:    FixtureTest{Name: "visible pass"},
		Command: "bash -c echo hi",
		Stdout:  "hello",
		Display: &DisplayOptions{
			ShowPassed:  true,
			ShowCommand: false,
			ShowStdout:  OutputNever,
			ShowStderr:  OutputOnFailure,
		},
	}

	hidden := stripFixtureANSI(result.Pretty().String())
	if strings.Contains(hidden, "bash -c") || strings.Contains(hidden, "hello") {
		t.Fatalf("display options should hide command/stdout, got %q", hidden)
	}

	result.Display.ShowCommand = true
	result.Display.ShowStdout = OutputAlways
	shown := stripFixtureANSI(result.Pretty().String())
	if !strings.Contains(shown, "bash -c") || !strings.Contains(shown, "hello") {
		t.Fatalf("display options should show command/stdout, got %q", shown)
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
