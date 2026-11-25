// Package proposal keeps proposed patches as commits under refs, built from a
// temporary index so the working tree and the real index are never touched.
package proposal

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/flanksource/gavel/internal/safefs"
)

type CreateOptions struct {
	Dir string
	// Base defaults to HEAD.
	Base string
	// Patch is a unified diff; Edits are Edit-tool-style changes. Exactly one is set.
	Patch   []byte
	Edits   []Edit
	Message string
	Ref     string
}

type Result struct {
	Base   string
	Commit string
	Files  []FileStat
}

// FileStat is one changed path; a binary file is listed with zero counts.
type FileStat struct {
	Path      string
	Additions int
	Deletions int
}

// Create commits Patch or Edits onto Base and creates Ref at it; an existing
// Ref is an error.
func Create(ctx context.Context, opts CreateOptions) (Result, error) {
	switch {
	case len(opts.Patch) > 0 && len(opts.Edits) > 0:
		return Result{}, errors.New("proposal: a patch or edits, not both")
	case len(opts.Patch) == 0 && len(opts.Edits) == 0:
		return Result{}, errors.New("proposal: empty patch and no edits")
	}
	if opts.Message == "" {
		return Result{}, errors.New("proposal: commit message is required")
	}
	dir, err := topLevel(ctx, opts.Dir)
	if err != nil {
		return Result{}, err
	}
	if err := checkRef(ctx, dir, opts.Ref); err != nil {
		return Result{}, err
	}
	base, err := resolveBase(ctx, dir, opts.Base)
	if err != nil {
		return Result{}, err
	}
	change := func(git call) error { return applyPatch(ctx, git, base, opts.Patch) }
	if len(opts.Edits) > 0 {
		change = func(git call) error { return applyEdits(ctx, git, base, opts.Edits) }
	}
	commit, err := commitChange(ctx, dir, base, opts.Message, change)
	if err != nil {
		return Result{}, err
	}
	if _, err := (call{dir: dir}).run(ctx, "update-ref", "--no-deref", "-m", opts.Message, opts.Ref, commit, ""); err != nil {
		return Result{}, fmt.Errorf("proposal: create %s: %w", opts.Ref, err)
	}
	files, err := Stat(ctx, dir, base, commit)
	if err != nil {
		return Result{}, err
	}
	return Result{Base: base, Commit: commit, Files: files}, nil
}

func resolveBase(ctx context.Context, dir, base string) (string, error) {
	if base == "" {
		base = "HEAD"
	}
	sha, err := call{dir: dir}.trimmed(ctx, "rev-parse", "--verify", "--end-of-options", base+"^{commit}")
	if err != nil {
		return "", fmt.Errorf("proposal: resolve base %q: %w", base, err)
	}
	return sha, nil
}

// commitChange runs change on base's tree in a temporary index under the git
// dir and commits the result with base as its parent; a change leaving the
// tree as it was is an error.
func commitChange(ctx context.Context, dir, base, message string, change func(call) error) (_ string, err error) {
	gitDir, err := call{dir: dir}.trimmed(ctx, "rev-parse", "--path-format=absolute", "--git-dir")
	if err != nil {
		return "", err
	}
	tmp, err := os.MkdirTemp(gitDir, "gavel-proposal-")
	if err != nil {
		return "", fmt.Errorf("proposal: make a temporary index: %w", err)
	}
	defer func() {
		if rmErr := safefs.SafeRemove(tmp, safefs.RemoveContext{AllowedRoot: gitDir}); rmErr != nil {
			err = errors.Join(err, fmt.Errorf("proposal: remove the temporary index: %w", rmErr))
		}
	}()
	git := call{dir: dir, env: []string{"GIT_INDEX_FILE=" + filepath.Join(tmp, "index")}}
	if _, err := git.run(ctx, "read-tree", base); err != nil {
		return "", fmt.Errorf("proposal: read the tree of %s: %w", base, err)
	}
	if err := change(git); err != nil {
		return "", err
	}
	tree, err := git.trimmed(ctx, "write-tree")
	if err != nil {
		return "", fmt.Errorf("proposal: write the tree: %w", err)
	}
	baseTree, err := git.trimmed(ctx, "rev-parse", base+"^{tree}")
	if err != nil {
		return "", fmt.Errorf("proposal: resolve the tree of %s: %w", base, err)
	}
	if tree == baseTree {
		return "", fmt.Errorf("proposal: the proposed edits change nothing on %s", base)
	}
	commit, err := git.trimmed(ctx, "-c", "user.name=gavel", "-c", "user.email=gavel@gavel.invalid", "commit-tree", tree, "-p", base, "-m", message)
	if err != nil {
		return "", fmt.Errorf("proposal: commit the tree: %w", err)
	}
	return commit, nil
}

// Delete removes ref.
func Delete(ctx context.Context, dir, ref string) error {
	if err := checkRef(ctx, dir, ref); err != nil {
		return err
	}
	if _, err := (call{dir: dir}).run(ctx, "update-ref", "-d", ref); err != nil {
		return fmt.Errorf("proposal: delete %s: %w", ref, err)
	}
	return nil
}
