package ui

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/flanksource/commons/logger"
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
	runs, err := snapshots.ListRuns(workDir, snapshots.ListRunsOptions{Since: started.Add(-time.Nanosecond)})
	if err != nil || len(runs) == 0 {
		return
	}
	snapshot, err := readProjectActionSnapshot(runs[0].Path)
	if err == nil {
		run.LoadSnapshot(*snapshot)
	}
}

// serveProjectActionTraces reopens the store the finished run left, read-only,
// so its Traces sections and Processes tab read it; the project-runs
// MultiServer closes it as it evicts the run. A store that does not open is
// the run's error.
func serveProjectActionTraces(run *testui.Server) {
	trace, err := testui.TraceRunOf(run.Snapshot().Metadata)
	if err == nil && trace != nil {
		err = run.ReopenTraces(*trace)
	}
	if err != nil {
		logger.Errorf("serve a project action run's trace store: %v", err)
		run.SetRunError(fmt.Errorf("serve the run's trace store: %w", err))
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
