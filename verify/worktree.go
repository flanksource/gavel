package verify

import (
	"fmt"
	"os/exec"
	"strings"
)

// worktreeChanges returns the repo's current uncommitted changes as a
// final-state signal for black-box verification: the raw `git status` porcelain
// text (for {{worktree}}) and the list of changed paths (for {{changedFiles}}).
// It is computed lazily — only when the verify template references those
// variables — so a git failure is surfaced loudly (CW-2) rather than silently
// emptying the prompt for templates that never asked for it.
func worktreeChanges(workDir string) (files []string, worktree string, err error) {
	cmd := exec.Command("git", "status", "--porcelain")
	if workDir != "" {
		cmd.Dir = workDir
	}
	out, err := cmd.Output()
	if err != nil {
		return nil, "", fmt.Errorf("git status in %q: %w", workDir, err)
	}
	worktree = strings.TrimRight(string(out), "\n")
	if worktree == "" {
		return nil, "", nil
	}
	for _, line := range strings.Split(worktree, "\n") {
		if len(line) < 4 {
			continue
		}
		// porcelain XY prefix + path; renames render as "old -> new".
		path := strings.TrimSpace(line[3:])
		if i := strings.Index(path, " -> "); i >= 0 {
			path = path[i+len(" -> "):]
		}
		files = append(files, path)
	}
	return files, worktree, nil
}

// templateWants reports whether the verify template references a Handlebars
// variable, so its (possibly expensive) value is computed only on demand.
func templateWants(template, variable string) bool {
	return strings.Contains(template, variable)
}
