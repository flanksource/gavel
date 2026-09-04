package utils

import (
	"bufio"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/go-git/go-billy/v5/osfs"
	"github.com/go-git/go-git/v5/plumbing/format/gitignore"
)

type walkStopFn func(root, path string, d fs.DirEntry) (bool, error)

func FindGitRoot(dir string) string {
	dir, _ = filepath.Abs(dir)
	for {
		// A worktree root has `.git` as a directory (normal clone) OR as a file
		// (linked worktree / submodule — the file holds a `gitdir:` pointer).
		// Requiring a directory silently disabled gitignore filtering in
		// worktrees, so accept either.
		if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
}

// GitRoot resolves dir to the root of the working tree it belongs to, keeping
// dir itself when it is not inside a repository.
//
// Anything that crosses the agent/commit boundary is repo-relative: the paths
// an agent is recorded as editing, and the paths `git status` reports as dirty,
// which git anchors on the root however deep it is invoked. A run launched in a
// subdirectory that names that subdirectory as its repo therefore describes its
// own edits in a namespace nothing else uses.
func GitRoot(dir string) string {
	if root := FindGitRoot(dir); root != "" {
		return root
	}
	return dir
}

func FindNearestGoModRoot(dir string) string {
	return FindNearestProjectRoot(dir, []string{"go.mod"})
}

// FindAllProjectRoots returns every directory under root that contains one of
// the given marker filenames. Nested projects are included, while dependency
// trees and nested git repositories are pruned. Respects gitignore via
// WalkGitIgnored. Results are absolute paths in shallowest-first walk order.
func FindAllProjectRoots(root string, markers []string) []string {
	if len(markers) == 0 {
		return nil
	}
	root, _ = filepath.Abs(root)
	var roots []string
	_ = WalkGitIgnored(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil || !d.IsDir() {
			return nil
		}
		if path != root {
			if d.Name() == "node_modules" {
				return fs.SkipDir
			}
			if _, err := os.Stat(filepath.Join(path, ".git")); err == nil {
				return fs.SkipDir
			}
		}
		for _, marker := range markers {
			if info, err := os.Stat(filepath.Join(path, marker)); err == nil && !info.IsDir() {
				roots = append(roots, path)
				break
			}
		}
		return nil
	})
	return roots
}

// FindNearestProjectRoot walks up from dir looking for the first directory
// that contains any of the given marker filenames as a regular file. The
// search is bounded by the enclosing git root (inclusive) to avoid escaping
// the repository when a marker only exists further up in the filesystem.
// Returns "" when no marker is found within the git root.
func FindNearestProjectRoot(dir string, markers []string) string {
	if len(markers) == 0 {
		return ""
	}
	dir, _ = filepath.Abs(dir)
	gitRoot := FindGitRoot(dir)
	for {
		for _, marker := range markers {
			if info, err := os.Stat(filepath.Join(dir, marker)); err == nil && !info.IsDir() {
				return dir
			}
		}
		if gitRoot != "" && dir == gitRoot {
			return ""
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
}

// IsWithin reports whether path is equal to root or nested inside it. Paths
// are compared as cleaned, absolute strings; any `..` traversal that would
// escape root returns false.
func IsWithin(path, root string) bool {
	if path == root {
		return true
	}
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return false
	}
	return rel != "." && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}

func LoadIgnorePatterns(path string, domain []string) []gitignore.Pattern {
	f, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer f.Close()

	var patterns []gitignore.Pattern
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		patterns = append(patterns, gitignore.ParsePattern(line, domain))
	}
	return patterns
}

// WalkGitIgnored walks a directory tree like filepath.WalkDir but skips entries
// matched by .gitignore patterns. The allowList contains entry names that should
// never be skipped even if gitignored (e.g. ".todos", ".codex").
func WalkGitIgnored(root string, fn fs.WalkDirFunc, allowList ...string) error {
	return walkGitIgnored(root, fn, nil, allowList...)
}

// WalkGitIgnoredBounded walks a directory tree like WalkGitIgnored but stops
// descending into nested project roots. Only descendant directories are treated
// as boundaries; the starting root itself is always traversed.
func WalkGitIgnoredBounded(root string, fn fs.WalkDirFunc, allowList ...string) error {
	return walkGitIgnored(root, fn, stopAtNestedProjectRoot, allowList...)
}

