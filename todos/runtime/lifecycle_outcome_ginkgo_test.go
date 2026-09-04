package runtime

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/flanksource/captain/pkg/api"
	captaindb "github.com/flanksource/captain/pkg/database"
	commonsdb "github.com/flanksource/commons-db/db"
	"github.com/flanksource/gavel/internal/database"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/lifecycle"
	"github.com/flanksource/gavel/todos/native"
	"github.com/flanksource/gavel/todos/types"
	"github.com/flanksource/gavel/verify"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// The host's OnOutcome is the single status writer for a step run. These specs
// drive it against the native runtime: the attempt lands on Captain's prompt
// run, a plan step persists its revision, and the status the lifecycle chose is
// what the todo projects afterwards.
var _ = Describe("lifecycle outcomes on the native runtime", Ordered, func() {
	var (
		ctx      context.Context
		provider *Provider
		host     *lifecycle.Host
	)

	issueOf := func(todo *types.TODO) *native.Issue {
		GinkgoHelper()
		id, err := uuid.Parse(todo.ID)
		Expect(err).NotTo(HaveOccurred())
		issue, err := provider.repository.GetIssue(ctx, id)
		Expect(err).NotTo(HaveOccurred())
		return issue
	}

	stepNamed := func(name string) lifecycle.Step {
		GinkgoHelper()
		step, ok := host.Def.Definition().Step(name)
		Expect(ok).To(BeTrue(), "step %s", name)
		return step
	}

	// admit is what Host.RunStep does before dispatch: the run is prepared under
	// its lifecycle identity and its runtime recorded.
	admit := func(todo *types.TODO, step string, mode types.RunMode, session string) todos.RunPreparationResult {
		GinkgoHelper()
		admission, err := provider.PrepareRun(ctx, todo, todos.RunPreparation{Mode: mode, Prompt: step, ExecutorName: "headless-claude"})
		Expect(err).NotTo(HaveOccurred())
		Expect(provider.RecordRunStart(ctx, todo, todos.RunStartMetadata{
			SessionID: session, Mode: string(mode), Driver: "headless-claude", Agent: "claude",
			Provider: "anthropic", RuntimeMode: "agent", ResolvedModel: "claude-sonnet", Effort: "medium",
		})).To(Succeed())
		return admission
	}

	// outcomeOf classifies a synthetic result the way RunStep does.
	outcomeOf := func(todo *types.TODO, step lifecycle.Step, execution *todos.ExecutionResult, facts lifecycle.StepResult) (*lifecycle.StepOutcome, string) {
		GinkgoHelper()
		lc, err := host.Context(ctx, todo)
		Expect(err).NotTo(HaveOccurred())
		status, err := host.Def.Outcome(step, lc, facts)
		Expect(err).NotTo(HaveOccurred())
		return &lifecycle.StepOutcome{Step: step, Status: status, Result: facts, Execution: execution}, status
	}

	lifecycleEvents := func(todo *types.TODO) []native.Event {
		GinkgoHelper()
		events, err := provider.repository.ListEvents(ctx, issueOf(todo).ID)
		Expect(err).NotTo(HaveOccurred())
		var matched []native.Event
		for _, event := range events {
			if event.Kind == lifecycle.EventLifecycleOutcome {
				matched = append(matched, event)
			}
		}
		return matched
	}

	payloadOf := func(event native.Event) map[string]any {
		GinkgoHelper()
		var payload map[string]any
		Expect(json.Unmarshal(event.Payload, &payload)).To(Succeed())
		return payload
	}

	BeforeAll(func() {
		if os.Getenv("GAVEL_DB_EMBEDDED_TEST") == "" {
			Skip("set GAVEL_DB_EMBEDDED_TEST=1 to run embedded-postgres native runtime tests")
		}
		ctx = context.Background()
		dsn, stop, err := commonsdb.StartEmbedded(commonsdb.EmbeddedConfig{
			DataDir: filepath.Join(GinkgoT().TempDir(), "postgres"), Database: "gavel_todo_lifecycle_outcome",
		})
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(func() { Expect(stop()).To(Succeed()) })

		GinkgoT().Setenv(database.EnvDSN, dsn)
		GinkgoT().Setenv(database.EnvDisable, "")
		GinkgoT().Setenv(database.LegacyEnvDSN, "")
		GinkgoT().Setenv(database.LegacyEnvDisable, "")
		GinkgoT().Setenv("HOME", GinkgoT().TempDir())

		opened, err := database.Open(ctx, database.WithMigrations())
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(func() { Expect(opened.Close()).To(Succeed()) })

		workDir := GinkgoT().TempDir()
		provider, err = New(ctx, opened.Gorm(), WorkspaceOptions{
			Name: "lifecycle-outcome", RootPath: workDir, Repositories: []string{"acme/lifecycle-outcome"},
		})
		Expect(err).NotTo(HaveOccurred())

		def, err := lifecycle.Default()
		Expect(err).NotTo(HaveOccurred())
		engine, err := lifecycle.New(def)
		Expect(err).NotTo(HaveOccurred())
		host = &lifecycle.Host{Provider: provider, Def: engine, Config: verify.DefaultGavelConfig(), WorkDir: workDir, Kind: lifecycle.HostCLI}
	})

	var planned *types.TODO

	It("lands a new plan in review with its revision persisted", func() {
		todo, err := provider.Create(ctx, todos.CreateRequest{Title: "Plan the widget", Body: "Design it", Status: types.StatusDraft})
		Expect(err).NotTo(HaveOccurred())
		admit(todo, "plan", types.ModePlan, "sess-plan-new")
		const markdown = "# Plan\n\n1. Build the widget."
		execution := &todos.ExecutionResult{
			Success: true, ExecutorName: "headless-claude", EndStatus: types.EndCompleted,
			Summary: "The agent created a plan.", Plan: &types.PlanResult{Status: types.PlanNew, Content: markdown},
		}
		facts := lifecycle.StepResult{
			Run:      lifecycle.RunFacts{State: lifecycle.RunSucceeded},
			Envelope: lifecycle.Envelope{Summary: execution.Summary, EndStatus: "completed"},
			Plan:     &lifecycle.PlanFacts{Status: "new", Content: markdown},
		}
		outcome, status := outcomeOf(todo, stepNamed("plan"), execution, facts)
		Expect(status).To(Equal(string(types.StatusReview)))

		Expect(host.OnOutcome(ctx, todo, stepNamed("plan"), outcome, status)).To(Succeed())

		Expect(todo.Status).To(Equal(types.StatusReview))
		issue := issueOf(todo)
		Expect(issue.SelectedPlanID).NotTo(BeNil())
		plan, err := provider.captain.GetPlan(ctx, *issue.SelectedPlanID)
		Expect(err).NotTo(HaveOccurred())
		Expect(plan.LatestRevision).NotTo(BeNil())
		Expect(plan.LatestRevision.Revision).To(Equal(1))
		Expect(plan.LatestRevision.PlanMarkdown).To(Equal(markdown))
		events := lifecycleEvents(todo)
		Expect(events).To(HaveLen(1))
		Expect(payloadOf(events[0])).To(HaveKeyWithValue("status", "review"))
		planned = todo
	})

	It("keeps an unchanged approved plan and returns the todo to pending", func() {
		approved, err := provider.ApprovePlan(ctx, planned, "moshe", "looks right")
		Expect(err).NotTo(HaveOccurred())
		Expect(approved.Status).To(Equal(types.StatusPending))
		admit(approved, "plan", types.ModePlan, "sess-plan-unchanged")
		execution := &todos.ExecutionResult{
			Success: true, ExecutorName: "headless-claude", EndStatus: types.EndCompleted,
			Summary: "The existing plan is unchanged.", Plan: &types.PlanResult{Status: types.PlanUnchanged},
		}
		facts := lifecycle.StepResult{
			Run:      lifecycle.RunFacts{State: lifecycle.RunSucceeded},
			Envelope: lifecycle.Envelope{Summary: execution.Summary, EndStatus: "completed"},
			Plan:     &lifecycle.PlanFacts{Status: "unchanged"},
		}
		outcome, status := outcomeOf(approved, stepNamed("plan"), execution, facts)
		Expect(status).To(Equal(string(types.StatusPending)))

		Expect(host.OnOutcome(ctx, approved, stepNamed("plan"), outcome, status)).To(Succeed())

		Expect(approved.Status).To(Equal(types.StatusPending))
		plan, err := provider.captain.GetPlan(ctx, *issueOf(approved).SelectedPlanID)
		Expect(err).NotTo(HaveOccurred())
		Expect(plan.LatestRevision.Revision).To(Equal(1), "an unchanged plan appends no revision")
		Expect(lifecycleEvents(approved)).To(HaveLen(2))
	})

	It("parks a run that asks questions, and a failed resume leaves it parked", func() {
		todo, err := provider.Create(ctx, todos.CreateRequest{Title: "Ask about the database", Body: "Migrate it", Status: types.StatusPending})
		Expect(err).NotTo(HaveOccurred())
		admit(todo, "plan", types.ModePlan, "sess-ask")
		execution := &todos.ExecutionResult{
			Success: true, ExecutorName: "headless-claude", EndStatus: types.EndAsk,
			Summary:   "Which database should the migration target?",
			Questions: []types.AgentQuestion{{Text: "Which database should the migration target?"}},
		}
		facts := lifecycle.StepResult{
			Run:       lifecycle.RunFacts{State: lifecycle.RunWaiting},
			Envelope:  lifecycle.Envelope{Summary: execution.Summary, EndStatus: "ask"},
			Questions: []any{map[string]any{"text": execution.Questions[0].Text}},
		}
		outcome, status := outcomeOf(todo, stepNamed("plan"), execution, facts)
		Expect(status).To(Equal(string(types.StatusAsk)))

		Expect(host.OnOutcome(ctx, todo, stepNamed("plan"), outcome, status)).To(Succeed())

		Expect(todo.Status).To(Equal(types.StatusAsk))
		Expect(todo.Questions).To(HaveLen(1))
		run, err := provider.captain.GetPromptRun(ctx, *issueOf(todo).ActivePromptRunID)
		Expect(err).NotTo(HaveOccurred())
		Expect(run.State).To(Equal(captaindb.PromptRunStateWaiting))

		// Admission is the seam a resume fails at before anything is dispatched:
		// a mismatched continuation is refused, and the todo stays parked rather
		// than being marked failed for a run that never started (55fa405b).
		_, err = provider.PrepareRun(ctx, todo, todos.RunPreparation{
			Mode: types.ModeRun, Prompt: "run", ExecutorName: "headless-claude", Resume: true,
		})
		Expect(err).To(MatchError(todos.ErrRunResumeModeMismatch))
		Expect(provider.reloadTODO(ctx, todo, todo.CWD)).To(Succeed())
		Expect(todo.Status).To(Equal(types.StatusAsk))
		Expect(lifecycleEvents(todo)).To(HaveLen(1))
	})

	It("verifies a run whose definition of done passed", func() {
		todo, err := provider.Create(ctx, todos.CreateRequest{
			Title: "Ship the widget", Body: "Build it\n\n## Verification\n\n```bash\ntrue\n```\n", Status: types.StatusPending,
		})
		Expect(err).NotTo(HaveOccurred())
		admit(todo, "run", types.ModeRun, "sess-run-pass")
		report := &api.VerifyReport{Ran: true, Passed: true}
		execution := &todos.ExecutionResult{
			Success: true, ExecutorName: "headless-claude", EndStatus: types.EndCompleted, Summary: "Built the widget.",
			DoD: &todos.DoDOutcome{Ran: true, Passed: true, Report: report},
		}
		facts := lifecycle.StepResult{
			Run:      lifecycle.RunFacts{State: lifecycle.RunSucceeded},
			Envelope: lifecycle.Envelope{Summary: execution.Summary, EndStatus: "completed"},
			Verify:   report,
		}
		outcome, status := outcomeOf(todo, stepNamed("run"), execution, facts)
		Expect(status).To(Equal(string(types.StatusVerified)))

		Expect(host.OnOutcome(ctx, todo, stepNamed("run"), outcome, status)).To(Succeed())

		Expect(todo.Status).To(Equal(types.StatusVerified))
		Expect(issueOf(todo).Status).To(Equal(native.StatusVerified))
		Expect(lifecycleEvents(todo)).To(HaveLen(1))
	})

	It("leaves a run whose definition of done failed unverified", func() {
		todo, err := provider.Create(ctx, todos.CreateRequest{
			Title: "Ship the gadget", Body: "Build it\n\n## Verification\n\n```bash\nfalse\n```\n", Status: types.StatusPending,
		})
		Expect(err).NotTo(HaveOccurred())
		admit(todo, "run", types.ModeRun, "sess-run-fail")
		report := &api.VerifyReport{Ran: true, Passed: false, Reason: "1 check failed"}
		execution := &todos.ExecutionResult{
			Success: true, ExecutorName: "headless-claude", EndStatus: types.EndCompleted, Summary: "Built the gadget.",
			DoD: &todos.DoDOutcome{Ran: true, Passed: false, Report: report},
		}
		facts := lifecycle.StepResult{
			Run:      lifecycle.RunFacts{State: lifecycle.RunSucceeded},
			Envelope: lifecycle.Envelope{Summary: execution.Summary, EndStatus: "completed"},
			Verify:   report,
		}
		outcome, status := outcomeOf(todo, stepNamed("run"), execution, facts)
		Expect(status).To(Equal(string(types.StatusUnverified)))

		Expect(host.OnOutcome(ctx, todo, stepNamed("run"), outcome, status)).To(Succeed())

		Expect(todo.Status).To(Equal(types.StatusUnverified))
		issue := issueOf(todo)
		Expect(issue.Status).To(Equal(native.StatusOpen))
		Expect(issue.ExecutionState).To(Equal(native.ExecutionVerificationFailed))
	})

	It("never writes a status for a triage verdict of done", func() {
		todo, err := provider.Create(ctx, todos.CreateRequest{Title: "Triage me", Body: "Already shipped", Status: types.StatusPending})
		Expect(err).NotTo(HaveOccurred())
		admit(todo, "triage", types.ModePlan, "sess-triage")
		execution := &todos.ExecutionResult{
			Success: true, ExecutorName: "headless-claude", EndStatus: types.EndCompleted, Summary: "This already landed.",
			Triage: &types.TriageEnvelope{
				ResultEnvelope: types.ResultEnvelope{Summary: "This already landed.", EndStatus: types.EndCompleted},
				Verdict:        types.VerdictDone, Status: string(types.StatusCompleted),
			},
		}
		facts := lifecycle.StepResult{
			Run:      lifecycle.RunFacts{State: lifecycle.RunSucceeded},
			Envelope: lifecycle.Envelope{Summary: execution.Summary, EndStatus: "completed"},
		}
		outcome, status := outcomeOf(todo, stepNamed("triage"), execution, facts)
		Expect(status).To(Equal(lifecycle.OutcomeKeep))

		Expect(host.OnOutcome(ctx, todo, stepNamed("triage"), outcome, status)).To(Succeed())

		Expect(todo.Status).To(Equal(types.StatusPending), "a done verdict is a claim the definition of done must prove")
		Expect(issueOf(todo).Status).To(Equal(native.StatusOpen))
		events := lifecycleEvents(todo)
		Expect(events).To(HaveLen(1))
		Expect(payloadOf(events[0])).To(HaveKeyWithValue("status", lifecycle.OutcomeKeep))
	})

	It("lists the run history oldest first, each run under its step with the status it landed", func() {
		todo, err := provider.Create(ctx, todos.CreateRequest{
			Title: "Trace the widget", Body: "Build it\n\n## Verification\n\n```bash\ntrue\n```\n", Status: types.StatusDraft,
		})
		Expect(err).NotTo(HaveOccurred())
		planAdmission := admit(todo, "plan", types.ModePlan, "sess-history-plan")
		planExecution := &todos.ExecutionResult{
			Success: true, ExecutorName: "headless-claude", EndStatus: types.EndCompleted,
			Summary: "The agent created a plan.", Plan: &types.PlanResult{Status: types.PlanNew, Content: "# Plan\n\n1. Trace it."},
		}
		planOutcome, status := outcomeOf(todo, stepNamed("plan"), planExecution, lifecycle.StepResult{
			Run:      lifecycle.RunFacts{State: lifecycle.RunSucceeded},
			Envelope: lifecycle.Envelope{Summary: planExecution.Summary, EndStatus: "completed"},
			Plan:     &lifecycle.PlanFacts{Status: "new", Content: "# Plan\n\n1. Trace it."},
		})
		planOutcome.Admission = planAdmission
		Expect(host.OnOutcome(ctx, todo, stepNamed("plan"), planOutcome, status)).To(Succeed())
		Expect(todo.Status).To(Equal(types.StatusReview))

		approved, err := provider.ApprovePlan(ctx, todo, "moshe", "go")
		Expect(err).NotTo(HaveOccurred())
		runAdmission := admit(approved, "run", types.ModeRun, "sess-history-run")
		report := &api.VerifyReport{Ran: true, Passed: true}
		runExecution := &todos.ExecutionResult{
			Success: true, ExecutorName: "headless-claude", EndStatus: types.EndCompleted, Summary: "Traced it.",
			DoD: &todos.DoDOutcome{Ran: true, Passed: true, Report: report},
		}
		runOutcome, status := outcomeOf(approved, stepNamed("run"), runExecution, lifecycle.StepResult{
			Run:      lifecycle.RunFacts{State: lifecycle.RunSucceeded},
			Envelope: lifecycle.Envelope{Summary: runExecution.Summary, EndStatus: "completed"},
			Verify:   report,
		})
		runOutcome.Admission = runAdmission
		Expect(host.OnOutcome(ctx, approved, stepNamed("run"), runOutcome, status)).To(Succeed())
		Expect(approved.Status).To(Equal(types.StatusVerified))

		history, err := provider.RunHistory(ctx, approved)

		Expect(err).NotTo(HaveOccurred())
		Expect(history).To(HaveLen(2))
		Expect(history[0].Step).To(Equal("plan"))
		Expect(history[0].PromptRunID).To(Equal(planAdmission.PromptRunID.String()))
		Expect(history[0].State).To(Equal(string(captaindb.PromptRunStateSucceeded)))
		Expect(history[0].Outcome).To(Equal("review"))
		Expect(history[0].FinishedAt).NotTo(BeNil())
		Expect(history[1].Step).To(Equal("run"))
		Expect(history[1].PromptRunID).To(Equal(runAdmission.PromptRunID.String()))
		Expect(history[1].Outcome).To(Equal("verified"))

		states, err := host.Steps(ctx, approved)
		Expect(err).NotTo(HaveOccurred())
		for _, state := range states {
			if state.Step.Name != "run" {
				continue
			}
			Expect(state.LastRun).NotTo(BeNil())
			Expect(state.LastRun.Outcome).To(Equal("verified"))
			Expect(state.LastRun.PromptRunID).To(Equal(runAdmission.PromptRunID.String()))
			Expect(state.Done).To(BeTrue())
		}
	})
})
