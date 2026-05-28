package commit

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/flanksource/commons/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/flanksource/gavel/verify"
)

// withStubbedTidyRecording stubs runGoModTidy to capture the modDirs it was
// invoked with (in order) and applies an optional side effect per call. Returns
// a pointer to the captured slice so assertions can inspect call order.
func withStubbedTidyRecording(t *testing.T, sideEffect func(modDir string) error) *[]string {
	t.Helper()
	var (
		mu    sync.Mutex
		calls []string
	)
	prev := runGoModTidy
	runGoModTidy = func(modDir string) error {
		mu.Lock()
		calls = append(calls, modDir)
		mu.Unlock()
		if sideEffect != nil {
			return sideEffect(modDir)
		}
		return nil
	}
	t.Cleanup(func() { runGoModTidy = prev })
	return &calls
}

// withFindGoModRoots stubs the directory walker so tests don't depend on
// .gitignore traversal correctness. Pass the absolute dirs to return.
func withFindGoModRoots(t *testing.T, dirs []string) {
	t.Helper()
	prev := findGoModRoots
	findGoModRoots = func(string) []string { return dirs }
	t.Cleanup(func() { findGoModRoots = prev })
}

// captureLogger redirects the global logger to a buffer for the duration of
// the test. Returns the buffer so tests can match on emitted lines.
func captureLogger(t *testing.T) *strings.Builder {
	t.Helper()
	var buf strings.Builder
	prev := logger.StandardLogger()
	logger.Use(&buf)
	t.Cleanup(func() { logger.SetLogger(prev) })
	return &buf
}

func writeGoMod(t *testing.T, dir, module string) {
	t.Helper()
	require.NoError(t, os.MkdirAll(dir, 0o755))
	contents := fmt.Sprintf("module %s\n\ngo 1.22\n", module)
	require.NoError(t, os.WriteFile(filepath.Join(dir, "go.mod"), []byte(contents), 0o644))
}

func ptrBool(b bool) *bool { return &b }

func TestTidyEnabled(t *testing.T) {
	cases := []struct {
		name string
		flag string
		cfg  *bool
		want bool
	}{
		{"default (no flag, no config) is on", "", nil, true},
		{"config false disables", "", ptrBool(false), false},
		{"config true enables", "", ptrBool(true), true},
		{"flag false overrides config true", "false", ptrBool(true), false},
		{"flag true overrides config false", "true", ptrBool(false), true},
		{"flag whitespace tolerated", "  TRUE  ", ptrBool(false), true},
		{"unknown flag falls back to config", "maybe", ptrBool(false), false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			opts := Options{TidyFlag: c.flag, Config: verify.CommitConfig{Tidy: verify.CommitTidyConfig{Enabled: c.cfg}}}
			assert.Equal(t, c.want, tidyEnabled(opts))
		})
	}
}

func TestApplyGoModTidy_OnByDefault(t *testing.T) {
	repo := initCommitRepo(t)
	writeGoMod(t, repo, "example.com/app")
	withFindGoModRoots(t, []string{repo})
	calls := withStubbedTidyRecording(t, nil)

	src := stagedSource{Files: []string{"README.md"}}
	got, err := applyGoModTidy(context.Background(), Options{WorkDir: repo}, src)
	require.NoError(t, err)
	assert.Equal(t, []string{repo}, *calls)
	assert.Equal(t, src.Files, got.Files, "no changes -> source returned unchanged")
}

func TestApplyGoModTidy_DisabledByConfig(t *testing.T) {
	repo := initCommitRepo(t)
	writeGoMod(t, repo, "example.com/app")
	withFindGoModRoots(t, []string{repo})
	calls := withStubbedTidyRecording(t, nil)

	opts := Options{
		WorkDir: repo,
		Config:  verify.CommitConfig{Tidy: verify.CommitTidyConfig{Enabled: ptrBool(false)}},
	}
	_, err := applyGoModTidy(context.Background(), opts, stagedSource{})
	require.NoError(t, err)
	assert.Empty(t, *calls, "tidy disabled -> runGoModTidy never called")
}

