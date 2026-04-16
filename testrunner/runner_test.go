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

func writeGoTestPackage(t *testing.T, repoRoot, pkgDir, modulePath string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(repoRoot, ".git"), 0o755); err != nil {
		t.Fatalf("create git dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(repoRoot, "go.mod"), []byte("module "+modulePath+"\n\ngo 1.22\n"), 0o644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}
	fullPkgDir := filepath.Join(repoRoot, pkgDir)
	if err := os.MkdirAll(fullPkgDir, 0o755); err != nil {
		t.Fatalf("create package dir: %v", err)
	}
	testFile := filepath.Join(fullPkgDir, "sample_test.go")
	content := `package sample

import "testing"

func TestPass(t *testing.T) {}
`
	if err := os.WriteFile(testFile, []byte(content), 0o644); err != nil {
		t.Fatalf("write test file: %v", err)
	}
}

func collectPackagePaths(tests []parsers.Test, seen map[string]bool) {
	for _, test := range tests {
		if test.PackagePath != "" {
			seen[test.PackagePath] = true
		}
		if len(test.Children) > 0 {
			collectPackagePaths(test.Children, seen)
		}
	}
}

func TestGroupPathsByGitRoot(t *testing.T) {
	// Create two fake git repos in temp dirs
	repoA := t.TempDir()
	repoB := t.TempDir()
	os.MkdirAll(filepath.Join(repoA, ".git"), 0755)
	os.MkdirAll(filepath.Join(repoB, ".git"), 0755)
	os.MkdirAll(filepath.Join(repoA, "pkg"), 0755)
	os.MkdirAll(filepath.Join(repoB, "cmd"), 0755)

	t.Run("single repo stays unchanged", func(t *testing.T) {
		groups, err := groupPathsByGitRoot(repoA, []string{"./pkg"})
		if err != nil {
			t.Fatalf("groupPathsByGitRoot returned error: %v", err)
		}
		if len(groups) != 1 {
			t.Fatalf("expected 1 group, got %d", len(groups))
		}
		if groups[0].workDir != repoA {
			t.Errorf("expected workDir=%s, got %s", repoA, groups[0].workDir)
		}
		if len(groups[0].paths) != 1 || groups[0].paths[0] != "./pkg" {
			t.Errorf("expected paths=[./pkg], got %v", groups[0].paths)
		}
	})

	t.Run("cross-repo subdir paths are split", func(t *testing.T) {
		// Simulate: cwd=repoA, paths=["./pkg", "<repoB>/cmd"]
		groups, err := groupPathsByGitRoot(repoA, []string{"./pkg", filepath.Join(repoB, "cmd")})
		if err != nil {
			t.Fatalf("groupPathsByGitRoot returned error: %v", err)
		}
		if len(groups) != 2 {
			t.Fatalf("expected 2 groups, got %d: %+v", len(groups), groups)
		}

		byRoot := make(map[string]testGroup)
		for _, g := range groups {
			byRoot[g.workDir] = g
		}

		groupA, ok := byRoot[repoA]
		if !ok {
			t.Fatalf("expected group for repoA=%s", repoA)
		}
		if len(groupA.paths) != 1 || groupA.paths[0] != "./pkg" {
			t.Errorf("repoA paths: expected [./pkg], got %v", groupA.paths)
		}

		groupB, ok := byRoot[repoB]
		if !ok {
			t.Fatalf("expected group for repoB=%s", repoB)
		}
		if len(groupB.paths) != 1 || groupB.paths[0] != "./cmd" {
			t.Errorf("repoB paths: expected [./cmd], got %v", groupB.paths)
		}
	})

	t.Run("cross-repo root path produces empty paths", func(t *testing.T) {
		// When the path IS the git root (e.g. "../otherrepo"), the runner
		// should get empty paths so it discovers from the root.
		groups, err := groupPathsByGitRoot(repoA, []string{repoB})
		if err != nil {
			t.Fatalf("groupPathsByGitRoot returned error: %v", err)
		}
		if len(groups) != 1 {
			t.Fatalf("expected 1 group, got %d: %+v", len(groups), groups)
		}
		if groups[0].workDir != repoB {
			t.Errorf("expected workDir=%s, got %s", repoB, groups[0].workDir)
		}
		if len(groups[0].paths) != 0 {
			t.Errorf("expected empty paths (discover from root), got %v", groups[0].paths)
		}
	})

	t.Run("relative cross-repo path", func(t *testing.T) {
		// Simulate: cwd=repoA, paths=["../otherrepo/cmd"] where otherrepo is repoB
		// We need to actually use a relative path, so put both repos under a shared parent
		parent := t.TempDir()
		rA := filepath.Join(parent, "repoA")
		rB := filepath.Join(parent, "repoB")
		os.MkdirAll(filepath.Join(rA, ".git"), 0755)
		os.MkdirAll(filepath.Join(rB, ".git"), 0755)
		os.MkdirAll(filepath.Join(rA, "src"), 0755)
		os.MkdirAll(filepath.Join(rB, "lib"), 0755)

		groups, err := groupPathsByGitRoot(rA, []string{"./src", "../repoB/lib"})
		if err != nil {
			t.Fatalf("groupPathsByGitRoot returned error: %v", err)
		}
		if len(groups) != 2 {
			t.Fatalf("expected 2 groups, got %d: %+v", len(groups), groups)
		}

		byRoot := make(map[string]testGroup)
		for _, g := range groups {
			byRoot[g.workDir] = g
		}

		if g, ok := byRoot[rA]; !ok {
			t.Error("missing group for rA")
		} else if len(g.paths) != 1 || g.paths[0] != "./src" {
			t.Errorf("rA paths: expected [./src], got %v", g.paths)
		}

		if g, ok := byRoot[rB]; !ok {
			t.Error("missing group for rB")
		} else if len(g.paths) != 1 || g.paths[0] != "./lib" {
			t.Errorf("rB paths: expected [./lib], got %v", g.paths)
		}
	})

	t.Run("empty paths returns single group with workdir", func(t *testing.T) {
		groups, err := groupPathsByGitRoot(repoA, nil)
		if err != nil {
			t.Fatalf("groupPathsByGitRoot returned error: %v", err)
		}
		if len(groups) != 1 {
			t.Fatalf("expected 1 group, got %d", len(groups))
		}
		if groups[0].workDir != repoA {
			t.Errorf("expected workDir=%s, got %s", repoA, groups[0].workDir)
		}
		if len(groups[0].paths) != 0 {
			t.Errorf("expected empty paths, got %v", groups[0].paths)
		}
	})

	t.Run("nested go module path gets its own workdir", func(t *testing.T) {
		repo := t.TempDir()
		if err := os.MkdirAll(filepath.Join(repo, ".git"), 0o755); err != nil {
			t.Fatalf("create repo git dir: %v", err)
		}
		nestedModule := filepath.Join(repo, "submodule")
		if err := os.MkdirAll(filepath.Join(nestedModule, "pkg"), 0o755); err != nil {
			t.Fatalf("create nested module dir: %v", err)
		}
		if err := os.WriteFile(filepath.Join(nestedModule, "go.mod"), []byte("module example.com/sub\n"), 0o644); err != nil {
			t.Fatalf("write nested go.mod: %v", err)
		}

		groups, err := groupPathsByGitRoot(repo, []string{"./submodule/pkg"})
		if err != nil {
			t.Fatalf("groupPathsByGitRoot returned error: %v", err)
		}
		if len(groups) != 1 {
			t.Fatalf("expected 1 group, got %d: %+v", len(groups), groups)
		}
		if groups[0].workDir != nestedModule {
			t.Fatalf("expected workDir=%s, got %s", nestedModule, groups[0].workDir)
		}
		if len(groups[0].paths) != 1 || groups[0].paths[0] != "./pkg" {
			t.Fatalf("expected paths=[./pkg], got %v", groups[0].paths)
		}
	})

	t.Run("discover from root adds nested go modules as separate groups", func(t *testing.T) {
		repo := t.TempDir()
		if err := os.MkdirAll(filepath.Join(repo, ".git"), 0o755); err != nil {
			t.Fatalf("create repo git dir: %v", err)
		}
		if err := os.WriteFile(filepath.Join(repo, "go.mod"), []byte("module example.com/root\n"), 0o644); err != nil {
			t.Fatalf("write root go.mod: %v", err)
		}
		nestedModule := filepath.Join(repo, "submodule")
		if err := os.MkdirAll(filepath.Join(nestedModule, "pkg"), 0o755); err != nil {
			t.Fatalf("create nested module dir: %v", err)
		}
		if err := os.WriteFile(filepath.Join(nestedModule, "go.mod"), []byte("module example.com/sub\n"), 0o644); err != nil {
			t.Fatalf("write nested go.mod: %v", err)
		}
		nestedRepo := filepath.Join(repo, "subrepo")
		if err := os.MkdirAll(filepath.Join(nestedRepo, ".git"), 0o755); err != nil {
			t.Fatalf("create nested repo git dir: %v", err)
		}
		if err := os.WriteFile(filepath.Join(nestedRepo, "go.mod"), []byte("module example.com/subrepo\n"), 0o644); err != nil {
			t.Fatalf("write nested repo go.mod: %v", err)
		}

		groups, err := groupPathsByGitRoot(repo, nil)
		if err != nil {
			t.Fatalf("groupPathsByGitRoot returned error: %v", err)
		}
		if len(groups) != 2 {
			t.Fatalf("expected 2 groups (root + nested module), got %d: %+v", len(groups), groups)
		}
		if groups[0].workDir != repo || len(groups[0].paths) != 0 {
			t.Fatalf("expected root discover group first, got %+v", groups[0])
		}
		if groups[1].workDir != nestedModule || len(groups[1].paths) != 0 {
			t.Fatalf("expected nested module discover group second, got %+v", groups[1])
		}
	})
}

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

func TestRunMultiRootKeepsSharedUpdatesOpenUntilAllRootsComplete(t *testing.T) {
	parent := t.TempDir()
	repoA := filepath.Join(parent, "repoA")
	repoB := filepath.Join(parent, "repoB")
	writeGoTestPackage(t, repoA, "pkg1", "example.com/repoA")
	writeGoTestPackage(t, repoB, "pkg2", "example.com/repoB")

	updates := make(chan []parsers.Test, 32)
	_, err := Run(RunOptions{
		WorkDir:       parent,
		StartingPaths: []string{filepath.Join(repoA, "pkg1"), filepath.Join(repoB, "pkg2")},
		Updates:       updates,
	})
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	seenPackages := make(map[string]bool)
	timeout := time.After(10 * time.Second)
	for {
		select {
		case batch, ok := <-updates:
			if !ok {
				if !seenPackages["./pkg1"] || !seenPackages["./pkg2"] {
					t.Fatalf("expected streamed updates for both multiroot packages, saw %v", seenPackages)
				}
				return
			}
			collectPackagePaths(batch, seenPackages)
		case <-timeout:
			t.Fatal("timed out waiting for multiroot updates channel to close")
		}
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

	mustWrite := func(rel, body string) {
		full := filepath.Join(tmpDir, rel)
		if err := os.MkdirAll(filepath.Dir(full), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(body), 0644); err != nil {
			t.Fatal(err)
		}
	}
	mustWrite("cli.fixture.md", "# cli")
	mustWrite("sub/nested.fixture.md", "# nested")
	mustWrite("fixtures/old.md", "# old style, should be ignored")
	mustWrite("readme.md", "# readme")

	found := discoverFixtures(tmpDir, nil, []string{"**/*.fixture.md"})

	foundMap := make(map[string]bool)
	for _, f := range found {
		foundMap[filepath.Base(f)] = true
	}

	if !foundMap["cli.fixture.md"] {
		t.Error("expected cli.fixture.md to be discovered")
	}
	if !foundMap["nested.fixture.md"] {
		t.Error("expected sub/nested.fixture.md to be discovered")
	}
	if foundMap["old.md"] {
		t.Error("fixtures/old.md should NOT be discovered under new glob")
	}
	if foundMap["readme.md"] {
		t.Error("readme.md should NOT be discovered")
	}
}

func TestDiscoverFixturesEmptyGlobs(t *testing.T) {
	tmpDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmpDir, "x.fixture.md"), []byte("# x"), 0644); err != nil {
		t.Fatal(err)
	}
	if got := discoverFixtures(tmpDir, nil, nil); len(got) != 0 {
		t.Errorf("expected no discovery with empty globs, got %v", got)
	}
}

