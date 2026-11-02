package runtime

import (
	"context"
	"encoding/json"

	"github.com/flanksource/captain/pkg/api"
	captaindb "github.com/flanksource/captain/pkg/database"
	"github.com/flanksource/captain/pkg/promptrun"
	"github.com/flanksource/commons-db/dbtest"
	"github.com/flanksource/gavel/internal/database"
	"github.com/flanksource/gavel/lifecycle"
	"github.com/flanksource/gavel/todos"
	lifecycle1 "github.com/flanksource/gavel/todos/lifecycle"
	"github.com/flanksource/gavel/todos/native"
	"github.com/flanksource/gavel/todos/runtime/runtimetest"
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
		host     *lifecycle1.Host
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

	// admit is what Host.RunStep leaves behind before its outcome is applied: the
	// run admitted under its lifecycle identity, and filed by Captain in the
	// terminal state the execution maps to.
	admit := func(todo *types.TODO, step string, mode types.RunMode, session string, execution *todos.ExecutionResult) todos.RunPreparationResult {
		GinkgoHelper()
		admission, err := runtimetest.Admit(ctx, provider, todo, todos.RunPreparation{Mode: mode, Prompt: step, ExecutorName: "headless-claude"},
			promptrun.Completed{
				Runtime: runtimetest.ClaudeCLI, Model: "claude-sonnet", ProviderSessionID: session,
				Outcome: lifecycle1.RunOutcome(execution, mode == types.ModeVerify),
			})
		Expect(err).NotTo(HaveOccurred())
		return admission
	}
	// succeeded is a run that returned a successful envelope with one event.
	succeeded := func(event, comment string, verify *api.VerifyReport) lifecycle.StepResult {
		return lifecycle.StepResult{
			Run: lifecycle.RunFacts{State: lifecycle1.RunSucceeded}, Verify: verify,
			Envelope: lifecycle.Envelope{Success: true, Event: event, Comment: comment},
		}
	}

	// outcomeOf classifies a synthetic result the way RunStep does.
	outcomeOf := func(todo *types.TODO, step lifecycle.Step, execution *todos.ExecutionResult, facts lifecycle.StepResult) (*lifecycle1.StepOutcome, string) {
		GinkgoHelper()
		lc, err := host.Context(ctx, todo)
		Expect(err).NotTo(HaveOccurred())
		status, err := host.Def.Outcome(step, lc, facts)
		Expect(err).NotTo(HaveOccurred())
		return &lifecycle1.StepOutcome{Step: step, Status: status, Result: facts, Execution: execution}, status
	}

	lifecycleEvents := func(todo *types.TODO) []native.Event {
		GinkgoHelper()
		events, err := provider.repository.ListEvents(ctx, issueOf(todo).ID)
		Expect(err).NotTo(HaveOccurred())
		var matched []native.Event
		for _, event := range events {
			if event.Kind == lifecycle1.EventLifecycleOutcome {
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
		ctx = context.Background()
		dsn := dbtest.ForGinkgo(dbtest.Options{Name: "gavel_todo_lifecycle_outcome"}).DSN()

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

		def, err := lifecycle1.Default()
		Expect(err).NotTo(HaveOccurred())
		engine, err := lifecycle.New(def)
		Expect(err).NotTo(HaveOccurred())
		host = &lifecycle1.Host{Provider: provider, Def: engine, Config: verify.DefaultGavelConfig(), WorkDir: workDir, Kind: lifecycle1.HostCLI}
	})

	var planned *types.TODO

	It("lands a new plan in review with its revision persisted", func() {
		todo, err := provider.Create(ctx, todos.CreateRequest{Title: "Plan the widget", Body: "Design it", Status: types.StatusDraft})
		Expect(err).NotTo(HaveOccurred())
		const markdown = "# Plan\n\n1. Build the widget."
		execution := &todos.ExecutionResult{
			Success: true, ExecutorName: "headless-claude",
			Summary: "The agent created a plan.", Plan: &types.PlanResult{Status: types.PlanNew, Content: markdown},
		}
		admit(todo, "plan", types.ModePlan, "sess-plan-new", execution)
		outcome, status := outcomeOf(todo, stepNamed("plan"), execution, succeeded("new", execution.Summary, nil))
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
		execution := &todos.ExecutionResult{
			Success: true, ExecutorName: "headless-claude",
			Summary: "The existing plan is unchanged.", Plan: &types.PlanResult{Status: types.PlanUnchanged},
		}
		admit(approved, "plan", types.ModePlan, "sess-plan-unchanged", execution)
		outcome, status := outcomeOf(approved, stepNamed("plan"), execution, succeeded("unchanged", execution.Summary, nil))
		Expect(status).To(Equal(string(types.StatusPending)))

		Expect(host.OnOutcome(ctx, approved, stepNamed("plan"), outcome, status)).To(Succeed())

		Expect(approved.Status).To(Equal(types.StatusPending))
		plan, err := provider.captain.GetPlan(ctx, *issueOf(approved).SelectedPlanID)
		Expect(err).NotTo(HaveOccurred())
		Expect(plan.LatestRevision.Revision).To(Equal(1), "an unchanged plan appends no revision")
		Expect(lifecycleEvents(approved)).To(HaveLen(2))
	})

	It("verifies a run whose definition of done passed", func() {
		todo, err := provider.Create(ctx, todos.CreateRequest{
			Title: "Ship the widget", Body: "Build it\n\n## Verification\n\n```bash\ntrue\n```\n", Status: types.StatusPending,
		})
		Expect(err).NotTo(HaveOccurred())
		report := &api.VerifyReport{Ran: true, Passed: true}
		execution := &todos.ExecutionResult{
			Success: true, ExecutorName: "headless-claude", Summary: "Built the widget.",
			DoD: &todos.DoDOutcome{Ran: true, Passed: true, Report: report},
		}
		admit(todo, "run", types.ModeRun, "sess-run-pass", execution)
		facts := succeeded("completed", execution.Summary, report)
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
		report := &api.VerifyReport{Ran: true, Passed: false, Reason: "1 check failed"}
		execution := &todos.ExecutionResult{
			Success: true, ExecutorName: "headless-claude", Summary: "Built the gadget.",
			DoD: &todos.DoDOutcome{Ran: true, Passed: false, Report: report},
		}
		admit(todo, "run", types.ModeRun, "sess-run-fail", execution)
		facts := succeeded("completed", execution.Summary, report)
		outcome, status := outcomeOf(todo, stepNamed("run"), execution, facts)
		Expect(status).To(Equal(string(types.StatusUnverified)))

		Expect(host.OnOutcome(ctx, todo, stepNamed("run"), outcome, status)).To(Succeed())

		Expect(todo.Status).To(Equal(types.StatusUnverified))
		issue := issueOf(todo)
		Expect(issue.Status).To(Equal(native.StatusOpen))
		Expect(issue.ExecutionState).To(Equal(native.ExecutionVerificationFailed))
	})

	// Triage writes through its tools during the run; its outcome is keep, and
	// the one lifecycle_outcome event records the event it reported and every
	// write its tools made.
	It("keeps the status for a triage event and records the run's tool writes", func() {
		todo, err := provider.Create(ctx, todos.CreateRequest{Title: "Triage me", Body: "Already shipped", Status: types.StatusPending})
		Expect(err).NotTo(HaveOccurred())
		const comment = "This already landed; labelled it."
		execution := &todos.ExecutionResult{
			Success: true, ExecutorName: "headless-claude", Summary: comment,
			Envelope:   &types.Envelope{Success: true, Event: "done", Comment: comment},
			ToolWrites: []todos.ToolWrite{{Name: "todo_labels", Ref: todo.ShortID}},
		}
		admit(todo, "triage", types.ModePlan, "sess-triage", execution)
		outcome, status := outcomeOf(todo, stepNamed("triage"), execution, succeeded("done", comment, nil))
		Expect(status).To(Equal(lifecycle.OutcomeKeep))

		Expect(host.OnOutcome(ctx, todo, stepNamed("triage"), outcome, status)).To(Succeed())

		Expect(todo.Status).To(Equal(types.StatusPending), "a done event is a claim the definition of done must prove")
		Expect(issueOf(todo).Status).To(Equal(native.StatusOpen))
		events := lifecycleEvents(todo)
		Expect(events).To(HaveLen(1))
		Expect(events[0].Body).To(ContainSubstring(comment))
		payload := payloadOf(events[0])
		Expect([]any{payload["status"], payload["envelope"], payload["tools"]}).To(Equal([]any{
			lifecycle.OutcomeKeep,
			map[string]any{"success": true, "event": "done"},
			[]any{map[string]any{"name": "todo_labels", "ref": todo.ShortID}},
		}))
	})

	It("lists the run history oldest first, each run under its step with the status it landed", func() {
		todo, err := provider.Create(ctx, todos.CreateRequest{
			Title: "Trace the widget", Body: "Build it\n\n## Verification\n\n```bash\ntrue\n```\n", Status: types.StatusDraft,
		})
		Expect(err).NotTo(HaveOccurred())
		planExecution := &todos.ExecutionResult{
			Success: true, ExecutorName: "headless-claude",
			Summary: "The agent created a plan.", Plan: &types.PlanResult{Status: types.PlanNew, Content: "# Plan\n\n1. Trace it."},
		}
		planAdmission := admit(todo, "plan", types.ModePlan, "sess-history-plan", planExecution)
		planOutcome, status := outcomeOf(todo, stepNamed("plan"), planExecution, succeeded("new", planExecution.Summary, nil))
		planOutcome.Admission = planAdmission
		Expect(host.OnOutcome(ctx, todo, stepNamed("plan"), planOutcome, status)).To(Succeed())
		Expect(todo.Status).To(Equal(types.StatusReview))

		approved, err := provider.ApprovePlan(ctx, todo, "moshe", "go")
		Expect(err).NotTo(HaveOccurred())
		report := &api.VerifyReport{Ran: true, Passed: true}
		runExecution := &todos.ExecutionResult{
			Success: true, ExecutorName: "headless-claude", Summary: "Traced it.",
			DoD: &todos.DoDOutcome{Ran: true, Passed: true, Report: report},
		}
		runAdmission := admit(approved, "run", types.ModeRun, "sess-history-run", runExecution)
		runOutcome, status := outcomeOf(approved, stepNamed("run"), runExecution, succeeded("completed", runExecution.Summary, report))
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
