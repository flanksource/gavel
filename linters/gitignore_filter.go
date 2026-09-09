package linters

import (
	"github.com/flanksource/gavel/models"
	"github.com/flanksource/gavel/utils"
)

// FilterViolationsByGitIgnoreInResults removes gitignored violations from each
// result in place, using each result's own WorkDir to resolve every Git exclude
// source. Returns the total number of filtered violations.
func FilterViolationsByGitIgnoreInResults(results []*LinterResult) (int, error) {
	filtered := 0
	for _, result := range results {
		if result == nil || len(result.Violations) == 0 {
			continue
		}
		before := len(result.Violations)
		violations, err := FilterViolationsByGitIgnore(result.Violations, result.WorkDir)
		if err != nil {
			return 0, err
		}
		result.Violations = violations
		filtered += before - len(result.Violations)
	}
	return filtered, nil
}

// FilterViolationsByGitIgnore removes violations whose File is matched by
// Git's repository, info, global, or XDG exclude rules for workDir.
func FilterViolationsByGitIgnore(violations []models.Violation, workDir string) ([]models.Violation, error) {
	if len(violations) == 0 {
		return violations, nil
	}

	seen := make(map[string]bool, len(violations))
	var paths []string
	for _, v := range violations {
		if v.File != "" && !seen[v.File] {
			seen[v.File] = true
			paths = append(paths, v.File)
		}
	}

	ignored, err := utils.GitIgnoredPaths(paths, workDir)
	if err != nil {
		return nil, err
	}

	var result []models.Violation
	for _, v := range violations {
		if _, excluded := ignored[v.File]; v.File == "" || !excluded {
			result = append(result, v)
		}
	}
	return result, nil
}
