package lint

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	deps "github.com/flanksource/deps"
	"github.com/flanksource/gavel/linters"
	"github.com/flanksource/gavel/linters/golangci"
	"github.com/flanksource/gavel/linters/reactdoctor"
	"github.com/flanksource/gavel/linters/tsc"
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

	groups := GroupFilesByGitRoot(Options{
		WorkDir: subdir,
		Files:   []string{"foo.go"},
	})
	if len(groups) != 1 {
		t.Fatalf("expected 1 group, got %d", len(groups))
	}
	if groups[0].GitRoot != repo {
		t.Fatalf("gitRoot = %q, want %q", groups[0].GitRoot, repo)
	}
	want := filepath.Join("sub", "foo.go")
	if len(groups[0].Files) != 1 || groups[0].Files[0] != want {
		t.Fatalf("files = %v, want [%s]", groups[0].Files, want)
	}
}

func TestNormalizeRootArgLeavesFilesAloneInsideGitRepo(t *testing.T) {
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

	opts, err := NormalizeRootArg(Options{
		Files: []string{subdir},
	})
	if err != nil {
		t.Fatalf("NormalizeRootArg: %v", err)
	}
	if opts.WorkDir != "" {
		t.Fatalf("WorkDir = %q, want empty (downstream resolves per linter)", opts.WorkDir)
	}
	if len(opts.Files) != 1 || opts.Files[0] != subdir {
		t.Fatalf("Files = %v, want [%s]", opts.Files, subdir)
	}
}

