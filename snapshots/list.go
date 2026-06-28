package snapshots

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/flanksource/gavel/testrunner/parsers"
	testui "github.com/flanksource/gavel/testrunner/ui"
)

// Run kinds reported by ListRuns, derived from a snapshot's tests/lint content.
const (
	RunKindTest     = "test"
	RunKindLint     = "lint"
	RunKindTestLint = "test+lint"
)

// RunInfo is the per-run summary ListRuns derives from a single
// .gavel/run-<timestamp>.json file. It carries enough to render a run list row
// and locate the snapshot on disk (Path) without re-reading the file.
type RunInfo struct {
	RunID      string    `json:"runId"`             // file stem, e.g. run-2026-06-28T10-43-42Z
	Path       string    `json:"path"`              // absolute path to the run-*.json
	Started    time.Time `json:"started,omitempty"` // run start (UTC)
	Ended      time.Time `json:"ended,omitempty"`   // run end (UTC)
	Kind       string    `json:"kind"`              // test | lint | test+lint
	Repo       string    `json:"repo,omitempty"`    // git repo name (Git.Repo)
	SHA        string    `json:"sha,omitempty"`     // git HEAD sha at run time
	Frameworks []string  `json:"frameworks,omitempty"`

	Passed  int `json:"passed"`
	Failed  int `json:"failed"`
	Skipped int `json:"skipped"`
	Warned  int `json:"warned"`
	Total   int `json:"total"`

	LintViolations int `json:"lintViolations"`
	LintLinters    int `json:"lintLinters"`
}

// ListRuns enumerates the per-run snapshots in workDir/.gavel (run-*.json),
// returning one RunInfo per run sorted newest-first. When since is non-zero,
// runs whose start time is at or before since are skipped — run-*.json files
// are immutable once written, so the start time is a sound incremental
// watermark. A missing .gavel directory yields (nil, nil): the normal "never
// run here" state, mirroring history.loadRunSnapshots.
func ListRuns(workDir string, since time.Time) ([]RunInfo, error) {
	if workDir == "" {
		return nil, errors.New("snapshots.ListRuns: workDir is required")
	}
	dir := filepath.Join(workDir, Dir)
	entries, err := os.ReadDir(dir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, fmt.Errorf("read %s: %w", dir, err)
	}

	var runs []RunInfo
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasPrefix(name, PerRunPrefix) || !strings.HasSuffix(name, ".json") {
			continue
		}
		path := filepath.Join(dir, name)
		snap, err := loadSnapshot(path)
		if err != nil {
			return nil, err
		}
		started := runStartTime(*snap, name, path, entry)
		if !since.IsZero() && !started.After(since) {
			continue
		}
		runs = append(runs, runInfo(*snap, name, path, started))
	}

	sort.Slice(runs, func(i, j int) bool {
		if runs[i].Started.Equal(runs[j].Started) {
			return runs[i].Path > runs[j].Path
		}
		return runs[i].Started.After(runs[j].Started)
	})
	return runs, nil
}

func loadSnapshot(path string) (*testui.Snapshot, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read run snapshot %s: %w", path, err)
	}
	var snap testui.Snapshot
	if err := json.Unmarshal(data, &snap); err != nil {
		return nil, fmt.Errorf("decode run snapshot %s: %w", path, err)
	}
	return &snap, nil
}

// runInfo builds a RunInfo from a decoded snapshot. Counts come from the leaf
// roll-up (Tests.Sum); lint counts skip linters that did not run.
func runInfo(snap testui.Snapshot, name, path string, started time.Time) RunInfo {
	sum := parsers.Tests(snap.Tests).Sum()

	info := RunInfo{
		RunID:   strings.TrimSuffix(name, ".json"),
		Path:    path,
		Started: started,
		Passed:  sum.Passed,
		Failed:  sum.Failed,
		Skipped: sum.Skipped,
		Warned:  sum.Warned,
		Total:   sum.Total,
	}
	for _, lr := range snap.Lint {
		if lr == nil || lr.Skipped {
			continue
		}
		info.LintLinters++
		info.LintViolations += len(lr.Violations)
	}
	if snap.Metadata != nil {
		info.Ended = snap.Metadata.Ended.UTC()
		info.Frameworks = snap.Metadata.Frameworks
	}
	if snap.Git != nil {
		info.Repo = snap.Git.Repo
		info.SHA = snap.Git.SHA
	}
	info.Kind = runKind(sum.Total, info.LintLinters, snap.Status.LintRun)
	return info
}

// runKind classifies a run from its leaf-test count, the number of linters that
// ran, and the snapshot's LintRun flag (a lint run with zero violations still
// ran linters and must read as "lint", not "test").
func runKind(testTotal, lintLinters int, lintRun bool) string {
	hasTests := testTotal > 0
	hasLint := lintLinters > 0 || lintRun
	switch {
	case hasTests && hasLint:
		return RunKindTestLint
	case hasLint:
		return RunKindLint
	default:
		return RunKindTest
	}
}

// runStartTime resolves a run's start time, preferring the recorded metadata,
// then the filename timestamp, then the file mtime — the same precedence
// history.snapshotTime uses.
func runStartTime(snap testui.Snapshot, name, path string, entry os.DirEntry) time.Time {
	if snap.Metadata != nil && !snap.Metadata.Started.IsZero() {
		return snap.Metadata.Started.UTC()
	}
	stem := strings.TrimSuffix(strings.TrimPrefix(name, PerRunPrefix), ".json")
	if ts, err := time.Parse(PerRunTimestampLayout, stem); err == nil {
		return ts.UTC()
	}
	if info, err := entry.Info(); err == nil {
		return info.ModTime().UTC()
	}
	if info, err := os.Stat(path); err == nil {
		return info.ModTime().UTC()
	}
	return time.Time{}
}