func walkGitIgnored(root string, fn fs.WalkDirFunc, stop walkStopFn, allowList ...string) error {
	root, _ = filepath.Abs(root)
	gitRoot := FindGitRoot(root)
	if gitRoot == "" {
		return filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return fn(path, d, err)
			}
			if d.Name() == ".git" && d.IsDir() {
				return fs.SkipDir
			}
			if stop != nil {
				skip, err := stop(root, path, d)
				if err != nil {
					return fn(path, d, err)
				}
				if skip {
					return fs.SkipDir
				}
			}
			return fn(path, d, err)
		})
	}

	allowed := make(map[string]bool, len(allowList))
	for _, name := range allowList {
		allowed[name] = true
	}

	var patterns []gitignore.Pattern
	patterns = append(patterns, LoadIgnorePatterns(filepath.Join(gitRoot, ".git", "info", "exclude"), nil)...)
	rel, _ := filepath.Rel(gitRoot, root)
	patterns = append(patterns, LoadIgnorePatterns(filepath.Join(gitRoot, ".gitignore"), nil)...)
	if rel != "." {
		parts := strings.Split(filepath.ToSlash(rel), "/")
		for i := range parts {
			d := filepath.Join(gitRoot, filepath.Join(parts[:i+1]...))
			domain := parts[:i+1]
			patterns = append(patterns, LoadIgnorePatterns(filepath.Join(d, ".gitignore"), domain)...)
		}
	}

	return filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return fn(path, d, err)
		}

		if d.Name() == ".git" && d.IsDir() {
			return fs.SkipDir
		}

		// Load .gitignore when entering a directory
		if d.IsDir() && path != root {
			dirRel, _ := filepath.Rel(gitRoot, path)
			domain := strings.Split(filepath.ToSlash(dirRel), "/")
			patterns = append(patterns, LoadIgnorePatterns(filepath.Join(path, ".gitignore"), domain)...)
		}

		// Check if this entry or any ancestor is in the allowList
		pathRel, _ := filepath.Rel(gitRoot, path)
		pathParts := strings.Split(filepath.ToSlash(pathRel), "/")
		for _, part := range pathParts {
			if allowed[part] {
				return fn(path, d, err)
			}
		}

		// Check if this path is gitignored
		if gitignore.NewMatcher(patterns).Match(pathParts, d.IsDir()) {
			if d.IsDir() {
				return fs.SkipDir
			}
			return nil
		}

		if stop != nil {
			skip, err := stop(root, path, d)
			if err != nil {
				return fn(path, d, err)
			}
			if skip {
				return fs.SkipDir
			}
		}

		return fn(path, d, err)
	})
}

func stopAtNestedProjectRoot(root, path string, d fs.DirEntry) (bool, error) {
	if !d.IsDir() || path == root {
		return false, nil
	}

	// `.git` is a directory in a normal clone and a file holding a `gitdir:`
	// pointer in a linked worktree or submodule. Both are separate checkouts,
	// so both bound the walk.
	if _, err := os.Stat(filepath.Join(path, ".git")); err == nil {
		return true, nil
	} else if !os.IsNotExist(err) {
		return false, err
	}

	if _, err := os.Stat(filepath.Join(path, "go.mod")); err == nil {
		return true, nil
	} else if err != nil && !os.IsNotExist(err) {
		return false, err
	}

	return false, nil
}

// FilterGitIgnored returns the subset of absolute paths that are not matched
// by .gitignore patterns. Uses go-git's ReadPatterns to recursively load all
// .gitignore files. If no git root is found, all paths are returned.
func FilterGitIgnored(paths []string, dir string) []string {
	kept, _ := PartitionGitIgnored(paths, dir)
	return kept
}

// PartitionGitIgnored splits absolute paths into those NOT matched by
// .gitignore (kept) and those matched (ignored). It honors !-negation via
// go-git's matcher. Fail-open: if no git root is found or no patterns load,
// every path is kept and ignored is empty.
func PartitionGitIgnored(paths []string, dir string) (kept, ignored []string) {
	dir, _ = filepath.Abs(dir)
	gitRoot := FindGitRoot(dir)
	if gitRoot == "" {
		return paths, nil
	}

	fs := osfs.New(gitRoot)
	patterns, err := gitignore.ReadPatterns(fs, nil)
	if err != nil || len(patterns) == 0 {
		return paths, nil
	}

	matcher := gitignore.NewMatcher(patterns)
	for _, p := range paths {
		rel, err := filepath.Rel(gitRoot, p)
		if err != nil {
			kept = append(kept, p)
			continue
		}
		parts := strings.Split(filepath.ToSlash(rel), "/")
		if matcher.Match(parts, false) {
			ignored = append(ignored, p)
		} else {
			kept = append(kept, p)
		}
	}
	return kept, ignored
}

