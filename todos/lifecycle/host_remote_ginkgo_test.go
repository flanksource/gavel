// Tests a host whose runs execute elsewhere: the host hands its RemoteRunner the prompt run
// and the data gavel's hooks are rebuilt from, and classifies what comes back as a local run.
package lifecycle_test

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"

	captainai "github.com/flanksource/captain/pkg/ai"
	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/captain/pkg/promptrun"
	"github.com/flanksource/gavel/commit"
	"github.com/flanksource/gavel/pr/model"
	"github.com/flanksource/gavel/pr/provider"
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
	built, err := hooks.Build()
	Expect(err).NotTo(HaveOccurred())
	in := promptrun.Input{
		Resolved: run.Input.Resolved, Config: run.Input.Config, Provider: l.agent,
		Hooks: built, CallerOwnsCommits: run.Input.CallerOwnsCommits,
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

// capturingRunner keeps the remote run it is handed and runs nothing.
type capturingRunner struct{ runs []lifecycle.RemoteRun }

func (c *capturingRunner) Run(_ context.Context, run lifecycle.RemoteRun) (promptrun.Result, error) {
	c.runs = append(c.runs, run)
	return promptrun.Result{}, errors.New("not run")
}

// linkedPR is GitHub for the one open PR a todo is linked to.
type linkedPR struct{ provider.Client }

const linkedPRHead = "0123456789abcdef0123456789abcdef01234567"

func (linkedPR) Status(context.Context, int, model.StatusOptions) (*model.StatusSnapshot, error) {
	return &model.StatusSnapshot{PR: &model.PRInfo{
		Number: 7, URL: "https://github.com/acme/app/pull/7", State: "OPEN",
		HeadRefOID: linkedPRHead, HeadRefName: "topic", HeadRepositoryURL: "https://github.com/acme/app",
	}}, nil
}

var _ = Describe("a remote run on a linked PR", func() {
	It("resolves the PR through the host's client and hands the sandbox the PR's data", func() {
		GinkgoT().Setenv("HOME", GinkgoT().TempDir())
		host := newHost(&fakeProvider{plan: todos.PlanState{Exists: true, Approved: true, Content: "# plan", Revision: 2}})
		runner := &capturingRunner{}
		host.Remote = runner
		host.PRs = func(_ context.Context, url string) (provider.Client, int, error) {
			Expect(url).To(Equal("https://github.com/acme/app/pull/7"))
			return linkedPR{}, 7, nil
		}
		todo := hostTodo()
		todo.PR = &types.PR{Number: 7, URL: "https://github.com/acme/app/pull/7"}

		_, _ = host.RunStep(context.Background(), todo, stepNamed(host.Def, "run"), lifecycle.RunOptions{})

		Expect(runner.runs).To(HaveLen(1))
		pr := runner.runs[0].Hooks.PR
		Expect(pr).NotTo(BeNil())
		Expect(pr.URL).To(Equal("https://github.com/acme/app/pull/7"))
		Expect(pr.Number).To(Equal(7))
		Expect(pr.Target).To(Equal(provider.RunTarget{RepositoryURL: "https://github.com/acme/app", Branch: "topic", HeadSHA: linkedPRHead}))
		Expect(pr.Write).To(BeTrue(), "a run step writes to the PR")
		Expect(pr.WorkDir).To(Equal(host.WorkDir))
	})

	It("rebuilds the PR hook inside a sandbox job", func() {
		GinkgoT().Setenv("GAVEL_SERVER", "https://gavel.example")
		GinkgoT().Setenv("GAVEL_RUN_TOKEN", "run-token")
		GinkgoT().Setenv("GAVEL_RUN_ID", "run-1")
		hooks := lifecycle.RunHooks{PR: &lifecycle.PRHooks{
			URL: "https://github.com/acme/app/pull/7", Number: 7, Write: true, WorkDir: "/workspace/src",
			Target: provider.RunTarget{RepositoryURL: "https://github.com/acme/app", Branch: "topic", HeadSHA: linkedPRHead},
		}}
		built, err := hooks.Build()
		Expect(err).NotTo(HaveOccurred())
		var names []string
		for _, hook := range built {
			if named, ok := hook.(interface{ Name() string }); ok {
				names = append(names, named.Name())
			}
		}
		Expect(names).To(ContainElement("gavel-pr-turn"))
	})
})
