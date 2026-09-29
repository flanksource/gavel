package runtime

import (
	"context"
	"fmt"

	captainai "github.com/flanksource/captain/pkg/ai"
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

	_ "github.com/flanksource/gavel/fixtures/verifier"
)

// provenanceAgent is the streaming provider a step is dispatched through in
// place of a real agent: one turn that announces its session and returns a
// completed envelope.
type provenanceAgent struct{ sessionID string }

func (a provenanceAgent) GetModel() string { return "scripted-model" }
func (a provenanceAgent) GetRuntime() api.Runtime {
	return api.RuntimeOf(api.Anthropic, api.ModeAgent)
}
func (a provenanceAgent) Execute(context.Context, api.Spec) (*api.Response, error) {
	return nil, fmt.Errorf("the scripted agent only streams")
}
func (a provenanceAgent) ExecuteStream(context.Context, api.Spec) (<-chan captainai.Event, error) {
	events := []captainai.Event{
		{Kind: captainai.EventSystem, SessionID: a.sessionID},
		{Kind: captainai.EventText, Text: `{"summary":"Persisted it.","endStatus":"completed"}`},
		{Kind: captainai.EventResult, Success: true, Usage: &api.Usage{InputTokens: 7, OutputTokens: 3}},
	}
	ch := make(chan captainai.Event, len(events))
	for _, event := range events {
		ch <- event
	}
	close(ch)
	return ch, nil
}

// A todo run is filed by Captain, not by gavel: promptrun admits it with the
// resolution trace in its metadata, Link attaches it to the todo in the same
// transaction, and Captain files the terminal state gavel's envelope maps to.
var _ = Describe("todo run provenance", func() {
	var (
		provider *Provider
		host     *lifecycle.Host
	)
	BeforeEach(func(ctx SpecContext) {
		db := dbtest.ForGinkgo(dbtest.Options{Name: "gavel_run_provenance"})
		GinkgoT().Setenv(database.EnvDSN, db.DSN())
		GinkgoT().Setenv(database.EnvDisable, "")
		GinkgoT().Setenv(database.LegacyEnvDisable, "")
		GinkgoT().Setenv("HOME", GinkgoT().TempDir())
		opened, err := database.Open(ctx, database.WithMigrations())
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(func() { Expect(opened.Close()).To(Succeed()) })
		workDir := GinkgoT().TempDir()
		provider, err = New(ctx, opened.Gorm(), WorkspaceOptions{
			Name: "run-provenance", RootPath: workDir, Repositories: []string{"acme/run-provenance"},
		})
		Expect(err).NotTo(HaveOccurred())

		// The built-in run step checks out a worktree and commits; there is no
		// repository here to act on. Its definition of done stays.
		def, err := lifecycle.Default()
		Expect(err).NotTo(HaveOccurred())
		for i := range def.Steps {
			if def.Steps[i].Name == "run" {
				def.Steps[i].Spec = &api.Spec{Workflow: &api.Workflow{Verify: &api.Verify{
					Fixture: "{{subject.verification.document}}", Scope: api.VerifyScopeAll,
				}}}
			}
		}
		engine, err := lifecycle.New(def)
		Expect(err).NotTo(HaveOccurred())
		host = &lifecycle.Host{Provider: provider, Def: engine, Config: verify.DefaultGavelConfig(), WorkDir: workDir, Kind: lifecycle.HostCLI}
	})

	It("files the run with its spec trace and links it to the todo", func(ctx SpecContext) {
		const fixture = "```bash\ntrue\n```"
		todo, err := provider.Create(ctx, todos.CreateRequest{
			Title: "Persist the spec trace", Body: "Record how the spec was resolved", Status: types.StatusPending, Verification: fixture,
		})
		Expect(err).NotTo(HaveOccurred())
		run, ok := host.Def.Definition().Step("run")
		Expect(ok).To(BeTrue())

		outcome, err := host.RunStep(ctx, todo, run, lifecycle.RunOptions{Provider: provenanceAgent{sessionID: uuid.NewString()}})

		Expect(err).NotTo(HaveOccurred())
		Expect(outcome.Result.Run.State).To(Equal(lifecycle.RunSucceeded))
		runID := outcome.Admission.PromptRunID
		Expect(runID).NotTo(Equal(uuid.Nil))
		filed, err := provider.Captain().GetPromptRun(ctx, runID)
		Expect(err).NotTo(HaveOccurred())
		Expect(filed.Metadata).To(HaveKeyWithValue("specTrace", Not(BeEmpty())), "the Raw tab reads the resolution trace from here")
		Expect(filed).To(And(
			HaveField("Origin", runOrigin),
			HaveField("SpecProfile", "run"),
			HaveField("VerificationMarkdown", fixture),
			HaveField("State", captaindb.PromptRunStateSucceeded),
		))
		links, err := provider.Repository().ListPromptRuns(ctx, uuid.MustParse(todo.ID))
		Expect(err).NotTo(HaveOccurred())
		Expect(links).To(ContainElement(And(HaveField("PromptRunID", runID), HaveField("StepKind", native.StepRun))),
			"todo_issue_prompt_runs links the run to its todo")
		issue, err := provider.Repository().GetIssue(ctx, uuid.MustParse(todo.ID))
		Expect(err).NotTo(HaveOccurred())
		Expect(issue.ActivePromptRunID).To(Equal(&runID))
	})
})
