package snapshots

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/flanksource/gavel/blob"
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
// run-<timestamp>.json snapshot. It carries enough to render a run list row;
// ReadRun(workDir, RunID) reads the snapshot itself.
type RunInfo struct {
	RunID      string    `json:"runId"`             // file stem, e.g. run-2026-06-28T10-43-42Z
	Path       string    `json:"path"`              // where the snapshot is stored (see Location)
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

// LastRun returns the summary of the snapshot last.json points at. A missing pointer is the
// normal "never run" state and returns nil without an error.
func LastRun(workDir string) (*RunInfo, error) {
	if workDir == "" {
		return nil, errors.New("snapshots.LastRun: workDir is required")
	}
	store, err := OpenStore(workDir)
	if err != nil {
		return nil, err
	}
	pointer, err := loadPointer(store, PointerLast)
	if err != nil || pointer == nil {
		return nil, err
	}
	// Save records `.gavel/<name>.json`. A pointer naming anything else is
	// corrupt or hand-edited, so it is rejected before the snapshot behind it
	// is read.
	key, err := pointerKey(pointer)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", Location(workDir, PointerLast+".json"), err)
	}
	snap, err := loadRun(store, workDir, key)
	if err != nil {
		return nil, err
	}
	var modTime time.Time
	if objects, err := store.List(context.Background(), key); err == nil && len(objects) > 0 {
		modTime = objects[0].ModTime
	}
	info := runInfo(*snap, PointerLast+".json", Location(workDir, key), runStartTime(*snap, PointerLast+".json", modTime))
	return &info, nil
}

// ListRunsOptions narrows ListRuns.
type ListRunsOptions struct {
	// Since skips runs whose start time is at or before it — run-*.json
	// snapshots are immutable once written, so the start time is a sound
	// incremental watermark.
	Since time.Time
	// Before skips runs whose RunID sorts at or after it: the cursor for the
	// next page is the oldest RunID of the previous one.
	Before string
	// Limit stops after this many runs, taken newest-first by RunID (the start
	// timestamp), so only the snapshots of one page are read. 0 means no limit.
	Limit int
}

// ListRuns enumerates workDir's per-run snapshots (run-*.json), returning one RunInfo per run
// sorted newest-first. A workspace that never ran yields no runs.
func ListRuns(workDir string, opts ListRunsOptions) ([]RunInfo, error) {
	if workDir == "" {
		return nil, errors.New("snapshots.ListRuns: workDir is required")
	}
	store, err := OpenStore(workDir)
	if err != nil {
		return nil, err
	}
	objects, err := runObjects(store)
	if err != nil {
		return nil, err
	}
	index, err := loadIndex(store)
	if err != nil {
		return nil, err
	}

	var runs []RunInfo
	for _, object := range objects {
		if opts.Limit > 0 && len(runs) >= opts.Limit {
			break
		}
		if opts.Before != "" && strings.TrimSuffix(object.Key, ".json") >= opts.Before {
			continue
		}
		// Decide the cutoff from the listing before reading the snapshot. A run
		// snapshot carries the whole test tree and lint violation set — the
		// workspaces on one developer machine total 320MB across 1171 files — so
		// parsing every run only to discard it by start time made an incremental
		// sweep cost the same as a full one.
		//
		// Only the modification time can gate this. Stores may round it down to
		// the nearest second, so snapshots from the watermark's second still need
		// their recorded start time checked below. Older ones can be skipped safely.
		// The name's timestamp is NOT usable here even though runStartTime
		// prefers it: PerRunTimestampLayout has the same second precision.
		if !opts.Since.IsZero() && !object.ModTime.IsZero() && !object.ModTime.UTC().Add(time.Second).After(opts.Since) {
			continue
		}
		info, err := summarizeRun(store, index, workDir, object)
		if err != nil {
			return nil, err
		}
		if !opts.Since.IsZero() && !info.Started.After(opts.Since) {
			continue
		}
		runs = append(runs, info)
	}
	if err := index.save(store); err != nil {
		return nil, err
	}

	sort.Slice(runs, func(i, j int) bool { return newerRun(runs[i], runs[j]) })
	return runs, nil
}

// Run is one per-run snapshot and its summary.
type Run struct {
	Info     RunInfo
	Snapshot testui.Snapshot
}

// LoadRuns returns every per-run snapshot of workDir with its summary, newest first, in
// ListRuns' order.
func LoadRuns(workDir string) ([]Run, error) {
	if workDir == "" {
		return nil, errors.New("snapshots.LoadRuns: workDir is required")
	}
	store, err := OpenStore(workDir)
	if err != nil {
		return nil, err
	}
	objects, err := runObjects(store)
	if err != nil {
		return nil, err
	}
	var runs []Run
	for _, object := range objects {
		snap, err := loadRun(store, workDir, object.Key)
		if err != nil {
			return nil, err
		}
		info := runInfo(*snap, object.Key, Location(workDir, object.Key), runStartTime(*snap, object.Key, object.ModTime))
		runs = append(runs, Run{Info: info, Snapshot: *snap})
	}
	sort.Slice(runs, func(i, j int) bool { return newerRun(runs[i].Info, runs[j].Info) })
	return runs, nil
}

