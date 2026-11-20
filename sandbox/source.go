// Source bundles give a sandbox job its repository: a git bundle of one ref from the pushed bare
// repository, or from a mirror of its remote fetched with the credentials of the installation or
// git connection it comes from.
package sandbox

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	commonsgit "github.com/flanksource/commons-db/git"
	"github.com/flanksource/gavel/internal/safefs"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Repo is the repository a run checks out.
type Repo struct {
	ID     uuid.UUID
	Org    uuid.UUID
	Source string // push, github, gitlab, azure_devops or git
	// Name is owner/name for a GitHub repository, its provider's full name for a connection's.
	Name           string
	GitHubRepoID   int64
	InstallationID int64
	// ConnectionID is the git connection the repository comes from; nil for none.
	ConnectionID *uuid.UUID
	// RemoteURL is where a connection's repository is cloned from.
	RemoteURL string
	// DefaultBranch is the branch a connection's sync recorded as the repository's default.
	DefaultBranch string
}

// RepoColumns selects a repos row as a Repo.
const RepoColumns = `id, org_id AS org, source, name, COALESCE(github_repo_id, 0) AS git_hub_repo_id,
	COALESCE(installation_id, 0) AS installation_id, connection_id, COALESCE(remote_url, '') AS remote_url,
	COALESCE(default_branch, '') AS default_branch`

// ErrRepoNotFound is a repository the organization doesn't have.
var ErrRepoNotFound = errors.New("repository not found")

// LoadRepo is org's repository id.
func LoadRepo(ctx context.Context, db *gorm.DB, org, id uuid.UUID) (Repo, error) {
	var repo Repo
	result := db.WithContext(ctx).Raw(`SELECT `+RepoColumns+` FROM repos WHERE id = ? AND org_id = ?`, id, org).Scan(&repo)
	if result.Error != nil {
		return Repo{}, fmt.Errorf("read repository %s: %w", id, result.Error)
	}
	if result.RowsAffected == 0 {
		return Repo{}, fmt.Errorf("%w: organization %s has no repository %s", ErrRepoNotFound, org, id)
	}
	return repo, nil
}

// Sources bundles repositories for sandbox jobs.
type Sources struct {
	// DataDir holds pushed repositories (repos/<org>/<repo>.git) and the mirrors of every other
	// (mirrors/<org>/<repo>.git).
	DataDir string
	// GitHubURL is where GitHub repositories are cloned from; empty is https://github.com.
	GitHubURL string
	// Token returns a token for fetching a GitHub repository from the installation recorded on
	// it, or from the server's one GitHub login when it serves without the App.
	Token func(ctx context.Context, installationID int64) (string, error)
	// Connection is the remote of a repository a git connection brings, with that connection's
	// credentials; nil serves no connection's repositories. A GitHub repository its installation
	// also grants is fetched with the installation's token instead.
	Connection func(ctx context.Context, repo Repo) (*commonsgit.Repository, error)

	mirrors   sync.Map // mirror path → *sync.Mutex
	checkouts sync.Map // workspace path → *sync.Mutex
}

// Bundle writes a git bundle of ref in repo to w and returns the commit ref names.
func (s *Sources) Bundle(ctx context.Context, repo Repo, ref string, w io.Writer) (string, error) {
	if strings.HasPrefix(ref, "-") {
		return "", fmt.Errorf("ref %q can't start with \"-\"", ref)
	}
	dir, err := s.sourceDir(ctx, repo)
	if err != nil {
		return "", err
	}
	sha, err := runGit(ctx, nil, "-C", dir, "rev-parse", "--verify", "--end-of-options", ref+"^{commit}")
	if err != nil {
		return "", fmt.Errorf("resolve %s in %s: %w", ref, repo.ID, err)
	}
	if _, err := runGit(ctx, w, "-C", dir, "bundle", "create", "-", ref); err != nil {
		return "", fmt.Errorf("bundle %s of %s: %w", ref, repo.ID, err)
	}
	return strings.TrimSpace(sha), nil
}

