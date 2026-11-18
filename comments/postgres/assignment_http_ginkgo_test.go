package postgres_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"

	"github.com/flanksource/commons-db/dbtest"
	"github.com/flanksource/gavel/comments"
	"github.com/flanksource/gavel/comments/postgres"
	"github.com/flanksource/gavel/internal/database"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("comment assignment over HTTP", func() {
	It("assigns an actor and clears it with JSON null through the generated route", func(ctx SpecContext) {
		db := dbtest.ForGinkgo(dbtest.Options{Name: "comments_assignment_http"})
		Expect(database.Migrate(ctx, db.DSN())).To(Succeed())
		store, err := postgres.New(db.Gorm())
		Expect(err).NotTo(HaveOccurred())
		host := access{tenant: "tenant-assignment"}
		svc, err := comments.New(comments.Options{Store: store, Access: host, Targets: host, Workflows: []comments.Workflow{comments.HumanReview()}})
		Expect(err).NotTo(HaveOccurred())
		handler := commentsAPI(svc)
		thread, err := svc.Create(ctx, comments.CreateRequest{
			Target: comments.Target{URI: "git://repo/widget.go", ScopeURI: "git://repo", Kind: "code"},
			Body:   "Check assignment", WorkflowID: "human-review", RequestID: "assignment-http",
		})
		Expect(err).NotTo(HaveOccurred())
		reviewer := &comments.Actor{Kind: "user", ID: "reviewer", Label: "Reviewer"}
		for _, assignee := range []*comments.Actor{reviewer, nil} {
			data, err := json.Marshal(map[string]any{"expectedVersion": thread.Version, "assignee": assignee})
			Expect(err).NotTo(HaveOccurred())
			request := httptest.NewRequest(http.MethodPost, "/api/v1/comments/"+thread.ID+"/assign", bytes.NewReader(data))
			request.Header.Set("Content-Type", "application/json")
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, request)
			Expect(response.Code).To(Equal(http.StatusOK), response.Body.String())
			var updated comments.Comment
			Expect(json.Unmarshal(response.Body.Bytes(), &updated)).To(Succeed())
			Expect(updated.Assignee).To(Equal(assignee))
			Expect(updated.Version).To(Equal(thread.Version + 1))
			stored, err := svc.Get(ctx, thread.ID)
			Expect(err).NotTo(HaveOccurred())
			Expect(stored.Assignee).To(Equal(assignee))
			thread = updated
		}
	})
})
