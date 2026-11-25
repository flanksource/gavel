package postgres_test

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/flanksource/commons-db/dbtest"
	"github.com/flanksource/gavel/comments"
	"github.com/flanksource/gavel/comments/postgres"
	"github.com/flanksource/gavel/internal/database"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("listing comments by path", func() {
	const scope = "gavel://project/acme"
	locations := map[string]comments.Location{
		"a.go": {ScopeURI: scope, TargetURI: scope + "/files/a.go"},
		"dir":  {ScopeURI: scope, TargetPrefix: scope + "/files/dir/"},
		".":    {ScopeURI: scope},
	}
	locate := func(_ context.Context, path string) (comments.Location, error) {
		if l, ok := locations[path]; ok {
			return l, nil
		}
		return comments.Location{}, fmt.Errorf("%w: %s is not inside a project", comments.ErrInvalid, path)
	}

	It("filters a file exactly, a directory by prefix and the project root not at all", func(ctx SpecContext) {
		db := dbtest.ForGinkgo(dbtest.Options{Name: "comment_paths"})
		Expect(database.Migrate(ctx, db.DSN())).To(Succeed())
		store, err := postgres.New(db.Gorm())
		Expect(err).NotTo(HaveOccurred())
		a := access{tenant: "paths"}
		svc, err := comments.New(comments.Options{Store: store, Access: a, Targets: a, Workflows: []comments.Workflow{comments.HumanReview()}, Locate: locate})
		Expect(err).NotTo(HaveOccurred())
		for _, file := range []string{"a.go", "dir/b.go", "dir2/c.go"} {
			_, err = svc.Create(ctx, comments.CreateRequest{Target: comments.Target{URI: scope + "/files/" + file, ScopeURI: scope, Kind: "code"}, Body: "Check " + file, WorkflowID: "human-review", RequestID: file})
			Expect(err).NotTo(HaveOccurred())
		}
		listed := func(path string) []string {
			page, err := svc.Page(ctx, comments.Filter{Path: path, Limit: 20})
			Expect(err).NotTo(HaveOccurred())
			uris := []string{}
			for _, c := range page.Items {
				uris = append(uris, strings.TrimPrefix(c.Target.URI, scope+"/files/"))
			}
			return uris
		}
		Expect(listed("a.go")).To(ConsistOf("a.go"))
		Expect(listed("dir")).To(ConsistOf("dir/b.go"), "dir/ must not match dir2/")
		Expect(listed(".")).To(ConsistOf("a.go", "dir/b.go", "dir2/c.go"))

		_, err = svc.Page(ctx, comments.Filter{Path: "elsewhere", Limit: 20})
		Expect(err).To(MatchError(ContainSubstring("elsewhere is not inside a project")))
		_, err = svc.Page(ctx, comments.Filter{Path: "a.go", TargetURI: scope + "/files/a.go", Limit: 20})
		Expect(err).To(MatchError(comments.ErrInvalid))
		_, err = svc.Page(ctx, comments.Filter{Path: "a.go", ScopeURI: "gavel://project/other", Limit: 20})
		Expect(err).To(MatchError(ContainSubstring("not in scope gavel://project/other")))
		_, err = svc.Page(ctx, comments.Filter{Limit: 20})
		Expect(err).To(MatchError(ContainSubstring("scope or path is required")))

		bare, err := comments.New(comments.Options{Store: store, Access: a, Targets: a, Workflows: []comments.Workflow{comments.HumanReview()}})
		Expect(err).NotTo(HaveOccurred())
		_, err = bare.Page(ctx, comments.Filter{Path: "a.go", Limit: 20})
		Expect(err).To(MatchError(ContainSubstring("cannot list comments by path")))
	})
})

var _ = Describe("comment read markers and attachments", func() {
	It("tracks unread versions per actor and preserves binary attachments with tenant isolation", func(ctx SpecContext) {
		db := dbtest.ForGinkgo(dbtest.Options{Name: "comment_resources"})
		Expect(database.Migrate(ctx, db.DSN())).To(Succeed())
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
