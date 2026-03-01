package testrunner

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/flanksource/gavel/testrunner/parsers"
)

func TestRunnerDetectAndRun(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a simple test file
	testFile := filepath.Join(tmpDir, "simple_test.go")
	content := `package main

import "testing"

func TestPass(t *testing.T) {
	if 1+1 != 2 {
		t.Error("math broken")
	}
}
`
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	runner := &TestOrchestrator{
		RunOptions: RunOptions{
			WorkDir: tmpDir,
		},
		registry: DefaultRegistry(tmpDir),
	}
	frameworks := []Framework{GoTest}

	results, err := runner.detectAndRun(frameworks, nil, nil)

	if err != nil {
		t.Logf("Got expected result (may have failures or errors): %v", err)
	}
	if len(results) > 0 {
		totalTests := 0
		totalPassed := 0
		for _, result := range results {
			summary := result.Sum()
			totalTests += len(result.Tests)
			totalPassed += summary.Passed
		}
		t.Logf("Got %d tests, %d passed", totalTests, totalPassed)
	}
}

func TestRunnerRun(t *testing.T) {
	tmpDir := t.TempDir()
	todosDir := filepath.Join(tmpDir, ".todos")

	// Create a simple test file
	testFile := filepath.Join(tmpDir, "simple_test.go")
	content := `package main

import "testing"

func TestPass(t *testing.T) {
	t.Skip("skipped")
}
`
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	opts := RunOptions{
		TodosDir:  todosDir,
		WorkDir:   tmpDir,
		SyncTodos: false,
	}

	results, err := Run(opts)
	if err != nil {
		t.Logf("Run completed with error: %v", err)
	}

	// Type assert results to []parsers.Test
	if tests, ok := results.([]parsers.Test); ok && len(tests) > 0 {
		totalTests := len(tests)
		totalPassed := 0
		totalFailed := 0
		for _, test := range tests {
			if test.Passed {
				totalPassed++
			}
			if test.Failed {
				totalFailed++
			}
		}
		t.Logf("Run completed with %d total tests, %d passed, %d failed", totalTests, totalPassed, totalFailed)
	}
}

func TestRunnerNoTests(t *testing.T) {
	tmpDir := t.TempDir()

	_, err := Run(RunOptions{WorkDir: tmpDir})
	if err == nil {
		t.Error("expected error for no tests")
	}
}

func TestDiscoverFixtures(t *testing.T) {
	tmpDir := t.TempDir()

	// Create fixtures/ subdirectory with a .md file
	fixturesDir := filepath.Join(tmpDir, "fixtures")
	if err := os.MkdirAll(fixturesDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(fixturesDir, "test.md"), []byte("# Test"), 0644); err != nil {
		t.Fatal(err)
	}

	// Create fixture*.md in root
	if err := os.WriteFile(filepath.Join(tmpDir, "fixture-cli.md"), []byte("# CLI"), 0644); err != nil {
		t.Fatal(err)
	}

	// Create a non-matching file
	if err := os.WriteFile(filepath.Join(tmpDir, "readme.md"), []byte("# Readme"), 0644); err != nil {
		t.Fatal(err)
	}

	found := discoverFixtures(tmpDir)

	if len(found) != 2 {
		t.Fatalf("expected 2 fixture files, got %d: %v", len(found), found)
	}

	// Verify expected files are found
	foundMap := make(map[string]bool)
	for _, f := range found {
		foundMap[filepath.Base(f)] = true
	}
	if !foundMap["test.md"] {
		t.Error("expected fixtures/test.md to be discovered")
	}
	if !foundMap["fixture-cli.md"] {
		t.Error("expected fixture-cli.md to be discovered")
	}
	if foundMap["readme.md"] {
		t.Error("readme.md should not be discovered")
	}
}
