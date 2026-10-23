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
	"github.com/flanksource/gavel/utils"
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

// LastRun returns the summary of the snapshot referenced by
// .gavel/last.json. A missing pointer is the normal "never run" state and
// returns nil without an error.
func LastRun(workDir string) (*RunInfo, error) {
	if workDir == "" {
		return nil, errors.New("snapshots.LastRun: workDir is required")
	}
	pointer, err := LoadPointer(workDir, PointerLast)
	if err != nil || pointer == nil {
		return nil, err
	}
	// Save records a workDir-relative `.gavel/<name>.json`. A pointer naming
	// anything else is corrupt or hand-edited, so it is rejected before the
	// snapshot behind it is stat'd or decoded.
	path, err := utils.ResolveWithin(workDir, pointer.Path)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", filepath.Join(Dir, PointerLast+".json"), err)
	}
	snap, err := LoadByPointer(workDir, pointer)
	if err != nil {
		return nil, err
	}
	info := runInfo(*snap, PointerLast+".json", path, runStartTime(*snap, PointerLast+".json", path, nil))
	return &info, nil
}

// ListRunsOptions narrows ListRuns.
type ListRunsOptions struct {
	// Since skips runs whose start time is at or before it — run-*.json files
	// are immutable once written, so the start time is a sound incremental
	// watermark.
	Since time.Time
	// Before skips runs whose RunID sorts at or after it: the cursor for the
	// next page is the oldest RunID of the previous one.
	Before string
	// Limit stops after this many runs, taken newest-first by RunID (the start
	// timestamp), so only the files of one page are read. 0 means no limit.
	Limit int
}

// ListRuns enumerates the per-run snapshots in workDir/.gavel (run-*.json),
// returning one RunInfo per run sorted newest-first. A missing .gavel
// directory yields (nil, nil): the normal "never run here" state, mirroring
// history.loadRunSnapshots.
func ListRuns(workDir string, opts ListRunsOptions) ([]RunInfo, error) {
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
	since := opts.Since
	index, err := loadIndex(dir)
	if err != nil {
		return nil, err
	}

	var runs []RunInfo
	// os.ReadDir sorts by name, so walking it backwards visits runs newest-first.
	for i := len(entries) - 1; i >= 0 && (opts.Limit <= 0 || len(runs) < opts.Limit); i-- {
		entry := entries[i]
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasPrefix(name, PerRunPrefix) || !strings.HasSuffix(name, ".json") {
			continue
		}
		if opts.Before != "" && strings.TrimSuffix(name, ".json") >= opts.Before {
			continue
		}
		path := filepath.Join(dir, name)
		// Decide the cutoff from the dirent before opening the file. A run
		// snapshot carries the whole test tree and lint violation set — the
		// workspaces on one developer machine total 320MB across 1171 files — so
		// parsing every run only to discard it by start time made an incremental
		// sweep cost the same as a full one.
		//
		// Only the mtime can gate this. Filesystems may round mtimes down to the
		// nearest second, so files from the watermark's second still need their
		// recorded start time checked below. Older files can be skipped safely.
		// The filename timestamp is NOT usable here even though runStartTime
		// prefers it: PerRunTimestampLayout has the same second precision.
		if !since.IsZero() {
			if modTime := runFileModTime(path, entry); !modTime.IsZero() && !modTime.Add(time.Second).After(since) {
				continue
			}
		}
		info, err := summarizeRun(index, entry, path)
		if err != nil {
			return nil, err
		}
		if !since.IsZero() && !info.Started.After(since) {
			continue
		}
		runs = append(runs, info)
	}
	if err := index.save(dir); err != nil {
		return nil, err
	}

	sort.Slice(runs, func(i, j int) bool {
		if runs[i].Started.Equal(runs[j].Started) {
			return runs[i].Path > runs[j].Path
		}
		return runs[i].Started.After(runs[j].Started)
	})
	return runs, nil
}

