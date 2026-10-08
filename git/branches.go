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

// BranchRef is one local branch as `git for-each-ref refs/heads` lists it:
// everything about a branch that is read from its ref alone, without
// comparing it to another commit.
type BranchRef struct {
	Name         string
	Head         string
	Worktree     string
	LastCommitAt time.Time
}

// ListBranches lists every local branch of repo, sorted by name.
func ListBranches(repo string) ([]BranchRef, error) {
	out, err := gitOutput(repo, "for-each-ref", "--format=%(refname:short)%00%(objectname)%00%(worktreepath)%00%(committerdate:iso-strict)", "refs/heads")
	if err != nil {
		return nil, err
	}
	var branches []BranchRef
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if line == "" {
			continue
		}
		fields := strings.Split(line, "\x00")
		if len(fields) != 4 {
			return nil, fmt.Errorf("git for-each-ref in %s: unexpected line %q", repo, line)
		}
		branch := BranchRef{Name: fields[0], Head: fields[1], Worktree: fields[2]}
		if branch.LastCommitAt, err = parseCommitTime(fields[3]); err != nil {
			return nil, fmt.Errorf("branch %s in %s: %w", branch.Name, repo, err)
		}
		branches = append(branches, branch)
	}
	return branches, nil
}

// RangeCompare is head measured against base: the commits only in head
// (Ahead) and only in base (Behind), their best common ancestor, and the
// footprint of MergeBase..head. Diff is zero when head has nothing base lacks.
// It is a pure function of the two commits, so it can be cached by them.
type RangeCompare struct {
	MergeBase string
	Ahead     int
	Behind    int
	Diff      DiffStat
}

// CompareRange compares the commit head against the commit base. Both must be
// full commit hashes; neither is resolved as a ref. A hash naming no commit in
// repo is ErrCommitNotFound.
func CompareRange(repo, base, head string) (RangeCompare, error) {
	for _, sha := range []string{base, head} {
		if !IsValidCommitHash(sha) {
			return RangeCompare{}, fmt.Errorf("invalid commit hash %q", sha)
		}
		if err := requireCommit(repo, sha); err != nil {
			return RangeCompare{}, err
		}
	}
	behind, ahead, err := aheadBehind(repo, base, head)
	if err != nil {
		return RangeCompare{}, err
	}
	mergeBase, err := MergeBase(repo, base, head)
	if err != nil {
		return RangeCompare{}, fmt.Errorf("merge base of %s and %s in %s: %w", base, head, repo, err)
	}
	compare := RangeCompare{MergeBase: mergeBase, Ahead: ahead, Behind: behind}
	if ahead == 0 {
		return compare, nil
	}
	if compare.Diff, err = RangeDiffStat(repo, mergeBase, head); err != nil {
		return RangeCompare{}, err
	}
	return compare, nil
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

// RevParse resolves rev to a full commit hash.
func RevParse(repo, rev string) (string, error) {
	out, err := gitOutput(repo, "rev-parse", "--verify", "--quiet", "--end-of-options", rev+"^{commit}")
	if err != nil {
		return "", fmt.Errorf("resolve %q in %s: %w", rev, repo, err)
	}
	return strings.TrimSpace(string(out)), nil
}

// aheadBehind counts the commits only in base (behind) and only in head
// (ahead).
func aheadBehind(repo, base, head string) (behind, ahead int, err error) {
	out, err := gitOutput(repo, "rev-list", "--left-right", "--count", base+"..."+head)
	if err != nil {
		return 0, 0, err
	}
	fields := strings.Fields(string(out))
	if len(fields) != 2 {
		return 0, 0, fmt.Errorf("git rev-list --left-right --count %s...%s in %s: unexpected output %q", base, head, repo, out)
	}
	if behind, err = strconv.Atoi(fields[0]); err != nil {
		return 0, 0, fmt.Errorf("parse behind count %q: %w", fields[0], err)
	}
	if ahead, err = strconv.Atoi(fields[1]); err != nil {
		return 0, 0, fmt.Errorf("parse ahead count %q: %w", fields[1], err)
	}
	return behind, ahead, nil
}
