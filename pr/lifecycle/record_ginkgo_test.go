package lifecycle

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	captainai "github.com/flanksource/captain/pkg/ai"
	"github.com/flanksource/captain/pkg/api"
	captaindb "github.com/flanksource/captain/pkg/database"
	"github.com/flanksource/captain/pkg/promptrun"
	"github.com/flanksource/commons-db/dbtest"
	git "github.com/flanksource/commons-db/git"
	. "github.com/onsi/ginkgo/v2"
	g "github.com/onsi/gomega"
)

type envelopeProvider struct{ envelope json.RawMessage }

func (p envelopeProvider) GetModel() string { return "claude-sonnet-5" }
func (p envelopeProvider) GetRuntime() api.Runtime {
	return api.RuntimeOf(api.Anthropic, api.ModeAgent)
}
func (p envelopeProvider) Execute(context.Context, api.Spec) (*api.Response, error) {
	return nil, errors.New("fixture provider only streams")
}
func (p envelopeProvider) ExecuteStream(_ context.Context, _ api.Spec) (<-chan captainai.Event, error) {
	events := make(chan captainai.Event, 2)
	events <- captainai.Event{Kind: captainai.EventText, Text: string(p.envelope)}
	events <- captainai.Event{Kind: captainai.EventResult, Success: true, StructuredData: p.envelope}
	close(events)
	return events, nil
}

var _ = Describe("native PR lifecycle admission", func() {
	It("records the lifecycle outcome through Captain when a completed turn reports failure", func(ctx SpecContext) {
		lease := dbtest.ForGinkgo(dbtest.Options{Name: "pr_lifecycle_outcome"})
		db, err := captaindb.Open(ctx, captaindb.WithDSN(lease.DSN()), captaindb.WithMigrations())
		g.Expect(err).NotTo(g.HaveOccurred())
		DeferCleanup(func() { g.Expect(db.Close()).To(g.Succeed()) })
		def, err := Default()
		g.Expect(err).NotTo(g.HaveOccurred())
		host := Host{DB: db, Definition: def, Dir: GinkgoT().TempDir()}
		snapshot := Snapshot{PR: git.PullRequest{Number: 7, URL: "https://git.example/acme/cli/pull/7", State: git.PullRequestOpen, HeadSHA: "head-a"}}
		action := Action{Step: def.Steps[0], Event: Event{ID: "conflict:head-a", Type: "git.conflict"}}
		provider := envelopeProvider{envelope: json.RawMessage(`{"success":false,"event":"blocked","comment":"The required input is missing."}`)}
		result, err := promptrun.Run(ctx, promptrun.Input{Resolved: api.ResolvedSpec{Spec: api.Spec{Prompt: api.Prompt{User: "Return the fixture envelope."}}}, Provider: provider, Record: host.record(snapshot, action, nil), Timeout: time.Minute})
		g.Expect(err).NotTo(g.HaveOccurred())
		stored, err := db.GetPromptRun(ctx, result.PromptRunID)
		g.Expect(err).NotTo(g.HaveOccurred())
		g.Expect(stored.State).To(g.Equal(captaindb.PromptRunStateFailed))
		g.Expect(stored.ResultJSON).To(g.HaveKeyWithValue("event", "blocked"))
		g.Expect(stored.ResultText).To(g.Equal("The required input is missing."))
	})
	It("pins the source context, rejects an overlapping run and permits retry after cancellation", func(ctx SpecContext) {
		lease := dbtest.ForGinkgo(dbtest.Options{Name: "pr_lifecycle_admission"})
		db, err := captaindb.Open(ctx, captaindb.WithDSN(lease.DSN()), captaindb.WithMigrations())
		g.Expect(err).NotTo(g.HaveOccurred())
		DeferCleanup(func() { g.Expect(db.Close()).To(g.Succeed()) })
		def, err := Default()
		g.Expect(err).NotTo(g.HaveOccurred())
		host := Host{DB: db, Definition: def, Dir: GinkgoT().TempDir()}
		snapshot := Snapshot{PR: git.PullRequest{Number: 7, URL: "https://git.example/acme/cli/pull/7", HeadSHA: "head-a"}}
		action := Action{Step: def.Steps[0], Event: Event{ID: "conflict:head-a", Type: "git.conflict"}}
		input := promptrun.Input{Resolved: api.ResolvedSpec{}, Record: host.record(snapshot, action, nil), Timeout: time.Minute}
		first, err := promptrun.Admit(ctx, input)
		g.Expect(err).NotTo(g.HaveOccurred())
		stored, err := db.GetPromptRun(ctx, first.ID)
		g.Expect(err).NotTo(g.HaveOccurred())
		g.Expect(stored.Metadata).To(g.HaveKeyWithValue("headSHA", "head-a"))
		g.Expect(stored.Metadata).To(g.HaveKey("context"))
		input.Record = host.record(snapshot, action, nil)
		_, err = promptrun.Admit(ctx, input)
		g.Expect(errors.Is(err, captaindb.ErrPromptRunConflict)).To(g.BeTrue(), "%v", err)
		_, err = promptrun.Cancel(context.Background(), db, first.ID, "Stopped in test")
		g.Expect(err).NotTo(g.HaveOccurred())
		_, err = promptrun.Admit(ctx, input)
		g.Expect(err).NotTo(g.HaveOccurred())
	})
})
