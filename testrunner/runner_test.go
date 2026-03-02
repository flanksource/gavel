package testrunner

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/flanksource/clicky/task"
	"github.com/flanksource/gavel/fixtures"
	"github.com/flanksource/gavel/testrunner/parsers"
)

func TestStripExitStatus(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{name: "empty", input: "", expected: ""},
		{name: "no exit status", input: "some error output", expected: "some error output"},
		{name: "only exit status", input: "exit status 1", expected: ""},
		{name: "exit status at end", input: "some error\nexit status 1", expected: "some error"},
		{name: "exit status 2", input: "error output\nexit status 2", expected: "error output"},
		{name: "exit status with trailing newline", input: "error\nexit status 1\n", expected: "error"},
		{name: "exit status in middle", input: "before\nexit status 1\nafter", expected: "before\n\nafter"},
		{name: "multiple exit statuses", input: "exit status 1\nexit status 2", expected: ""},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := stripExitStatus(tc.input); got != tc.expected {
				t.Errorf("stripExitStatus(%q) = %q, want %q", tc.input, got, tc.expected)
			}
		})
	}
}

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

func TestFixtureNodeToTests(t *testing.T) {
	tests := []struct {
		name     string
		node     *fixtures.FixtureNode
		expected []parsers.Test
	}{
		{
			name: "leaf node with passed result",
			node: &fixtures.FixtureNode{
				Name: "echo test",
				Type: fixtures.TestNode,
				Results: &fixtures.FixtureResult{
					Name:     "echo test",
					Status:   task.StatusPASS,
					Duration: 100 * time.Millisecond,
					Stdout:   "hello\n",
					Stderr:   "warn\n",
					Error:    "",
				},
			},
			expected: []parsers.Test{{
				Name:      "echo test",
				Framework: "fixture",
				Duration:  100 * time.Millisecond,
				Stdout:    "hello\n",
				Stderr:    "warn\n",
				Passed:    true,
				Failed:    false,
			}},
		},
		{
			name: "leaf node with failed result",
			node: &fixtures.FixtureNode{
				Name: "bad test",
				Type: fixtures.TestNode,
				Results: &fixtures.FixtureResult{
					Name:     "bad test",
					Status:   task.StatusFAIL,
					Duration: 50 * time.Millisecond,
					Stdout:   "out",
					Stderr:   "err",
					Error:    "exit code 1",
				},
			},
			expected: []parsers.Test{{
				Name:      "bad test",
				Framework: "fixture",
				Duration:  50 * time.Millisecond,
				Stdout:    "out",
				Stderr:    "err",
				Failed:    true,
				Passed:    false,
				Message:   "exit code 1",
			}},
		},
		{
			name: "section node wraps children",
			node: &fixtures.FixtureNode{
				Name: "section",
				Type: fixtures.SectionNode,
				Children: []*fixtures.FixtureNode{
					{
						Name: "child test",
						Type: fixtures.TestNode,
						Results: &fixtures.FixtureResult{
							Name:   "child test",
							Status: task.StatusPASS,
						},
					},
				},
			},
			expected: []parsers.Test{{
				Name: "section",
				Children: parsers.Tests{{
					Name:      "child test",
					Framework: "fixture",
					Passed:    true,
				}},
			}},
		},
		{
			name: "node without results or section type returns children directly",
			node: &fixtures.FixtureNode{
				Name: "root",
				Type: fixtures.NodeType(99), // unknown type
				Children: []*fixtures.FixtureNode{
					{
						Name: "test1",
						Type: fixtures.TestNode,
						Results: &fixtures.FixtureResult{
							Name:   "test1",
							Status: task.StatusPASS,
						},
					},
				},
			},
			expected: []parsers.Test{{
				Name:      "test1",
				Framework: "fixture",
				Passed:    true,
			}},
		},
		{
			name: "ERR status maps to failed",
			node: &fixtures.FixtureNode{
				Name: "err test",
				Type: fixtures.TestNode,
				Results: &fixtures.FixtureResult{
					Name:   "err test",
					Status: task.StatusERR,
					Error:  "timeout",
				},
			},
			expected: []parsers.Test{{
				Name:      "err test",
				Framework: "fixture",
				Failed:    true,
				Message:   "timeout",
			}},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := fixtureNodeToTests(tc.node)
			if len(got) != len(tc.expected) {
				t.Fatalf("expected %d tests, got %d: %+v", len(tc.expected), len(got), got)
			}
			for i, exp := range tc.expected {
				if got[i].Name != exp.Name {
					t.Errorf("test[%d].Name = %q, want %q", i, got[i].Name, exp.Name)
				}
				if got[i].Framework != exp.Framework {
					t.Errorf("test[%d].Framework = %q, want %q", i, got[i].Framework, exp.Framework)
				}
				if got[i].Failed != exp.Failed {
					t.Errorf("test[%d].Failed = %v, want %v", i, got[i].Failed, exp.Failed)
				}
				if got[i].Passed != exp.Passed {
					t.Errorf("test[%d].Passed = %v, want %v", i, got[i].Passed, exp.Passed)
				}
				if got[i].Message != exp.Message {
					t.Errorf("test[%d].Message = %q, want %q", i, got[i].Message, exp.Message)
				}
				if got[i].Stdout != exp.Stdout {
					t.Errorf("test[%d].Stdout = %q, want %q", i, got[i].Stdout, exp.Stdout)
				}
				if got[i].Stderr != exp.Stderr {
					t.Errorf("test[%d].Stderr = %q, want %q", i, got[i].Stderr, exp.Stderr)
				}
				if got[i].Duration != exp.Duration {
					t.Errorf("test[%d].Duration = %v, want %v", i, got[i].Duration, exp.Duration)
				}
				if len(exp.Children) > 0 && len(got[i].Children) != len(exp.Children) {
					t.Errorf("test[%d].Children length = %d, want %d", i, len(got[i].Children), len(exp.Children))
				}
			}
		})
	}
}
