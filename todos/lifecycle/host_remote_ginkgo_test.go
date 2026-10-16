// Tests a host whose runs execute elsewhere: the host hands its RemoteRunner the prompt run
// and the data gavel's hooks are rebuilt from, and classifies what comes back as a local run.
package lifecycle_test

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"

	captainai "github.com/flanksource/captain/pkg/ai"
	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/captain/pkg/promptrun"
	"github.com/flanksource/gavel/commit"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/lifecycle"
	"github.com/flanksource/gavel/todos/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// loopbackRunner executes a remote run in this process the way a sandbox job does: gavel's
// hooks rebuilt from their data, nothing recorded, the agent supplied by the test.
type loopbackRunner struct {
	agent captainai.Provider
	runs  []lifecycle.RemoteRun
}

func (l *loopbackRunner) Run(ctx context.Context, run lifecycle.RemoteRun) (promptrun.Result, error) {
	l.runs = append(l.runs, run)
	data, err := json.Marshal(run.Hooks)
	Expect(err).NotTo(HaveOccurred())
	var hooks lifecycle.RunHooks
	Expect(json.Unmarshal(data, &hooks)).To(Succeed(), "the hook data crosses to the sandbox as JSON")
	in := promptrun.Input{
		Resolved: run.Input.Resolved, Config: run.Input.Config, Provider: l.agent,
		Hooks: hooks.Build(), CallerOwnsCommits: run.Input.CallerOwnsCommits,
		Verify: run.Input.Verify, Timeout: run.Input.Timeout, Repo: run.Input.Repo,
		OnEvent: run.Input.OnEvent,
	}
	return promptrun.Run(ctx, in)
}

var _ = Describe("a host whose runs execute elsewhere", func() {
	var (
		host *lifecycle.Host
		ctx  context.Context
	)
	BeforeEach(func() {
		GinkgoT().Setenv("HOME", GinkgoT().TempDir())
		host = newHost(&fakeProvider{plan: todos.PlanState{Exists: true, Approved: true, Content: "# plan", Revision: 2}})
		ctx = context.Background()
		// The run step's commit hook commits in the work dir, so it is a repository.
		for _, args := range [][]string{{"init", "-q", "-b", "main"}, {"commit", "-q", "--allow-empty", "-m", "init"}} {
			cmd := exec.Command("git", args...)
			cmd.Dir = host.WorkDir
			cmd.Env = append(os.Environ(), "GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@example.com", "GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@example.com")
			out, err := cmd.CombinedOutput()
			Expect(err).NotTo(HaveOccurred(), string(out))
		}
		def := host.Def.Definition()
		for i := range def.Steps {
			if def.Steps[i].Name == "run" {
				def.Steps[i].Spec = &api.Spec{Workflow: &api.Workflow{
					Verify:  &api.Verify{Fixture: "{{subject.verification.document}}", Scope: api.VerifyScopeAll},
					Commits: []api.Commit{{On: api.CommitOnTurn, Mode: api.CommitModeCommit, Stage: api.CommitStageWorktree}},
				}}
			}
		}
		engine, err := lifecycle.New(def)
		Expect(err).NotTo(HaveOccurred())
		host.Def = engine
	})

	It("hands the run and its hook data to the runner and classifies the result", func() {
		runner := &loopbackRunner{agent: &scriptedProvider{events: scriptedTurn(`{"summary":"Built it.","endStatus":"completed"}`, true)}}
		host.Remote = runner
		todo := hostTodo()

		outcome, err := host.RunStep(ctx, todo, stepNamed(host.Def, "run"), lifecycle.RunOptions{})

		Expect(err).NotTo(HaveOccurred())
		Expect(outcome.Status).To(Equal("verified"), outcome.Execution.ErrorMessage)
		Expect(outcome.Execution.Summary).To(Equal("Built it."))
		Expect(outcome.Execution.CostUSD).To(Equal(0.25))
		Expect(runner.runs).To(HaveLen(1))
		hooks := runner.runs[0].Hooks
		Expect(hooks.Commits).To(HaveLen(1))
		Expect(hooks.Meta).To(Equal(commit.AgentRunMetadata{IssueID: todo.ID}))
		Expect(hooks.Env).To(HaveKeyWithValue(commit.EnvIssueID, todo.ID))
		Expect(hooks.Labels).To(HaveKeyWithValue("todo", todo.ID))
		Expect(runner.runs[0].Input.Resolved.Spec.Prompt.User).To(ContainSubstring("Implement the thing"))
	})

	It("refuses a step whose tools read todos through the provider", func() {
		host.Remote = &loopbackRunner{}
		_, err := host.RunStep(ctx, hostTodo(), stepNamed(host.Def, "triage.new"), lifecycle.RunOptions{})
		Expect(err).To(MatchError("step triage.new reads todos while it runs, which a run executing elsewhere can't do yet"))
	})

	It("refuses to resume a session", func() {
		host.Remote = &loopbackRunner{}
		_, err := host.RunStep(ctx, hostTodo(), stepNamed(host.Def, "run"), lifecycle.RunOptions{Resume: true})
		Expect(err).To(MatchError("step run resumes an agent session, which a run executing elsewhere can't do yet: start a fresh run"))
	})

	It("refuses a run in a terminal on this host", func() {
		host.Remote = &loopbackRunner{}
		_, err := host.RunStep(ctx, hostTodo(), stepNamed(host.Def, "run"), lifecycle.RunOptions{Request: api.Spec{Model: api.Model{Mode: api.ModeCmux}}})
		Expect(err).To(MatchError("step run runs in cmux, a terminal on this host, which a run executing elsewhere can't use: choose the agent or cli mode"))
	})

	It("refuses a step on a todo linked to a PR", func() {
		host.Remote = &loopbackRunner{}
		todo := hostTodo()
		todo.PR = &types.PR{Number: 7, URL: "https://github.com/acme/app/pull/7"}
		_, err := host.RunStep(ctx, todo, stepNamed(host.Def, "run"), lifecycle.RunOptions{})
		Expect(err).To(MatchError(ContainSubstring("can't run elsewhere yet")))
	})
})

var _ = Describe("a host for a hosted repository", func() {
	It("reads the repository's configuration alone", func() {
		home, repo := GinkgoT().TempDir(), GinkgoT().TempDir()
		GinkgoT().Setenv("HOME", home)
		Expect(os.WriteFile(filepath.Join(home, ".gavel.yaml"), []byte("ssh:\n  cmd: make deploy-from-the-host\n"), 0o644)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(repo, ".gavel.yaml"), []byte("ssh:\n  cmd: make check\n"), 0o644)).To(Succeed())

		host, err := lifecycle.NewRepositoryHost(&fakeProvider{}, repo, lifecycle.HostDashboard)

		Expect(err).NotTo(HaveOccurred())
		Expect(host.Config.SSH.Cmd).To(Equal("make check"))
		Expect(host.WorkDir).To(Equal(repo))
		Expect(*host.Saved).To(BeZero(), "the serve host's Captain defaults don't apply")
	})
})
