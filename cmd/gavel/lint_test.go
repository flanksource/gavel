package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/flanksource/clicky"
	"github.com/flanksource/gavel/linters/eslint"
	"github.com/flanksource/gavel/verify"
)

func TestShouldRunLinterSkipsGolangciWithoutGoMod(t *testing.T) {
	workDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(workDir, ".git"), 0o755); err != nil {
		t.Fatalf("create .git: %v", err)
	}

	ok, reason := shouldRunLinter(workDir, verify.GavelConfig{}, "golangci-lint", false, false, false)
	if ok {
		t.Fatal("expected golangci-lint to be skipped without a go.mod")
	}
	if reason != "no go.mod found" {
		t.Fatalf("unexpected skip reason: %q", reason)
	}
}

func TestExecuteLintersJSCPDOptIn(t *testing.T) {
	t.Run("jscpd disabled by default", func(t *testing.T) {
		workDir := t.TempDir()
		if err := os.MkdirAll(filepath.Join(workDir, ".git"), 0o755); err != nil {
			t.Fatalf("create .git: %v", err)
		}
		if err := os.WriteFile(filepath.Join(workDir, "example.go"), []byte("package example\n"), 0o644); err != nil {
			t.Fatalf("write example.go: %v", err)
		}

		results, err := executeLinters(LintOptions{
			WorkDir: workDir,
			Timeout: "1s",
		})
		if err != nil {
			t.Fatalf("executeLinters: %v", err)
		}
		for _, result := range results {
			if result != nil && result.Linter == "jscpd" {
				t.Fatalf("expected jscpd to be absent when not enabled, got %+v", result)
			}
		}
	})

	t.Run("jscpd enabled via config participates in selection", func(t *testing.T) {
		workDir := t.TempDir()
		if err := os.MkdirAll(filepath.Join(workDir, ".git"), 0o755); err != nil {
			t.Fatalf("create .git: %v", err)
		}
		if err := os.WriteFile(filepath.Join(workDir, ".gavel.yaml"), []byte("lint:\n  linters:\n    jscpd:\n      enabled: true\n"), 0o644); err != nil {
			t.Fatalf("write .gavel.yaml: %v", err)
		}
		if err := os.WriteFile(filepath.Join(workDir, "example.go"), []byte("package example\n"), 0o644); err != nil {
			t.Fatalf("write example.go: %v", err)
		}

		results, err := executeLinters(LintOptions{
			WorkDir: workDir,
			Timeout: "1s",
		})
		if err != nil {
			t.Fatalf("executeLinters: %v", err)
		}
		found := false
		for _, result := range results {
			if result != nil && result.Linter == "jscpd" {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("expected jscpd result when enabled via config")
		}
	})
}

func TestExecuteLintersSelectsESLintForESLintConfigFiles(t *testing.T) {
	clicky.ClearGlobalTasks()
	t.Cleanup(clicky.ClearGlobalTasks)

	workDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(workDir, ".git"), 0o755); err != nil {
		t.Fatalf("create .git: %v", err)
	}
	if err := os.WriteFile(filepath.Join(workDir, "eslint.config.mjs"), []byte("export default [];\n"), 0o644); err != nil {
		t.Fatalf("write eslint config: %v", err)
	}

	results, err := executeLinters(LintOptions{
		WorkDir: workDir,
		Timeout: "1s",
	})
	if err != nil {
		t.Fatalf("executeLinters: %v", err)
	}

	found := false
	for _, result := range results {
		if result != nil && result.Linter == "eslint" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected eslint to be selected when eslint.config.mjs is present")
	}
}

func TestGroupFilesByGitRootUsesWorkDirForImplicitRun(t *testing.T) {
	repo := t.TempDir()
	subdir := filepath.Join(repo, "subdir")
	if err := os.MkdirAll(filepath.Join(repo, ".git"), 0o755); err != nil {
		t.Fatalf("create .git: %v", err)
	}
	if err := os.MkdirAll(subdir, 0o755); err != nil {
		t.Fatalf("create subdir: %v", err)
	}

	groups := groupFilesByGitRoot(LintOptions{WorkDir: subdir})
	if len(groups) != 1 {
		t.Fatalf("expected 1 group, got %d", len(groups))
	}
	if groups[0].gitRoot != subdir {
		t.Fatalf("expected implicit group root %q, got %q", subdir, groups[0].gitRoot)
	}
}

