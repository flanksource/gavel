package linters

import (
	"testing"

	"github.com/flanksource/gavel/models"
	"github.com/stretchr/testify/assert"
)

const fingerprintWorkDir = "/repo"

func fingerprinted(violations ...models.Violation) []string {
	result := &LinterResult{Linter: "vale", WorkDir: fingerprintWorkDir, Violations: violations}
	AssignFingerprints(result)
	prints := make([]string, len(result.Violations))
	for i, v := range result.Violations {
		prints[i] = v.Fingerprint
	}
	return prints
}

func finding(file string, line int, rule, message string) models.Violation {
	return models.Violation{File: file, Line: line, Rule: &models.Rule{Method: rule}, Message: models.StringPtr(message)}
}

func TestAssignFingerprints(t *testing.T) {
	const (
		file    = "docs/guide.md"
		rule    = "Vale.Repetition"
		message = "'is' is repeated!"
	)

	t.Run("ignores the line, so a finding keeps its fingerprint when code shifts above it", func(t *testing.T) {
		assert.Equal(t, fingerprinted(finding(file, 3, rule, message)), fingerprinted(finding(file, 30, rule, message)))
	})

	t.Run("treats a relative and an absolute path to the same file alike", func(t *testing.T) {
		assert.Equal(t, fingerprinted(finding(file, 3, rule, message)), fingerprinted(finding(fingerprintWorkDir+"/"+file, 3, rule, message)))
	})

	t.Run("collapses whitespace in the message", func(t *testing.T) {
		assert.Equal(t, fingerprinted(finding(file, 3, rule, message)), fingerprinted(finding(file, 3, rule, "  'is'   is\nrepeated! ")))
	})

	t.Run("changes with the rule, the file or the message", func(t *testing.T) {
		base := fingerprinted(finding(file, 3, rule, message))[0]
		others := fingerprinted(
			finding(file, 3, "Vale.Spelling", message),
			finding("docs/other.md", 3, rule, message),
			finding(file, 3, rule, "'and' is repeated!"),
		)
		assert.NotContains(t, others, base)
	})

	t.Run("numbers repeats of one finding in line order, so each stays distinct", func(t *testing.T) {
		prints := fingerprinted(finding(file, 9, rule, message), finding(file, 4, rule, message))
		assert.NotEqual(t, prints[0], prints[1])
		assert.Equal(t, fingerprinted(finding(file, 4, rule, message))[0], prints[1], "the earlier line keeps the unnumbered fingerprint")
	})

	t.Run("keys a symbol-anchored finding on the symbol, not the file it lives in", func(t *testing.T) {
		anchored := func(file string) models.Violation {
			v := finding(file, 3, rule, message)
			v.Anchor = &models.ViolationAnchor{Symbol: "sym-retry"}
			return v
		}
		assert.Equal(t, fingerprinted(anchored("pkg/retry/retry.go")), fingerprinted(anchored("pkg/backoff/retry.go")))
	})

	t.Run("separates findings in different markdown sections of one file", func(t *testing.T) {
		section := func(heading string) models.Violation {
			v := finding(file, 3, rule, message)
			v.Anchor = &models.ViolationAnchor{Section: heading}
			return v
		}
		prints := fingerprinted(section("Install"), section("Usage"))
		assert.NotEqual(t, prints[0], prints[1])
	})
}
