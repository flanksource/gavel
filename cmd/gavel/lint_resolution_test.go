package main

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/flanksource/gavel/linters"
	"github.com/flanksource/gavel/linters/golangci"
	"github.com/flanksource/gavel/models"
)

func TestGroupFilesByGitRootResolvesRelativeFilesFromWorkDir(t *testing.T) {
	repo := t.TempDir()
	subdir := filepath.Join(repo, "sub")
	if err := os.MkdirAll(filepath.Join(repo, ".git"), 0o755); err != nil {
		t.Fatalf("create .git: %v", err)
	}
	if err := os.MkdirAll(subdir, 0o755); err != nil {
		t.Fatalf("create subdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(subdir, "foo.go"), []byte("package sub\n"), 0o644); err != nil {
		t.Fatalf("write foo.go: %v", err)
	}

	other := t.TempDir()
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(other); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(cwd)
	})

	groups := groupFilesByGitRoot(LintOptions{
		WorkDir: subdir,
		Files:   []string{"foo.go"},
	})
	if len(groups) != 1 {
		t.Fatalf("expected 1 group, got %d", len(groups))
	}
	if groups[0].gitRoot != repo {
		t.Fatalf("gitRoot = %q, want %q", groups[0].gitRoot, repo)
	}
	want := filepath.Join("sub", "foo.go")
	if len(groups[0].files) != 1 || groups[0].files[0] != want {
		t.Fatalf("files = %v, want [%s]", groups[0].files, want)
	}
}

func TestNormalizeLintRootArgPromotesSingleDirectoryToWorkDir(t *testing.T) {
	repo := t.TempDir()
	subdir := filepath.Join(repo, "sub")
	if err := os.MkdirAll(filepath.Join(repo, ".git"), 0o755); err != nil {
		t.Fatalf("create .git: %v", err)
	}
	if err := os.MkdirAll(subdir, 0o755); err != nil {
		t.Fatalf("create subdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(repo, "go.mod"), []byte("module example.com/repo\n"), 0o644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}

	opts, err := normalizeLintRootArg(LintOptions{
		Files: []string{subdir},
	})
	if err != nil {
		t.Fatalf("normalizeLintRootArg: %v", err)
	}
	if opts.WorkDir != repo {
		t.Fatalf("WorkDir = %q, want %q", opts.WorkDir, repo)
	}
	if len(opts.Files) != 0 {
		t.Fatalf("Files = %v, want []", opts.Files)
	}
}

func TestGroupFilesByGitRootFallsBackToGitRootForRelativePaths(t *testing.T) {
	repo := t.TempDir()
	subdir := filepath.Join(repo, "sub")
	if err := os.MkdirAll(filepath.Join(repo, ".git"), 0o755); err != nil {
		t.Fatalf("create .git: %v", err)
	}
	if err := os.MkdirAll(subdir, 0o755); err != nil {
		t.Fatalf("create subdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(repo, "root.go"), []byte("package root\n"), 0o644); err != nil {
		t.Fatalf("write root.go: %v", err)
	}

	groups := groupFilesByGitRoot(LintOptions{
		WorkDir: subdir,
		Files:   []string{"root.go"},
	})
	if len(groups) != 1 {
		t.Fatalf("expected 1 group, got %d", len(groups))
	}
	if groups[0].gitRoot != repo {
		t.Fatalf("gitRoot = %q, want %q", groups[0].gitRoot, repo)
	}
	if len(groups[0].files) != 1 || groups[0].files[0] != "root.go" {
		t.Fatalf("files = %v, want [root.go]", groups[0].files)
	}
}

func TestResolveLinterExecutableDryRunUsesGitRootInstallPath(t *testing.T) {
	repo := t.TempDir()
	t.Setenv("PATH", t.TempDir())

	got, reason, err := resolveLinterExecutable(context.Background(), golangci.NewGolangciLint(repo), repo, true, true)
	if err != nil {
		t.Fatalf("resolveLinterExecutable: %v", err)
	}
	if reason != "" {
		t.Fatalf("unexpected skip reason: %q", reason)
	}
	want := filepath.Join(repo, ".gavel", executableFileName("golangci-lint"))
	if got != want {
		t.Fatalf("command = %q, want %q", got, want)
	}
}

func TestResolveLinterExecutableUsesInstalledGolangciBinary(t *testing.T) {
	repo := t.TempDir()
	t.Setenv("PATH", t.TempDir())

	installed := filepath.Join(repo, ".gavel", executableFileName("golangci-lint"))
	if err := os.MkdirAll(filepath.Dir(installed), 0o755); err != nil {
		t.Fatalf("mkdir .gavel: %v", err)
	}
	if err := os.WriteFile(installed, []byte("#!/bin/sh\nexit 0\n"), 0o755); err != nil {
		t.Fatalf("write installed golangci: %v", err)
	}

	got, reason, err := resolveLinterExecutable(context.Background(), golangci.NewGolangciLint(repo), repo, false, false)
	if err != nil {
		t.Fatalf("resolveLinterExecutable: %v", err)
	}
	if reason != "" {
		t.Fatalf("unexpected skip reason: %q", reason)
	}
	if got != installed {
		t.Fatalf("command = %q, want %q", got, installed)
	}
}

func TestApplyGroupIgnoresUsesGroupConfig(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	rootA := t.TempDir()
	rootB := t.TempDir()
	for _, root := range []string{rootA, rootB} {
		if err := os.MkdirAll(filepath.Join(root, ".git"), 0o755); err != nil {
			t.Fatalf("create .git in %s: %v", root, err)
		}
	}
	if err := os.WriteFile(filepath.Join(rootA, ".gavel.yaml"), []byte("lint:\n  ignore:\n    - source: golangci-lint\n      rule: errcheck\n      file: pkg/foo.go\n"), 0o644); err != nil {
		t.Fatalf("write rootA .gavel.yaml: %v", err)
	}
	if err := os.WriteFile(filepath.Join(rootB, ".gavel.yaml"), []byte("lint:\n  ignore:\n    - source: golangci-lint\n      rule: errcheck\n      file: pkg/bar.go\n"), 0o644); err != nil {
		t.Fatalf("write rootB .gavel.yaml: %v", err)
	}

	errcheck := &models.Rule{Method: "errcheck"}
	makeResults := func(file string) []*linters.LinterResult {
		return []*linters.LinterResult{{
			Linter:  "golangci-lint",
			WorkDir: rootA,
			Violations: []models.Violation{{
				Source: "golangci-lint",
				File:   file,
				Rule:   errcheck,
			}},
		}}
	}

	resultsA := makeResults("pkg/foo.go")
	if err := applyGroupIgnores(rootA, resultsA); err != nil {
		t.Fatalf("applyGroupIgnores rootA: %v", err)
	}
	if len(resultsA[0].Violations) != 0 {
		t.Fatalf("expected rootA ignore to filter pkg/foo.go, got %v", resultsA[0].Violations)
	}

	resultsB := makeResults("pkg/foo.go")
	if err := applyGroupIgnores(rootB, resultsB); err != nil {
		t.Fatalf("applyGroupIgnores rootB: %v", err)
	}
	if len(resultsB[0].Violations) != 1 {
		t.Fatalf("expected rootB config to leave pkg/foo.go alone, got %v", resultsB[0].Violations)
	}
}
