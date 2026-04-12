package linters

import (
	"path/filepath"

	"github.com/flanksource/gavel/verify"
)

// FilterIgnoredViolations removes violations matching any ignore rule from results in place.
// Violation file paths are relativized against each result's WorkDir before matching,
// so that relative ignore-rule globs match absolute violation paths.
// Returns the total number of filtered violations.
func FilterIgnoredViolations(results []*LinterResult, rules []verify.LintIgnoreRule) int {
	if len(rules) == 0 {
		return 0
	}
	filtered := 0
	for _, result := range results {
		if result == nil {
			continue
		}
		kept := result.Violations[:0]
		for _, v := range result.Violations {
			if result.WorkDir != "" && filepath.IsAbs(v.File) {
				if rel, err := filepath.Rel(result.WorkDir, v.File); err == nil {
					v.File = rel
				}
			}
			match := false
			for _, rule := range rules {
				if rule.MatchesViolation(v) {
					match = true
					break
				}
			}
			if match {
				filtered++
			} else {
				kept = append(kept, v)
			}
		}
		result.Violations = kept
	}
	return filtered
}
