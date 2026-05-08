package linters

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/flanksource/gavel/verify"
)

// FilterIgnoredViolations removes violations matching any ignore rule from results in place.
// Violation file paths are relativized against each result's WorkDir before matching,
// so that relative ignore-rule globs match absolute violation paths.
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
			if result.WorkDir != "" && filepath.IsAbs(v.File) {
				if rel, err := filepath.Rel(result.WorkDir, v.File); err == nil {
					v.File = rel
				}
			}
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

// FilterViolationsByUserScope drops violations whose file is not equal to, or
// a descendant of, any path in scopes. scopes may contain absolute or
// workDir-relative paths and may name files or directories. workDir is used
// to resolve relative entries in scopes; each result's own WorkDir is used to
// resolve relative violation file paths. Returns the total number of dropped
// violations. A nil/empty scopes list is a no-op.
//
// This is the final user-intent filter applied after linters return: even
// when an underlying tool (e.g. tsc) ignores per-file args and reports the
// whole project, this restricts the surfaced violations to the paths the
// user actually asked about.
func FilterViolationsByUserScope(results []*LinterResult, workDir string, scopes []string) int {
	if len(scopes) == 0 {
		return 0
	}
	absScopes := make([]string, 0, len(scopes))
	for _, s := range scopes {
		if s == "" {
			continue
		}
		abs := s
		if !filepath.IsAbs(abs) {
			base := workDir
			if base == "" {
				base, _ = os.Getwd()
			}
			abs = filepath.Join(base, abs)
		}
		absScopes = append(absScopes, filepath.Clean(abs))
	}
	if len(absScopes) == 0 {
		return 0
	}

	dropped := 0
	for _, result := range results {
		if result == nil {
			continue
		}
		kept := result.Violations[:0]
		for _, v := range result.Violations {
			if violationMatchesAnyScope(v.File, result.WorkDir, workDir, absScopes) {
				kept = append(kept, v)
			} else {
				dropped++
			}
		}
		result.Violations = kept
	}
	return dropped
}

// violationMatchesAnyScope reports whether a violation file lies inside any
// scope. A violation file may be:
//   - absolute, or
//   - relative to the linter's own project root (result.WorkDir), or
//   - relative to the outer git root / invocation workDir.
//
// Different linters anchor their reported paths differently (tsc anchors to
// the git root; eslint/golangci anchor to their project root). Rather than
// pick one and silently drop the others, we accept a match against any
// plausible base.
func violationMatchesAnyScope(file, resultWorkDir, callerWorkDir string, scopes []string) bool {
	if file == "" {
		return false
	}
	if filepath.IsAbs(file) {
		return pathMatchesAnyScope(filepath.Clean(file), scopes)
	}
	bases := make([]string, 0, 2)
	if resultWorkDir != "" {
		bases = append(bases, resultWorkDir)
	}
	if callerWorkDir != "" && callerWorkDir != resultWorkDir {
		bases = append(bases, callerWorkDir)
	}
	if len(bases) == 0 {
		cwd, _ := os.Getwd()
		bases = append(bases, cwd)
	}
	for _, base := range bases {
		candidate := filepath.Clean(filepath.Join(base, file))
		if pathMatchesAnyScope(candidate, scopes) {
			return true
		}
	}
	return false
}

// pathMatchesAnyScope reports whether absPath equals or is a descendant of
// any scope path. Scope entries naming a file match only that exact file;
// entries naming a directory match the directory itself and everything under
// it. Non-existent scope entries are treated as directories (a typo-tolerant
// prefix match) so the filter is useful even when the user typed a path that
// does not yet exist on disk.
func pathMatchesAnyScope(absPath string, scopes []string) bool {
	for _, scope := range scopes {
		if absPath == scope {
			return true
		}
		info, err := os.Stat(scope)
		isDir := err != nil || info.IsDir()
		if !isDir {
			continue
		}
		prefix := scope + string(filepath.Separator)
		if strings.HasPrefix(absPath, prefix) {
			return true
		}
	}
	return false
}
