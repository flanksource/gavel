// Tests source bundles: a pushed repository is bundled from its bare repository, a GitHub one from
// a mirror fetched with the installation token, which is never written to the mirror's config.
package sandbox

import (
	"bytes"
	"context"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func git(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@example.com", "GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@example.com")
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, "git %v: %s", args, out)
	return strings.TrimSpace(string(out))
}

// commitFile commits content to file on main in work and returns the commit.
func commitFile(t *testing.T, work, file, content string) string {
	t.Helper()
	require.NoError(t, os.WriteFile(filepath.Join(work, file), []byte(content), 0o644))
	git(t, work, "add", file)
	git(t, work, "commit", "-m", "update "+file)
	return git(t, work, "rev-parse", "HEAD")
}

// unbundle clones a bundle and returns the content of file at its main branch.
func unbundle(t *testing.T, bundle []byte, file string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "src.bundle")
	require.NoError(t, os.WriteFile(path, bundle, 0o644))
	git(t, dir, "clone", "-q", "--branch", "main", path, "src")
	data, err := os.ReadFile(filepath.Join(dir, "src", file))
	require.NoError(t, err)
	return string(data)
}

func newWork(t *testing.T) string {
	t.Helper()
	work := t.TempDir()
	git(t, work, "init", "-q", "-b", "main")
	return work
}

func TestBundleAPushedRepository(t *testing.T) {
	dataDir := t.TempDir()
	org, repo := uuid.New(), uuid.New()
	bare := filepath.Join(dataDir, "repos", org.String(), repo.String()+".git")
	work := newWork(t)
	sha := commitFile(t, work, "README.md", "pushed")
	git(t, work, "clone", "-q", "--bare", work, bare)

	var bundle bytes.Buffer
	got, err := (&Sources{DataDir: dataDir}).Bundle(context.Background(), Repo{ID: repo, Org: org, Source: "push"}, "refs/heads/main", &bundle)
	require.NoError(t, err)
	assert.Equal(t, sha, got)
	assert.Equal(t, "pushed", unbundle(t, bundle.Bytes(), "README.md"))
}

func TestBundleAGitHubRepositoryFromItsMirror(t *testing.T) {
	dataDir, remotes := t.TempDir(), t.TempDir()
	org, repo := uuid.New(), uuid.New()
	work := newWork(t)
	commitFile(t, work, "README.md", "first")
	remote := filepath.Join(remotes, "acme", "app.git")
	git(t, work, "clone", "-q", "--bare", work, remote)

	var tokens []int64
	sources := &Sources{
		DataDir:   dataDir,
		GitHubURL: "file://" + remotes,
		Token: func(_ context.Context, installation int64) (string, error) {
			tokens = append(tokens, installation)
			return "installation-token", nil
		},
	}
	github := Repo{ID: repo, Org: org, Source: "github", Name: "acme/app", GitHubRepoID: 42, InstallationID: 7}

	var bundle bytes.Buffer
	_, err := sources.Bundle(context.Background(), github, "refs/heads/main", &bundle)
	require.NoError(t, err)
	assert.Equal(t, "first", unbundle(t, bundle.Bytes(), "README.md"))

	// A new commit on GitHub is fetched into the existing mirror.
	second := commitFile(t, work, "README.md", "second")
	git(t, work, "push", "-q", remote, "main")
	bundle.Reset()
	got, err := sources.Bundle(context.Background(), github, "refs/heads/main", &bundle)
	require.NoError(t, err)
	assert.Equal(t, second, got)
	assert.Equal(t, "second", unbundle(t, bundle.Bytes(), "README.md"))

	assert.Equal(t, []int64{7, 7}, tokens, "every fetch uses a fresh installation token")
	mirror := filepath.Join(dataDir, "mirrors", org.String(), "42.git")
	config, err := os.ReadFile(filepath.Join(mirror, "config"))
	require.NoError(t, err)
	assert.NotContains(t, string(config), "installation-token", "the token is never written to the mirror's config")
}

func TestAnInterruptedMirrorCloneLeavesNoMirror(t *testing.T) {
	dataDir, remotes := t.TempDir(), t.TempDir()
	org := uuid.New()
	work := newWork(t)
	commitFile(t, work, "README.md", "first")
	git(t, work, "clone", "-q", "--bare", work, filepath.Join(remotes, "acme", "app.git"))
	// A remote that never answers, so the clone is killed partway, as a cancelled request kills it.
	t.Setenv("GIT_CONFIG_COUNT", "1")
	t.Setenv("GIT_CONFIG_KEY_0", "protocol.ext.allow")
	t.Setenv("GIT_CONFIG_VALUE_0", "always")
	sources := &Sources{
		DataDir:   dataDir,
		GitHubURL: "ext::sh -c sleep% 2;:",
		Token:     func(context.Context, int64) (string, error) { return "installation-token", nil },
	}
	github := Repo{ID: uuid.New(), Org: org, Source: "github", Name: "acme/app", GitHubRepoID: 42, InstallationID: 7}

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()
	_, err := sources.Bundle(ctx, github, "refs/heads/main", io.Discard)
	require.Error(t, err)

	sources.GitHubURL = "file://" + remotes
	var bundle bytes.Buffer
	_, err = sources.Bundle(context.Background(), github, "refs/heads/main", &bundle)
	require.NoError(t, err, "the next request mirrors the repository whole")
	assert.Equal(t, "first", unbundle(t, bundle.Bytes(), "README.md"))
}

func TestBundleRefusesOptionLikeRefsAndUnknownRefs(t *testing.T) {
	dataDir := t.TempDir()
	org, repo := uuid.New(), uuid.New()
	work := newWork(t)
	commitFile(t, work, "README.md", "x")
	git(t, work, "clone", "-q", "--bare", work, filepath.Join(dataDir, "repos", org.String(), repo.String()+".git"))
	sources := &Sources{DataDir: dataDir}
	pushed := Repo{ID: repo, Org: org, Source: "push"}

	_, err := sources.Bundle(context.Background(), pushed, "--all", &bytes.Buffer{})
	require.EqualError(t, err, `ref "--all" can't start with "-"`)

	_, err = sources.Bundle(context.Background(), pushed, "refs/heads/missing", &bytes.Buffer{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), `resolve refs/heads/missing in `+repo.String())

	_, err = sources.Bundle(context.Background(), Repo{ID: repo, Org: org, Source: "svn"}, "refs/heads/main", &bytes.Buffer{})
	require.EqualError(t, err, `repository `+repo.String()+` has source "svn", not push or github`)
}
