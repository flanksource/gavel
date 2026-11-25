package linters

import "github.com/flanksource/gavel/models"

// SelectViolations returns copies of the results holding only the open
// violations whose fingerprint is in fingerprints, dropping results left with
// none. Ignored and dismissed violations are never selected: it is what a
// scoped AI fix is asked to repair.
func SelectViolations(results []*LinterResult, fingerprints map[string]bool) []*LinterResult {
	var selected []*LinterResult
	for _, result := range results {
		if result == nil {
			continue
		}
		var kept []models.Violation
		for _, v := range result.Violations {
			if fingerprints[v.Fingerprint] {
				kept = append(kept, v)
			}
		}
		if len(kept) == 0 {
			continue
		}
		copy := *result
		copy.Violations, copy.Ignored, copy.Dismissed = kept, nil, nil
		selected = append(selected, &copy)
	}
	return selected
}
