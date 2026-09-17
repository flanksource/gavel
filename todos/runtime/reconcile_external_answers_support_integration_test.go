package runtime

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/flanksource/captain/pkg/api"
	captaindb "github.com/flanksource/captain/pkg/database"
	"github.com/flanksource/commons-db/dbtest"
	"github.com/flanksource/gavel/internal/database"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/lifecycle"
	"github.com/flanksource/gavel/todos/native"
	"github.com/flanksource/gavel/todos/types"
	"github.com/flanksource/gavel/verify"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"gorm.io/gorm"
)

// The answer-reconciliation specs stand in for Captain: they write the rows
// Captain and its monitor write — the answer as a user message on the
// transcript, the resumed turn's messages after it, and the turn itself as an
// origin=captain prompt run on that transcript — and assert what a gavel read
// makes of them.
const (
	askedQuestion    = "Which database should the migration target?"
	captainAnswer    = "Target the staging database."
	followupSummary  = "Migrated the staging database."
	followupQuestion = "Should the staging data be anonymised first?"
	askSequence      = 40
)

type answerFixture struct {
	db         *gorm.DB
	options    WorkspaceOptions
	dispatcher *Provider
	reader     *Provider
	host       *lifecycle.Host
	sequence   int64
}

type parkedAsk struct {
	todo               *types.TODO
	runID              uuid.UUID
	executionSessionID uuid.UUID
}

func newAnswerFixture(ctx SpecContext) *answerFixture {
	GinkgoHelper()
	handle := dbtest.ForGinkgo(dbtest.Options{Name: "gavel_external_answers"})
	GinkgoT().Setenv(database.EnvDSN, handle.DSN())
	GinkgoT().Setenv(database.EnvDisable, "")
	GinkgoT().Setenv(database.LegacyEnvDisable, "")
	GinkgoT().Setenv("HOME", GinkgoT().TempDir())
	opened, err := database.Open(ctx, database.WithMigrations())
	Expect(err).NotTo(HaveOccurred())
	DeferCleanup(func() { Expect(opened.Close()).To(Succeed()) })

	f := &answerFixture{db: opened.Gorm(), sequence: askSequence}
	f.options = WorkspaceOptions{
		Name: "external-answers", RootPath: GinkgoT().TempDir(), Repositories: []string{"acme/external-answers"},
	}
	f.dispatcher, err = New(ctx, f.db, f.options)
	Expect(err).NotTo(HaveOccurred())
	// Reads come from a provider of their own, as every dashboard request and
	// CLI command opens one: it never dispatched the run it reads.
	f.reader = f.newProvider(ctx)
	def, err := lifecycle.Default()
	Expect(err).NotTo(HaveOccurred())
	engine, err := lifecycle.New(def)
	Expect(err).NotTo(HaveOccurred())
	f.host = &lifecycle.Host{
		Provider: f.dispatcher, Def: engine, Config: verify.DefaultGavelConfig(),
		WorkDir: f.options.RootPath, Kind: lifecycle.HostCLI,
	}
	return f
}

func (f *answerFixture) newProvider(ctx SpecContext) *Provider {
	GinkgoHelper()
	provider, err := New(ctx, f.db, f.options)
	Expect(err).NotTo(HaveOccurred())
	return provider
}

func (f *answerFixture) insertMessage(ctx SpecContext, sessionID uuid.UUID, role string, parts any) {
	GinkgoHelper()
	encoded, err := json.Marshal(parts)
	Expect(err).NotTo(HaveOccurred())
	f.sequence += 2
	// occurred_at is the provider's clock and deliberately far off here: the
	// reconciliation must order by the database's own clock, never this one.
	Expect(f.db.WithContext(ctx).Exec(`
		INSERT INTO captain_messages (session_id, sequence, role, parts, occurred_at)
		VALUES (?, ?, ?, CAST(? AS jsonb), ?)`,
		sessionID, f.sequence, role, string(encoded), time.Now().Add(48*time.Hour)).Error).To(Succeed())
}

func structuredOutput(envelope map[string]any) []map[string]any {
	return []map[string]any{{
		"type": "dynamic-tool", "toolName": "StructuredOutput", "toolCallId": "toolu_" + uuid.NewString()[:8],
		"state": "output-available", "input": envelope,
	}}
}

