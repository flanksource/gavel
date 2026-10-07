// Package gitstate computes a project repository's git state relative to its
// base branch — worktrees, their uncommitted changes and the unmerged branches —
// and memoizes it per project directory (see Service).
package gitstate

import (
	"context"
	"fmt"
	"strings"
	"time"

	gavelgit "github.com/flanksource/gavel/git"
	"github.com/flanksource/gavel/github/prcreate"
	"github.com/flanksource/gavel/status"
)

// Changes counts a worktree's uncommitted files by state, plus their line
// counts.
type Changes struct {
	Staged    int `json:"staged"`
	Unstaged  int `json:"unstaged"`
	Both      int `json:"both"`
	Untracked int `json:"untracked"`
	Conflict  int `json:"conflict"`
	Adds      int `json:"adds"`
	Dels      int `json:"dels"`
}

// Files is the number of uncommitted files: every file falls in exactly one
// state.
func (c Changes) Files() int {
	return c.Staged + c.Unstaged + c.Both + c.Untracked + c.Conflict
}

type Worktree struct {
	gavelgit.Worktree
	Changes Changes `json:"changes"`
	// Ahead counts the worktree branch's commits not in base; 0 for the base
	// branch itself and for a detached worktree.
	Ahead int `json:"ahead"`
	// LastCommitAt is the committer date of Head, zero for a prunable worktree.
	LastCommitAt time.Time `json:"lastCommitAt"`
	// TouchedAt is the newest modification time among the worktree's
	// uncommitted files; nil when it has none.
	TouchedAt *time.Time `json:"touchedAt,omitempty"`
}

// State is the git state of a project's repository relative to its base
// branch. CurrentBranch and BaseCheckedOut describe the primary checkout, which
// is where branches merge into. ComputedAt is when the computation started, so
// a reader can tell how old a memoized state is.
type State struct {
	Base           string                `json:"base"`
	CurrentBranch  string                `json:"currentBranch"`
	BaseCheckedOut bool                  `json:"baseCheckedOut"`
	Worktrees      []Worktree            `json:"worktrees"`
	Branches       []gavelgit.BranchInfo `json:"branches"`
	ComputedAt     time.Time             `json:"computedAt"`
}

// Summary is one project's row of /api/projects/git-summary: the unmerged and
// uncommitted line counts plus the linked worktree count (the primary checkout
// excluded) and unmerged branch count. Error is set, with the numbers left at 0,
// when the project could not be inspected.
type Summary struct {
	Name      string `json:"name"`
	Base      string `json:"base"`
	Adds      int    `json:"adds"`
	Dels      int    `json:"dels"`
	Worktrees int    `json:"worktrees"`
	Branches  int    `json:"branches"`
	Error     string `json:"error,omitempty"`
}

func (state State) Summary(name string) Summary {
	summary := Summary{Name: name, Base: state.Base, Branches: len(state.Branches)}
	for _, wt := range state.Worktrees {
		if !wt.Primary {
			summary.Worktrees++
		}
		summary.Adds += wt.Changes.Adds
		summary.Dels += wt.Changes.Dels
	}
	for _, branch := range state.Branches {
		summary.Adds += branch.Diff.Adds
		summary.Dels += branch.Diff.Dels
	}
	return summary
}

// Primary is the repository's primary checkout, listed first by git.
func (state State) Primary() (Worktree, bool) {
	if len(state.Worktrees) == 0 || !state.Worktrees[0].Primary {
		return Worktree{}, false
	}
	return state.Worktrees[0], true
}

// Base is the local branch project work merges into: the PR base (.gavel.yaml
// pr.base, else origin/main) without its remote prefix.
func Base(dir string) (string, error) {
	ref, err := prcreate.ResolveBase(dir, "")
	if err != nil {
		return "", err
	}
	return strings.TrimPrefix(strings.TrimPrefix(ref, "refs/heads/"), "origin/"), nil
}

// Compute reads the git state of the repository at dir live.
func Compute(ctx context.Context, dir string) (State, error) {
	started := time.Now().UTC()
	base, err := Base(dir)
	if err != nil {
		return State{}, err
	}
	worktrees, err := gavelgit.ListWorktrees(dir)
	if err != nil {
		return State{}, err
	}
	branches, err := gavelgit.UnmergedBranches(dir, base)
	if err != nil {
		return State{}, err
	}
	ahead := make(map[string]int, len(branches))
	for _, branch := range branches {
		ahead[branch.Name] = branch.Ahead
	}
	state := State{
		Base: base, CurrentBranch: worktrees[0].Branch, BaseCheckedOut: worktrees[0].Branch == base,
		Worktrees: make([]Worktree, 0, len(worktrees)), Branches: append([]gavelgit.BranchInfo{}, branches...),
		ComputedAt: started,
	}
	for _, wt := range worktrees {
		view := Worktree{Worktree: wt, Ahead: ahead[wt.Branch]}
		if !wt.Prunable {
			if view.Changes, view.TouchedAt, err = WorktreeChanges(ctx, wt.Path); err != nil {
				return State{}, err
			}
			if view.LastCommitAt, err = gavelgit.CommitTime(wt.Path, wt.Head); err != nil {
				return State{}, err
			}
		}
		state.Worktrees = append(state.Worktrees, view)
	}
	return state, nil
}

// WorktreeChanges counts the uncommitted changes of the worktree at path and
// returns the newest modification time among them, in UTC; nil when no changed
// file exists on disk (none changed, or all deleted).
func WorktreeChanges(ctx context.Context, path string) (Changes, *time.Time, error) {
	result, err := status.GatherBase(path, status.Options{NoRepomap: true, NoResults: true, Context: ctx})
	if err != nil {
		return Changes{}, nil, fmt.Errorf("gather status of worktree %s: %w", path, err)
	}
	var touched *time.Time
	for _, file := range result.Files {
		if at := file.ModifiedAt.UTC(); !file.ModifiedAt.IsZero() && (touched == nil || at.After(*touched)) {
			touched = &at
		}
	}
	counts := result.Counts()
	return Changes{
		Staged: counts.Staged, Unstaged: counts.Unstaged, Both: counts.Both, Untracked: counts.Untracked,
		Conflict: counts.Conflict, Adds: counts.Adds, Dels: counts.Dels,
	}, touched, nil
}
