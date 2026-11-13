package linters

import "github.com/flanksource/gavel/models"

// DismissViolations moves the violations whose fingerprint is in fingerprints
// into each result's Dismissed: hidden for this run only, unlike a lint.ignore
// rule, and kept out of verdicts and AI fixes. Returns how many moved.
func DismissViolations(results []*LinterResult, fingerprints map[string]bool) int {
	return moveViolations(results, fingerprints, func(r *LinterResult) (*[]models.Violation, *[]models.Violation) {
		return &r.Violations, &r.Dismissed
	})
}

// RestoreViolations undoes DismissViolations for fingerprints.
func RestoreViolations(results []*LinterResult, fingerprints map[string]bool) int {
	return moveViolations(results, fingerprints, func(r *LinterResult) (*[]models.Violation, *[]models.Violation) {
		return &r.Dismissed, &r.Violations
	})
}

func moveViolations(results []*LinterResult, fingerprints map[string]bool, lists func(*LinterResult) (from, to *[]models.Violation)) int {
	if len(fingerprints) == 0 {
		return 0
	}
	moved := 0
	for _, result := range results {
		if result == nil {
			continue
		}
		from, to := lists(result)
		kept := (*from)[:0]
		for _, v := range *from {
			if fingerprints[v.Fingerprint] {
				*to = append(*to, v)
				moved++
			} else {
				kept = append(kept, v)
			}
		}
		*from = kept
	}
	return moved
}
