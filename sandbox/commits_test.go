// Tests PushRunCommits: the commits a sandbox job's run made come back as a git bundle, and serve
// creates the branch it names for them, in the pushed repository or on GitHub, after checking
// they continue from the commit the run started at.
package sandbox

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// runCommits commits on a new branch of a clone of source and returns a bundle of the branch's
// commits since base, and its tip.
func runCommits(t *testing.T, source, base string, branches ...string) ([]byte, string) {
	t.Helper()
	work := t.TempDir()
	git(t, work, "clone", "-q", source, ".")
	var tip string
	for _, branch := range branches {
		git(t, work, "checkout", "-q", "-b", branch, base)
		tip = commitFile(t, work, "fix.txt", "fixed on "+branch)
	}
	args := append([]string{"bundle", "create", filepath.Join(work, "run.bundle")}, branches...)
	git(t, work, append(args, "^"+base)...)
	data, err := os.ReadFile(filepath.Join(work, "run.bundle"))
	require.NoError(t, err)
	return data, tip
}

func TestPushRunCommitsToAPushedRepository(t *testing.T) {
	dataDir := t.TempDir()
	org, repo := uuid.New(), uuid.New()
	bare := filepath.Join(dataDir, "repos", org.String(), repo.String()+".git")
	work := newWork(t)
	base := commitFile(t, work, "README.md", "first")
	git(t, work, "clone", "-q", "--bare", work, bare)
	sources := &Sources{DataDir: dataDir}
	pushed := Repo{ID: repo, Org: org, Source: "push", Name: "app"}

	bundle, tip := runCommits(t, bare, base, "worktree-run")
	head, err := sources.PushRunCommits(context.Background(), pushed, base, "gavel/todo-run", bytes.NewReader(bundle))
	require.NoError(t, err)
	assert.Equal(t, tip, head)
	assert.Equal(t, tip, git(t, bare, "rev-parse", "refs/heads/gavel/todo-run"))
	assert.Empty(t, git(t, bare, "for-each-ref", "refs/gavel/"), "nothing is left behind but the branch")

	_, err = sources.PushRunCommits(context.Background(), pushed, base, "gavel/todo-run", bytes.NewReader(bundle))
	require.ErrorIs(t, err, ErrBranchExists)
	assert.EqualError(t, err, "create branch gavel/todo-run of app: the branch already exists")

	elsewhere := newWork(t)
	commitFile(t, elsewhere, "README.md", "unrelated")
	git(t, elsewhere, "bundle", "create", filepath.Join(elsewhere, "x.bundle"), "main")
	unrelated, err := os.ReadFile(filepath.Join(elsewhere, "x.bundle"))
	require.NoError(t, err)
	_, err = sources.PushRunCommits(context.Background(), pushed, base, "gavel/unrelated", bytes.NewReader(unrelated))
	require.EqualError(t, err, "the run's commits on main don't continue from "+base+", the commit it started at")

	two, _ := runCommits(t, bare, base, "one", "two")
	_, err = sources.PushRunCommits(context.Background(), pushed, base, "gavel/two", bytes.NewReader(two))
	require.EqualError(t, err, "the run's bundle has 2 branches; a run's commits are on one")

	_, err = sources.PushRunCommits(context.Background(), pushed, base, "main", bytes.NewReader(bundle))
	require.EqualError(t, err, `a run's commits go on a gavel/ branch, not "main"`)
}

func TestPushRunCommitsToGitHub(t *testing.T) {
	dataDir, remotes := t.TempDir(), t.TempDir()
	org, repo := uuid.New(), uuid.New()
	work := newWork(t)
	base := commitFile(t, work, "README.md", "first")
	remote := filepath.Join(remotes, "acme", "app.git")
	git(t, work, "clone", "-q", "--bare", work, remote)
	sources := &Sources{
		DataDir:   dataDir,
		GitHubURL: "file://" + remotes,
		Token:     func(context.Context, int64) (string, error) { return "installation-token", nil },
	}
	github := Repo{ID: repo, Org: org, Source: "github", Name: "acme/app", GitHubRepoID: 42, InstallationID: 7}
	_, err := sources.Checkout(context.Background(), github)
	require.NoError(t, err)

	bundle, tip := runCommits(t, remote, base, "worktree-run")
	head, err := sources.PushRunCommits(context.Background(), github, base, "gavel/todo-run", bytes.NewReader(bundle))
	require.NoError(t, err)
	assert.Equal(t, tip, head)
	assert.Equal(t, tip, git(t, remote, "rev-parse", "refs/heads/gavel/todo-run"), "the branch is on GitHub")
	_, err = sources.PushRunCommits(context.Background(), github, base, "gavel/todo-run", bytes.NewReader(bundle))
	require.ErrorIs(t, err, ErrBranchExists, "a branch already on GitHub is never moved")
	assert.EqualError(t, err, "create branch gavel/todo-run of acme/app on GitHub: the branch already exists")
	config, err := os.ReadFile(filepath.Join(dataDir, "mirrors", org.String(), "42.git", "config"))
	require.NoError(t, err)
	assert.NotContains(t, string(config), "installation-token")
}

