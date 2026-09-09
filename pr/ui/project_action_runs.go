package ui

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/flanksource/gavel/snapshots"
	testui "github.com/flanksource/gavel/testrunner/ui"
)

func streamProjectActionSnapshots(workDir string, started time.Time, run *testui.Server, stop <-chan struct{}) {
	ticker := time.NewTicker(250 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-stop:
			loadRunningProjectActionSnapshot(workDir, started, run)
			return
		case <-ticker.C:
			loadRunningProjectActionSnapshot(workDir, started, run)
		}
	}
}

func loadRunningProjectActionSnapshot(workDir string, started time.Time, run *testui.Server) {
	path := filepath.Join(workDir, snapshots.Dir, snapshots.RunningName)
	snapshot, err := readProjectActionSnapshot(path)
	if err != nil || snapshot.Metadata == nil || snapshot.Metadata.Started.Before(started) {
		return
	}
	run.LoadSnapshot(*snapshot)
}

func loadCompletedProjectActionSnapshot(workDir string, started time.Time, run *testui.Server) {
	runs, err := snapshots.ListRuns(workDir, started.Add(-time.Nanosecond))
	if err != nil || len(runs) == 0 {
		return
	}
	snapshot, err := readProjectActionSnapshot(runs[0].Path)
	if err == nil {
		run.LoadSnapshot(*snapshot)
	}
}

func readProjectActionSnapshot(path string) (*testui.Snapshot, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var snapshot testui.Snapshot
	if err := json.Unmarshal(data, &snapshot); err != nil {
		return nil, err
	}
	return &snapshot, nil
}
