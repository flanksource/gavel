package snapshots

import (
	"time"

	"github.com/flanksource/gavel/testrunner/parsers"
)

// TestsOfFailedRunsSince returns the test trees of every run in workDir that
// started after since and recorded at least one failed test, newest first.
func TestsOfFailedRunsSince(workDir string, since time.Time) ([]parsers.Test, error) {
	runs, err := ListRuns(workDir, ListRunsOptions{Since: since})
	if err != nil {
		return nil, err
	}
	var tests []parsers.Test
	for _, run := range runs {
		if run.Failed == 0 {
			continue
		}
		snap, err := LoadRun(workDir, run.RunID)
		if err != nil {
			return nil, err
		}
		tests = append(tests, snap.Tests...)
	}
	return tests, nil
}