func textParts(text string) []map[string]any {
	return []map[string]any{{"type": "text", "text": text}}
}

func askEnvelope(question string) map[string]any {
	return map[string]any{
		"summary": "The agent is waiting for answers.", "endStatus": "ask",
		"questions": []map[string]any{{"text": question, "options": []string{"staging", "production"}}},
	}
}

// parkAsk runs a todo's step to an ask exactly as the host records one, with
// the asking StructuredOutput already ingested on the transcript.
func (f *answerFixture) parkAsk(ctx SpecContext, title, stepName string, mode types.RunMode) parkedAsk {
	GinkgoHelper()
	todo, err := f.dispatcher.Create(ctx, todos.CreateRequest{Title: title, Body: "Migrate it", Status: types.StatusPending})
	Expect(err).NotTo(HaveOccurred())
	admission, err := f.dispatcher.PrepareRun(ctx, todo, todos.RunPreparation{Mode: mode, Prompt: stepName, ExecutorName: "claude"})
	Expect(err).NotTo(HaveOccurred())
	DeferCleanup(func() { f.dispatcher.ownership.stop(admission.PromptRunID) })
	Expect(f.dispatcher.RecordRunStart(ctx, todo, todos.RunStartMetadata{
		SessionID: uuid.NewString(), Provider: "claude", Mode: string(mode),
	})).To(Succeed())
	run, err := f.dispatcher.Captain().GetPromptRun(ctx, admission.PromptRunID)
	Expect(err).NotTo(HaveOccurred())
	Expect(run.ExecutionSessionID).NotTo(BeNil())
	f.insertMessage(ctx, *run.ExecutionSessionID, "assistant", structuredOutput(askEnvelope(askedQuestion)))

	step, ok := f.host.Def.Definition().Step(stepName)
	Expect(ok).To(BeTrue())
	execution := &todos.ExecutionResult{
		Success: true, ExecutorName: "claude", EndStatus: types.EndAsk,
		Summary: "The agent is waiting for answers.", Questions: []types.AgentQuestion{{Text: askedQuestion}},
	}
	facts := lifecycle.StepResult{
		Run:       lifecycle.RunFacts{State: lifecycle.RunWaiting},
		Envelope:  lifecycle.Envelope{Summary: execution.Summary, EndStatus: string(types.EndAsk)},
		Questions: []any{map[string]any{"text": askedQuestion}},
	}
	lc, err := f.host.Context(ctx, todo)
	Expect(err).NotTo(HaveOccurred())
	status, err := f.host.Def.Outcome(step, lc, facts)
	Expect(err).NotTo(HaveOccurred())
	Expect(status).To(Equal(string(types.StatusAsk)))
	outcome := &lifecycle.StepOutcome{Step: step, Status: status, Result: facts, Execution: execution, Admission: admission}
	Expect(f.host.OnOutcome(ctx, todo, step, outcome, status)).To(Succeed())
	Expect(todo.Status).To(Equal(types.StatusAsk))
	return parkedAsk{todo: todo, runID: admission.PromptRunID, executionSessionID: *run.ExecutionSessionID}
}

func (f *answerFixture) parkRunAsk(ctx SpecContext, title string) parkedAsk {
	GinkgoHelper()
	return f.parkAsk(ctx, title, "run", types.ModeRun)
}

func (f *answerFixture) answerInCaptain(ctx SpecContext, parked parkedAsk) {
	GinkgoHelper()
	f.insertMessage(ctx, parked.executionSessionID, "user", textParts(captainAnswer))
}

// captainTurn records a Captain turn the way Captain's chat persists one: the
// row is created and finished in one go once the turn is over, carrying no
// result text for a chat turn.
func (f *answerFixture) captainTurn(ctx SpecContext, parked parkedAsk, state captaindb.PromptRunState, errorText string) {
	GinkgoHelper()
	f.captainTurnFor(ctx, parked, captainAnswer, state, errorText)
}

