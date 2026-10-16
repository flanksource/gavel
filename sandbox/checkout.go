// Workspace checkouts: serve's read-only working tree of a repository's default branch, which
// hosted todos read their configuration, lifecycle and prompts from. Serve never runs code in it.
package sandbox

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// Checkout is a working tree of a repository's default branch at one commit.
type Checkout struct {
	Dir string
	SHA string
}

// Checkout brings the working tree at <DataDir>/workspaces/<org>/<repo id> up to date with repo's
// default branch, cloning it on first use. Symlinks are checked out as plain files holding their
// target, so nothing in the tree reads outside it.
func (s *Sources) Checkout(ctx context.Context, repo Repo) (Checkout, error) {
	source, err := s.sourceDir(ctx, repo)
	if err != nil {
		return Checkout{}, err
	}
	if _, err := runGit(ctx, nil, "-C", source, "rev-parse", "--verify", "--quiet", "HEAD^{commit}"); err != nil {
		return Checkout{}, fmt.Errorf("repository %s has no default branch yet: push a branch to it first", repo.label())
	}
	dir := filepath.Join(s.DataDir, "workspaces", repo.Org.String(), repo.ID.String())
	lock, _ := s.checkouts.LoadOrStore(dir, &sync.Mutex{})
	lock.(*sync.Mutex).Lock()
	defer lock.(*sync.Mutex).Unlock()

	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(filepath.Dir(dir), 0o755); err != nil {
			return Checkout{}, fmt.Errorf("create the workspace directory for %s: %w", repo.label(), err)
		}
		if _, err := runGit(ctx, nil, "-c", "core.symlinks=false", "clone", "--quiet", "--no-checkout", source, dir); err != nil {
			return Checkout{}, fmt.Errorf("clone the workspace of %s: %w", repo.label(), err)
		}
		if _, err := runGit(ctx, nil, "-C", dir, "config", "core.symlinks", "false"); err != nil {
			return Checkout{}, fmt.Errorf("configure the workspace of %s: %w", repo.label(), err)
		}
	} else if _, err := runGit(ctx, nil, "-C", dir, "fetch", "--quiet", "--prune", "origin", "+refs/heads/*:refs/remotes/origin/*"); err != nil {
		return Checkout{}, fmt.Errorf("refresh the workspace of %s: %w", repo.label(), err)
	}
	head, err := runGit(ctx, nil, "-C", source, "rev-parse", "HEAD^{commit}")
	if err != nil {
		return Checkout{}, fmt.Errorf("resolve the default branch of %s: %w", repo.label(), err)
	}
	sha := strings.TrimSpace(head)
	if _, err := runGit(ctx, nil, "-C", dir, "checkout", "--quiet", "--force", "--detach", sha); err != nil {
		return Checkout{}, fmt.Errorf("check out the default branch of %s: %w", repo.label(), err)
	}
	return Checkout{Dir: dir, SHA: sha}, nil
}

// sourceDir is the repository Sources reads repo from: its pushed bare repository, or its GitHub
// mirror, fetched first.
func (s *Sources) sourceDir(ctx context.Context, repo Repo) (string, error) {
	switch repo.Source {
	case "push":
		return filepath.Join(s.DataDir, "repos", repo.Org.String(), repo.ID.String()+".git"), nil
	case "github":
		return s.updateMirror(ctx, repo)
	default:
		return "", fmt.Errorf("repository %s has source %q, not push or github", repo.ID, repo.Source)
	}
}

// label names repo in messages: its name, or its ID when it has none.
func (r Repo) label() string {
	if r.Name != "" {
		return r.Name
	}
	return r.ID.String()
}
