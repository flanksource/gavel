package snapshots

import (
	"testing"

	"github.com/flanksource/gavel/linters"
	"github.com/flanksource/gavel/models"
	"github.com/stretchr/testify/assert"
)

func lintRun(open []string, ignored []string, dismissed []string) []*linters.LinterResult {
	result := &linters.LinterResult{Linter: "vale"}
	for _, fp := range open {
		result.Violations = append(result.Violations, models.Violation{Fingerprint: fp})
	}
	for _, fp := range ignored {
		result.Ignored = append(result.Ignored, linters.IgnoredViolation{Violation: models.Violation{Fingerprint: fp}})
	}
	for _, fp := range dismissed {
		result.Dismissed = append(result.Dismissed, models.Violation{Fingerprint: fp})
	}
	return []*linters.LinterResult{result}
}

func TestDiffLint(t *testing.T) {
	prev := lintRun([]string{"persisting", "fixed", "now-ignored", "now-dismissed"}, nil, nil)
	cur := lintRun([]string{"persisting", "new"}, []string{"now-ignored"}, []string{"now-dismissed"})

	diff, err := DiffLint(prev, cur)

	assert.NoError(t, err)
	assert.Equal(t, LintDiff{
		New:        []string{"new"},
		Fixed:      []string{"fixed"},
		Persisting: []string{"persisting"},
		Ignored:    []string{"now-ignored"},
		Dismissed:  []string{"now-dismissed"},
	}, diff)
}

func TestDiffLintRejectsUnfingerprintedViolations(t *testing.T) {
	_, err := DiffLint(nil, lintRun([]string{""}, nil, nil))
	assert.EqualError(t, err, "vale violation at :0 has no fingerprint; lint results must pass through lint.FilterConfiguredViolations")
}
