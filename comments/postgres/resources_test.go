package postgres_test

import (
	"io"
	"strings"

	"github.com/flanksource/commons-db/dbtest"
	"github.com/flanksource/gavel/comments"
	"github.com/flanksource/gavel/comments/postgres"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("comment read markers and attachments", func() {
	It("tracks unread versions per actor and preserves binary attachments with tenant isolation", func(ctx SpecContext) {
		db := dbtest.ForGinkgo(dbtest.Options{Name: "comment_resources"})
		Expect(postgres.Provision(ctx, db.DSN())).To(Succeed())
		store, err := postgres.New(db.Gorm())
		Expect(err).NotTo(HaveOccurred())
		a := access{tenant: "resources"}
		svc, err := comments.New(comments.Options{Store: store, Access: a, Targets: a, Workflows: []comments.Workflow{comments.HumanReview()}, Blobs: store, AttachmentLimit: 1024})
		Expect(err).NotTo(HaveOccurred())
		c, err := svc.Create(ctx, comments.CreateRequest{Target: comments.Target{URI: "file://repo/page.tsx", ScopeURI: "git://repo", Kind: "code"}, Body: "Check layout", WorkflowID: "human-review", RequestID: "resources"})
		Expect(err).NotTo(HaveOccurred())
		f := comments.Filter{ScopeURI: c.Target.ScopeURI, Limit: 20, Unread: true}
		page, err := svc.Page(ctx, f)
		Expect(err).NotTo(HaveOccurred())
		Expect(page.Total).To(Equal(int64(1)))
		_, err = svc.MarkRead(ctx, c.ID, c.Version)
		Expect(err).NotTo(HaveOccurred())
		page, err = svc.Page(ctx, f)
		Expect(err).NotTo(HaveOccurred())
		Expect(page.Total).To(BeZero())
		c, err = svc.Upload(ctx, comments.Upload{CommentID: c.ID, ExpectedVersion: c.Version, Name: "evidence.txt", MediaType: "text/plain", Body: strings.NewReader("captured evidence")})
		Expect(err).NotTo(HaveOccurred())
		Expect(c.Attachments).To(HaveLen(1))
		c, err = svc.Get(ctx, c.ID)
		Expect(err).NotTo(HaveOccurred())
		Expect(c.Attachments).To(HaveLen(1))
		meta, reader, err := svc.Download(ctx, c.ID, c.Attachments[0].ID)
		Expect(err).NotTo(HaveOccurred())
		data, err := io.ReadAll(reader)
		Expect(err).NotTo(HaveOccurred())
		Expect(reader.Close()).To(Succeed())
		Expect(string(data)).To(Equal("captured evidence"))
		Expect(meta.Size).To(Equal(int64(len(data))))
		page, err = svc.Page(ctx, f)
		Expect(err).NotTo(HaveOccurred())
		Expect(page.Total).To(Equal(int64(1)))
		_, err = svc.MarkRead(ctx, c.ID, c.Version)
		Expect(err).NotTo(HaveOccurred())
		marker, err := svc.MarkRead(ctx, c.ID, 1)
		Expect(err).NotTo(HaveOccurred())
		Expect(marker.Version).To(Equal(c.Version))
		_, err = svc.MarkRead(ctx, c.ID, c.Version+1)
		Expect(err).To(MatchError(comments.ErrInvalid))
		_, err = store.Open(ctx, "other-tenant", meta.ID)
		Expect(err).To(MatchError(comments.ErrNotFound))
		_, err = svc.Upload(ctx, comments.Upload{CommentID: c.ID, ExpectedVersion: c.Version, Name: "../unsafe.txt", MediaType: "text/plain", Body: strings.NewReader("unsafe")})
		Expect(err).To(MatchError(comments.ErrInvalid))
	})
})
