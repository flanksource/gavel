package testrunner

import (
	"path/filepath"
	"strings"
)

// applyIgnorePatterns returns the subset of pkgs that does not match any
// pattern in ignore. Patterns are repo-relative package paths. A bare
// directory pattern matches that directory and every package below it
// ("./bench" hides "./bench" and "./bench/sub"). Patterns may also use the
// Go-style "./bench/..." or "./bench/**" suffixes; both are treated as
// recursive directory matches. Empty ignore returns pkgs unchanged.
func applyIgnorePatterns(pkgs, ignore []string) []string {
	if len(ignore) == 0 || len(pkgs) == 0 {
		return pkgs
	}
	normalized := make([]string, 0, len(ignore))
	for _, p := range ignore {
		if n := normalizeIgnorePattern(p); n != "" {
			normalized = append(normalized, n)
		}
	}
	if len(normalized) == 0 {
		return pkgs
	}
	out := pkgs[:0:0]
	for _, pkg := range pkgs {
		if !ignoreMatches(pkg, normalized) {
			out = append(out, pkg)
		}
	}
	return out
}

// normalizeIgnorePattern strips trailing /... and /** suffixes and collapses
// the path through filepath.Clean. The returned string is a directory
// prefix used as both an exact match and a path-prefix match.
func normalizeIgnorePattern(p string) string {
	p = strings.TrimSpace(p)
	if p == "" {
		return ""
	}
	for _, suffix := range []string{"/...", "/**", "/*"} {
		p = strings.TrimSuffix(p, suffix)
	}
	p = filepath.ToSlash(filepath.Clean(p))
	if p == "." {
		return ""
	}
	if !strings.HasPrefix(p, "./") && !filepath.IsAbs(p) {
		p = "./" + p
	}
	return p
}

// ignoreMatches reports whether pkg matches any normalized ignore prefix.
// pkg is expected to look like "./foo/bar" (from getRelativePath). Match is
// either equality or a path-prefix check ("./bench" matches "./bench/sub").
func ignoreMatches(pkg string, normalized []string) bool {
	clean := filepath.ToSlash(filepath.Clean(pkg))
	if !strings.HasPrefix(clean, "./") && !filepath.IsAbs(clean) {
		clean = "./" + clean
	}
	for _, pat := range normalized {
		if clean == pat {
			return true
		}
		if strings.HasPrefix(clean, pat+"/") {
			return true
		}
	}
	return false
}
