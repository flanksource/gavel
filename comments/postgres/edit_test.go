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

var _ = Describe("embedded comment editing", func() {
	It("rolls back stale batches and preserves rich reply data through edits and deletion", func(ctx SpecContext) {
		db := dbtest.ForGinkgo(dbtest.Options{Name: "comments_edit"})
		Expect(database.Migrate(ctx, db.DSN())).To(Succeed())
		store, err := postgres.New(db.Gorm())
		Expect(err).NotTo(HaveOccurred())
		a := access{tenant: "embedding"}
		svc, err := comments.New(comments.Options{Store: store, Access: a, Targets: a, Workflows: []comments.Workflow{comments.HumanReview()}})
		Expect(err).NotTo(HaveOccurred())
		c, err := svc.Create(ctx, comments.CreateRequest{Target: comments.Target{URI: "example://workspace/page", ScopeURI: "example://workspace", Kind: "page"}, Body: "Review", WorkflowID: "human-review", RequestID: "edit"})
		Expect(err).NotTo(HaveOccurred())
		_, err = svc.Batch(ctx, []comments.ActionRequest{{ID: c.ID, ExpectedVersion: 1, Action: "metadata", Content: json.RawMessage(`{"anchor":"section"}`)}, {ID: c.ID, ExpectedVersion: 1, Action: "close"}})
		Expect(err).To(MatchError(comments.ErrConflict))
		c, err = svc.Get(ctx, c.ID)
		Expect(err).NotTo(HaveOccurred())
		Expect(c.Version).To(Equal(int64(1)))
		Expect(c.Content).To(BeEmpty())
		c, err = svc.Act(ctx, comments.ActionRequest{ID: c.ID, ExpectedVersion: 1, Action: "reply", Body: "First", Content: json.RawMessage(`{"sessionId":"session-a"}`)})
		Expect(err).NotTo(HaveOccurred())
		message := c.Messages[0]
		c, err = svc.Act(ctx, comments.ActionRequest{ID: c.ID, ExpectedVersion: 2, Action: "edit-reply", MessageID: message.ID, Body: "Edited"})
		Expect(err).NotTo(HaveOccurred())
		Expect(c.Messages[0].Body).To(Equal("Edited"))
		Expect(c.Messages[0].Content).To(MatchJSON(message.Content))
		_, err = svc.Act(ctx, comments.ActionRequest{ID: c.ID, ExpectedVersion: 3, Action: "delete-reply", MessageID: "foreign-message"})
		Expect(err).To(MatchError(comments.ErrNotFound))
		target := c.Target
		target.URI = "example://workspace/renamed"
		batch, err := svc.Batch(ctx, []comments.ActionRequest{{ID: c.ID, ExpectedVersion: 3, Action: "retarget", Target: &target}, {ID: c.ID, ExpectedVersion: 4, Action: "delete"}})
		Expect(err).NotTo(HaveOccurred())
		Expect(batch).To(HaveLen(2))
		_, err = svc.Get(ctx, c.ID)
		Expect(err).To(MatchError(comments.ErrNotFound))
		rows, err := svc.List(ctx, comments.Filter{ScopeURI: target.ScopeURI, Limit: 20})
		Expect(err).NotTo(HaveOccurred())
		Expect(rows).To(BeEmpty())
		_, err = svc.Act(ctx, comments.ActionRequest{ID: c.ID, ExpectedVersion: 5, Action: "restore"})
		Expect(err).NotTo(HaveOccurred())
		c, err = svc.Get(ctx, c.ID)
		Expect(err).NotTo(HaveOccurred())
		Expect(c.Target.URI).To(Equal(target.URI))
		Expect(c.Messages[0].Body).To(Equal("Edited"))
		c, err = svc.Act(ctx, comments.ActionRequest{ID: c.ID, ExpectedVersion: c.Version, Action: "reply", ParentID: message.ID, Body: "Follow-up"})
		Expect(err).NotTo(HaveOccurred())
		_, err = svc.Act(ctx, comments.ActionRequest{ID: c.ID, ExpectedVersion: c.Version, Action: "delete-reply", MessageID: message.ID})
		Expect(err).To(MatchError(comments.ErrConflict))
	})
})
