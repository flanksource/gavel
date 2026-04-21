package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/linters"
	"github.com/flanksource/gavel/testrunner"
	"github.com/flanksource/gavel/testrunner/parsers"
	testui "github.com/flanksource/gavel/testrunner/ui"
	"github.com/flanksource/gavel/utils"
)

func buildTestSnapshot(
	opts testrunner.RunOptions,
	tests []parsers.Test,
	lint []*linters.LinterResult,
	started time.Time,
	ended time.Time,
	diagnostics *testui.DiagnosticsSnapshot,
) testui.Snapshot {
	if started.IsZero() {
		started = time.Now().UTC()
	}
	if ended.IsZero() {
		ended = time.Now().UTC()
	}

	return testui.Snapshot{
		Metadata: &testui.SnapshotMetadata{
			Version:  version,
			Started:  started.UTC(),
			Ended:    ended.UTC(),
			Kind:     "initial",
			Sequence: 1,
			Args:     snapshotArgs(opts),
		},
		Git: snapshotGitInfo(opts.WorkDir),
		Status: testui.SnapshotStatus{
			Running:              false,
			LintRun:              opts.Lint,
			DiagnosticsAvailable: diagnostics != nil,
		},
		Tests:       tests,
		Lint:        lint,
		Diagnostics: diagnostics,
	}
}

func captureFinalDiagnostics(enabled bool, rootPID int) *testui.DiagnosticsSnapshot {
	if !enabled {
		return nil
	}
	snapshot, err := testui.NewDiagnosticsManager(rootPID, nil).Snapshot()
	if err != nil {
		logger.Warnf("Diagnostics capture failed: %v", err)
		return nil
	}
	return snapshot
}

// installTimeoutDiagnosticsHook wires testrunner.captureGlobalDiagnostics to
// this package's diagnostics capture so the per-package timeout supervisor can
// snapshot process/goroutine state before killing a subprocess. The hook is a
// no-op when --diagnostics is disabled; always reinstalled per run so
// concurrent `gavel test` invocations stay idempotent.
func installTimeoutDiagnosticsHook(opts testrunner.RunOptions) {
	rootPID := os.Getpid()
	enabled := opts.Diagnostics
	var once sync.Once
	testrunner.SetCaptureGlobalDiagnostics(func() {
		once.Do(func() {
			_ = captureFinalDiagnostics(enabled, rootPID)
		})
	})
}

func snapshotArgs(opts testrunner.RunOptions) map[string]any {
	return map[string]any{
		"sync_todos":     opts.SyncTodos,
		"starting_paths": append([]string(nil), opts.StartingPaths...),
		"extra_args":     append([]string(nil), opts.ExtraArgs...),
		"show_passed":    opts.ShowPassed,
		"show_stdout":    string(opts.ShowStdout),
		"show_stderr":    string(opts.ShowStderr),
		"todos_dir":      opts.TodosDir,
		"todo_template":  opts.TodoTemplate,
		"work_dir":       opts.WorkDir,
		"dry_run":        opts.DryRun,
		"recursive":      opts.Recursive,
		"nodes":          opts.Nodes,
		"ui":             opts.UI,
		"addr":           opts.Addr,
		"diagnostics":    opts.Diagnostics,
		"skip_hooks":     opts.SkipHooks,
		"auto_stop":      durationString(opts.AutoStop),
		"idle_timeout":   durationString(opts.IdleTimeout),
		"timeout":        durationString(opts.Timeout),
		"lint_timeout":   durationString(opts.LintTimeout),
		"test_timeout":   durationString(opts.TestTimeout),
		"lint":           opts.Lint,
		"cache":          opts.Cache,
		"changed":        opts.Changed,
		"since":          opts.Since,
		"bench":          opts.Bench,
		"fixtures":       opts.Fixtures,
		"fixture_files":  append([]string(nil), opts.FixtureFiles...),
	}
}

func durationString(d time.Duration) string {
	if d <= 0 {
		return ""
	}
	return d.String()
}

func snapshotGitInfo(workDir string) *testui.SnapshotGit {
	if workDir == "" {
		return nil
	}

	root := utils.FindGitRoot(workDir)
	if root == "" {
		abs, err := filepath.Abs(workDir)
		if err == nil {
			root = abs
		} else {
			root = workDir
		}
	}

	git := &testui.SnapshotGit{
		Repo: filepath.Base(root),
		Root: root,
	}
	if utils.FindGitRoot(workDir) != "" {
		git.SHA = gitHeadSHA(root)
	}
	return git
}

func gitHeadSHA(workDir string) string {
	cmd := exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = workDir
	out, err := cmd.Output()
	if err != nil {
		logger.V(2).Infof("git rev-parse HEAD in %s failed: %v", workDir, err)
		return ""
	}
	return strings.TrimSpace(string(out))
}
