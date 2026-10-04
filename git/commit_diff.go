package git

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

// CommitDiffOptions selects the changes CommitDiff and CommitFiles read. An
// empty Base is the single commit Head; a set Base is the Base..Head range (a
// run branch's setup..head). File optionally narrows either to one repository
// path, a file or a directory.
type CommitDiffOptions struct {
	Base string
	Head string
	File string
}

// CommitDiffResult is a plain unified diff (no colour, no diffstat) ready for a
// unified-diff parser. Truncated reports that Diff was capped at the size limit;
// Binary reports that every file in the untruncated diff is binary.
type CommitDiffResult struct {
	Diff      string
	Truncated bool
	Binary    bool
}

// Validate trims the options and rejects anything that must not reach git: a
// malformed Head or Base hash, or a File git would read as something other
// than a literal repository path.
func (o CommitDiffOptions) Validate() error {
	_, err := o.normalize()
	return err
}

func (o CommitDiffOptions) normalize() (CommitDiffOptions, error) {
	o = CommitDiffOptions{Base: strings.TrimSpace(o.Base), Head: strings.TrimSpace(o.Head), File: strings.TrimSpace(o.File)}
	if !IsValidCommitHash(o.Head) {
		return o, fmt.Errorf("invalid commit hash %q", o.Head)
	}
	if o.Base != "" && !IsValidCommitHash(o.Base) {
		return o, fmt.Errorf("invalid base commit hash %q", o.Base)
	}
	if o.File != "" {
		if err := validateDiffPath(o.File); err != nil {
			return o, err
		}
	}
	return o, nil
}

// CommitDiff returns the plain unified diff of a commit or a Base..Head range,
// optionally narrowed to File, capped at the TruncateDiff limit. A well-formed
// hash whose commit is not in the repository is ErrCommitNotFound.
func CommitDiff(dir string, opts CommitDiffOptions) (CommitDiffResult, error) {
	out, err := commitDiffOutput(dir, opts)
	if err != nil {
		return CommitDiffResult{}, err
	}
	diff, truncated := TruncateDiff(out)
	return CommitDiffResult{Diff: diff, Truncated: truncated, Binary: IsBinaryDiff(out)}, nil
}

// commitDiffOutput validates opts, confirms every named commit exists, then runs
// `git show` (single commit) or `git diff` (range). `--` always terminates the
// revisions so File can never be read as one.
func commitDiffOutput(dir string, opts CommitDiffOptions) (string, error) {
	opts, err := opts.normalize()
	if err != nil {
		return "", err
	}
	args := []string{"show", "--format=", "-M", "--no-color", "--no-ext-diff", "--patch", opts.Head}
	if opts.Base != "" {
		if err := requireCommit(dir, opts.Base); err != nil {
			return "", err
		}
		args = []string{"diff", "-M", "--no-color", "--no-ext-diff", opts.Base, opts.Head}
	}
	if err := requireCommit(dir, opts.Head); err != nil {
		return "", err
	}
	args = append(args, "--")
	if opts.File != "" {
		args = append(args, opts.File)
	}
	out, err := gitOutput(dir, args...)
	if err != nil {
		return "", err
	}
	return string(out), nil
}

// requireCommit reports a sha that does not resolve to a commit in dir as
// ErrCommitNotFound. `rev-parse -q --verify` exits 1 for a missing or
// non-commit object (full or abbreviated) and 128 for anything else, such as a
// dir that is not a repository, which is surfaced as is.
func requireCommit(dir, sha string) error {
	_, err := gitOutput(dir, "rev-parse", "-q", "--verify", "--end-of-options", sha+"^{commit}")
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) && exitErr.ExitCode() == 1 {
		return fmt.Errorf("%w: %s in %s", ErrCommitNotFound, sha, dir)
	}
	return err
}
