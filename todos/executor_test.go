package todos

import (
	"testing"

	"github.com/flanksource/clicky/task"
	"github.com/flanksource/gavel/fixtures"
)

func TestAllPassed_AllSuccess(t *testing.T) {
	// Input: []fixtures.FixtureResult with all IsOK() == true
	results := []fixtures.FixtureResult{
		{Status: task.StatusPASS},
		{Status: task.StatusSuccess},
	}

	if !AllPassed(results) {
		t.Error("Expected AllPassed to return true for all successful results")
	}
}

func TestAllPassed_OneFailed(t *testing.T) {
	// Input: []fixtures.FixtureResult with one IsOK() == false
	results := []fixtures.FixtureResult{
		{Status: task.StatusPASS},
		{Status: task.StatusFailed},
	}

	if AllPassed(results) {
		t.Error("Expected AllPassed to return false when one result failed")
	}
}

func TestAllPassed_Empty(t *testing.T) {
	// Input: Empty slice
	results := []fixtures.FixtureResult{}

	if !AllPassed(results) {
		t.Error("Expected AllPassed to return true for empty results")
	}
}