func TestNormalizeRootArgPromotesSingleDirectoryWhenNotInGitRepo(t *testing.T) {
	root := t.TempDir()
	subdir := filepath.Join(root, "sub")
	if err := os.MkdirAll(subdir, 0o755); err != nil {
		t.Fatalf("create subdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte("module example.com/repo\n"), 0o644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}

	opts, err := NormalizeRootArg(Options{
		Files: []string{subdir},
	})
	if err != nil {
		t.Fatalf("NormalizeRootArg: %v", err)
	}
	if opts.WorkDir != root {
		t.Fatalf("WorkDir = %q, want %q", opts.WorkDir, root)
	}
	if len(opts.Files) != 0 {
		t.Fatalf("Files = %v, want []", opts.Files)
	}
}

func TestResolveLinterInvocationsTscFindsParentTsconfig(t *testing.T) {
	repo := t.TempDir()
	if err := os.MkdirAll(filepath.Join(repo, ".git"), 0o755); err != nil {
		t.Fatalf("create .git: %v", err)
	}
	if err := os.WriteFile(filepath.Join(repo, "tsconfig.json"), []byte("{}\n"), 0o644); err != nil {
		t.Fatalf("write tsconfig.json: %v", err)
	}
	scope := filepath.Join(repo, "frontend", "src")
	if err := os.MkdirAll(scope, 0o755); err != nil {
		t.Fatalf("create scope: %v", err)
	}

	invs := resolveLinterInvocations(tsc.NewTSC(repo), Options{
		WorkDir: repo,
		Files:   []string{filepath.Join("frontend", "src")},
	})

	if len(invs) != 1 {
		t.Fatalf("invocations = %d, want 1: %+v", len(invs), invs)
	}
	if invs[0].projectRoot != repo {
		t.Fatalf("projectRoot = %q, want %q", invs[0].projectRoot, repo)
	}
	want := filepath.Join("frontend", "src")
	if len(invs[0].files) != 1 || invs[0].files[0] != want {
		t.Fatalf("files = %v, want [%s]", invs[0].files, want)
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

	groups := GroupFilesByGitRoot(Options{
		WorkDir: subdir,
		Files:   []string{"root.go"},
	})
	if len(groups) != 1 {
		t.Fatalf("expected 1 group, got %d", len(groups))
	}
	if groups[0].GitRoot != repo {
		t.Fatalf("gitRoot = %q, want %q", groups[0].GitRoot, repo)
	}
	if len(groups[0].Files) != 1 || groups[0].Files[0] != "root.go" {
		t.Fatalf("files = %v, want [root.go]", groups[0].Files)
	}
}

func TestResolveLinterExecutableDryRunUsesGitRootInstallPath(t *testing.T) {
	repo := t.TempDir()
	t.Setenv("PATH", t.TempDir())

	got, reason, err := resolveLinterExecutable(context.Background(), golangci.NewGolangciLint(repo), repo, repo, true, true)
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

	got, reason, err := resolveLinterExecutable(context.Background(), golangci.NewGolangciLint(repo), repo, repo, false, false)
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

func TestResolveLinterExecutableReinstallsInvalidGolangciBinary(t *testing.T) {
	repo := t.TempDir()
	t.Setenv("PATH", t.TempDir())

	installed := filepath.Join(repo, ".gavel", executableFileName("golangci-lint"))
	if err := os.MkdirAll(filepath.Dir(installed), 0o755); err != nil {
		t.Fatalf("mkdir .gavel: %v", err)
	}
	if err := os.WriteFile(installed, []byte("#!/bin/sh\nexit 126\n"), 0o755); err != nil {
		t.Fatalf("write stale golangci: %v", err)
	}

	oldInstall := installGolangciLint
	installGolangciLint = func(ctx context.Context, packageName, version string, opts ...deps.InstallOption) (*deps.InstallResult, error) {
		if err := os.WriteFile(installed, []byte("#!/bin/sh\nexit 0\n"), 0o755); err != nil {
			return nil, err
		}
		return &deps.InstallResult{BinDir: filepath.Dir(installed)}, nil
	}
	t.Cleanup(func() { installGolangciLint = oldInstall })

	got, reason, err := resolveLinterExecutable(context.Background(), golangci.NewGolangciLint(repo), repo, repo, true, false)
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

func TestResolveLinterExecutableReactDoctorPrefersPNPXForPNPMLock(t *testing.T) {
	repo := t.TempDir()
	if err := os.WriteFile(filepath.Join(repo, "pnpm-lock.yaml"), nil, 0o644); err != nil {
		t.Fatalf("write pnpm lock: %v", err)
	}
	binDir := t.TempDir()
	pnpx := writeExecutable(t, binDir, "pnpx")
	writeExecutable(t, binDir, "npx")
	writeExecutable(t, binDir, "react-doctor")
	t.Setenv("PATH", binDir)

	got, reason, err := resolveLinterExecutable(context.Background(), reactdoctor.NewReactDoctor(repo), repo, repo, false, false)
	if err != nil {
		t.Fatalf("resolveLinterExecutable: %v", err)
	}
	if reason != "" {
		t.Fatalf("unexpected skip reason: %q", reason)
	}
	if got != pnpx {
		t.Fatalf("command = %q, want %q", got, pnpx)
	}
}

func TestResolveLinterExecutableReactDoctorPrefersNPXForNPMLock(t *testing.T) {
	repo := t.TempDir()
	if err := os.WriteFile(filepath.Join(repo, "package-lock.json"), nil, 0o644); err != nil {
		t.Fatalf("write package lock: %v", err)
	}
	binDir := t.TempDir()
	npx := writeExecutable(t, binDir, "npx")
	writeExecutable(t, binDir, "react-doctor")
	t.Setenv("PATH", binDir)

	got, reason, err := resolveLinterExecutable(context.Background(), reactdoctor.NewReactDoctor(repo), repo, repo, false, false)
	if err != nil {
		t.Fatalf("resolveLinterExecutable: %v", err)
	}
	if reason != "" {
		t.Fatalf("unexpected skip reason: %q", reason)
	}
	if got != npx {
		t.Fatalf("command = %q, want %q", got, npx)
	}
}

func TestResolveLinterExecutableReactDoctorFallsBackToBinary(t *testing.T) {
	repo := t.TempDir()
	if err := os.WriteFile(filepath.Join(repo, "pnpm-lock.yaml"), nil, 0o644); err != nil {
		t.Fatalf("write pnpm lock: %v", err)
	}
	binDir := t.TempDir()
	binary := writeExecutable(t, binDir, "react-doctor")
	t.Setenv("PATH", binDir)

	got, reason, err := resolveLinterExecutable(context.Background(), reactdoctor.NewReactDoctor(repo), repo, repo, false, false)
	if err != nil {
		t.Fatalf("resolveLinterExecutable: %v", err)
	}
	if reason != "" {
		t.Fatalf("unexpected skip reason: %q", reason)
	}
	if got != binary {
		t.Fatalf("command = %q, want %q", got, binary)
	}
}

func TestResolveLinterExecutableReactDoctorWithoutLockUsesBinary(t *testing.T) {
	repo := t.TempDir()
	binDir := t.TempDir()
	writeExecutable(t, binDir, "npx")
	binary := writeExecutable(t, binDir, "react-doctor")
	t.Setenv("PATH", binDir)

	got, reason, err := resolveLinterExecutable(context.Background(), reactdoctor.NewReactDoctor(repo), repo, repo, false, false)
	if err != nil {
		t.Fatalf("resolveLinterExecutable: %v", err)
	}
	if reason != "" {
		t.Fatalf("unexpected skip reason: %q", reason)
	}
	if got != binary {
		t.Fatalf("command = %q, want %q", got, binary)
	}
}

func writeExecutable(t *testing.T, dir, name string) string {
	t.Helper()

	path := filepath.Join(dir, executableFileName(name))
	if err := os.WriteFile(path, []byte("#!/bin/sh\nexit 0\n"), 0o755); err != nil {
		t.Fatalf("write executable %s: %v", name, err)
	}
	return path
}

func TestApplyPostLintFiltersUsesGroupConfig(t *testing.T) {
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
	applyPostLintFilters(resultsA, rootA, nil)
	if len(resultsA[0].Violations) != 0 {
		t.Fatalf("expected rootA ignore to filter pkg/foo.go, got %v", resultsA[0].Violations)
	}

	resultsB := makeResults("pkg/foo.go")
	applyPostLintFilters(resultsB, rootB, nil)
	if len(resultsB[0].Violations) != 1 {
		t.Fatalf("expected rootB config to leave pkg/foo.go alone, got %v", resultsB[0].Violations)
	}
}
