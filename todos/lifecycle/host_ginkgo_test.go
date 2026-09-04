package lifecycle_test

import (
	"context"
	"time"

	"github.com/flanksource/captain/pkg/api"
	captaindb "github.com/flanksource/captain/pkg/database"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/lifecycle"
	"github.com/flanksource/gavel/todos/types"
	"github.com/flanksource/gavel/verify"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// fakeProvider is the slice of todos.Provider the host reads and writes. The
// embedded nil interface makes any other call a loud panic rather than a
// silent success.
type fakeProvider struct {
	todos.Provider
	plan       todos.PlanState
	runs       []todos.StepRunRecord
	backlog    types.TODOS
	states     []todos.StateUpdate
	attempts   []*todos.ExecutionResult
	comments   []string
	events     []todos.Event
	iterations []captaindb.UpsertPromptRunIterationInput
}

// RecordRunIterations captures the rows the host files for a finished run,
// stamped with the run they belong to as the native runtime stamps them.
func (f *fakeProvider) RecordRunIterations(_ context.Context, promptRunID uuid.UUID, records []captaindb.UpsertPromptRunIterationInput) error {
	for _, record := range records {
		record.PromptRunID = promptRunID
		f.iterations = append(f.iterations, record)
	}
	return nil
}

func (f *fakeProvider) PlanState(context.Context, *types.TODO) (todos.PlanState, error) {
	return f.plan, nil
}

func (f *fakeProvider) RunHistory(context.Context, *types.TODO) ([]todos.StepRunRecord, error) {
	return f.runs, nil
}

func (f *fakeProvider) List(context.Context, todos.DiscoveryFilters) (types.TODOS, error) {
	return f.backlog, nil
}

func (f *fakeProvider) UpdateState(_ context.Context, todo *types.TODO, update todos.StateUpdate) error {
	if update.Status != nil {
		todo.Status = *update.Status
	}
	f.states = append(f.states, update)
	return nil
}

// SaveAttempt reloads the todo's attempt count the way the native runtime does
// (attempts are counted from the run links it just wrote), so a state update
// that still pointed into the todo would carry the reloaded value.
func (f *fakeProvider) SaveAttempt(_ context.Context, todo *types.TODO, result *todos.ExecutionResult) error {
	f.attempts = append(f.attempts, result)
	todo.Attempts = 99
	return nil
}

func (f *fakeProvider) Comment(_ context.Context, _ *types.TODO, body string) error {
	f.comments = append(f.comments, body)
	return nil
}

func (f *fakeProvider) AppendEvent(_ context.Context, _ *types.TODO, event todos.Event) error {
	f.events = append(f.events, event)
	return nil
}

const hostIssueID = "11111111-2222-3333-4444-555555555555"

func hostTodo() *types.TODO {
	return &types.TODO{
		ID: hostIssueID, Version: 3, ExecutionState: "idle", Labels: []string{"area/todos"},
		MarkdownBody:         "Implement the thing",
		VerificationMarkdown: "### command: it works\n\n```bash\ntrue\n```",
		TODOFrontmatter:      types.TODOFrontmatter{Status: types.StatusPending, Priority: types.PriorityHigh, Attempts: 2},
	}
}

func newHost(provider *fakeProvider) *lifecycle.Host {
	GinkgoHelper()
	engine := defaultEngine()
	return &lifecycle.Host{
		Provider: provider, Def: engine, Config: verify.DefaultGavelConfig(),
		WorkDir: GinkgoT().TempDir(), Kind: lifecycle.HostCLI,
	}
}

// graderFor mirrors the verify chain the host resolves for the definition of
// done's acceptance-criteria grader: `.gavel.yaml todos.verify` over `ai:`.
func graderFor(host *lifecycle.Host) api.Spec {
	GinkgoHelper()
	resolved, err := lifecycle.ResolveLayers(lifecycle.LayerInput{Config: host.Config, Step: lifecycle.StepVerify, Host: host.Kind})
	Expect(err).NotTo(HaveOccurred())
	return resolved.Spec
}

var _ = Describe("Host", func() {
	var (
		provider *fakeProvider
		host     *lifecycle.Host
		ctx      context.Context
	)
	BeforeEach(func() {
		// The definition of done loads the user's own ~/.gavel.yaml as a layer;
		// the specs assert against the built-in defaults, not this machine's.
		GinkgoT().Setenv("HOME", GinkgoT().TempDir())
		provider = &fakeProvider{plan: todos.PlanState{Exists: true, Approved: true, Content: "# plan", Revision: 2}}
		host = newHost(provider)
		ctx = context.Background()
	})

	Describe("Subject", func() {
		It("projects the todo onto every declared field", func() {
			todo := hostTodo()
			dod, err := todos.BuildDefinitionOfDone(todos.DefinitionOfDoneOptions{
				WorkDir: host.WorkDir, Todos: []*types.TODO{todo}, Grader: graderFor(host),
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(dod.Declared()).To(BeTrue())

			subject, err := host.Subject(ctx, todo)

			Expect(err).NotTo(HaveOccurred())
			Expect(subject).To(Equal(map[string]any{
				"id": hostIssueID, "status": "pending", "priority": "high", "labels": []string{"area/todos"},
				"body": "Implement the thing", "attempts": 2,
				"execution":    map[string]any{"state": "idle"},
				"plan":         map[string]any{"exists": true, "approved": true, "content": "# plan", "path": "", "revision": 2},
				"verification": map[string]any{"exists": true, "document": dod.Fixture},
			}))
		})

		It("reports no verification for a todo without a definition of done", func() {
			todo := hostTodo()
			todo.VerificationMarkdown = ""
			todo.Labels = nil

			subject, err := host.Subject(ctx, todo)

			Expect(err).NotTo(HaveOccurred())
			Expect(subject["verification"]).To(Equal(map[string]any{"exists": false, "document": ""}))
			Expect(subject["labels"]).To(Equal([]string{}), "a declared list is never nil")
		})

		It("requires a grader model once the todo has acceptance criteria to grade", func() {
			todo := hostTodo()
			todo.AcceptanceCriteria = []types.AcceptanceCriterion{{Text: "the chart renders"}}

			_, err := host.Subject(ctx, todo)

			Expect(err).To(MatchError(ContainSubstring("acceptance-criteria grader: model: model name is required")))
		})

		It("refuses a provider that cannot report plan state", func() {
			host.Provider = &struct{ todos.Provider }{}
			_, err := host.Subject(ctx, hostTodo())
			Expect(err).To(MatchError(ContainSubstring("does not expose plan state")))
		})
	})

	Describe("Runs", func() {
		It("carries every recorded run oldest first, with its outcome and prompt run", func() {
			finished := time.Now()
			provider.runs = []todos.StepRunRecord{
				{Step: "plan", State: "succeeded", Outcome: "review", PromptRunID: "run-1", FinishedAt: &finished},
				{Step: "run", State: "failed", Outcome: "failed", PromptRunID: "run-2", FinishedAt: &finished},
			}

			runs, err := host.Runs(ctx, hostTodo())

			Expect(err).NotTo(HaveOccurred())
			Expect(runs).To(Equal([]lifecycle.StepRun{
				{Step: "plan", State: "succeeded", Outcome: "review", PromptRunID: "run-1", FinishedAt: &finished},
				{Step: "run", State: "failed", Outcome: "failed", PromptRunID: "run-2", FinishedAt: &finished},
			}))
		})

		It("refuses a provider that cannot report run history", func() {
			host.Provider = &struct {
				todos.Provider
				todos.PlanStateProvider
			}{PlanStateProvider: provider}
			_, err := host.Context(ctx, hostTodo())
			Expect(err).To(MatchError(ContainSubstring("does not expose run history")))
		})
	})

	Describe("VerifyDocument", func() {
		It("equals the definition-of-done builder's output for the verify chain's grader", func() {
			todo := hostTodo()
			want, err := todos.BuildDefinitionOfDone(todos.DefinitionOfDoneOptions{
				WorkDir: host.WorkDir, Todos: []*types.TODO{todo}, Grader: graderFor(host),
			})
			Expect(err).NotTo(HaveOccurred())

			got, err := host.VerifyDocument(todo)

			Expect(err).NotTo(HaveOccurred())
			Expect(got).To(Equal(want))
		})
	})

	Describe("Next", func() {
		It("implements a pending todo whose plan is approved", func() {
			step, ok, err := host.Next(ctx, hostTodo())
			Expect(err).NotTo(HaveOccurred())
			Expect(ok).To(BeTrue())
			Expect(step.Name).To(Equal("run"))
		})

		It("verifies a todo whose last run landed unchecked", func() {
			todo := hostTodo()
			finished := time.Now()
			provider.runs = []todos.StepRunRecord{{Step: "run", State: "succeeded", FinishedAt: &finished}}

			step, ok, err := host.Next(ctx, todo)

			Expect(err).NotTo(HaveOccurred())
			Expect(ok).To(BeTrue())
			Expect(step.Name).To(Equal(lifecycle.StepVerify))
		})

		It("plans a draft that has no plan", func() {
			provider.plan = todos.PlanState{}
			todo := hostTodo()
			todo.Status = types.StatusDraft

			step, ok, err := host.Next(ctx, todo)

			Expect(err).NotTo(HaveOccurred())
			Expect(ok).To(BeTrue())
			Expect(step.Name).To(Equal("plan"))
		})
	})

	Describe("Hooks", func() {
		It("orders the commit pipeline, the run environment and the spec recorder", func() {
			req := api.Spec{Workflow: &api.Workflow{Commits: []api.Commit{{On: "run"}, {On: "turn"}}}}
			meta := todos.RunStartMetadata{SessionID: "session-1"}

			hooks := host.Hooks(hostTodo(), req, meta, func(todos.RunStartMetadata) {})

			var names []string
			for _, hook := range hooks {
				names = append(names, hook.(interface{ Name() string }).Name())
			}
			Expect(names).To(Equal([]string{"commit:run", "commit:turn", "gavel-run-env", "gavel-spec-recorder"}))
		})

		It("declares no commit hook for a spec without commit policies", func() {
			hooks := host.Hooks(hostTodo(), api.Spec{}, todos.RunStartMetadata{}, func(todos.RunStartMetadata) {})

			var names []string
			for _, hook := range hooks {
				names = append(names, hook.(interface{ Name() string }).Name())
			}
			Expect(names).To(Equal([]string{"gavel-run-env", "gavel-spec-recorder"}))
		})
	})

	Describe("OnOutcome", func() {
		var (
			todo    *types.TODO
			run     lifecycle.Step
			outcome *lifecycle.StepOutcome
		)
		BeforeEach(func() {
			todo = hostTodo()
			run = stepNamed(host.Def, "run")
			outcome = &lifecycle.StepOutcome{
				Step:   run,
				Result: lifecycle.StepResult{Run: lifecycle.RunFacts{State: lifecycle.RunSucceeded, Iterations: 2, CostUSD: 0.5}},
				Execution: &todos.ExecutionResult{
					Success: true, Summary: "Built it.", EndStatus: types.EndCompleted, TokensUsed: 10, CostUSD: 0.5,
					Runtime:   todos.RunStartMetadata{ResolvedModel: "claude-sonnet"},
					Questions: []types.AgentQuestion{{Text: "stale question from an earlier ask"}},
				},
			}
		})

		It("persists the attempt, then writes the status once, then records one event", func() {
			Expect(host.OnOutcome(ctx, todo, run, outcome, "verified")).To(Succeed())

			Expect(provider.attempts).To(HaveLen(1))
			Expect(provider.states).To(HaveLen(1))
			update := provider.states[0]
			Expect(*update.Status).To(Equal(types.StatusVerified))
			Expect(*update.Attempts).To(Equal(3), "the attempt count is the todo's plus one, not whatever SaveAttempt reloaded")
			Expect(*update.RunMode).To(Equal(types.ModeRun))
			Expect(*update.LastRunSummary).To(Equal("Built it."))
			Expect(*update.Questions).To(BeEmpty(), "questions persist only for an ask outcome")
			Expect(todo.Status).To(Equal(types.StatusVerified))
			Expect(todo.LLM.Model).To(Equal("claude-sonnet"))
			Expect(provider.events).To(HaveLen(1))
			Expect(provider.events[0].Kind).To(Equal(lifecycle.EventLifecycleOutcome))
			Expect(provider.events[0].Payload).To(HaveKeyWithValue("status", "verified"))
		})

		It("writes no status when the step keeps it", func() {
			Expect(host.OnOutcome(ctx, todo, stepNamed(host.Def, "triage"), outcome, lifecycle.OutcomeKeep)).To(Succeed())

			Expect(provider.states).To(HaveLen(1))
			Expect(provider.states[0].Status).To(BeNil())
			Expect(todo.Status).To(Equal(types.StatusPending))
			Expect(provider.events[0].Payload).To(HaveKeyWithValue("status", lifecycle.OutcomeKeep))
		})

		It("records the status a triage verdict assigned when the step kept it", func() {
			outcome.Execution.Triage = &types.TriageEnvelope{
				ResultEnvelope: types.ResultEnvelope{Summary: "Shelve it.", EndStatus: types.EndCompleted},
				Verdict:        types.VerdictRetire, Status: string(types.StatusDraft), Comment: "obsolete",
			}

			Expect(host.OnOutcome(ctx, todo, stepNamed(host.Def, "triage"), outcome, lifecycle.OutcomeKeep)).To(Succeed())

			Expect(todo.Status).To(Equal(types.StatusDraft))
			Expect(provider.states).To(HaveLen(2), "the triage write, then the attempt's own update")
			Expect(*provider.states[0].Status).To(Equal(types.StatusDraft))
			Expect(provider.states[1].Status).To(BeNil(), "the step kept the status; only the verdict wrote one")
			Expect(provider.events).To(HaveLen(1))
			Expect(provider.events[0].Payload).To(HaveKeyWithValue("status", "draft"), "the event names the status actually written")
			Expect(provider.events[0].Body).To(ContainSubstring("`triage` → `draft`"))
		})

		It("keeps the questions of a run that asks", func() {
			outcome.Execution.EndStatus = types.EndAsk

			Expect(host.OnOutcome(ctx, todo, run, outcome, "ask")).To(Succeed())

			Expect(*provider.states[0].Questions).To(HaveLen(1))
			Expect(todo.Status).To(Equal(types.StatusAsk))
		})

		It("refuses a status the lifecycle does not know", func() {
			err := host.OnOutcome(ctx, todo, run, outcome, "done")
			Expect(err).To(MatchError(ContainSubstring(`"done" is not a todo status`)))
			Expect(provider.attempts).To(BeEmpty())
		})
	})

	Describe("RunStep", func() {
		It("refuses a step the lifecycle does not declare", func() {
			_, err := host.RunStep(ctx, hostTodo(), lifecycle.Step{Name: "deploy"}, lifecycle.RunOptions{})
			Expect(err).To(MatchError(ContainSubstring(`step "deploy" is not part of lifecycle todos`)))
		})

		It("refuses a prompt reference that is neither built-in nor a file", func() {
			def := host.Def.Definition()
			def.Steps[3].Prompt = "run"
			engine, err := lifecycle.New(def)
			Expect(err).NotTo(HaveOccurred())
			host.Def = engine

			_, err = host.RunStep(ctx, hostTodo(), def.Steps[3], lifecycle.RunOptions{})

			Expect(err).To(MatchError(ContainSubstring(`prompt "run" is neither todos.<name> nor file:<path>`)))
		})

		It("refuses a built-in prompt whose envelope disagrees with the step", func() {
			def := host.Def.Definition()
			def.Steps[3].Prompt = "todos.plan"
			engine, err := lifecycle.New(def)
			Expect(err).NotTo(HaveOccurred())
			host.Def = engine

			_, err = host.RunStep(ctx, hostTodo(), def.Steps[3], lifecycle.RunOptions{})

			Expect(err).To(MatchError(ContainSubstring("returns a plan envelope, the step declares result")))
		})
	})
})