// summarizeRun returns a run's RunInfo from the index, decoding the run file
// only when the index has no entry for it at its current size and mtime.
func summarizeRun(index *runIndex, entry os.DirEntry, path string) (RunInfo, error) {
	stat, err := entry.Info()
	if err != nil {
		return RunInfo{}, fmt.Errorf("stat run snapshot %s: %w", path, err)
	}
	if info, ok := index.lookup(entry.Name(), path, stat); ok {
		return info, nil
	}
	snap, err := LoadRunFile(path)
	if err != nil {
		return RunInfo{}, err
	}
	info := runInfo(*snap, entry.Name(), path, runStartTime(*snap, entry.Name(), path, entry))
	index.store(entry.Name(), stat, info)
	return info, nil
}

// LatestRuns returns the absolute paths of the newest run-*.json that ran tests
// and the newest that ran linters, newest first — a single path when one run
// did both. Runs are visited newest-first by filename (the start timestamp) and
// the scan stops once both are found, so the long tail of older multi-MB runs
// is never read. A missing .gavel directory yields (nil, nil).
func LatestRuns(workDir string) ([]string, error) {
	if workDir == "" {
		return nil, errors.New("snapshots.LatestRuns: workDir is required")
	}
	dir := filepath.Join(workDir, Dir)
	names, err := filepath.Glob(filepath.Join(dir, PerRunPrefix+"*.json"))
	if err != nil {
		return nil, fmt.Errorf("list %s: %w", dir, err)
	}
	sort.Sort(sort.Reverse(sort.StringSlice(names)))

	var testPath, lintPath string
	var paths []string
	for _, path := range names {
		hasTests, hasLint, err := runContents(path)
		if err != nil {
			return nil, err
		}
		if (hasTests && testPath == "") || (hasLint && lintPath == "") {
			paths = append(paths, path)
		}
		if hasTests && testPath == "" {
			testPath = path
		}
		if hasLint && lintPath == "" {
			lintPath = path
		}
		if testPath != "" && lintPath != "" {
			break
		}
	}
	return paths, nil
}

// runContents reports whether a run snapshot holds tests and lint results,
// decoding into empty element types so the test tree and violations are
// skipped rather than materialised.
func runContents(path string) (hasTests, hasLint bool, err error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return false, false, fmt.Errorf("read run snapshot %s: %w", path, err)
	}
	var shape struct {
		Status struct {
			LintRun bool `json:"lint_run"`
		} `json:"status"`
		Tests []struct{} `json:"tests"`
		Lint  []struct {
			Skipped bool `json:"skipped"`
		} `json:"lint"`
	}
	if err := json.Unmarshal(data, &shape); err != nil {
		return false, false, fmt.Errorf("decode run snapshot %s: %w", path, err)
	}
	hasLint = shape.Status.LintRun
	for _, l := range shape.Lint {
		hasLint = hasLint || !l.Skipped
	}
	return len(shape.Tests) > 0, hasLint, nil
}

// LoadRunFile reads and decodes one run snapshot file.
func LoadRunFile(path string) (*testui.Snapshot, error) {
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
// tests; lint counts skip linters that did not run.
func runInfo(snap testui.Snapshot, name, path string, started time.Time) RunInfo {
	sum := leafSum(snap.Tests)

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

// leafSum totals the leaf tests only. Tests.Sum also counts a package or suite
// whose status rolled up from its children, so one failing test read as five.
func leafSum(tests []parsers.Test) parsers.TestSummary {
	var sum parsers.TestSummary
	for _, test := range tests {
		if len(test.Children) == 0 {
			sum = sum.Add(test.Sum())
			continue
		}
		sum = sum.Add(leafSum(test.Children))
	}
	return sum
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
	return runFileModTime(path, entry)
}

// runFileModTime reads a run file's mtime without opening it, preferring the
// dirent os.ReadDir already stat'd. Run files are written once and never
// rewritten, so this is an upper bound on the run's start time — which is what
// lets ListRuns skip a file without decoding it.
func runFileModTime(path string, entry os.DirEntry) time.Time {
	if entry != nil {
		if info, err := entry.Info(); err == nil {
			return info.ModTime().UTC()
		}
	}
	if info, err := os.Stat(path); err == nil {
		return info.ModTime().UTC()
	}
	return time.Time{}
}