func TestDiscoverFixturesWithStartingPaths(t *testing.T) {
	tmpDir := t.TempDir()

	// Layout:
	//   root.fixture.md                (top-level fixture)
	//   sub/sub.fixture.md             (fixture under sub/)
	//   other/other.fixture.md         (sibling fixture, excluded when path=sub)
	mustWrite := func(rel, body string) {
		full := filepath.Join(tmpDir, rel)
		if err := os.MkdirAll(filepath.Dir(full), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(body), 0644); err != nil {
			t.Fatal(err)
		}
	}
	mustWrite("root.fixture.md", "# root")
	mustWrite("sub/sub.fixture.md", "# sub")
	mustWrite("other/other.fixture.md", "# other")

	globs := []string{"**/*.fixture.md"}
	bases := func(paths []string) map[string]bool {
		m := make(map[string]bool)
		for _, p := range paths {
			m[filepath.Base(p)] = true
		}
		return m
	}

	t.Run("subdir path excludes siblings", func(t *testing.T) {
		got := bases(discoverFixtures(tmpDir, []string{"sub"}, globs))
		if !got["sub.fixture.md"] {
			t.Error("expected sub/sub.fixture.md to be discovered")
		}
		if got["other.fixture.md"] {
			t.Error("other/other.fixture.md should NOT be discovered for path=sub")
		}
		if got["root.fixture.md"] {
			t.Error("root.fixture.md should NOT be discovered for path=sub")
		}
	})

	t.Run("absolute starting path", func(t *testing.T) {
		got := bases(discoverFixtures(tmpDir, []string{filepath.Join(tmpDir, "sub")}, globs))
		if !got["sub.fixture.md"] || got["other.fixture.md"] {
			t.Errorf("absolute path filter wrong: %v", got)
		}
	})

	t.Run("multiple starting paths union", func(t *testing.T) {
		got := bases(discoverFixtures(tmpDir, []string{"sub", "other"}, globs))
		if !got["sub.fixture.md"] || !got["other.fixture.md"] {
			t.Errorf("expected union of sub and other fixtures, got %v", got)
		}
	})

	t.Run("empty starting paths falls back to workdir", func(t *testing.T) {
		got := bases(discoverFixtures(tmpDir, nil, globs))
		if !got["root.fixture.md"] {
			t.Error("expected root.fixture.md under no-arg discovery")
		}
	})
}

