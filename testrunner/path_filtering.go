package testrunner

import (
	"path/filepath"
	"strings"
)

// filterIgnoredGroups drops testGroups whose workDir is matched by an
// ignore pattern relative to baseWorkDir. This is needed because nested-go-mod
// expansion spawns independent groups (each with its own WorkDir) before
// per-package discovery happens, so package-level filtering would never see
// them. baseWorkDir is the original (top-level) workdir; patterns are
// resolved against it.
func filterIgnoredGroups(baseWorkDir string, groups []testGroup, ignore []string) []testGroup {
	if len(ignore) == 0 || len(groups) == 0 {
		return groups
	}
	normalized := make([]string, 0, len(ignore))
	for _, p := range ignore {
		if n := normalizeIgnorePattern(p); n != "" {
			normalized = append(normalized, n)
		}
	}
	if len(normalized) == 0 {
		return groups
	}
	baseAbs, err := filepath.Abs(baseWorkDir)
	if err != nil {
		return groups
	}
	out := groups[:0:0]
	for _, g := range groups {
		gAbs, err := filepath.Abs(g.workDir)
		if err != nil {
			out = append(out, g)
			continue
		}
		rel, err := filepath.Rel(baseAbs, gAbs)
		if err != nil || strings.HasPrefix(rel, "..") {
			// Group lives outside the base workdir — leave it alone.
			out = append(out, g)
			continue
		}
		if rel == "." {
			out = append(out, g)
			continue
		}
		pkgLike := "./" + filepath.ToSlash(rel)
		if !ignoreMatches(pkgLike, normalized) {
			out = append(out, g)
		}
	}
	return out
}

// expandRecursiveWildcards translates Go-style recursive wildcards
// ("./...", "...") into the empty starting-path convention, which downstream
// code already interprets as "discover from WorkDir recursively". Any path
// of the form "./<dir>/..." is rewritten to "./<dir>" so the directory is
// preserved but the wildcard suffix is dropped (DiscoverPackages walks
// recursively by default).
//
// Returns a new slice; input is not mutated. An input that contains only
// "./..." (or equivalents) becomes nil.
func expandRecursiveWildcards(paths []string) []string {
	if len(paths) == 0 {
		return paths
	}
	out := make([]string, 0, len(paths))
	for _, p := range paths {
		switch p {
		case "./...", "...", "./", ".":
			// Equivalent to "no starting path" — let the orchestrator
			// discover from WorkDir.
			continue
		}
		p = strings.TrimSuffix(p, "/...")
		out = append(out, p)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

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
