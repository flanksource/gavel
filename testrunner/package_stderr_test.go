package testrunner

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/flanksource/clicky/exec"
	"github.com/flanksource/gavel/testrunner/parsers"
	"github.com/flanksource/gavel/testrunner/runners"
)

const vitestReportWithAFailure = `{
  "numTotalTests": 2,
  "testResults": [{
    "name": "src/sum.test.ts",
    "status": "failed",
    "message": "",
    "assertionResults": [
      {"title": "adds", "ancestorTitles": ["sum"], "status": "passed", "duration": 1, "failureMessages": []},
      {"title": "subtracts", "ancestorTitles": ["sum"], "status": "failed", "duration": 1, "failureMessages": []}
    ]
  }]
}`

// The package's stderr is stored once, on its suite result. Copying it onto
// every passing test made a 400 KB stderr cost 400 KB per test in the snapshot,
// twice with each test's attempt; a failing test with none of its own keeps it,
// since it may hold the reason it failed.
func TestParseTestResultsKeepsPackageStderrOffPassingTests(t *testing.T) {
	workDir := t.TempDir()
	report := filepath.Join(workDir, "vitest-report.json")
	if err := os.WriteFile(report, []byte(vitestReportWithAFailure), 0o644); err != nil {
		t.Fatalf("write report: %v", err)
	}
	const packageStderr = "Warning: An update to App inside a test was not wrapped in act(...)."
	o := &TestOrchestrator{RunOptions: RunOptions{WorkDir: workDir}}
	testRun := &runners.TestRun{
		Framework:  parsers.Vitest,
		Parser:     parsers.NewJestJSON(workDir, parsers.Vitest),
		Process:    exec.NewExec("vitest", "run"),
		ReportPath: report,
	}

	results, err := o.parseTestResults(testRun, &exec.ExecResult{Stderr: packageStderr}, "./")
	if err != nil {
		t.Fatalf("parseTestResults: %v", err)
	}
	if len(results) != 1 || results[0].Stderr != packageStderr {
		t.Fatalf("suite stderr = %q, want the package stderr", results[0].Stderr)
	}
	byName := map[string]string{}
	for _, test := range results[0].Tests {
		byName[test.Name] = test.Stderr
	}
	want := map[string]string{"adds": "", "subtracts": packageStderr}
	if len(byName) != len(want) || byName["adds"] != want["adds"] || byName["subtracts"] != want["subtracts"] {
		t.Fatalf("test stderr = %q, want %q", byName, want)
	}
}
