// Tests PushRunCommits: the commits a sandbox job's run made come back as a git bundle, and serve
// creates the branch it names for them, in the pushed repository or on GitHub, after checking
// they continue from the commit the run started at.
package sandbox

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
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

func TestPushPRHeadRefusesABaseThatIsNotACommitSHA(t *testing.T) {
	sources := &Sources{DataDir: t.TempDir()}
	github := Repo{ID: uuid.New(), Org: uuid.New(), Source: "github", Name: "acme/app"}
	for _, base := range []string{"", "main", "HEAD~1", "--upload-pack=touch /tmp/pwned", strings.Repeat("g", 40)} {
		_, err := sources.PushPRHead(context.Background(), github, "https://github.com/acme/app", "topic", base, strings.NewReader(""))
		require.EqualError(t, err, fmt.Sprintf("the PR's base %q is not a commit SHA", base))
	}
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

func TestLeaveOutPutsFilesBackAsTheRunFoundThem(t *testing.T) {
	dataDir, remotes := t.TempDir(), t.TempDir()
	org, repo := uuid.New(), uuid.New()
	work := newWork(t)
	commitFile(t, work, "README.md", "first")
	base := commitFile(t, work, "go.mod", "module app")
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

	// The run edits go.mod, adds fix.go and notes.md, and deletes README.md.
	run := t.TempDir()
	git(t, run, "clone", "-q", remote, ".")
	git(t, run, "checkout", "-q", "-b", "run")
	commitFile(t, run, "go.mod", "module app\n\ngo 1.25")
	commitFile(t, run, "fix.go", "package app")
	commitFile(t, run, "notes.md", "scratch")
	commitFile(t, run, "my notes.md", "more scratch")
	git(t, run, "rm", "-q", "README.md")
	git(t, run, "commit", "-q", "-m", "drop the readme")
	head := git(t, run, "rev-parse", "HEAD")
	git(t, run, "push", "-q", "origin", "HEAD:refs/heads/gavel/todo-run")

	pushed := Repo{ID: uuid.New(), Org: org, Source: "push", Name: "tool"}
	_, err = sources.LeaveOut(context.Background(), pushed, "gavel/todo-run", base, head, []string{"notes.md"})
	require.EqualError(t, err, `repository tool has source "push"; only a GitHub repository's runs land`)
	_, err = sources.LeaveOut(context.Background(), github, "gavel/todo-run", base, head, []string{"main.go"})
	require.ErrorIs(t, err, ErrInvalidLeaveOut)
	require.EqualError(t, err, `can't leave those files out: the run didn't change main.go`)
	_, err = sources.LeaveOut(context.Background(), github, "gavel/todo-run", base, head, []string{"go.mod", "fix.go", "notes.md", "my notes.md", "README.md"})
	require.ErrorIs(t, err, ErrInvalidLeaveOut)
	require.EqualError(t, err, "can't leave those files out: that is every file the run changed, leaving nothing to land")

	tip, err := sources.LeaveOut(context.Background(), github, "gavel/todo-run", base, head, []string{"notes.md", "my notes.md", "README.md", "go.mod"})
	require.NoError(t, err)
	assert.Equal(t, tip, git(t, remote, "rev-parse", "refs/heads/gavel/todo-run"), "the branch on GitHub carries it")
	assert.Equal(t, head, git(t, remote, "rev-parse", tip+"^"), "on top of the run's commits")
	assert.Equal(t, "Leave out notes.md, my notes.md, README.md, go.mod", git(t, remote, "log", "-1", "--format=%s", tip))
	assert.Equal(t, "fix.go", git(t, remote, "diff", "--name-only", base, tip), "only the kept file differs from where the run started")

	_, err = sources.LeaveOut(context.Background(), github, "gavel/todo-run", base, head, []string{"notes.md"})
	require.Error(t, err, "a branch that moved past head isn't pushed over")
	assert.Contains(t, err.Error(), "leave out files on branch gavel/todo-run of acme/app")
	assert.Equal(t, tip, git(t, remote, "rev-parse", "refs/heads/gavel/todo-run"))

	require.Error(t, sources.ResetRunBranch(context.Background(), github, "gavel/todo-run", head, base), "a branch not at the given tip isn't moved")
	require.NoError(t, sources.ResetRunBranch(context.Background(), github, "gavel/todo-run", tip, head))
	assert.Equal(t, head, git(t, remote, "rev-parse", "refs/heads/gavel/todo-run"), "back at the run's head")
}
