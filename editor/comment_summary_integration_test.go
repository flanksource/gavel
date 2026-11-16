package editor

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"os"
	"path/filepath"

	"github.com/flanksource/commons-db/dbtest"
	"github.com/flanksource/gavel/comments"
	"github.com/flanksource/gavel/comments/postgres"
	"github.com/flanksource/gavel/internal/database"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Editor comment summary", func() {
	It("counts unread comments per file and image, lists recent activity newest first, and marks a file's comments read through its target", func(ctx SpecContext) {
		db := dbtest.ForGinkgo(dbtest.Options{Name: "editor_comment_summary"})
		Expect(database.Migrate(ctx, db.DSN())).To(Succeed())
		dir := GinkgoT().TempDir()
		useProject(dir)
		paths := []string{filepath.Join(dir, "a.yaml"), filepath.Join(dir, "b.yaml")}
		for _, path := range paths {
			Expect(os.WriteFile(path, []byte("name: x\n"), 0600)).To(Succeed())
		}
		imagePath := filepath.Join(dir, "c.png")
		Expect(os.WriteFile(imagePath, []byte("\x89PNG\r\n\x1a\npixels"), 0600)).To(Succeed())
		files, err := NewStore(StoreOptions{Paths: append(paths, imagePath)})
		Expect(err).NotTo(HaveOccurred())
		docs := map[string]Document{}
		for _, entry := range files.Files() {
			doc, err := files.commentable(entry.ID)
			Expect(err).NotTo(HaveOccurred())
			docs[filepath.Base(doc.Path)] = doc
		}
		storage, err := postgres.New(db.Gorm())
		Expect(err).NotTo(HaveOccurred())
		host := commentHost{store: files}
		svc, err := comments.New(comments.Options{Store: storage, Access: host, Targets: host, Workflows: []comments.Workflow{comments.HumanReview()}, Blobs: storage, AttachmentLimit: 1024})
		Expect(err).NotTo(HaveOccurred())
		create := func(doc Document, body string, selector CommentSelector) comments.Comment {
			raw, err := json.Marshal(selector)
			Expect(err).NotTo(HaveOccurred())
			place := mustTarget(doc)
			c, err := svc.Create(ctx, comments.CreateRequest{Target: comments.Target{ScopeURI: place.Scope, URI: place.URI, Kind: selector.Kind, Revision: doc.Revision, Selector: raw}, Body: body, WorkflowID: "human-review", RequestID: "summary-" + body})
			Expect(err).NotTo(HaveOccurred())
			return c
		}
		whole := CommentSelector{Kind: "document", View: "source", Version: 1}
		a, b, c := docs["a.yaml"], docs["b.yaml"], docs["c.png"]
		create(a, "Review A", whole)
		onB := create(b, "Review B", whole)
		_, err = svc.Act(ctx, comments.ActionRequest{ID: onB.ID, ExpectedVersion: onB.Version, Action: "reply", Body: "Replied on B"})
		Expect(err).NotTo(HaveOccurred())
		create(c, "Whole image", CommentSelector{Kind: "document", View: "preview", Version: 1})
		create(c, "This box", CommentSelector{Kind: "region", View: "preview", Version: 1, Region: &comments.Region{X: 0.1, Y: 0.2, W: 0.3, H: 0.4}})

		server := &Server{Store: files, CommentService: func(context.Context) (*comments.Service, error) { return svc, nil }}
		summary := func() CommentSummary {
			response := httptest.NewRecorder()
			server.commentSummary(response, httptest.NewRequest("GET", "/api/comments", nil))
			Expect(response.Code).To(Equal(200), response.Body.String())
			var out CommentSummary
			Expect(json.Unmarshal(response.Body.Bytes(), &out)).To(Succeed())
			return out
		}
		recent := func(s CommentSummary) []string {
			out := []string{}
			for _, r := range s.Recent {
				state := "read"
				if r.Unread {
					state = "unread"
				}
				out = append(out, r.FileID+":"+r.Body+":"+state)
			}
			return out
		}
		before := summary()

		placeA := mustTarget(a)
		Expect(before.Targets[a.ID]).To(Equal(comments.ReadTarget{ScopeURI: placeA.Scope, URI: placeA.URI}))
		Expect(svc.MarkTargetRead(ctx, before.Targets[a.ID])).To(Equal(1))
		after := summary()

		Expect(before.Files).To(Equal(map[string]FileCommentCounts{a.ID: {Total: 1, Unread: 1}, b.ID: {Total: 1, Unread: 1}, c.ID: {Total: 2, Unread: 2}}))
		Expect(recent(before)).To(Equal([]string{c.ID + ":This box:unread", c.ID + ":Whole image:unread", b.ID + ":Replied on B:unread", a.ID + ":Review A:unread"}))
		Expect(after.Files).To(Equal(map[string]FileCommentCounts{a.ID: {Total: 1, Unread: 0}, b.ID: {Total: 1, Unread: 1}, c.ID: {Total: 2, Unread: 2}}))
		Expect(recent(after)).To(Equal([]string{c.ID + ":This box:unread", c.ID + ":Whole image:unread", b.ID + ":Replied on B:unread", a.ID + ":Review A:read"}))
	})
})
