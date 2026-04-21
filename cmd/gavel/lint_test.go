package main

import (
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/flanksource/clicky"
	"github.com/flanksource/gavel/linters/eslint"
	"github.com/flanksource/gavel/linters/golangci"
	"github.com/flanksource/gavel/linters/markdownlint"
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

func TestResolveLinterInvocationsBucketsByProjectRoot(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".git"), 0o755); err != nil {
		t.Fatalf("create .git: %v", err)
	}

	backend := filepath.Join(root, "backend")
	backendPkg := filepath.Join(backend, "pkg")
	frontend := filepath.Join(root, "frontend")
	frontendSrc := filepath.Join(frontend, "src")
	if err := os.MkdirAll(backendPkg, 0o755); err != nil {
		t.Fatalf("mkdir backend: %v", err)
	}
	if err := os.MkdirAll(frontendSrc, 0o755); err != nil {
		t.Fatalf("mkdir frontend: %v", err)
	}
	if err := os.WriteFile(filepath.Join(backend, "go.mod"), []byte("module example.com/backend\n"), 0o644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}
	if err := os.WriteFile(filepath.Join(frontend, "package.json"), []byte("{}"), 0o644); err != nil {
		t.Fatalf("write package.json: %v", err)
	}
	backendFile := filepath.Join(backendPkg, "foo.go")
	frontendFile := filepath.Join(frontendSrc, "index.ts")
	if err := os.WriteFile(backendFile, []byte("package pkg\n"), 0o644); err != nil {
		t.Fatalf("write foo.go: %v", err)
	}
	if err := os.WriteFile(frontendFile, []byte("export const x = 1;\n"), 0o644); err != nil {
		t.Fatalf("write index.ts: %v", err)
	}

	opts := LintOptions{
		WorkDir: root,
		Files:   []string{backendFile, frontendFile},
	}

	t.Run("go files route to go.mod root", func(t *testing.T) {
		invs := resolveLinterInvocations(golangci.NewGolangciLint(root), opts)
		if len(invs) != 1 {
			t.Fatalf("expected 1 golangci invocation, got %d", len(invs))
		}
		if invs[0].projectRoot != backend {
			t.Fatalf("expected projectRoot=%q, got %q", backend, invs[0].projectRoot)
		}
		if len(invs[0].files) != 1 || invs[0].files[0] != filepath.Join("pkg", "foo.go") {
			t.Fatalf("expected files=[pkg/foo.go] relative to backend, got %v", invs[0].files)
		}
	})

	t.Run("ts files route to package.json root", func(t *testing.T) {
		invs := resolveLinterInvocations(eslint.NewESLint(root), opts)
		if len(invs) != 1 {
			t.Fatalf("expected 1 eslint invocation, got %d", len(invs))
		}
		if invs[0].projectRoot != frontend {
			t.Fatalf("expected projectRoot=%q, got %q", frontend, invs[0].projectRoot)
		}
		if len(invs[0].files) != 1 || invs[0].files[0] != filepath.Join("src", "index.ts") {
			t.Fatalf("expected files=[src/index.ts] relative to frontend, got %v", invs[0].files)
		}
	})

	t.Run("non-rooted linter keeps workdir and files", func(t *testing.T) {
		invs := resolveLinterInvocations(markdownlint.NewMarkdownlint(root), opts)
		if len(invs) != 1 {
			t.Fatalf("expected 1 markdownlint invocation, got %d", len(invs))
		}
		if invs[0].projectRoot != root {
			t.Fatalf("expected projectRoot=%q, got %q", root, invs[0].projectRoot)
		}
		if len(invs[0].files) != 2 {
			t.Fatalf("expected 2 files passed through unchanged, got %d", len(invs[0].files))
		}
	})

	t.Run("no files: fans out across every discovered project root", func(t *testing.T) {
		// Add a second go.mod under tools/ so golangci has two roots to find.
		tools := filepath.Join(root, "tools")
		if err := os.MkdirAll(tools, 0o755); err != nil {
			t.Fatalf("mkdir tools: %v", err)
		}
		if err := os.WriteFile(filepath.Join(tools, "go.mod"), []byte("module example.com/tools\n"), 0o644); err != nil {
			t.Fatalf("write tools/go.mod: %v", err)
		}

		invs := resolveLinterInvocations(golangci.NewGolangciLint(root), LintOptions{WorkDir: root})
		if len(invs) != 2 {
			t.Fatalf("expected 2 golangci invocations (backend, tools), got %d", len(invs))
		}
		got := []string{invs[0].projectRoot, invs[1].projectRoot}
		sort.Strings(got)
		want := []string{backend, tools}
		sort.Strings(want)
		if got[0] != want[0] || got[1] != want[1] {
			t.Fatalf("expected roots=%v, got %v", want, got)
		}
		for _, inv := range invs {
			if len(inv.files) != 0 {
				t.Fatalf("expected empty files for whole-root invocation, got %v", inv.files)
			}
		}
	})

	t.Run("files with no project root are dropped", func(t *testing.T) {
		orphan := filepath.Join(root, "orphan.go")
		if err := os.WriteFile(orphan, []byte("package orphan\n"), 0o644); err != nil {
			t.Fatalf("write orphan: %v", err)
		}
		invs := resolveLinterInvocations(golangci.NewGolangciLint(root), LintOptions{
			WorkDir: root,
			Files:   []string{orphan},
		})
		if len(invs) != 0 {
			t.Fatalf("expected 0 invocations for file without go.mod, got %+v", invs)
		}
	})
}
