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

// ListRuns enumerates workDir's per-run snapshots (run-*.json), returning one RunInfo per run
// sorted newest-first. When since is non-zero, runs whose start time is at or before since are
// skipped — run-*.json snapshots are immutable once written, so the start time is a sound
// incremental watermark. A workspace that never ran yields no runs.
func ListRuns(workDir string, since time.Time) ([]RunInfo, error) {
	if workDir == "" {
		return nil, errors.New("snapshots.ListRuns: workDir is required")
	}
	store, err := OpenStore(workDir)
	if err != nil {
		return nil, err
	}
	objects, err := store.List(context.Background(), PerRunPrefix)
	if err != nil {
		return nil, err
	}

	var runs []RunInfo
	for _, object := range objects {
		name := object.Key
		if strings.Contains(name, "/") || !strings.HasSuffix(name, ".json") {
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
		if !since.IsZero() && !object.ModTime.IsZero() && !object.ModTime.UTC().Add(time.Second).After(since) {
			continue
		}
		snap, err := loadRun(store, workDir, name)
		if err != nil {
			return nil, err
		}
		started := runStartTime(*snap, name, object.ModTime)
		if !since.IsZero() && !started.After(since) {
			continue
		}
		runs = append(runs, runInfo(*snap, name, Location(workDir, name), started))
	}

	sort.Slice(runs, func(i, j int) bool {
		if runs[i].Started.Equal(runs[j].Started) {
			return runs[i].Path > runs[j].Path
		}
		return runs[i].Started.After(runs[j].Started)
	})
	return runs, nil
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
