package land_test

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/flanksource/captain/pkg/api"
	captaindb "github.com/flanksource/captain/pkg/database"
	"github.com/flanksource/captain/pkg/promptrun"
	"github.com/flanksource/commons-db/dbtest"
	"github.com/flanksource/gavel/internal/database"
	"github.com/flanksource/gavel/todos"
	todoruntime "github.com/flanksource/gavel/todos/runtime"
	"github.com/flanksource/gavel/todos/runtime/runtimetest"
	"github.com/flanksource/gavel/todos/types"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const (
	snapshotMessage = "chore(setup): snapshot uncommitted changes"
	agentMessage    = "feat: add feature.txt"
	shellBranch     = "shell/0a1b2c3d"
)

// landFixture is a todo whose run step worked in a shell/* worktree branched
// off repo's main: a setup snapshot commit (wip.txt) then the agent's commit
// (feature.txt), recorded on the run as Captain's teardown would.
type landFixture struct {
	provider *todoruntime.Provider
	todo     *types.TODO
	repo     string
	bare     string
	worktree string
	state    api.WorktreeState
}

func git(dir string, args ...string) string {
	GinkgoHelper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GIT_CONFIG_GLOBAL=/dev/null", "GIT_CONFIG_SYSTEM=/dev/null")
	out, err := cmd.CombinedOutput()
	Expect(err).NotTo(HaveOccurred(), "git %s:\n%s", strings.Join(args, " "), out)
	return strings.TrimSpace(string(out))
}

func commitFile(dir, name, content, message string) string {
	GinkgoHelper()
	Expect(os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644)).To(Succeed())
	git(dir, "add", name)
	git(dir, "commit", "-q", "-m", message)
	return git(dir, "rev-parse", "HEAD")
}

func branchExists(repo, branch string) bool {
	return exec.Command("git", "-C", repo, "rev-parse", "--verify", "--quiet", "refs/heads/"+branch).Run() == nil
}

func newLandFixture(ctx context.Context) *landFixture {
	GinkgoHelper()
	handle := dbtest.ForGinkgo(dbtest.Options{Name: "gavel_todos_land"})
	GinkgoT().Setenv(database.EnvDSN, handle.DSN())
	GinkgoT().Setenv(database.EnvDisable, "")
	GinkgoT().Setenv(database.LegacyEnvDSN, "")
	GinkgoT().Setenv(database.LegacyEnvDisable, "")
	GinkgoT().Setenv("HOME", GinkgoT().TempDir())
	opened, err := database.Open(ctx, database.WithMigrations())
	Expect(err).NotTo(HaveOccurred())
	DeferCleanup(func() { Expect(opened.Close()).To(Succeed()) })

	root := GinkgoT().TempDir()
	f := &landFixture{repo: filepath.Join(root, "work"), bare: filepath.Join(root, "origin.git")}
	git(root, "init", "-q", "--bare", "-b", "main", f.bare)
	git(root, "init", "-q", "-b", "main", f.repo)
	git(f.repo, "config", "user.email", "test@example.com")
	git(f.repo, "config", "user.name", "test")
	git(f.repo, "remote", "add", "origin", f.bare)
	commitFile(f.repo, "README.md", "hello\n", "chore: initial")
	git(f.repo, "push", "-q", "-u", "origin", "main")

	f.provider, err = todoruntime.New(ctx, opened.Gorm(), todoruntime.WorkspaceOptions{
		Name: "land", RootPath: f.repo, Repositories: []string{"acme/land"},
	})
	Expect(err).NotTo(HaveOccurred())
	f.todo, err = f.provider.Create(ctx, todos.CreateRequest{Title: "Add the feature", Status: types.StatusPending})
	Expect(err).NotTo(HaveOccurred())
	return f
}

// runWorktree reproduces what the run step leaves behind: the shell branch with
// a setup snapshot and the agent's commit, its worktree removed unless kept.
func (f *landFixture) runWorktree(keep bool) {
	GinkgoHelper()
	f.worktree = filepath.Join(filepath.Dir(f.repo), "shell-worktree")
	base := git(f.repo, "rev-parse", "HEAD")
	git(f.repo, "worktree", "add", "-q", "-b", shellBranch, f.worktree, base)
	setup := commitFile(f.worktree, "wip.txt", "in progress\n", snapshotMessage)
	head := commitFile(f.worktree, "feature.txt", "feature\n", agentMessage)
	f.state = api.WorktreeState{
		Repo: f.repo, Path: f.worktree, Branch: shellBranch, Base: base, Setup: setup, Head: head, Removed: !keep,
	}
	if !keep {
		git(f.repo, "worktree", "remove", f.worktree)
	}
}

// recordRun files a finished run step for the todo carrying f.state as its
// workspace, the way Captain's recorder persists it at finish.
func (f *landFixture) recordRun(ctx context.Context) uuid.UUID {
	GinkgoHelper()
	admission, err := runtimetest.Admit(ctx, f.provider, f.todo,
		todos.RunPreparation{Mode: types.ModeRun, Prompt: "run", ExecutorName: "claude"},
		promptrun.Completed{Runtime: runtimetest.ClaudeCLI, Outcome: promptrun.Outcome{State: captaindb.PromptRunStateSucceeded}})
	Expect(err).NotTo(HaveOccurred())
	run, err := f.provider.Captain().GetPromptRun(ctx, admission.PromptRunID)
	Expect(err).NotTo(HaveOccurred())
	state := f.state
	_, err = f.provider.Captain().UpdatePromptRun(ctx, captaindb.UpdatePromptRunInput{
		ID: run.ID, ExpectedVersion: run.Version,
		Workspace: &api.WorkspaceRecord{
			Cwd: f.worktree, Worktree: &state,
			Commits: []api.CommitRecord{{SHA: state.Head, Message: agentMessage}},
		},
	})
	Expect(err).NotTo(HaveOccurred())
	return run.ID
}
