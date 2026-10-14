package runtime

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	captainai "github.com/flanksource/captain/pkg/ai"
	"github.com/flanksource/captain/pkg/api"
	captaindb "github.com/flanksource/captain/pkg/database"
	"github.com/flanksource/captain/pkg/promptrun"
	"github.com/flanksource/commons-db/dbtest"
	"github.com/flanksource/gavel/internal/database"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/lifecycle"
	"github.com/flanksource/gavel/todos/native"
	"github.com/flanksource/gavel/todos/runtime/runtimetest"
	"github.com/flanksource/gavel/todos/types"
	"github.com/flanksource/gavel/verify"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	_ "github.com/flanksource/gavel/fixtures/verifier"
)

// editingAgent is one completed turn that leaves an edit in the run's working
// tree, so the run's commit policy has a change to commit once the turn ends.
type editingAgent struct{ sessionID string }

func (a editingAgent) GetModel() string { return "scripted-model" }
func (a editingAgent) GetRuntime() api.Runtime {
	return api.RuntimeOf(api.Anthropic, api.ModeAgent)
}
func (a editingAgent) Execute(context.Context, api.Spec) (*api.Response, error) {
	return nil, fmt.Errorf("the scripted agent only streams")
}
func (a editingAgent) ExecuteStream(_ context.Context, req api.Spec) (<-chan captainai.Event, error) {
	if err := os.WriteFile(filepath.Join(req.Cwd(), "widget.txt"), []byte("built\n"), 0o644); err != nil {
		return nil, err
	}
	events := []captainai.Event{
		{Kind: captainai.EventSystem, SessionID: a.sessionID},
		{Kind: captainai.EventText, Text: `{"summary":"Built it.","endStatus":"completed"}`},
		{Kind: captainai.EventResult, Success: true, Usage: &api.Usage{InputTokens: 7, OutputTokens: 3}},
	}
	ch := make(chan captainai.Event, len(events))
	for _, event := range events {
		ch <- event
	}
	close(ch)
	return ch, nil
}

func gitRepo(dir string) {
	GinkgoHelper()
	for _, args := range [][]string{
		{"init", "-q", "-b", "main"},
		{"config", "user.email", "agent@example.com"},
		{"config", "user.name", "Agent"},
		{"commit", "-q", "--allow-empty", "-m", "initial"},
	} {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		out, err := cmd.CombinedOutput()
		Expect(err).NotTo(HaveOccurred(), string(out))
	}
}

// outcomeProvider is a native runtime over its own database, rooted at workDir.
func outcomeProvider(ctx context.Context, name, workDir string) *Provider {
	GinkgoHelper()
	db := dbtest.ForGinkgo(dbtest.Options{Name: "gavel_" + name})
	GinkgoT().Setenv(database.EnvDSN, db.DSN())
	GinkgoT().Setenv(database.EnvDisable, "")
	GinkgoT().Setenv(database.LegacyEnvDisable, "")
	GinkgoT().Setenv("HOME", GinkgoT().TempDir())
	opened, err := database.Open(ctx, database.WithMigrations())
	Expect(err).NotTo(HaveOccurred())
	DeferCleanup(func() { Expect(opened.Close()).To(Succeed()) })
	provider, err := New(ctx, opened.Gorm(), WorkspaceOptions{
		Name: name, RootPath: workDir, Repositories: []string{"acme/" + name},
	})
	Expect(err).NotTo(HaveOccurred())
	return provider
}