// runObjects lists the store's run-*.json snapshots newest-first by key (the
// start timestamp).
func runObjects(store blob.Store) ([]blob.Object, error) {
	objects, err := store.List(context.Background(), PerRunPrefix)
	if err != nil {
		return nil, err
	}
	var runs []blob.Object
	for _, object := range objects {
		if strings.Contains(object.Key, "/") || !strings.HasSuffix(object.Key, ".json") {
			continue
		}
		runs = append(runs, object)
	}
	sort.Slice(runs, func(i, j int) bool { return runs[i].Key > runs[j].Key })
	return runs, nil
}

func newerRun(a, b RunInfo) bool {
	if a.Started.Equal(b.Started) {
		return a.Path > b.Path
	}
	return a.Started.After(b.Started)
}

// summarizeRun returns a run's RunInfo from the index, decoding the snapshot
// only when the index has no entry for it at its current size and mtime.
func summarizeRun(store blob.Store, index *runIndex, workDir string, object blob.Object) (RunInfo, error) {
	location := Location(workDir, object.Key)
	if info, ok := index.lookup(object, location); ok {
		return info, nil
	}
	snap, err := loadRun(store, workDir, object.Key)
	if err != nil {
		return RunInfo{}, err
	}
	info := runInfo(*snap, object.Key, location, runStartTime(*snap, object.Key, object.ModTime))
	index.store(object, info)
	return info, nil
}

// LatestRuns returns the locations of the newest run-*.json that ran tests
// and the newest that ran linters, newest first — a single location when one
// run did both. Runs are visited newest-first by key (the start timestamp) and
// the scan stops once both are found, so the long tail of older multi-MB runs
// is never read. A workspace that never ran yields no runs.
func LatestRuns(workDir string) ([]string, error) {
	if workDir == "" {
		return nil, errors.New("snapshots.LatestRuns: workDir is required")
	}
	store, err := OpenStore(workDir)
	if err != nil {
		return nil, err
	}
	objects, err := runObjects(store)
	if err != nil {
		return nil, err
	}

	var testKey, lintKey string
	var paths []string
	for _, object := range objects {
		hasTests, hasLint, err := runContents(store, workDir, object.Key)
		if err != nil {
			return nil, err
		}
		if (hasTests && testKey == "") || (hasLint && lintKey == "") {
			paths = append(paths, Location(workDir, object.Key))
		}
		if hasTests && testKey == "" {
			testKey = object.Key
		}
		if hasLint && lintKey == "" {
			lintKey = object.Key
		}
		if testKey != "" && lintKey != "" {
			break
		}
	}
	return paths, nil
}

// runContents reports whether a run snapshot holds tests and lint results,
// decoding into empty element types so the test tree and violations are
// skipped rather than materialised.
func runContents(store blob.Store, workDir, key string) (hasTests, hasLint bool, err error) {
	data, err := readKey(store, key)
	if err != nil {
		return false, false, fmt.Errorf("read run snapshot %s: %w", Location(workDir, key), err)
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
		return false, false, fmt.Errorf("decode run snapshot %s: %w", Location(workDir, key), err)
	}
	hasLint = shape.Status.LintRun
	for _, l := range shape.Lint {
		hasLint = hasLint || !l.Skipped
	}
	return len(shape.Tests) > 0, hasLint, nil
}

// LoadRun reads and decodes the snapshot runID names (see ReadRun).
func LoadRun(workDir, runID string) (*testui.Snapshot, error) {
	data, err := ReadRun(workDir, runID)
	if err != nil {
		return nil, fmt.Errorf("read run snapshot %s: %w", runID, err)
	}
	var snap testui.Snapshot
	if err := json.Unmarshal(data, &snap); err != nil {
		return nil, fmt.Errorf("decode run snapshot %s: %w", runID, err)
	}
	return &snap, nil
}

func loadRun(store blob.Store, workDir, key string) (*testui.Snapshot, error) {
	data, err := readKey(store, key)
	if err != nil {
		return nil, fmt.Errorf("read run snapshot %s: %w", Location(workDir, key), err)
	}
	var snap testui.Snapshot
	if err := json.Unmarshal(data, &snap); err != nil {
		return nil, fmt.Errorf("decode run snapshot %s: %w", Location(workDir, key), err)
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
// then the name's timestamp, then the snapshot's modification time — the same
// precedence history.snapshotTime uses.
func runStartTime(snap testui.Snapshot, name string, modTime time.Time) time.Time {
	if snap.Metadata != nil && !snap.Metadata.Started.IsZero() {
		return snap.Metadata.Started.UTC()
	}
	stem := strings.TrimSuffix(strings.TrimPrefix(name, PerRunPrefix), ".json")
	if ts, err := time.Parse(PerRunTimestampLayout, stem); err == nil {
		return ts.UTC()
	}
	return modTime.UTC()
}

// Summarize is the RunInfo of snap, stored under key at location: its counts, kind and start
// time, which falls back to key's timestamp and then modTime.
func Summarize(snap testui.Snapshot, key, location string, modTime time.Time) RunInfo {
	return runInfo(snap, key, location, runStartTime(snap, key, modTime))
}
