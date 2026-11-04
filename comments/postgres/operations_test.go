package postgres_test

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/flanksource/commons-db/dbtest"
	"github.com/flanksource/gavel/comments"
	"github.com/flanksource/gavel/comments/postgres"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"time"
)

type uncertainDelivery struct{ calls int }

func (e *uncertainDelivery) Execute(context.Context, comments.Operation) (json.RawMessage, error) {
	e.calls++
	return nil, errors.New("connection lost after send")
}

var _ = Describe("durable publication admission", func() {
	It("retains uncertain delivery without replaying it", func(ctx SpecContext) {
		db := dbtest.ForGinkgo(dbtest.Options{Name: "comments_outbox"})
		Expect(postgres.Provision(ctx, db.DSN())).To(Succeed())
		store, err := postgres.New(db.Gorm())
		Expect(err).NotTo(HaveOccurred())
		a := access{tenant: "tenant-outbox"}
		svc, err := comments.New(comments.Options{Store: store, Access: a, Targets: a, Workflows: []comments.Workflow{comments.HumanReview()}})
		Expect(err).NotTo(HaveOccurred())
		c, err := svc.Create(ctx, comments.CreateRequest{Target: comments.Target{URI: "https://github.com/acme/project/pull/1", Kind: "pull-request", ScopeURI: "git://repo"}, Body: "Please inspect", WorkflowID: "human-review", RequestID: "draft"})
		Expect(err).NotTo(HaveOccurred())
		req := comments.QueueRequest{CommentID: c.ID, ExpectedVersion: 1, Kind: "publish", RequestID: "publish-request", Payload: json.RawMessage(`{"provider":"github"}`)}
		op, err := svc.Queue(ctx, req)
		Expect(err).NotTo(HaveOccurred())
		again, err := svc.Queue(ctx, req)
		Expect(err).NotTo(HaveOccurred())
		Expect(again.ID).To(Equal(op.ID))
		executor := &uncertainDelivery{}
		op, err = svc.Process(ctx, executor, time.Minute)
		Expect(err).NotTo(HaveOccurred())
		Expect(op.State).To(Equal("unknown"))
		Expect(op.Error).To(ContainSubstring("connection lost"))
		_, err = svc.Process(ctx, executor, time.Minute)
		Expect(err).To(MatchError(comments.ErrNotFound))
		Expect(executor.calls).To(Equal(1))
		persisted, err := store.Operation(ctx, "tenant-outbox", op.ID)
		Expect(err).NotTo(HaveOccurred())
		Expect(persisted.State).To(Equal("unknown"))
		req.RequestID = "expired-request"
		queued, err := svc.Queue(ctx, req)
		Expect(err).NotTo(HaveOccurred())
		claimed, err := store.Claim(ctx, a.tenant, time.Minute)
		Expect(err).NotTo(HaveOccurred())
		Expect(claimed.ID).To(Equal(queued.ID))
		Expect(db.Gorm().Exec("UPDATE comment_operations SET lease_until=now()-interval '1 second' WHERE id=?", claimed.ID).Error).To(Succeed())
		_, err = store.Claim(ctx, a.tenant, time.Minute)
		Expect(err).To(MatchError(comments.ErrNotFound))
		persisted, err = store.Operation(ctx, a.tenant, claimed.ID)
		Expect(err).NotTo(HaveOccurred())
		Expect(persisted.State).To(Equal("unknown"))
	})
})
