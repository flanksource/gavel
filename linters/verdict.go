package linters

// HasFailures uses the same severity and execution verdict as test reports.
func HasFailures(results []*LinterResult) bool {
	for _, result := range results {
		if result != nil && result.Test().Sum().Failed > 0 {
			return true
		}
	}
	return false
}