func TestResolveFixtureGlobs(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("disabled by default", func(t *testing.T) {
		if got := resolveFixtureGlobs(RunOptions{WorkDir: tmpDir}); got != nil {
			t.Errorf("expected nil (disabled), got %v", got)
		}
	})

	t.Run("flag enables default glob", func(t *testing.T) {
		got := resolveFixtureGlobs(RunOptions{WorkDir: tmpDir, Fixtures: true})
		if len(got) != 1 || got[0] != "**/*.fixture.md" {
			t.Errorf("expected [**/*.fixture.md], got %v", got)
		}
	})

	t.Run("flag-supplied globs win", func(t *testing.T) {
		custom := []string{"tests/**/*.md"}
		got := resolveFixtureGlobs(RunOptions{WorkDir: tmpDir, Fixtures: true, FixtureFiles: custom})
		if len(got) != 1 || got[0] != "tests/**/*.md" {
			t.Errorf("expected flag globs, got %v", got)
		}
	})

	t.Run("config enables and supplies globs", func(t *testing.T) {
		cfgDir := t.TempDir()
		cfgYAML := "fixtures:\n  enabled: true\n  files:\n    - specs/*.fixture.md\n"
		if err := os.WriteFile(filepath.Join(cfgDir, ".gavel.yaml"), []byte(cfgYAML), 0644); err != nil {
			t.Fatal(err)
		}
		got := resolveFixtureGlobs(RunOptions{WorkDir: cfgDir})
		if len(got) != 1 || got[0] != "specs/*.fixture.md" {
			t.Errorf("expected [specs/*.fixture.md] from config, got %v", got)
		}
	})
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
				Name:      "section",
				Framework: "fixture",
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
