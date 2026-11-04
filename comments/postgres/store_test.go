package postgres_test

import (
	"context"
	"github.com/flanksource/commons-db/dbtest"
	"github.com/flanksource/gavel/comments"
	"github.com/flanksource/gavel/comments/postgres"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"testing"
)

func TestPostgres(t *testing.T) { RegisterFailHandler(Fail); RunSpecs(t, "Comments Postgres") }

type access struct{ tenant string }

func (a access) Principal(context.Context) (comments.Principal, error) {
	return comments.Principal{TenantID: a.tenant, Actor: comments.Actor{Kind: "user", ID: "reviewer"}}, nil
}
func (a access) Authorize(context.Context, comments.Principal, string, comments.Target) error {
	return nil
}
func (a access) Resolve(_ context.Context, _ comments.Principal, t comments.Target) (comments.Target, error) {
	return t, nil
}

var _ = Describe("transactional comments", func() {
	It("rejects stale and cross-tenant writes, preserves outcomes in audit and deduplicates creation", func(ctx SpecContext) {
		db := dbtest.ForGinkgo(dbtest.Options{Name: "comments"})
		Expect(db.Gorm().Exec("CREATE TABLE host_sentinel (value text NOT NULL)").Error).NotTo(HaveOccurred())
		Expect(db.Gorm().Exec("INSERT INTO host_sentinel VALUES ('preserved')").Error).NotTo(HaveOccurred())
		Expect(postgres.Provision(ctx, db.DSN())).To(Succeed())
		var sentinel string
		Expect(db.Gorm().Raw("SELECT value FROM host_sentinel").Scan(&sentinel).Error).NotTo(HaveOccurred())
		Expect(sentinel).To(Equal("preserved"))
		Expect(postgres.Provision(ctx, db.DSN())).To(Succeed())
		store, err := postgres.New(db.Gorm())
		Expect(err).NotTo(HaveOccurred())
		a := access{tenant: "tenant-a"}
		svc, err := comments.New(comments.Options{Store: store, Access: a, Targets: a, Workflows: []comments.Workflow{comments.HumanReview()}})
		Expect(err).NotTo(HaveOccurred())
		req := comments.CreateRequest{Target: comments.Target{URI: "git://repo/file.go", ScopeURI: "git://repo", Kind: "code", Revision: "commit-a"}, Body: "Check the error", WorkflowID: "human-review", RequestID: "request-a"}
		c, err := svc.Create(ctx, req)
		Expect(err).NotTo(HaveOccurred())
		again, err := svc.Create(ctx, req)
		Expect(err).NotTo(HaveOccurred())
		Expect(again.ID).To(Equal(c.ID))
		req.Body = "different request"
		_, err = svc.Create(ctx, req)
		Expect(err).To(MatchError(comments.ErrConflict))
		reply, err := svc.Act(ctx, comments.ActionRequest{ID: c.ID, ExpectedVersion: 1, Action: "reply", Body: "Investigated"})
		Expect(err).NotTo(HaveOccurred())
		Expect(reply.Messages).To(HaveLen(1))
		_, err = svc.Act(ctx, comments.ActionRequest{ID: c.ID, ExpectedVersion: 1, Action: "close"})
		Expect(err).To(MatchError(comments.ErrConflict))
		_, err = store.Get(ctx, "tenant-b", c.ID)
		Expect(err).To(MatchError(comments.ErrNotFound))
		closed, err := svc.Act(ctx, comments.ActionRequest{ID: c.ID, ExpectedVersion: 2, Action: "close", Summary: "Reviewed"})
		Expect(err).NotTo(HaveOccurred())
		Expect(closed.Outcome.Summary).To(Equal("Reviewed"))
		reopened, err := svc.Act(ctx, comments.ActionRequest{ID: c.ID, ExpectedVersion: 3, Action: "reopen"})
		Expect(err).NotTo(HaveOccurred())
		Expect(reopened.Outcome).To(BeNil())
		events, err := store.Events(ctx, "tenant-a", "git://repo", 0, 100)
		Expect(err).NotTo(HaveOccurred())
		Expect(events).To(HaveLen(4))
		Expect(string(events[3].Data)).To(ContainSubstring("Reviewed"))
	})
})
