package linters

import (
	"github.com/flanksource/gavel/models"
	"github.com/flanksource/gavel/utils"
)

// FilterViolationsByGitIgnoreInResults removes gitignored violations from each
// result in place, using each result's own WorkDir to discover .gitignore
// patterns. Returns the total number of filtered violations.
func FilterViolationsByGitIgnoreInResults(results []*LinterResult) int {
	filtered := 0
	for _, result := range results {
		if result == nil || len(result.Violations) == 0 {
			continue
		}
		before := len(result.Violations)
		result.Violations = FilterViolationsByGitIgnore(result.Violations, result.WorkDir)
		filtered += before - len(result.Violations)
	}
	return filtered
}

// FilterViolationsByGitIgnore removes violations whose File is matched by
// .gitignore patterns found in the git repository containing workDir.
func FilterViolationsByGitIgnore(violations []models.Violation, workDir string) []models.Violation {
	if len(violations) == 0 {
		return violations
	}

	seen := make(map[string]bool, len(violations))
	var paths []string
	for _, v := range violations {
		if v.File != "" && !seen[v.File] {
			seen[v.File] = true
			paths = append(paths, v.File)
		}
	}

	kept := make(map[string]bool, len(paths))
	for _, p := range utils.FilterGitIgnored(paths, workDir) {
		kept[p] = true
	}

	var result []models.Violation
	for _, v := range violations {
		if v.File == "" || kept[v.File] {
			result = append(result, v)
		}
	}
	return result
}
