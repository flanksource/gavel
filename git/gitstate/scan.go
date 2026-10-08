package gitstate

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	gavelgit "github.com/flanksource/gavel/git"
	"github.com/google/uuid"
)

// pollGitArgs are the global options of every git command a background scan
// runs. --no-optional-locks keeps a poller from taking the index lock a
// person's own git command needs; the file monitor and untracked cache make a
// status of an unchanged worktree cost a few milliseconds.
var pollGitArgs = []string{"--no-optional-locks", "-c", "core.fsmonitor=true", "-c", "core.untrackedCache=true"}

// repoLocation is where a repository lives: its primary checkout and the
// directory holding the refs and worktree metadata all its worktrees share.
type repoLocation struct {
	RootDir   string
	CommonDir string
}

// locateRepo resolves the repository containing dir.
func locateRepo(ctx context.Context, dir string) (repoLocation, error) {
	worktrees, err := gavelgit.ListWorktrees(dir)
	if err != nil {
		return repoLocation{}, err
	}
	root, err := filepath.EvalSymlinks(worktrees[0].Path)
	if err != nil {
		return repoLocation{}, fmt.Errorf("resolve primary checkout of %s: %w", dir, err)
	}
	common, err := runGit(ctx, root, "rev-parse", "--path-format=absolute", "--git-common-dir")
	if err != nil {
		return repoLocation{}, err
	}
	commonDir, err := filepath.EvalSymlinks(strings.TrimSpace(string(common)))
	if err != nil {
		return repoLocation{}, fmt.Errorf("resolve git common dir of %s: %w", root, err)
	}
	return repoLocation{RootDir: root, CommonDir: commonDir}, nil
}

// refsResult is what a ref scan found: the live worktrees, and whether the
// stored rows changed.
type refsResult struct {
	Changed   bool
	Worktrees []gavelgit.Worktree
}

// scanRefs brings the repo's ref-level rows up to date. A fingerprint equal
// to the stored one ends the scan without comparing anything; otherwise only
// the (base, head) pairs that have no cached comparison yet run git.
func scanRefs(ctx context.Context, store *Store, repoID uuid.UUID, root string) (refsResult, error) {
	base, err := Base(root)
	if err != nil {
		return refsResult{}, err
	}
	worktrees, err := gavelgit.ListWorktrees(root)
	if err != nil {
		return refsResult{}, err
	}
	branches, err := gavelgit.ListBranches(root)
	if err != nil {
		return refsResult{}, err
	}
	fingerprint := refsFingerprint(base, worktrees, branches)
	stored, err := store.RefsFingerprint(ctx, repoID)
	if err != nil {
		return refsResult{}, err
	}
	if fingerprint == stored {
		return refsResult{Worktrees: worktrees}, store.MarkRefsScanned(ctx, repoID)
	}

	update := RefsUpdate{
		Fingerprint: fingerprint, BaseBranch: base, CurrentBranch: worktrees[0].Branch,
		BaseCheckedOut: worktrees[0].Branch == base, Branches: branches,
	}
	byName := make(map[string]gavelgit.BranchRef, len(branches))
	for _, branch := range branches {
		byName[branch.Name] = branch
	}
	baseRef, ok := byName[base]
	if !ok {
		return refsResult{}, fmt.Errorf("base branch %q not found in %s", base, root)
	}
	update.BaseSHA = baseRef.Head
	if err := saveMissingRanges(ctx, store, repoID, root, baseRef.Head, branches, base); err != nil {
		return refsResult{}, err
	}
	for _, wt := range worktrees {
		ref := WorktreeRef{Worktree: wt}
		switch {
		case wt.Prunable || wt.Head == "":
		case byName[wt.Branch].Head == wt.Head:
			ref.LastCommitAt = byName[wt.Branch].LastCommitAt
		default:
			if ref.LastCommitAt, err = gavelgit.CommitTime(wt.Path, wt.Head); err != nil {
				return refsResult{}, err
			}
		}
		update.Worktrees = append(update.Worktrees, ref)
	}
	if err := store.ApplyRefs(ctx, repoID, update); err != nil {
		return refsResult{}, err
	}
	return refsResult{Changed: true, Worktrees: worktrees}, nil
}

