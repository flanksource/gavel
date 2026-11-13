package editor

import (
	"bytes"
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

var _ = Describe("Editor comment persistence", func() {
	It("round trips threads and attachments and resolves shifted patch context without rewriting captures", func(ctx SpecContext) {
		db := dbtest.ForGinkgo(dbtest.Options{Name: "editor_comments"})
		Expect(database.Migrate(ctx, db.DSN())).To(Succeed())
		dir := GinkgoT().TempDir()
		useProject(dir)
		path := filepath.Join(dir, "settings.yaml")
		source := "header\nselected\nfooter"
		Expect(os.WriteFile(path, []byte(source), 0600)).To(Succeed())
		files, err := NewStore([]string{path})
		Expect(err).NotTo(HaveOccurred())
		doc, err := files.Read(files.Files()[0].ID)
		Expect(err).NotTo(HaveOccurred())
		storage, err := postgres.New(db.Gorm())
		Expect(err).NotTo(HaveOccurred())
		host := commentHost{store: files}
		svc, err := comments.New(comments.Options{Store: storage, Access: host, Targets: host, Workflows: []comments.Workflow{comments.HumanReview()}, Blobs: storage, AttachmentLimit: 1024})
		Expect(err).NotTo(HaveOccurred())
		patch, err := comments.CapturePatch(source, 2, 2)
		Expect(err).NotTo(HaveOccurred())
		selector, err := json.Marshal(CommentSelector{Kind: "code-range", View: "source", Version: 1, Patch: &patch})
		Expect(err).NotTo(HaveOccurred())
		place := mustTarget(doc)
		root, err := svc.Create(ctx, comments.CreateRequest{Target: comments.Target{ScopeURI: place.Scope, URI: place.URI, Kind: "code-range", Revision: doc.Revision, Selector: selector}, Body: "Review selected", WorkflowID: "human-review", RequestID: "editor-capture"})
		Expect(err).NotTo(HaveOccurred())
		root, err = svc.Act(ctx, comments.ActionRequest{ID: root.ID, ExpectedVersion: 1, Action: "reply", Body: "Reviewed"})
		Expect(err).NotTo(HaveOccurred())
		Expect(root.Messages).To(HaveLen(1))
		root, err = svc.Upload(ctx, comments.Upload{CommentID: root.ID, ExpectedVersion: root.Version, Name: "screenshot.png", MediaType: "image/png", Body: bytes.NewBufferString("image fixture")})
		Expect(err).NotTo(HaveOccurred())
		Expect(root.Attachments).To(HaveLen(1))
		_, err = svc.Act(ctx, comments.ActionRequest{ID: root.ID, ExpectedVersion: 1, Action: "resolve"})
		Expect(err).To(MatchError(comments.ErrConflict))
		_, err = files.Write(Document{File: doc.File, Source: "inserted\n" + source, Revision: doc.Revision})
		Expect(err).NotTo(HaveOccurred())
		server := &Server{Store: files, CommentService: func(context.Context) (*comments.Service, error) { return svc, nil }}
		response := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/api/file/"+doc.ID+"/comments", nil)
		req.SetPathValue("id", doc.ID)
		server.listComments(response, req)
		Expect(response.Code).To(Equal(200))
		Expect(response.Body.String()).To(ContainSubstring(`"startLine":3`))
		stored, err := svc.Get(ctx, root.ID)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(stored.Target.Selector)).To(MatchJSON(selector))
	})
})
