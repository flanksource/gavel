package git

import (
	"fmt"
	"strings"
)

// Worktree is one entry of `git worktree list`: the primary checkout or a
// linked worktree. Branch is the short branch name, empty when Detached.
// Prunable worktrees are registered but their directory is gone.
type Worktree struct {
	Path     string `json:"path"`
	Branch   string `json:"branch"`
	Head     string `json:"head"`
	Primary  bool   `json:"primary"`
	Detached bool   `json:"detached"`
	Prunable bool   `json:"prunable"`
}

// ListWorktrees returns every worktree of the repository containing repo, the
// primary checkout first, parsed from `git worktree list --porcelain`.
func ListWorktrees(repo string) ([]Worktree, error) {
	out, err := gitOutput(repo, "worktree", "list", "--porcelain")
	if err != nil {
		return nil, err
	}
	worktrees, err := parseWorktreeList(string(out))
	if err != nil {
		return nil, fmt.Errorf("git worktree list in %s: %w", repo, err)
	}
	if len(worktrees) == 0 {
		return nil, fmt.Errorf("git worktree list in %s: no worktrees listed", repo)
	}
	worktrees[0].Primary = true
	return worktrees, nil
}

func parseWorktreeList(out string) ([]Worktree, error) {
	var worktrees []Worktree
	for _, block := range strings.Split(strings.TrimSpace(out), "\n\n") {
		if strings.TrimSpace(block) == "" {
			continue
		}
		var wt Worktree
		for _, line := range strings.Split(block, "\n") {
			key, value, _ := strings.Cut(line, " ")
			switch key {
			case "worktree":
				wt.Path = value
			case "HEAD":
				wt.Head = value
			case "branch":
				wt.Branch = strings.TrimPrefix(value, "refs/heads/")
			case "detached":
				wt.Detached = true
			case "prunable":
				wt.Prunable = true
			}
		}
		if wt.Path == "" {
			return nil, fmt.Errorf("entry without a worktree path: %q", block)
		}
		worktrees = append(worktrees, wt)
	}
	return worktrees, nil
}