func saveMissingRanges(ctx context.Context, store *Store, repoID uuid.UUID, root, baseSHA string, branches []gavelgit.BranchRef, base string) error {
	keys := make([]RangeKey, 0, len(branches))
	for _, branch := range branches {
		if branch.Name != base {
			keys = append(keys, RangeKey{Base: baseSHA, Head: branch.Head})
		}
	}
	missing, err := store.MissingRanges(ctx, repoID, keys)
	if err != nil {
		return err
	}
	ranges := make([]Range, 0, len(missing))
	for _, key := range missing {
		compare, err := gavelgit.CompareRange(root, key.Base, key.Head)
		if err != nil {
			return err
		}
		ranges = append(ranges, Range{RangeKey: key, RangeCompare: compare})
	}
	return store.SaveRanges(ctx, repoID, ranges)
}

func refsFingerprint(base string, worktrees []gavelgit.Worktree, branches []gavelgit.BranchRef) string {
	hash := sha256.New()
	fmt.Fprintf(hash, "base %s\n", base)
	for _, wt := range worktrees {
		fmt.Fprintf(hash, "worktree %s %s %s %t %t %t\n", wt.Path, wt.Branch, wt.Head, wt.Primary, wt.Detached, wt.Prunable)
	}
	for _, branch := range branches {
		fmt.Fprintf(hash, "branch %s %s %s %s\n", branch.Name, branch.Head, branch.Worktree, branch.LastCommitAt.Format("2006-01-02T15:04:05Z"))
	}
	return hex.EncodeToString(hash.Sum(nil))
}

// scanStatus brings one worktree's uncommitted state up to date. A
// fingerprint of HEAD, the porcelain status and the stat of every listed file
// equal to the stored one ends the scan before the line counts are read. It
// writes nothing when it fails.
func scanStatus(ctx context.Context, store *Store, repoID uuid.UUID, path string) (bool, error) {
	fingerprint, err := statusFingerprint(ctx, path)
	if err != nil {
		return false, err
	}
	stored, err := store.StatusFingerprint(ctx, repoID, path)
	if err != nil {
		return false, err
	}
	if fingerprint == stored {
		return false, store.MarkStatusScanned(ctx, repoID, path)
	}
	changes, touched, files, err := WorktreeChanges(ctx, path, pollGitArgs...)
	if err != nil {
		return false, err
	}
	return true, store.ApplyStatus(ctx, repoID, path, StatusUpdate{Fingerprint: fingerprint, Changes: changes, TouchedAt: touched, Files: files})
}

func statusFingerprint(ctx context.Context, path string) (string, error) {
	head, err := runGit(ctx, path, append(pollGitArgs, "rev-parse", "--verify", "--quiet", "HEAD")...)
	// --quiet exits 1, and only 1, when HEAD names no commit: an unborn branch,
	// whose status still counts.
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) && exitErr.ExitCode() == 1 {
		head, err = nil, nil
	}
	if err != nil {
		return "", err
	}
	porcelain, err := runGit(ctx, path, append(pollGitArgs, "status", "--porcelain=v1", "-z", "--untracked-files=all")...)
	if err != nil {
		return "", err
	}
	hash := sha256.New()
	hash.Write(head)
	hash.Write(porcelain)
	for _, file := range porcelainPaths(porcelain) {
		info, err := os.Lstat(filepath.Join(path, file))
		if errors.Is(err, fs.ErrNotExist) {
			continue
		}
		if err != nil {
			return "", fmt.Errorf("stat changed file %s in worktree %s: %w", file, path, err)
		}
		fmt.Fprintf(hash, "%s %d %d\n", file, info.ModTime().UnixNano(), info.Size())
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

// porcelainPaths lists the paths of `git status --porcelain=v1 -z` output.
// A rename or copy record is followed by its source path, which is skipped.
func porcelainPaths(raw []byte) []string {
	var paths []string
	records := bytes.Split(raw, []byte{0})
	for i := 0; i < len(records); i++ {
		record := records[i]
		if len(record) < 4 {
			continue
		}
		paths = append(paths, string(record[3:]))
		if record[0] == 'R' || record[0] == 'C' {
			i++
		}
	}
	return paths
}

func runGit(ctx context.Context, dir string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = dir
	var stderr strings.Builder
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git %s in %s: %w: %s", strings.Join(args, " "), dir, err, strings.TrimSpace(stderr.String()))
	}
	return out, nil
}
