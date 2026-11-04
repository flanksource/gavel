package postgres_test

import (
	"encoding/json"
	"github.com/flanksource/commons-db/dbtest"
	"github.com/flanksource/gavel/comments"
	commentlint "github.com/flanksource/gavel/comments/lint"
	"github.com/flanksource/gavel/comments/postgres"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"sync"
)

var _ = Describe("comment concurrency and findings", func() {
	It("accepts one concurrent writer and materializes structured findings idempotently", func(ctx SpecContext) {
		db := dbtest.ForGinkgo(dbtest.Options{Name: "comments_concurrency"})
		Expect(postgres.Provision(ctx, db.DSN())).To(Succeed())
		store, err := postgres.New(db.Gorm())
		Expect(err).NotTo(HaveOccurred())
		a := access{tenant: "tenant-concurrent"}
		svc, err := comments.New(comments.Options{Store: store, Access: a, Targets: a, Workflows: []comments.Workflow{comments.HumanReview()}})
		Expect(err).NotTo(HaveOccurred())
		content := json.RawMessage(`{"lint":{"violations":[{"rule":"check-error","line":7}]},"capture":{"source":"return err"}}`)
		req := commentlint.MaterializeRequest{RunID: "execution-a", WorkflowID: "human-review", Findings: []commentlint.Finding{{ID: "finding-a", Target: comments.Target{URI: "git://repo/a.go", ScopeURI: "git://repo", Kind: "code"}, Body: "Handle error", Content: content}}}
		rows, err := commentlint.Materialize(ctx, svc, req)
		Expect(err).NotTo(HaveOccurred())
		again, err := commentlint.Materialize(ctx, svc, req)
		Expect(err).NotTo(HaveOccurred())
		Expect(again[0].ID).To(Equal(rows[0].ID))
		var wg sync.WaitGroup
		results := make(chan error, 2)
		for range 2 {
			wg.Add(1)
			go func() {
				defer wg.Done()
				_, err := svc.Act(ctx, comments.ActionRequest{ID: rows[0].ID, ExpectedVersion: 1, Action: "reply", Body: "Follow-up"})
				results <- err
			}()
		}
		wg.Wait()
		close(results)
		success, conflict := 0, 0
		for err := range results {
			switch err {
			case nil:
				success++
			case comments.ErrConflict:
				conflict++
			default:
				Expect(err).NotTo(HaveOccurred())
			}
		}
		Expect([]int{success, conflict}).To(Equal([]int{1, 1}))
		got, err := svc.Get(ctx, rows[0].ID)
		Expect(err).NotTo(HaveOccurred())
		Expect(got.Messages).To(HaveLen(1))
		Expect(got.Content).To(MatchJSON(content))
		Expect(got.Links).To(Equal([]comments.Link{{Relation: "execution", URI: "gavel://execution/execution-a"}}))
		events, err := svc.Events(ctx, "git://repo", 0, 100)
		Expect(err).NotTo(HaveOccurred())
		Expect(events).To(HaveLen(2))
	})
})
