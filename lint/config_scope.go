package lint

import (
	"path/filepath"
	"slices"

	"github.com/flanksource/gavel/linters"
)

func resultScopes(result *linters.LinterResult, workDir string, files []string) []string {
	scopes := slices.Clone(files)
	for _, file := range files {
		absolute := resolveLintPath(workDir, file)
		if filepath.Dir(absolute) != result.WorkDir {
			continue
		}
		for _, pattern := range linterConfigPatterns(result.Linter) {
			if directPatternMatch(filepath.Base(absolute), pattern) {
				return append(scopes, result.WorkDir)
			}
		}
	}
	return scopes
}
