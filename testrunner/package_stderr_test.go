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

// A test that fails with stderr of its own ("Failed to import test file …")
// never got the package stderr, and the tree was built from tests alone, so
// the process's own explanation (Vite's reload, a crash) was lost. A failed
// package keeps it on its own node, once.
func TestAttachFailedPackageStderrKeepsTheProcessStderrOnTheFailedPackage(t *testing.T) {
	const (
		failedPkg     = "./apps/storybook"
		passedPkg     = "./packages/ui"
		copiedPkg     = "./apps/playground"
		importFailure = "Failed to import test file src/Chat.stories.tsx"
		viteReload    = "[vite] new dependencies optimized: react-rnd\n[vite] ✨ optimized dependencies changed. reloading"
		passingNoise  = "Warning: An update to App inside a test was not wrapped in act(...)."
		copiedStderr  = "Error: Cannot find module './missing'"
	)
	results := parsers.TestSuiteResults{
		{Stderr: viteReload + "\nexit status 1", Tests: parsers.Tests{
			{Name: "Chat.stories.tsx", PackagePath: failedPkg, Failed: true, Stderr: importFailure},
			{Name: "Button.stories.tsx", PackagePath: failedPkg, Passed: true},
		}},
		{Stderr: passingNoise, Tests: parsers.Tests{
			{Name: "sum.test.ts", PackagePath: passedPkg, Passed: true},
		}},
		{Stderr: copiedStderr, Tests: parsers.Tests{
			{Name: "setup.test.ts", PackagePath: copiedPkg, Failed: true, Stderr: copiedStderr},
		}},
	}

	tree := attachFailedPackageStderr(parsers.BuildTestTree(results.All()), results)

	got := map[string]string{}
	var walk func(nodes []parsers.Test)
	walk = func(nodes []parsers.Test) {
		for _, node := range nodes {
			if len(node.Children) > 0 {
				got[node.PackagePath] = node.Stderr
				walk(node.Children)
			}
		}
	}
	walk(tree)
	want := map[string]string{
		"apps": "", "apps/storybook": viteReload, "apps/playground": "",
		"packages": "", "packages/ui": "",
	}
	for pkg, stderr := range want {
		if got[pkg] != stderr {
			t.Errorf("package %s stderr = %q, want %q", pkg, got[pkg], stderr)
		}
	}
}