// captainTurnFor records a Captain turn that sent prompt as its user text.
func (f *answerFixture) captainTurnFor(ctx SpecContext, parked parkedAsk, prompt string, state captaindb.PromptRunState, errorText string) {
	GinkgoHelper()
	captain := f.dispatcher.Captain()
	run, err := captain.CreatePromptRun(ctx, captaindb.CreatePromptRunInput{
		SessionID: parked.executionSessionID, Origin: "captain", PromptMarkdown: prompt,
	})
	Expect(err).NotTo(HaveOccurred())
	finished := captaindb.PromptRunPhaseFinished
	update := captaindb.UpdatePromptRunInput{ID: run.ID, ExpectedVersion: run.Version, Phase: &finished, State: &state}
	if errorText != "" {
		update.Error = &errorText
	}
	_, err = captain.UpdatePromptRun(ctx, update)
	Expect(err).NotTo(HaveOccurred())
}

func (f *answerFixture) liveProcess(ctx SpecContext, sessionID uuid.UUID) {
	GinkgoHelper()
	Expect(f.reader.Captain().UpsertSessionProcess(ctx, captaindb.SessionProcessInput{
		SessionID: sessionID, HostID: captaindb.LocalHostID(), BootID: "boot-external-answers",
		PID: 424242, ProcessStartedAt: time.Now().Add(-time.Minute), Status: "running", Source: "claude",
		SampledAt: time.Now(),
	})).To(Succeed())
}

// resumeFromDashboard admits a resumed turn with the answer as its message, as
// the dashboard's answer does through the lifecycle host.
func (f *answerFixture) resumeFromDashboard(ctx SpecContext, parked parkedAsk, answer string) {
	GinkgoHelper()
	f.resumeWith(ctx, f.dispatcher, parked, answer)
}

// resumeWith admits the resumed turn through the given provider.
func (f *answerFixture) resumeWith(ctx SpecContext, provider *Provider, parked parkedAsk, answer string) {
	GinkgoHelper()
	todo, err := provider.Get(ctx, parked.todo.ID)
	Expect(err).NotTo(HaveOccurred())
	admission, err := provider.PrepareRun(ctx, todo, todos.RunPreparation{
		Mode: types.ModeRun, Prompt: "run", ExecutorName: "claude", Resume: true,
		Spec: api.Spec{Prompt: api.Prompt{User: answer}},
	})
	Expect(err).NotTo(HaveOccurred())
	Expect(admission.PromptRunID).To(Equal(parked.runID))
	DeferCleanup(func() { provider.clearPrepared(uuid.MustParse(todo.ID), parked.runID) })
}

func (f *answerFixture) eventsOf(ctx SpecContext, todo *types.TODO, kind string) []native.Event {
	GinkgoHelper()
	events, err := f.reader.Repository().ListEvents(ctx, uuid.MustParse(todo.ID))
	Expect(err).NotTo(HaveOccurred())
	var matched []native.Event
	for _, event := range events {
		if event.Kind == kind {
			matched = append(matched, event)
		}
	}
	return matched
}

func payloadOf(event native.Event) map[string]any {
	GinkgoHelper()
	var payload map[string]any
	Expect(json.Unmarshal(event.Payload, &payload)).To(Succeed())
	return payload
}

func (f *answerFixture) gavelRun(ctx SpecContext, parked parkedAsk) *captaindb.PromptRun {
	GinkgoHelper()
	run, err := f.reader.Captain().GetPromptRun(ctx, parked.runID)
	Expect(err).NotTo(HaveOccurred())
	return run
}

// concurrentGets reads one todo from two fresh providers at once, as two
// dashboard requests racing to reconcile the same answer do.
func (f *answerFixture) concurrentGets(ctx SpecContext, ref string) []types.Status {
	GinkgoHelper()
	statuses := make([]types.Status, 2)
	errs := make([]error, 2)
	providers := []*Provider{f.newProvider(ctx), f.newProvider(ctx)}
	var wg sync.WaitGroup
	for i := range statuses {
		wg.Add(1)
		go func(i int) {
			defer GinkgoRecover()
			defer wg.Done()
			todo, err := providers[i].Get(ctx, ref)
			if errs[i] = err; err == nil {
				statuses[i] = todo.Status
			}
		}(i)
	}
	wg.Wait()
	Expect(errs).To(HaveEach(Not(HaveOccurred())))
	return statuses
}
