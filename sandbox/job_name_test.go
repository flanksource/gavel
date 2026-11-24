// Tests the names of runs' jobs: a run's job name, and the run a job name belongs to, refusing
// names gavel doesn't give.
package sandbox

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestRunOfJobReadsBackTheRunOfAJobName(t *testing.T) {
	run := uuid.New()
	got, ok := RunOfJob(JobName(run))
	assert.True(t, ok)
	assert.Equal(t, run, got)

	for _, name := range []string{"default-template-0", "gavel-run-", "gavel-run-not-a-uuid", run.String()} {
		_, ok := RunOfJob(name)
		assert.False(t, ok, name)
	}
}
