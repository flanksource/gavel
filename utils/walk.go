package utils

import (
	"bufio"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-billy/v5/osfs"
	"github.com/go-git/go-git/v5/plumbing/format/gitignore"
)

func FindGitRoot(dir string) string {
	dir, _ = filepath.Abs(dir)
	for {
		if info, err := os.Stat(filepath.Join(dir, ".git")); err == nil && info.IsDir() {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
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
	root, _ = filepath.Abs(root)
	gitRoot := FindGitRoot(root)
	if gitRoot == "" {
		return filepath.WalkDir(root, fn)
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

		return fn(path, d, err)
	})
}

// FilterGitIgnored returns the subset of absolute paths that are not matched
// by .gitignore patterns. Uses go-git's ReadPatterns to recursively load all
// .gitignore files. If no git root is found, all paths are returned.
func FilterGitIgnored(paths []string, dir string) []string {
	dir, _ = filepath.Abs(dir)
	gitRoot := FindGitRoot(dir)
	if gitRoot == "" {
		return paths
	}

	fs := osfs.New(gitRoot)
	patterns, err := gitignore.ReadPatterns(fs, nil)
	if err != nil || len(patterns) == 0 {
		return paths
	}

	matcher := gitignore.NewMatcher(patterns)
	var result []string
	for _, p := range paths {
		rel, err := filepath.Rel(gitRoot, p)
		if err != nil {
			result = append(result, p)
			continue
		}
		parts := strings.Split(filepath.ToSlash(rel), "/")
		if !matcher.Match(parts, false) {
			result = append(result, p)
		}
	}
	return result
}
