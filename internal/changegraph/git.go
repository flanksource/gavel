// Package changegraph resolves the set of packages affected by a file-level
// change, using git as the source of changes and `go list` as the dependency
// graph. See /Users/moshe/.claude/plans/shimmying-squishing-otter.md for the
// design context.
package changegraph

import (
	"bufio"
	"fmt"
	"os/exec"
	"sort"
	"strings"
)

// FileSet is a set of workdir-relative file paths (forward-slash separated).
// It is the input to Graph.AffectedPackages.
type FileSet map[string]struct{}

// NewFileSet creates an empty FileSet.
func NewFileSet() FileSet { return FileSet{} }

// Add inserts a path, normalizing to forward slashes.
func (fs FileSet) Add(path string) {
	path = strings.TrimSpace(path)
	if path == "" {
		return
	}
	fs[filepathToSlash(path)] = struct{}{}
}

// Union merges other into fs in place.
func (fs FileSet) Union(other FileSet) {
	for k := range other {
		fs[k] = struct{}{}
	}
}

// Has reports whether path is in the set.
func (fs FileSet) Has(path string) bool {
	_, ok := fs[filepathToSlash(path)]
	return ok
}

// Len returns the number of paths.
func (fs FileSet) Len() int { return len(fs) }

// Sorted returns the set as a sorted slice. Useful for deterministic output.
func (fs FileSet) Sorted() []string {
	out := make([]string, 0, len(fs))
	for k := range fs {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// DiffOptions selects which git views contribute to the resulting FileSet.
// All enabled sources are unioned together.
type DiffOptions struct {
	// Since is a ref. When set, include files changed from
	// merge-base(HEAD, Since)..HEAD. Empty means skip.
	Since string

	// IncludeStaged adds files from `git diff --cached --name-only`.
	IncludeStaged bool

	// IncludeUnstaged adds files from `git diff --name-only` (working tree
	// vs index).
	IncludeUnstaged bool

	// IncludeUntracked adds files from `git ls-files --others --exclude-standard`.
	IncludeUntracked bool
}

// ComputeFileSet runs git in workDir to produce the FileSet described by opts.
// All paths are returned workdir-relative with forward slashes. A missing git
// binary or non-repo workdir yields an error.
func ComputeFileSet(workDir string, opts DiffOptions) (FileSet, error) {
	fs := NewFileSet()

	if opts.Since != "" {
		base, err := mergeBase(workDir, "HEAD", opts.Since)
		if err != nil {
			return nil, fmt.Errorf("git merge-base HEAD %s: %w", opts.Since, err)
		}
		files, err := gitLines(workDir, "diff", "--name-only", base+"...HEAD")
		if err != nil {
			return nil, fmt.Errorf("git diff %s...HEAD: %w", base, err)
		}
		for _, f := range files {
			fs.Add(f)
		}
	}

	if opts.IncludeStaged {
		files, err := gitLines(workDir, "diff", "--cached", "--name-only")
		if err != nil {
			return nil, fmt.Errorf("git diff --cached: %w", err)
		}
		for _, f := range files {
			fs.Add(f)
		}
	}

	if opts.IncludeUnstaged {
		files, err := gitLines(workDir, "diff", "--name-only")
		if err != nil {
			return nil, fmt.Errorf("git diff: %w", err)
		}
		for _, f := range files {
			fs.Add(f)
		}
	}

	if opts.IncludeUntracked {
		files, err := gitLines(workDir, "ls-files", "--others", "--exclude-standard")
		if err != nil {
			return nil, fmt.Errorf("git ls-files --others: %w", err)
		}
		for _, f := range files {
			fs.Add(f)
		}
	}

	return fs, nil
}

// mergeBase resolves `git merge-base a b` to a commit sha.
func mergeBase(workDir, a, b string) (string, error) {
	cmd := exec.Command("git", "merge-base", a, b)
	cmd.Dir = workDir
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// gitLines runs `git <args...>` in workDir and returns one trimmed line per
// output line, dropping empties.
func gitLines(workDir string, args ...string) ([]string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = workDir
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	if err := cmd.Start(); err != nil {
		return nil, err
	}

	var lines []string
	scanner := bufio.NewScanner(stdout)
	// Allow long paths; default 64K is fine for any realistic file path set
	// but symlink chains could exceed it.
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			lines = append(lines, line)
		}
	}
	if err := scanner.Err(); err != nil {
		_ = cmd.Wait()
		return nil, err
	}
	if err := cmd.Wait(); err != nil {
		return nil, err
	}
	return lines, nil
}

// filepathToSlash is filepath.ToSlash inlined here to avoid a filepath import
// for a single call — the function is identical on every platform.
func filepathToSlash(p string) string {
	return strings.ReplaceAll(p, "\\", "/")
}
