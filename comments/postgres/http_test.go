package postgres_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"

	framework "github.com/flanksource/clicky"
	"github.com/flanksource/clicky/route"
	"github.com/flanksource/clicky/rpc"
	"github.com/flanksource/commons-db/dbtest"
	"github.com/flanksource/gavel/comments"
	api "github.com/flanksource/gavel/comments/clicky"
	"github.com/flanksource/gavel/comments/postgres"
	"github.com/flanksource/gavel/internal/database"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/cobra"
)

var _ = Describe("generated HTTP comments", func() {
	It("creates and closes through generated routes and returns stale writes as JSON conflicts", func(ctx SpecContext) {
		db := dbtest.ForGinkgo(dbtest.Options{Name: "comments_http"})
		Expect(database.Migrate(ctx, db.DSN())).To(Succeed())
		store, err := postgres.New(db.Gorm())
		Expect(err).NotTo(HaveOccurred())
		a := access{tenant: "tenant-http"}
		svc, err := comments.New(comments.Options{Store: store, Access: a, Targets: a, Workflows: []comments.Workflow{comments.HumanReview()}, Blobs: store, AttachmentLimit: 1024})
		Expect(err).NotTo(HaveOccurred())
		Expect(api.Register(func(context.Context) (*comments.Service, error) { return svc, nil })).To(Succeed())
		root := &cobra.Command{Use: "example"}
		framework.GenerateCLI(root)
		server := rpc.NewSwaggerServer(&rpc.ServeConfig{Executor: &rpc.ExecutorConfig{Enabled: true, PathPrefix: "/api/v1"}}, root, nil)
		router := route.NewRouter(nil)
		server.RegisterExecutionRoutes(router)
		router.MountGenerated("GET /api/v1/openapi.json", api.OpenAPI(http.HandlerFunc(server.HandleOpenAPIJSON), "/api/v1"))
		resolve := func(context.Context) (*comments.Service, error) { return svc, nil }
		router.Raw("POST /api/v1/comments/{id}/attachments", api.Attachments(resolve), route.Meta{Entity: "comments", Verb: "attach"})
		router.Raw("GET /api/v1/comments/{id}/attachments/{attachment}", api.Attachments(resolve), route.Meta{Entity: "comments", Verb: "download", ReadOnly: true})
		request := func(method, path string, body any) *httptest.ResponseRecorder {
			data, err := json.Marshal(body)
			Expect(err).NotTo(HaveOccurred())
			r := httptest.NewRequest(method, path, bytes.NewReader(data)).WithContext(context.Background())
			r.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, r)
			return w
		}
		w := request(http.MethodPost, "/api/v1/comments", comments.CreateRequest{Target: comments.Target{URI: "git://repo/widget.go", ScopeURI: "git://repo", Kind: "code"}, Body: "Check error propagation", WorkflowID: "human-review", RequestID: "http-request"})
		Expect(w.Code).To(BeNumerically(">=", 200), w.Body.String())
		Expect(w.Code).To(BeNumerically("<", 300), w.Body.String())
		rows, err := svc.List(ctx, comments.Filter{ScopeURI: "git://repo", Limit: 20})
		Expect(err).NotTo(HaveOccurred())
		Expect(rows).To(HaveLen(1))
		id := rows[0].ID
		doc := request(http.MethodGet, "/api/v1/openapi.json", nil)
		Expect(doc.Code).To(Equal(200), doc.Body.String())
		var document struct {
			Paths map[string]json.RawMessage `json:"paths"`
		}
		Expect(json.Unmarshal(doc.Body.Bytes(), &document)).To(Succeed())
		Expect(document.Paths).To(HaveKey("/api/v1/comments/{id}/attachments"))
		Expect(document.Paths).To(HaveKey("/api/v1/comments/{id}/deliver"))
		Expect(document.Paths).To(HaveKey("/api/v1/comments/routes"))
		Expect(string(document.Paths["/api/v1/comments"])).To(ContainSubstring(`"type":"object"`))
		w = request(http.MethodGet, "/api/v1/comments?scope="+url.QueryEscape("git://repo")+"&limit=20&unread=true", nil)
		Expect(w.Code).To(Equal(200), w.Body.String())
		Expect(w.Header().Get("X-Total-Count")).To(Equal("1"))
		w = request(http.MethodPost, "/api/v1/comments/"+id+"/read", map[string]any{"expectedVersion": 1})
		Expect(w.Code).To(Equal(200), w.Body.String())
		w = request(http.MethodGet, "/api/v1/comments?scope="+url.QueryEscape("git://repo")+"&limit=20&unread=true", nil)
		Expect(w.Code).To(Equal(200), w.Body.String())
		Expect(w.Header().Get("X-Total-Count")).To(Equal("0"))
		upload := httptest.NewRequest(http.MethodPost, "/api/v1/comments/"+id+"/attachments?name=evidence.txt&expectedVersion=1", bytes.NewBufferString("proof"))
		upload.Header.Set("Content-Type", "text/plain")
		w = httptest.NewRecorder()
		router.ServeHTTP(w, upload)
		Expect(w.Code).To(Equal(200), w.Body.String())
		var attached comments.Comment
		Expect(json.Unmarshal(w.Body.Bytes(), &attached)).To(Succeed())
		Expect(attached.Attachments).To(HaveLen(1))
		w = request(http.MethodGet, "/api/v1/comments/"+id+"/attachments/"+attached.Attachments[0].ID, nil)
		Expect(w.Code).To(Equal(200))
		Expect(w.Body.String()).To(Equal("proof"))
		w = request(http.MethodPost, "/api/v1/comments/"+id+"/close", map[string]any{"expectedVersion": 2, "summary": "Reviewed"})
		Expect(w.Code).To(Equal(200), w.Body.String())
		w = request(http.MethodPost, "/api/v1/comments/"+id+"/close", map[string]any{"expectedVersion": 1})
		Expect(w.Code).To(Equal(409), w.Body.String())
		Expect(json.Valid(w.Body.Bytes())).To(BeTrue())
	})
})
