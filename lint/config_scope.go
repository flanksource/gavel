package lint

import (
	"path/filepath"
	"slices"

	"github.com/flanksource/gavel/linters"
	"github.com/flanksource/gavel/utils"
)

func resultScopes(result *linters.LinterResult, workDir string, files []string) []string {
	scopes := slices.Clone(files)
	for _, file := range files {
		absolute := resolveLintPath(workDir, file)
		if filepath.Base(absolute) == ".gavel.yaml" && utils.IsWithin(result.WorkDir, filepath.Dir(absolute)) {
			return append(scopes, result.WorkDir)
		}
		if result.Linter == "uir" && utils.IsWithin(absolute, filepath.Join(result.WorkDir, ".uir.d")) {
			return append(scopes, result.WorkDir)
		}
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
