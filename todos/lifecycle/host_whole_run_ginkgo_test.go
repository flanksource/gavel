// Tests a local step whose sandbox runs whole runs elsewhere (substrate): the step's selection
// comes from the saved Captain config, and the run carries the data gavel's hooks are rebuilt from.
package lifecycle_test

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"

	captainai "github.com/flanksource/captain/pkg/ai"
	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/captain/pkg/captainconfig"
	"github.com/flanksource/captain/pkg/promptrun"
	"github.com/flanksource/captain/pkg/sandbox/substraterun"
	lifecycleengine "github.com/flanksource/gavel/lifecycle"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/lifecycle"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// environmentSandbox stands in for the substrate sandbox: it runs a whole run in this process as
// the environment does, with gavel's hooks rebuilt from the run's host data and the agent
// supplied by the test, and keeps the input it was handed.
type environmentSandbox struct {
	agent captainai.Provider
	runs  []promptrun.Input
}

func (s *environmentSandbox) Kind() api.SandboxKind { return api.SandboxSubstrate }
func (s *environmentSandbox) Prepare(context.Context, *api.Spec) (*api.SandboxSession, error) {
	return &api.SandboxSession{}, nil
}
func (s *environmentSandbox) Close() error              { return nil }
func (s *environmentSandbox) IsolatesWorkspace() bool   { return true }
func (s *environmentSandbox) ProvidesEgressProxy() bool { return true }
func (s *environmentSandbox) RunRemote(ctx context.Context, in promptrun.Input, started func(api.Spec) error) (promptrun.Result, error) {
	s.runs = append(s.runs, in)
	var hooks lifecycle.RunHooks
	Expect(json.Unmarshal(in.HostData, &hooks)).To(Succeed(), "the run carries gavel's hooks as data")
	built, _, err := hooks.Build()
	Expect(err).NotTo(HaveOccurred())
	Expect(started(in.Resolved.Spec)).To(Succeed())
	return promptrun.Run(ctx, promptrun.Input{
		Resolved: in.Resolved, Config: captainai.Config{Model: in.Config.Model}, Provider: s.agent,
		Hooks: built, CallerOwnsCommits: in.CallerOwnsCommits, Verify: in.Verify, Timeout: in.Timeout, Repo: in.Repo,
		OnEvent: in.OnEvent,
	})
}

var _ = Describe("a local step on a sandbox that runs it whole", func() {
	var (
		host    *lifecycle.Host
		sandbox *environmentSandbox
	)
	BeforeEach(func() {
		GinkgoT().Setenv("HOME", GinkgoT().TempDir())
		host = newHost(&fakeProvider{plan: todos.PlanState{Exists: true, Approved: true, Content: "# plan", Revision: 2}})
		for _, args := range [][]string{{"init", "-q", "-b", "main"}, {"commit", "-q", "--allow-empty", "-m", "init"}} {
			cmd := exec.Command("git", args...)
			cmd.Dir = host.WorkDir
			cmd.Env = append(os.Environ(), "GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@example.com", "GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@example.com")
			out, err := cmd.CombinedOutput()
			Expect(err).NotTo(HaveOccurred(), string(out))
		}
		host.Saved = &captainconfig.Config{Sandbox: captainconfig.SandboxDefaults{Backends: map[string]captainconfig.SandboxBackend{
			"beta-substrate": {Kind: string(api.SandboxSubstrate), Options: map[string]any{"envAPI": "env.example:443"}},
		}}}
		def := host.Def.Definition()
		for i := range def.Steps {
			if def.Steps[i].Name == "run" {
				// Its tools ask for approvals, which need a recorded run; this host records none.
				def.Steps[i].Tools = nil
				def.Steps[i].Spec = &api.Spec{
					Sandbox:  &api.SandboxRef{Backend: "beta-substrate"},
					Workflow: &api.Workflow{Verify: &api.Verify{Fixture: "{{subject.verification.document}}", Scope: api.VerifyScopeAll}},
				}
			}
		}
		engine, err := lifecycleengine.New(def)
		Expect(err).NotTo(HaveOccurred())
		host.Def = engine
		sandbox = &environmentSandbox{agent: &scriptedProvider{events: scriptedTurn(`{"success":true,"event":"completed","comment":"Built it."}`, true)}}
		api.RegisterSandbox(api.SandboxSubstrate, func(config api.SandboxConfig) (api.Sandbox, error) {
			Expect(config.Name).To(Equal("beta-substrate"))
			Expect(config.Options).To(HaveKeyWithValue("envAPI", "env.example:443"), "the backend's options come from the saved Captain config")
			return sandbox, nil
		})
		DeferCleanup(func() { api.RegisterSandbox(api.SandboxSubstrate, substraterun.Substrate) })
	})

	It("hands the run, with gavel's hook data, to the backend the step selects", func() {
		todo := hostTodo()

		outcome, err := host.RunStep(context.Background(), todo, stepNamed(host.Def, "run"), lifecycle.RunOptions{})

		Expect(err).NotTo(HaveOccurred())
		Expect(outcome.Status).To(Equal("verified"), outcome.Execution.ErrorMessage)
		Expect(outcome.Execution.Summary).To(Equal("Built it."))
		Expect(sandbox.runs).To(HaveLen(1))
		var hooks lifecycle.RunHooks
		Expect(json.Unmarshal(sandbox.runs[0].HostData, &hooks)).To(Succeed())
		Expect(hooks.Env).To(HaveKeyWithValue("GAVEL_ISSUE_ID", todo.ID))
		Expect(hooks.Labels).To(HaveKeyWithValue("todo", todo.ID))
	})
})