// GlobFilesBounded resolves doublestar patterns to absolute file paths using
// WalkGitIgnoredBounded, so matches never come from a gitignored directory or a
// nested checkout (scratch worktrees hold a full copy of the repo and would
// otherwise double-count every match). Relative patterns are matched against
// root-relative slash paths; patterns that are absolute or escape root are
// globbed directly, since the bounded walk never leaves root. Results are
// deduplicated and sorted.
func GlobFilesBounded(root string, patterns []string) ([]string, error) {
	if len(patterns) == 0 {
		return nil, nil
	}
	root, _ = filepath.Abs(root)

	var relative, external []string
	for _, pattern := range patterns {
		if filepath.IsAbs(pattern) {
			external = append(external, pattern)
			continue
		}
		// Patterns are compared against root-relative slash paths, so they need
		// the same cleaning filepath.Join would apply — "./pkg/**/*.md"
		// otherwise carries a literal "./" segment and matches nothing.
		cleaned := path.Clean(filepath.ToSlash(pattern))
		if cleaned == ".." || strings.HasPrefix(cleaned, "../") {
			external = append(external, filepath.Join(root, pattern))
			continue
		}
		if !doublestar.ValidatePattern(cleaned) {
			return nil, fmt.Errorf("invalid glob %q", pattern)
		}
		relative = append(relative, cleaned)
	}

	seen := map[string]bool{}
	var files []string
	add := func(file string) {
		if seen[file] {
			return
		}
		seen[file] = true
		files = append(files, file)
	}

	if len(relative) > 0 {
		err := WalkGitIgnoredBounded(root, func(file string, d fs.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				return err
			}
			rel, relErr := filepath.Rel(root, file)
			if relErr != nil {
				return nil
			}
			rel = filepath.ToSlash(rel)
			for _, pattern := range relative {
				if match, _ := doublestar.Match(pattern, rel); match {
					add(file)
					return nil
				}
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	}

	for _, pattern := range external {
		matches, err := doublestar.FilepathGlob(pattern)
		if err != nil {
			return nil, fmt.Errorf("invalid glob %q: %w", pattern, err)
		}
		for _, match := range matches {
			info, err := os.Stat(match)
			if err != nil {
				return nil, fmt.Errorf("stat %s: %w", match, err)
			}
			if !info.IsDir() {
				add(match)
			}
		}
	}

	sort.Strings(files)
	return files, nil
}

// GlobWalkGitIgnored walks root like WalkGitIgnored — pruning .gitignored
// directories and files before they are visited — and invokes fn for every
// non-ignored file whose root-relative slash path matches one of patterns
// (doublestar semantics, so "**/*.go" matches at any depth). extraIgnore holds
// additional gitignore-syntax patterns (e.g. the .gavel.yaml gitignore list)
// applied on top of the repository's .gitignore; a directory matching one of
// them is pruned, never descended. fn may return fs.SkipDir/fs.SkipAll to stop
// early.
func GlobWalkGitIgnored(root string, patterns, extraIgnore []string, fn func(rel string, d fs.DirEntry) error) error {
	root, _ = filepath.Abs(root)

	var extra gitignore.Matcher
	if len(extraIgnore) > 0 {
		ps := make([]gitignore.Pattern, 0, len(extraIgnore))
		for _, p := range extraIgnore {
			ps = append(ps, gitignore.ParsePattern(p, nil))
		}
		extra = gitignore.NewMatcher(ps)
	}

	return WalkGitIgnored(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil || path == root {
			return err
		}
		rel, relErr := filepath.Rel(root, path)
		if relErr != nil {
			return nil
		}
		rel = filepath.ToSlash(rel)

		if extra != nil && extra.Match(strings.Split(rel, "/"), d.IsDir()) {
			if d.IsDir() {
				return fs.SkipDir
			}
			return nil
		}
		if d.IsDir() {
			return nil
		}
		for _, pat := range patterns {
			if matched, _ := doublestar.Match(pat, rel); matched {
				return fn(rel, d)
			}
		}
		return nil
	})
}

// AnyGlobMatchGitIgnored reports whether any non-ignored file under root matches
// one of patterns, honoring .gitignore and extraIgnore the same way as
// GlobWalkGitIgnored. It stops at the first match.
func AnyGlobMatchGitIgnored(root string, patterns, extraIgnore []string) bool {
	if len(patterns) == 0 {
		return false
	}
	found := false
	_ = GlobWalkGitIgnored(root, patterns, extraIgnore, func(rel string, d fs.DirEntry) error {
		found = true
		return fs.SkipAll
	})
	return found
}