// updateMirror clones repo's mirror on first use and fetches it every time, from where its remote
// is now. The credentials are in that command's environment only, never written to git config.
func (s *Sources) updateMirror(ctx context.Context, repo Repo) (string, error) {
	dir := s.mirrorDir(repo)
	lock, _ := s.mirrors.LoadOrStore(dir, &sync.Mutex{})
	lock.(*sync.Mutex).Lock()
	defer lock.(*sync.Mutex).Unlock()

	remote, err := s.remote(ctx, repo)
	if err != nil {
		return "", err
	}
	defer remote.done()
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(filepath.Dir(dir), 0o755); err != nil {
			return "", fmt.Errorf("create the mirror directory for %s: %w", repo.label(), err)
		}
		if err := cloneWhole(ctx, remote.env, dir, "clone", "--quiet", "--mirror", remote.url); err != nil {
			return "", fmt.Errorf("mirror %s: %w", repo.label(), err)
		}
		return dir, nil
	}
	if _, err := runGitWith(ctx, remote.env, nil, "-C", dir, "remote", "set-url", "origin", remote.url); err != nil {
		return "", fmt.Errorf("point the mirror of %s at %s: %w", repo.label(), remote.url, err)
	}
	if _, err := runGitWith(ctx, remote.env, nil, "-C", dir, "fetch", "--quiet", "--prune", "origin"); err != nil {
		return "", fmt.Errorf("fetch %s: %w", repo.label(), err)
	}
	return dir, nil
}

// mirrorDir is where repo's mirror is.
func (s *Sources) mirrorDir(repo Repo) string {
	return filepath.Join(s.DataDir, "mirrors", repo.Org.String(), repo.ID.String()+".git")
}

// githubIDMirror is a mirror named by its GitHub repository id rather than by its repository.
var githubIDMirror = regexp.MustCompile(`^[0-9]+\.git$`)

// RemoveGitHubIDMirrors deletes the mirrors named by GitHub repository id that came before
// mirrors were kept under their repository; each is mirrored again on first use.
func (s *Sources) RemoveGitHubIDMirrors() error {
	root := filepath.Join(s.DataDir, "mirrors")
	orgs, err := os.ReadDir(root)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("list the mirrors: %w", err)
	}
	for _, org := range orgs {
		if !org.IsDir() {
			continue
		}
		mirrors, err := os.ReadDir(filepath.Join(root, org.Name()))
		if err != nil {
			return fmt.Errorf("list the mirrors of %s: %w", org.Name(), err)
		}
		for _, mirror := range mirrors {
			if !mirror.IsDir() || !githubIDMirror.MatchString(mirror.Name()) {
				continue
			}
			dir := filepath.Join(root, org.Name(), mirror.Name())
			if err := safefs.SafeRemove(dir, safefs.RemoveContext{AllowedRoot: root}); err != nil {
				return fmt.Errorf("remove the mirror %s: %w", dir, err)
			}
		}
	}
	return nil
}

// githubURL is where a GitHub repository its installation grants is cloned from and pushed to.
func (s *Sources) githubURL(repo Repo) string {
	base := strings.TrimRight(s.GitHubURL, "/")
	if base == "" {
		base = "https://github.com"
	}
	return base + "/" + repo.Name + ".git"
}

// remoteAccess is how git commands reach a repository's remote: its URL, and the environment
// that authenticates them to it. done removes the credential files the environment names.
type remoteAccess struct {
	url  string
	env  []string
	done func()
}

