package linters

import (
	"path/filepath"

	"github.com/flanksource/gavel/models"
	"github.com/flanksource/gavel/utils"
)

// FilterViolationsByGitIgnoreInResults removes gitignored violations from each
// result in place, using each result's own WorkDir to resolve every Git exclude
// source. Violations in a file listed in named (absolute paths the user asked
// to lint by name) are always kept. Returns the total number of filtered
// violations.
func FilterViolationsByGitIgnoreInResults(results []*LinterResult, named map[string]struct{}) (int, error) {
	filtered := 0
	for _, result := range results {
		if result == nil || len(result.Violations) == 0 {
			continue
		}
		before := len(result.Violations)
		violations, err := FilterViolationsByGitIgnore(result.Violations, result.WorkDir, named)
		if err != nil {
			return 0, err
		}
		result.Violations = violations
		filtered += before - len(result.Violations)
	}
	return filtered, nil
}

// FilterViolationsByGitIgnore removes violations whose File is matched by
// Git's repository, info, global, or XDG exclude rules for workDir. A violation
// whose File resolves (against workDir when relative) to a path in named is
// kept: naming a file asks for it to be linted even when it is gitignored.
func FilterViolationsByGitIgnore(violations []models.Violation, workDir string, named map[string]struct{}) ([]models.Violation, error) {
	if len(violations) == 0 {
		return violations, nil
	}

	seen := make(map[string]bool, len(violations))
	var paths []string
	for _, v := range violations {
		if v.File == "" || seen[v.File] {
			continue
		}
		seen[v.File] = true
		if !isNamedFile(v.File, workDir, named) {
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

func isNamedFile(file, workDir string, named map[string]struct{}) bool {
	if len(named) == 0 {
		return false
	}
	if !filepath.IsAbs(file) {
		file = filepath.Join(workDir, file)
	}
	_, ok := named[filepath.Clean(file)]
	return ok
}