func TestApplyGoModTidy_FlagOverridesConfig(t *testing.T) {
	repo := initCommitRepo(t)
	writeGoMod(t, repo, "example.com/app")
	withFindGoModRoots(t, []string{repo})
	calls := withStubbedTidyRecording(t, nil)

	opts := Options{
		WorkDir:  repo,
		TidyFlag: "false",
		Config:   verify.CommitConfig{Tidy: verify.CommitTidyConfig{Enabled: ptrBool(true)}},
	}
	_, err := applyGoModTidy(context.Background(), opts, stagedSource{})
	require.NoError(t, err)
	assert.Empty(t, *calls, "--tidy=false beats config.enabled=true")
}

func TestApplyGoModTidy_SingleModuleChanged(t *testing.T) {
	repo := initCommitRepo(t)
	writeGoMod(t, repo, "example.com/app")
	gitRun(t, repo, "add", "go.mod")
	gitRun(t, repo, "commit", "-m", "add go.mod")
	withFindGoModRoots(t, []string{repo})
	logBuf := captureLogger(t)

	calls := withStubbedTidyRecording(t, func(modDir string) error {
		// Simulate `go mod tidy` adding a require line.
		path := filepath.Join(modDir, "go.mod")
		return os.WriteFile(path, []byte("module example.com/app\n\ngo 1.22\n\nrequire example.com/dep v1.0.0\n"), 0o644)
	})

	// Stage something unrelated so readStagedSource has a populated set.
	writeFile(t, repo, "main.go", "package main\n")
	gitRun(t, repo, "add", "main.go")
	source, err := readStagedSource(repo)
	require.NoError(t, err)

	got, err := applyGoModTidy(context.Background(), Options{WorkDir: repo}, source)
	require.NoError(t, err)
	assert.Len(t, *calls, 1)
	assert.Contains(t, got.Files, "go.mod", "tidied go.mod should be staged into refreshed source")
	assert.Contains(t, logBuf.String(), "go mod tidy: updated ./go.mod")
}

func TestApplyGoModTidy_NoChange(t *testing.T) {
	repo := initCommitRepo(t)
	writeGoMod(t, repo, "example.com/app")
	gitRun(t, repo, "add", "go.mod")
	gitRun(t, repo, "commit", "-m", "add go.mod")
	withFindGoModRoots(t, []string{repo})
	logBuf := captureLogger(t)
	calls := withStubbedTidyRecording(t, nil) // no-op stub

	writeFile(t, repo, "main.go", "package main\n")
	gitRun(t, repo, "add", "main.go")
	source, err := readStagedSource(repo)
	require.NoError(t, err)
	originalFiles := append([]string(nil), source.Files...)

	got, err := applyGoModTidy(context.Background(), Options{WorkDir: repo}, source)
	require.NoError(t, err)
	assert.Len(t, *calls, 1)
	assert.Equal(t, originalFiles, got.Files, "unchanged files -> source returned as-is")
	assert.NotContains(t, logBuf.String(), "go mod tidy: updated", "no log line when nothing changed")
}

func TestApplyGoModTidy_NestedModules(t *testing.T) {
	repo := initCommitRepo(t)
	writeGoMod(t, repo, "example.com/app")
	nested := filepath.Join(repo, "services", "api")
	writeGoMod(t, nested, "example.com/app/services/api")
	gitRun(t, repo, "add", "go.mod", "services/api/go.mod")
	gitRun(t, repo, "commit", "-m", "add go mods")
	withFindGoModRoots(t, []string{repo, nested})
	logBuf := captureLogger(t)

	calls := withStubbedTidyRecording(t, func(modDir string) error {
		if modDir != nested {
			return nil
		}
		return os.WriteFile(filepath.Join(modDir, "go.mod"), []byte("module example.com/app/services/api\n\ngo 1.22\n\nrequire example.com/dep v1.0.0\n"), 0o644)
	})

	writeFile(t, repo, "main.go", "package main\n")
	gitRun(t, repo, "add", "main.go")
	source, err := readStagedSource(repo)
	require.NoError(t, err)

	got, err := applyGoModTidy(context.Background(), Options{WorkDir: repo}, source)
	require.NoError(t, err)
	assert.Equal(t, []string{repo, nested}, *calls)
	log := logBuf.String()
	assert.Contains(t, log, "go mod tidy: updated services/api/go.mod")
	assert.NotContains(t, log, "go mod tidy: updated ./go.mod", "root go.mod unchanged -> no log line")
	assert.Contains(t, got.Files, "services/api/go.mod")
	assert.NotContains(t, got.Files, "go.mod")
}

