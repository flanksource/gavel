package postgres_test

import (
	"encoding/json"
	"github.com/flanksource/commons-db/dbtest"
	"github.com/flanksource/gavel/comments"
	"github.com/flanksource/gavel/comments/postgres"
	"github.com/flanksource/gavel/internal/database"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("lint request metadata", func() {
	It("deduplicates admission and retains native references without claiming execution state", func(ctx SpecContext) {
		db := dbtest.ForGinkgo(dbtest.Options{Name: "comment_lint"})
		Expect(database.Migrate(ctx, db.DSN())).To(Succeed())
		store, err := postgres.New(db.Gorm())
		Expect(err).NotTo(HaveOccurred())
		a := access{tenant: "lint"}
		svc, err := comments.New(comments.Options{Store: store, Access: a, Targets: a, Workflows: []comments.Workflow{comments.HumanReview()}})
		Expect(err).NotTo(HaveOccurred())
		req := comments.LintRequest{Target: comments.Target{URI: "git://repo/a.go", ScopeURI: "git://repo", Kind: "code"}, RequestID: "lint-request", PromptID: "review", PromptVersion: "version-1", Selection: json.RawMessage(`{"paths":["a.go"]}`)}
		run, created, err := svc.BeginLint(ctx, req)
		Expect(err).NotTo(HaveOccurred())
		Expect(created).To(BeTrue())
		again, created, err := svc.BeginLint(ctx, req)
		Expect(err).NotTo(HaveOccurred())
		Expect(created).To(BeFalse())
		Expect(again.ID).To(Equal(run.ID))
		findings := json.RawMessage(`[{"file":"a.go","message":"Handle errors"}]`)
		_, err = svc.RecordLint(ctx, run.ID, comments.ExecutionRef{RunID: "captain-run", SessionID: "captain-session"}, "review", findings)
		Expect(err).NotTo(HaveOccurred())
		run, err = svc.LintRun(ctx, run.ID)
		Expect(err).NotTo(HaveOccurred())
		Expect(run.Executions).To(Equal([]comments.ExecutionRef{{RunID: "captain-run", SessionID: "captain-session"}}))
		Expect(run.Findings["review"]).To(MatchJSON(findings))
		_, err = store.LintRun(ctx, "another-tenant", run.ID)
		Expect(err).To(MatchError(comments.ErrNotFound))
		req.PromptVersion = "version-2"
		_, _, err = svc.BeginLint(ctx, req)
		Expect(err).To(MatchError(comments.ErrConflict))
	})
})
