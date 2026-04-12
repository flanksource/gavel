package linters

import (
	"github.com/flanksource/gavel/models"
	"github.com/flanksource/gavel/utils"
)

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