func TestApplyGoModTidy_FailsLoudly(t *testing.T) {
	repo := initCommitRepo(t)
	writeGoMod(t, repo, "example.com/app")
	withFindGoModRoots(t, []string{repo})
	_ = withStubbedTidyRecording(t, func(modDir string) error {
		return fmt.Errorf("network unreachable")
	})

	_, err := applyGoModTidy(context.Background(), Options{WorkDir: repo}, stagedSource{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "go mod tidy in .")
	assert.Contains(t, err.Error(), "network unreachable")
}

func TestApplyGoModTidy_NoGoModFiles(t *testing.T) {
	repo := initCommitRepo(t)
	withFindGoModRoots(t, []string{}) // empty
	calls := withStubbedTidyRecording(t, nil)
	logBuf := captureLogger(t)

	src := stagedSource{Files: []string{"README.md"}}
	got, err := applyGoModTidy(context.Background(), Options{WorkDir: repo}, src)
	require.NoError(t, err)
	assert.Empty(t, *calls, "no go.mod -> stub never called")
	assert.Equal(t, src.Files, got.Files)
	assert.Empty(t, logBuf.String())
}

func TestApplyGoModTidy_GoSumCreated(t *testing.T) {
	repo := initCommitRepo(t)
	writeGoMod(t, repo, "example.com/app")
	gitRun(t, repo, "add", "go.mod")
	gitRun(t, repo, "commit", "-m", "add go.mod")
	withFindGoModRoots(t, []string{repo})
	logBuf := captureLogger(t)

	withStubbedTidyRecording(t, func(modDir string) error {
		// Tidy writes a new go.sum and leaves go.mod alone.
		return os.WriteFile(filepath.Join(modDir, "go.sum"), []byte("example.com/dep v1.0.0 h1:abc=\n"), 0o644)
	})

	writeFile(t, repo, "main.go", "package main\n")
	gitRun(t, repo, "add", "main.go")
	source, err := readStagedSource(repo)
	require.NoError(t, err)

	got, err := applyGoModTidy(context.Background(), Options{WorkDir: repo}, source)
	require.NoError(t, err)
	log := logBuf.String()
	assert.Contains(t, log, "go mod tidy: updated ./go.sum")
	assert.NotContains(t, log, "go mod tidy: updated ./go.mod, go.sum", "go.mod unchanged -> only go.sum in log")
	assert.Contains(t, got.Files, "go.sum")
}

func TestFileChanged(t *testing.T) {
	cases := []struct {
		name                                   string
		before                                 []byte
		bExists                                bool
		after                                  []byte
		aExists                                bool
		want                                   bool
	}{
		{"both absent", nil, false, nil, false, false},
		{"absent -> present", nil, false, []byte("x"), true, true},
		{"present -> absent", []byte("x"), true, nil, false, true},
		{"same bytes", []byte{1, 2, 3}, true, []byte{1, 2, 3}, true, false},
		{"different bytes", []byte{1, 2, 3}, true, []byte{1, 2, 4}, true, true},
		{"different lengths", []byte{1, 2}, true, []byte{1, 2, 3}, true, true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			assert.Equal(t, c.want, fileChanged(c.before, c.bExists, c.after, c.aExists))
		})
	}
}
