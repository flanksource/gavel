package git

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// BranchInfo is a local branch with commits not yet in the base branch. Ahead
// and Behind count commits relative to base; Worktree is the path of the
// worktree that has it checked out, empty when none does. Diff is the
// footprint of merge-base(base, branch)..branch. LastCommitAt is the
// committer date of Head, in UTC.
type BranchInfo struct {
	Name         string    `json:"name"`
	Head         string    `json:"head"`
	Ahead        int       `json:"ahead"`
	Behind       int       `json:"behind"`
	Worktree     string    `json:"worktree"`
	Diff         DiffStat  `json:"diff"`
	LastCommitAt time.Time `json:"lastCommitAt"`
}

// UnmergedBranches lists the local branches of repo, other than base, that
// have at least one commit not reachable from base. base is a local branch
// name; a base that does not exist is an error.
func UnmergedBranches(repo, base string) ([]BranchInfo, error) {
	if _, err := gitOutput(repo, "rev-parse", "--verify", "--quiet", "refs/heads/"+base); err != nil {
		return nil, fmt.Errorf("base branch %q not found in %s: %w", base, repo, err)
	}
	out, err := gitOutput(repo, "for-each-ref", "--format=%(refname:short)%00%(objectname)%00%(worktreepath)%00%(committerdate:iso-strict)", "refs/heads")
	if err != nil {
		return nil, err
	}
	var branches []BranchInfo
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		fields := strings.Split(line, "\x00")
		if len(fields) != 4 {
			return nil, fmt.Errorf("git for-each-ref in %s: unexpected line %q", repo, line)
		}
		if fields[0] == base {
			continue
		}
		branch := BranchInfo{Name: fields[0], Head: fields[1], Worktree: fields[2]}
		if branch.LastCommitAt, err = parseCommitTime(fields[3]); err != nil {
			return nil, fmt.Errorf("branch %s in %s: %w", branch.Name, repo, err)
		}
		if branch.Behind, branch.Ahead, err = aheadBehind(repo, base, branch.Name); err != nil {
			return nil, err
		}
		if branch.Ahead == 0 {
			continue
		}
		if branch.Diff, err = branchDiffStat(repo, base, branch.Head); err != nil {
			return nil, err
		}
		branches = append(branches, branch)
	}
	return branches, nil
}

// CommitTime is the committer date of rev in repo, in UTC.
func CommitTime(repo, rev string) (time.Time, error) {
	out, err := gitOutput(repo, "log", "-1", "--format=%cI", rev, "--")
	if err != nil {
		return time.Time{}, fmt.Errorf("commit time of %s in %s: %w", rev, repo, err)
	}
	return parseCommitTime(strings.TrimSpace(string(out)))
}

func parseCommitTime(iso string) (time.Time, error) {
	at, err := time.Parse(time.RFC3339, iso)
	if err != nil {
		return time.Time{}, fmt.Errorf("parse committer date %q: %w", iso, err)
	}
	return at.UTC(), nil
}

// MergeBase is the full sha of the best common ancestor of a and b.
func MergeBase(repo, a, b string) (string, error) {
	out, err := gitOutput(repo, "merge-base", a, b)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func branchDiffStat(repo, base, head string) (DiffStat, error) {
	mergeBase, err := MergeBase(repo, "refs/heads/"+base, head)
	if err != nil {
		return DiffStat{}, err
	}
	return RangeDiffStat(repo, mergeBase, head)
}

// aheadBehind counts the commits only in base (behind) and only in branch
// (ahead).
func aheadBehind(repo, base, branch string) (behind, ahead int, err error) {
	out, err := gitOutput(repo, "rev-list", "--left-right", "--count", "refs/heads/"+base+"...refs/heads/"+branch)
	if err != nil {
		return 0, 0, err
	}
	fields := strings.Fields(string(out))
	if len(fields) != 2 {
		return 0, 0, fmt.Errorf("git rev-list --left-right --count %s...%s in %s: unexpected output %q", base, branch, repo, out)
	}
	if behind, err = strconv.Atoi(fields[0]); err != nil {
		return 0, 0, fmt.Errorf("parse behind count %q: %w", fields[0], err)
	}
	if ahead, err = strconv.Atoi(fields[1]); err != nil {
		return 0, 0, fmt.Errorf("parse ahead count %q: %w", fields[1], err)
	}
	return behind, ahead, nil
}
