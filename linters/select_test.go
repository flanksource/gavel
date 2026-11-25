package linters

import (
	"testing"

	"github.com/flanksource/gavel/models"
	"github.com/stretchr/testify/assert"
)

func TestSelectViolations(t *testing.T) {
	results := []*LinterResult{
		{Linter: "vale", WorkDir: "/repo", Violations: []models.Violation{{Fingerprint: "a"}, {Fingerprint: "b"}}, Dismissed: []models.Violation{{Fingerprint: "c"}}},
		{Linter: "markdownlint", WorkDir: "/repo", Violations: []models.Violation{{Fingerprint: "d"}}},
	}

	selected := SelectViolations(results, map[string]bool{"b": true, "c": true})

	assert.Equal(t, []*LinterResult{{Linter: "vale", WorkDir: "/repo", Violations: []models.Violation{{Fingerprint: "b"}}}}, selected)
	assert.Len(t, results[0].Violations, 2, "the input is left as it was")
}