func TestShouldSelectLinterIgnoresRepoRootConfigFromSubdir(t *testing.T) {
	repo := t.TempDir()
	subdir := filepath.Join(repo, "subdir")
	if err := os.MkdirAll(filepath.Join(repo, ".git"), 0o755); err != nil {
		t.Fatalf("create .git: %v", err)
	}
	if err := os.MkdirAll(subdir, 0o755); err != nil {
		t.Fatalf("create subdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(repo, "eslint.config.mjs"), []byte("export default [];\n"), 0o644); err != nil {
		t.Fatalf("write repo eslint config: %v", err)
	}
	if err := os.WriteFile(filepath.Join(subdir, "foo.ts"), []byte("const x = 1;\n"), 0o644); err != nil {
		t.Fatalf("write ts file: %v", err)
	}

	ok, _ := shouldSelectLinter(subdir, verify.GavelConfig{}, eslint.NewESLint(subdir), false)
	if ok {
		t.Fatal("expected eslint to be skipped when config exists only at repo root")
	}
}

func TestShouldSelectLinterUsesDirectCWDConfig(t *testing.T) {
	workDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(workDir, "eslint.config.mjs"), []byte("export default [];\n"), 0o644); err != nil {
		t.Fatalf("write eslint config: %v", err)
	}

	ok, reason := shouldSelectLinter(workDir, verify.GavelConfig{}, eslint.NewESLint(workDir), false)
	if !ok {
		t.Fatalf("expected eslint to be selected, got skip reason %q", reason)
	}
}

func TestShouldSelectLinterIgnoresNestedConfig(t *testing.T) {
	workDir := t.TempDir()
	nested := filepath.Join(workDir, "nested")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatalf("create nested dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(nested, "eslint.config.mjs"), []byte("export default [];\n"), 0o644); err != nil {
		t.Fatalf("write nested eslint config: %v", err)
	}

	ok, _ := shouldSelectLinter(workDir, verify.GavelConfig{}, eslint.NewESLint(workDir), false)
	if ok {
		t.Fatal("expected nested eslint config to be ignored")
	}
}

func TestShouldSelectLinterAllowsExplicitGavelEnablementWithoutToolConfig(t *testing.T) {
	repo := t.TempDir()
	subdir := filepath.Join(repo, "subdir")
	if err := os.MkdirAll(filepath.Join(repo, ".git"), 0o755); err != nil {
		t.Fatalf("create .git: %v", err)
	}
	if err := os.MkdirAll(subdir, 0o755); err != nil {
		t.Fatalf("create subdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(repo, ".gavel.yaml"), []byte("lint:\n  linters:\n    eslint:\n      enabled: true\n"), 0o644); err != nil {
		t.Fatalf("write .gavel.yaml: %v", err)
	}
	if err := os.WriteFile(filepath.Join(subdir, "foo.ts"), []byte("const x = 1;\n"), 0o644); err != nil {
		t.Fatalf("write ts file: %v", err)
	}

	cfg, err := verify.LoadGavelConfig(subdir)
	if err != nil {
		t.Fatalf("load gavel config: %v", err)
	}
	ok, reason := shouldSelectLinter(subdir, cfg, eslint.NewESLint(subdir), false)
	if !ok {
		t.Fatalf("expected eslint to be selected via inherited .gavel enablement, got %q", reason)
	}
}

func TestShouldSelectLinterExplicitFilesDoNotBypassConfigRequirement(t *testing.T) {
	workDir := t.TempDir()
	nested := filepath.Join(workDir, "nested")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatalf("create nested dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(nested, "foo.ts"), []byte("const x = 1;\n"), 0o644); err != nil {
		t.Fatalf("write nested ts file: %v", err)
	}

	ok, _ := shouldSelectLinter(workDir, verify.GavelConfig{}, eslint.NewESLint(workDir), false)
	if ok {
		t.Fatal("expected nested file to be insufficient without direct cwd config or enablement")
	}
}
