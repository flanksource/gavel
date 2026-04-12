package linters

import "github.com/flanksource/gavel/verify"

// FilterIgnoredViolations removes violations matching any ignore rule from results in place.
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