// A plan run Captain finished as succeeded can still fail once it is gavel's
// to persist. That failure is the attempt's outcome: recorded on the todo's
// history where the dashboard reads attempts, and projected as failed.
var _ = Describe("a run whose outcome could not be persisted", func() {
	It("records the failure as the attempt's outcome and lands the todo failed", func(ctx SpecContext) {
		workDir := GinkgoT().TempDir()
		provider := outcomeProvider(ctx, "outcome_not_persisted", workDir)
		def, err := lifecycle.Default()
		Expect(err).NotTo(HaveOccurred())
		engine, err := lifecycle.New(def)
		Expect(err).NotTo(HaveOccurred())
		host := &lifecycle.Host{Provider: provider, Def: engine, Config: verify.DefaultGavelConfig(), WorkDir: workDir, Kind: lifecycle.HostCLI}
		todo, err := provider.Create(ctx, todos.CreateRequest{Title: "Plan the widget", Body: "Design it", Status: types.StatusDraft})
		Expect(err).NotTo(HaveOccurred())
		missingPlan := filepath.Join(workDir, "missing-plan.md")
		execution := &todos.ExecutionResult{
			Success: true, ExecutorName: "headless-claude", EndStatus: types.EndCompleted,
			Summary: "The agent created a plan.", Plan: &types.PlanResult{Status: types.PlanNew, Path: missingPlan},
		}
		admission, err := runtimetest.Admit(ctx, provider, todo, todos.RunPreparation{Mode: types.ModePlan, Prompt: "plan", ExecutorName: "headless-claude"},
			promptrun.Completed{
				Runtime: runtimetest.ClaudeCLI, Model: "claude-sonnet", ProviderSessionID: "sess-plan-unpersisted",
				Outcome: lifecycle.RunOutcome(execution, false),
			})
		Expect(err).NotTo(HaveOccurred())
		plan, ok := engine.Definition().Step("plan")
		Expect(ok).To(BeTrue())
		outcome := &lifecycle.StepOutcome{
			Step: plan, Status: string(types.StatusReview), Execution: execution, Admission: admission,
			Result: lifecycle.StepResult{
				Run:      lifecycle.RunFacts{State: lifecycle.RunSucceeded},
				Envelope: lifecycle.Envelope{Summary: execution.Summary, EndStatus: "completed"},
				Plan:     &lifecycle.PlanFacts{Status: "new", Path: missingPlan},
			},
		}

		err = host.OnOutcome(ctx, todo, plan, outcome, outcome.Status)

		Expect(err).To(MatchError(ContainSubstring("reported plan path")))
		filed, getErr := provider.Captain().GetPromptRun(ctx, admission.PromptRunID)
		Expect(getErr).NotTo(HaveOccurred())
		Expect(filed.State).To(Equal(captaindb.PromptRunStateSucceeded), "Captain's record of the run is Captain's")
		Expect(todo.Status).To(Equal(types.StatusFailed))
		issue, getErr := provider.Repository().GetIssue(ctx, uuid.MustParse(todo.ID))
		Expect(getErr).NotTo(HaveOccurred())
		Expect(issue.ExecutionState).To(Equal(native.ExecutionFailed))
		events, getErr := provider.Repository().ListEvents(ctx, issue.ID)
		Expect(getErr).NotTo(HaveOccurred())
		var outcomes []map[string]any
		for _, event := range events {
			if event.Kind == lifecycle.EventLifecycleOutcome {
				var payload map[string]any
				Expect(json.Unmarshal(event.Payload, &payload)).To(Succeed())
				outcomes = append(outcomes, payload)
			}
		}
		Expect(outcomes).To(ConsistOf(And(
			HaveKeyWithValue("status", "failed"),
			HaveKeyWithValue("promptRunId", admission.PromptRunID.String()),
			HaveKeyWithValue("error", ContainSubstring("reported plan path")),
		)))
	})
})

// A run whose definition of done passed can still fail afterwards — its
// commit policy runs once verification is over. Such a run failed, and its
// todo must read failed, never unverified: verification is not what failed.
var _ = Describe("a run that fails after its definition of done passed", func() {
	It("lands the todo failed with the commit error, not unverified", func(ctx SpecContext) {
		workDir := GinkgoT().TempDir()
		gitRepo(workDir)
		provider := outcomeProvider(ctx, "run_post_verify_failure", workDir)

		// The run commits in place: no worktree, but the commit policy the
		// built-in step declares, which generates its message with a model this
		// environment has none of.
		def, err := lifecycle.Default()
		Expect(err).NotTo(HaveOccurred())
		for i := range def.Steps {
			if def.Steps[i].Name == "run" {
				def.Steps[i].Spec = &api.Spec{Workflow: &api.Workflow{
					Verify:  &api.Verify{Fixture: "{{subject.verification.document}}", Scope: api.VerifyScopeAll},
					Commits: []api.Commit{{On: api.CommitOnRun}},
				}}
			}
		}
		engine, err := lifecycle.New(def)
		Expect(err).NotTo(HaveOccurred())
		host := &lifecycle.Host{Provider: provider, Def: engine, Config: verify.DefaultGavelConfig(), WorkDir: workDir, Kind: lifecycle.HostCLI}
		todo, err := provider.Create(ctx, todos.CreateRequest{
			Title: "Build the widget", Body: "Build it", Status: types.StatusPending, Verification: "```bash\ntrue\n```",
		})
		Expect(err).NotTo(HaveOccurred())
		run, ok := host.Def.Definition().Step("run")
		Expect(ok).To(BeTrue())

		outcome, err := host.RunStep(ctx, todo, run, lifecycle.RunOptions{Provider: editingAgent{sessionID: uuid.NewString()}})

		Expect(err).NotTo(HaveOccurred())
		Expect(outcome.Result.Verify).To(HaveField("Passed", true), "the definition of done passed before the commit failed")
		Expect(outcome.Result.Run.Error).To(ContainSubstring(`run hook "commit:run"`))
		Expect(outcome.Status).To(Equal(string(types.StatusFailed)))
		filed, err := provider.Captain().GetPromptRun(ctx, outcome.Admission.PromptRunID)
		Expect(err).NotTo(HaveOccurred())
		Expect(filed).To(And(
			HaveField("State", captaindb.PromptRunStateFailed),
			HaveField("Phase", captaindb.PromptRunPhasePostRun),
		))

		Expect(host.OnOutcome(ctx, todo, run, outcome, outcome.Status)).To(Succeed())

		Expect(todo.Status).To(Equal(types.StatusFailed))
		issue, err := provider.Repository().GetIssue(ctx, uuid.MustParse(todo.ID))
		Expect(err).NotTo(HaveOccurred())
		Expect(issue.ExecutionState).To(Equal(native.ExecutionFailed))
	})
})