func TestPushPRHeadFastForwardsTheSourceBranch(t *testing.T) {
	dataDir, remotes := t.TempDir(), t.TempDir()
	org, repo := uuid.New(), uuid.New()
	work := newWork(t)
	base := commitFile(t, work, "README.md", "first")
	remote := filepath.Join(remotes, "acme", "app.git")
	git(t, work, "clone", "-q", "--bare", work, remote)
	git(t, remote, "branch", "topic", base)
	sources := &Sources{
		DataDir:   dataDir,
		GitHubURL: "file://" + remotes,
		Token:     func(context.Context, int64) (string, error) { return "installation-token", nil },
	}
	github := Repo{ID: repo, Org: org, Source: "github", Name: "acme/app", GitHubRepoID: 42, InstallationID: 7}
	_, err := sources.Checkout(context.Background(), github)
	require.NoError(t, err)

	bundle, tip := runCommits(t, remote, base, "topic-turn")
	head, err := sources.PushPRHead(context.Background(), github, "file://"+remote, "topic", base, bytes.NewReader(bundle))
	require.NoError(t, err)
	assert.Equal(t, tip, head)
	assert.Equal(t, tip, git(t, remote, "rev-parse", "refs/heads/topic"))

	stale, _ := runCommits(t, remote, base, "another-turn")
	_, err = sources.PushPRHead(context.Background(), github, "file://"+remote, "topic", base, bytes.NewReader(stale))
	require.Error(t, err, "a branch that moved since the job saw it isn't pushed over")
	assert.Contains(t, err.Error(), "push the PR's source branch topic")
	assert.Equal(t, tip, git(t, remote, "rev-parse", "refs/heads/topic"))
}

func TestARunBranchIsFoundByItsTip(t *testing.T) {
	dataDir := t.TempDir()
	org, repo, todo, run := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	bare := filepath.Join(dataDir, "repos", org.String(), repo.String()+".git")
	work := newWork(t)
	base := commitFile(t, work, "README.md", "first")
	git(t, work, "clone", "-q", "--bare", work, bare)
	sources := &Sources{DataDir: dataDir}
	pushed := Repo{ID: repo, Org: org, Source: "push", Name: "app"}

	branch := RunBranch(todo, run)
	assert.Equal(t, "gavel/"+todo.String()[:8]+"-"+run.String()[:8], branch)
	bundle, tip := runCommits(t, bare, base, "worktree-run")
	_, err := sources.PushRunCommits(context.Background(), pushed, base, branch, bytes.NewReader(bundle))
	require.NoError(t, err)
	other, _ := runCommits(t, bare, base, "other-run")
	_, err = sources.PushRunCommits(context.Background(), pushed, base, RunBranch(todo, uuid.New()), bytes.NewReader(other))
	require.NoError(t, err)

	found, err := sources.RunBranchAt(context.Background(), pushed, todo, tip)
	require.NoError(t, err)
	assert.Equal(t, branch, found, "the run branch whose tip is the run's head")

	_, err = sources.RunBranchAt(context.Background(), pushed, uuid.New(), tip)
	require.ErrorIs(t, err, ErrNoRunBranch)

	subjects, err := sources.CommitSubjects(context.Background(), pushed, base, tip)
	require.NoError(t, err)
	assert.Len(t, subjects, 1)
	_, err = sources.CommitSubjects(context.Background(), pushed, "--output=/tmp/x", tip)
	require.EqualError(t, err, `commit "--output=/tmp/x" can't start with "-"`)
}
