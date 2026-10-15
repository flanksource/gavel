// Source bundles give a sandbox job its repository: a git bundle of one ref from the pushed bare
// repository, or from a mirror of the GitHub repository fetched with its installation token.
package sandbox

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/google/uuid"
)

// Repo is the repository a run checks out.
type Repo struct {
	ID     uuid.UUID
	Org    uuid.UUID
	Source string // push or github
	// Name is owner/name for a GitHub repository.
	Name           string
	GitHubRepoID   int64
	InstallationID int64
}

// Sources bundles repositories for sandbox jobs.
type Sources struct {
	// DataDir holds pushed repositories (repos/<org>/<repo>.git) and GitHub mirrors
	// (mirrors/<org>/<github repo id>.git).
	DataDir string
	// GitHubURL is where GitHub repositories are cloned from; empty is https://github.com.
	GitHubURL string
	// Token returns an installation token for fetching a GitHub repository.
	Token func(ctx context.Context, installationID int64) (string, error)

	mirrors sync.Map // mirror path → *sync.Mutex
}

// Bundle writes a git bundle of ref in repo to w and returns the commit ref names.
func (s *Sources) Bundle(ctx context.Context, repo Repo, ref string, w io.Writer) (string, error) {
	if strings.HasPrefix(ref, "-") {
		return "", fmt.Errorf("ref %q can't start with \"-\"", ref)
	}
	var dir string
	switch repo.Source {
	case "push":
		dir = filepath.Join(s.DataDir, "repos", repo.Org.String(), repo.ID.String()+".git")
	case "github":
		var err error
		if dir, err = s.updateMirror(ctx, repo); err != nil {
			return "", err
		}
	default:
		return "", fmt.Errorf("repository %s has source %q, not push or github", repo.ID, repo.Source)
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

// updateMirror clones the GitHub repository's mirror on first use and fetches it every time. The
// installation token is passed as a header for that command only, never written to git config.
func (s *Sources) updateMirror(ctx context.Context, repo Repo) (string, error) {
	if s.Token == nil {
		return "", fmt.Errorf("no GitHub token source for %s", repo.Name)
	}
	dir := filepath.Join(s.DataDir, "mirrors", repo.Org.String(), strconv.FormatInt(repo.GitHubRepoID, 10)+".git")
	lock, _ := s.mirrors.LoadOrStore(dir, &sync.Mutex{})
	lock.(*sync.Mutex).Lock()
	defer lock.(*sync.Mutex).Unlock()

	token, err := s.Token(ctx, repo.InstallationID)
	if err != nil {
		return "", fmt.Errorf("get an installation token for %s: %w", repo.Name, err)
	}
	header := "http.extraHeader=Authorization: Basic " + base64.StdEncoding.EncodeToString([]byte("x-access-token:"+token))
	base := strings.TrimRight(s.GitHubURL, "/")
	if base == "" {
		base = "https://github.com"
	}
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(filepath.Dir(dir), 0o755); err != nil {
			return "", fmt.Errorf("create the mirror directory for %s: %w", repo.Name, err)
		}
		if _, err := runGit(ctx, nil, "-c", header, "clone", "--quiet", "--mirror", base+"/"+repo.Name+".git", dir); err != nil {
			return "", fmt.Errorf("mirror %s: %w", repo.Name, err)
		}
		return dir, nil
	}
	if _, err := runGit(ctx, nil, "-c", header, "-C", dir, "fetch", "--quiet", "--prune", "origin"); err != nil {
		return "", fmt.Errorf("fetch %s: %w", repo.Name, err)
	}
	return dir, nil
}

// runGit runs git, writing its output to stdout when set, or returning it otherwise. A failure
// carries git's stderr, which never holds the token: it is only in a -c argument.
func runGit(ctx context.Context, stdout io.Writer, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0")
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