// remote is repo's remote. A GitHub repository its installation grants, or that nothing but the
// server's GitHub login serves, is reached with that token; any other through its connection.
func (s *Sources) remote(ctx context.Context, repo Repo) (remoteAccess, error) {
	var remote *commonsgit.Repository
	var err error
	switch {
	case repo.Source == "github" && (repo.InstallationID != 0 || repo.ConnectionID == nil):
		remote, err = s.githubRemote(ctx, repo)
	case repo.ConnectionID == nil:
		return remoteAccess{}, fmt.Errorf("repository %s comes from no git connection any more; add one that brings it", repo.label())
	case s.Connection == nil:
		return remoteAccess{}, fmt.Errorf("repository %s comes from a git connection, which this server doesn't serve", repo.label())
	default:
		remote, err = s.Connection(ctx, repo)
	}
	if err != nil {
		return remoteAccess{}, err
	}
	env, done, err := remote.CommandEnv(ctx)
	if err != nil {
		return remoteAccess{}, fmt.Errorf("reach the remote of %s: %w", repo.label(), err)
	}
	return remoteAccess{url: remote.URL.Clone, env: env, done: done}, nil
}

// githubRemote is a GitHub repository's remote with a fresh token for it.
func (s *Sources) githubRemote(ctx context.Context, repo Repo) (*commonsgit.Repository, error) {
	if s.Token == nil {
		return nil, fmt.Errorf("no GitHub token source for %s", repo.Name)
	}
	token, err := s.Token(ctx, repo.InstallationID)
	if err != nil {
		return nil, fmt.Errorf("get an installation token for %s: %w", repo.Name, err)
	}
	// The provider is read from the URL, which a test's local remote isn't on GitHub; the
	// installation token's user is GitHub's either way.
	return commonsgit.NewRepository(commonsgit.RepositoryOptions{URL: s.githubURL(repo), Auth: commonsgit.Auth{Username: "x-access-token", Token: token}})
}

// cloneWhole runs the git clone args name into dir by way of a directory beside it, which moves
// into place once the clone completes: a clone cut off partway, as a cancelled request cuts it,
// never leaves dir half made. environ is git's whole environment (see runGitWith).
func cloneWhole(ctx context.Context, environ []string, dir string, args ...string) error {
	clone, err := os.MkdirTemp(filepath.Dir(dir), filepath.Base(dir)+".clone-")
	if err != nil {
		return err
	}
	defer safefs.SafeRemove(clone, safefs.RemoveContext{AllowedRoot: filepath.Dir(dir), Protected: []string{dir}}) //nolint:errcheck
	if _, err := runGitWith(ctx, environ, nil, append(args, clone)...); err != nil {
		return err
	}
	return os.Rename(clone, dir)
}

// runGit runs git, writing its output to stdout when set, or returning it otherwise. A failure
// carries git's stderr, which never holds a token: tokens are only in a remote's environment.
func runGit(ctx context.Context, stdout io.Writer, args ...string) (string, error) {
	return runGitEnv(ctx, nil, stdout, args...)
}

// runGitEnv is runGit with env added to git's environment.
func runGitEnv(ctx context.Context, env []string, stdout io.Writer, args ...string) (string, error) {
	return runGitWith(ctx, append(append(os.Environ(), "GIT_TERMINAL_PROMPT=0"), env...), stdout, args...)
}

// gitWaitDelay is how long a git command killed by its context is waited for after the kill.
const gitWaitDelay = 5 * time.Second

// runGitWith is runGit with environ as git's whole environment, as a remote's CommandEnv builds
// it; nil is runGit's.
func runGitWith(ctx context.Context, environ []string, stdout io.Writer, args ...string) (string, error) {
	if environ == nil {
		environ = append(os.Environ(), "GIT_TERMINAL_PROMPT=0")
	}
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Env = environ
	// A killed git can leave a remote helper holding its output open; stop waiting for it.
	cmd.WaitDelay = gitWaitDelay
	var out, stderr bytes.Buffer
	cmd.Stdout = &out
	if stdout != nil {
		cmd.Stdout = stdout
	}
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		if message := strings.TrimSpace(stderr.String()); message != "" {
			return "", fmt.Errorf("%w: %s", err, message)
		}
		return "", err
	}
	return out.String(), nil
}
