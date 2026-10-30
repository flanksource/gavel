package linters

import (
	"strings"
	"testing"

	"github.com/flanksource/clicky/api/icons"
	"github.com/flanksource/gavel/models"
	"github.com/stretchr/testify/assert"
)

func TestLinterResultPrettyStatusBySeverity(t *testing.T) {
	cases := map[string]struct {
		severities []models.ViolationSeverity
		status     string
	}{
		"info only passes":     {severities: []models.ViolationSeverity{models.SeverityInfo, models.SeverityInfo}, status: icons.Pass.Unicode},
		"a warning warns":      {severities: []models.ViolationSeverity{models.SeverityInfo, models.SeverityWarning}, status: "⚠"},
		"an error warns":       {severities: []models.ViolationSeverity{models.SeverityError}, status: "⚠"},
		"unset severity warns": {severities: []models.ViolationSeverity{""}, status: "⚠"},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			result := &LinterResult{Linter: "oipa-transaction", Success: true}
			for _, severity := range tc.severities {
				result.Violations = append(result.Violations, violation("tx/a", 0, "rule", "msg", severity))
			}
			assert.True(t, strings.HasPrefix(result.Pretty().String(), tc.status+" "), result.Pretty().String())
		})
	}
}

func TestViolationNodePrettyIconBySeverity(t *testing.T) {
	cases := map[models.ViolationSeverity]string{
		models.SeverityError:   "✗ ",
		models.SeverityWarning: "⚠ ",
		models.SeverityInfo:    "ℹ ",
	}
	for severity, icon := range cases {
		t.Run(string(severity), func(t *testing.T) {
			v := violation("tx/a", 0, "rule", "evaluated", severity)
			assert.True(t, strings.HasPrefix((&violationNode{v: &v}).Pretty().String(), icon))
		})
	}
}
