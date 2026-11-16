package postgres_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"

	captaincli "github.com/flanksource/captain/pkg/cli"
	framework "github.com/flanksource/clicky"
	"github.com/flanksource/clicky/route"
	"github.com/flanksource/clicky/rpc"
	"github.com/flanksource/commons-db/dbtest"
	"github.com/flanksource/gavel/comments"
	"github.com/flanksource/gavel/comments/agents"
	"github.com/flanksource/gavel/comments/agents/agentstest"
	api "github.com/flanksource/gavel/comments/clicky"
	"github.com/flanksource/gavel/comments/postgres"
	"github.com/flanksource/gavel/internal/database"
	"github.com/flanksource/gavel/projects"
	"github.com/flanksource/gavel/settings"
	"github.com/flanksource/gavel/verify"
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
		commentsAPI(svc)
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
		Expect(document.Paths).To(HaveKey("/api/v1/comments/agent-sessions"))
		Expect(document.Paths).To(HaveKey("/api/v1/comments/mentionables"))
		Expect(document.Paths).To(HaveKey("/api/v1/comments/agent"))
		Expect(document.Paths).To(HaveKey("/api/v1/comments/deliver"))
		Expect(string(document.Paths["/api/v1/comments/agent-sessions"])).To(And(ContainSubstring(`"managed"`), ContainSubstring(`"resumable"`)))
		Expect(document.Paths).NotTo(HaveKey("/api/v1/comments/routes"))
		Expect(string(document.Paths["/api/v1/comments"])).To(ContainSubstring(`"type":"object"`))
		var commentsPath struct {
			Post struct {
				RequestBody struct {
					Content map[string]struct {
						Schema struct {
							Properties map[string]struct {
								Items struct {
									Required []string `json:"required"`
								} `json:"items"`
							} `json:"properties"`
						} `json:"schema"`
					} `json:"content"`
				} `json:"requestBody"`
			} `json:"post"`
		}
		Expect(json.Unmarshal(document.Paths["/api/v1/comments"], &commentsPath)).To(Succeed())
		Expect(commentsPath.Post.RequestBody.Content["application/json"].Schema.Properties["attachments"].Items.Required).To(Equal([]string{"filename", "contentType", "data"}))
		Expect(string(document.Paths["/api/v1/comments"])).To(ContainSubstring(`"enum":["matched","ambiguous","outdated"]`))
		Expect(document.Paths).To(HaveKey("/api/v1/comments/read"))
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

// actorAccess is access acting as actor.
type actorAccess struct {
	access
	actor comments.Actor
}

func (a actorAccess) Principal(context.Context) (comments.Principal, error) {
	return comments.Principal{TenantID: a.tenant, Actor: a.actor}, nil
}

var _ = Describe("comment agent settings over HTTP", func() {
	It("saves and reads a page's comment agent, refusing hosts without agent settings and non-human writers", func(ctx SpecContext) {
		GinkgoT().Setenv("HOME", GinkgoT().TempDir())
		db := dbtest.ForGinkgo(dbtest.Options{Name: "comments_agent_http"})
		Expect(database.Migrate(ctx, db.DSN())).To(Succeed())
		store, err := postgres.New(db.Gorm())
		Expect(err).NotTo(HaveOccurred())
		configured, err := settings.NewDB(db.Gorm())
		Expect(err).NotTo(HaveOccurred())
		dir := GinkgoT().TempDir()
		routes := agents.WebRoutes(projects.NewDB(db.Gorm(), filepath.Join(dir, "no-projects.json")), configured)
		delivery, err := agents.Delivery(agents.Options{Bridge: &agentstest.Bridge{Dir: dir}, Groups: store, Workdir: func(context.Context, comments.Target) (string, error) { return dir, nil }})
		Expect(err).NotTo(HaveOccurred())
		host := func(a comments.Access, routes comments.RouteResolver) http.Handler {
			svc, err := comments.New(comments.Options{Store: store, Access: a, Targets: access{tenant: "tenant-agent"}, Workflows: []comments.Workflow{comments.HumanReview()}, Routes: routes, Delivery: delivery})
			Expect(err).NotTo(HaveOccurred())
			router := route.NewRouter(nil)
			api.RegisterDeliveryRoutes(router, func(context.Context) (*comments.Service, error) { return svc, nil })
			return router
		}
		person := host(access{tenant: "tenant-agent"}, routes)
		const scope, page = "https://example.test/site", "https://example.test/site/pages/home"
		call := func(handler http.Handler, method string, body any) *httptest.ResponseRecorder {
			data, err := json.Marshal(body)
			Expect(err).NotTo(HaveOccurred())
			path := "/api/v1/comments/agent"
			if method == http.MethodGet {
				path += "?scope=" + url.QueryEscape(scope) + "&uri=" + url.QueryEscape(body.(string))
			}
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, httptest.NewRequest(method, path, bytes.NewReader(data)))
			return w
		}
		save := func(handler http.Handler, patch any) *httptest.ResponseRecorder {
			return call(handler, http.MethodPut, map[string]any{"target": comments.Target{ScopeURI: scope, URI: page}, "layer": settings.Ref{Kind: settings.KindProject}, "page": true, "agent": patch})
		}
		type agentSettings struct {
			Settings settings.Resolved           `json:"settings"`
			Agent    verify.ResolvedCommentAgent `json:"agent"`
			Session  string                      `json:"session"`
		}
		settingsOf := func(w *httptest.ResponseRecorder) agentSettings {
			GinkgoHelper()
			Expect(w.Code).To(Equal(200), w.Body.String())
			var out agentSettings
			Expect(json.Unmarshal(w.Body.Bytes(), &out)).To(Succeed())
			Expect(out.Settings.Positions).NotTo(BeEmpty())
			return out
		}
		agentOf := func(w *httptest.ResponseRecorder) verify.ResolvedCommentAgent {
			GinkgoHelper()
			return settingsOf(w).Agent
		}
		enabled := true
		source := verify.CommentAgentSource{Layer: "db:project:", Page: "pages/home"}
		Expect(agentOf(save(person, &verify.CommentAgent{Enabled: &enabled, Group: "site-review"}))).To(Equal(verify.ResolvedCommentAgent{
			Agent: verify.CommentAgent{Enabled: &enabled, Group: "site-review"}, Provenance: map[string]verify.CommentAgentSource{"enabled": source, "group": source}, Found: true,
		}))
		Expect(agentOf(save(person, map[string]any{"group": nil, "edits": "propose"})).Agent).To(Equal(verify.CommentAgent{Enabled: &enabled, Edits: verify.CommentEditsPropose}), "a merge patch keeps absent fields and removes null ones")
		Expect(save(person, map[string]any{"edits": "rewrite"}).Code).To(Equal(http.StatusBadRequest))

		var mentionables []comments.Mentionable
		w := httptest.NewRecorder()
		person.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/comments/mentionables?scope="+url.QueryEscape(scope)+"&target="+url.QueryEscape(page), nil))
		Expect(w.Code).To(Equal(http.StatusOK), w.Body.String())
		Expect(json.Unmarshal(w.Body.Bytes(), &mentionables)).To(Succeed())
		Expect(mentionables).To(ContainElements(comments.Mentionable{ID: comments.MentionAgent, Name: "agent", Kind: "agent"}, comments.Mentionable{ID: "runtime:claude", Name: "claude", Kind: "agent"}))
		Expect(settingsOf(call(person, http.MethodGet, page)).Session).To(Equal(""))
		Expect(agentOf(save(person, nil)).Found).To(BeFalse())

		live := routes
		live.Session = "live-9"
		Expect(settingsOf(call(host(access{tenant: "tenant-agent"}, live), http.MethodGet, page)).Session).To(Equal("live-9"), "GET reports the --session that takes new threads")

		agent := host(actorAccess{access: access{tenant: "tenant-agent"}, actor: comments.Actor{Kind: "agent", ID: "claude:live-1"}}, routes)
		undelivered, err := comments.New(comments.Options{Store: store, Access: access{tenant: "tenant-agent"}, Targets: access{tenant: "tenant-agent"}, Workflows: []comments.Workflow{comments.HumanReview()}})
		Expect(err).NotTo(HaveOccurred())
		noDelivery := route.NewRouter(nil)
		api.RegisterDeliveryRoutes(noDelivery, func(context.Context) (*comments.Service, error) { return undelivered, nil })
		w = httptest.NewRecorder()
		noDelivery.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/comments/mentionables?scope="+url.QueryEscape(scope)+"&target="+url.QueryEscape(page), nil))
		Expect([]int{
			save(agent, &verify.CommentAgent{Enabled: &enabled}).Code,
			call(agent, http.MethodGet, page).Code,
			call(person, http.MethodGet, "https://elsewhere.test/home").Code,
			call(host(access{tenant: "tenant-agent"}, nil), http.MethodGet, page).Code,
			w.Code,
		}).To(Equal([]int{http.StatusForbidden, http.StatusOK, http.StatusBadRequest, http.StatusNotImplemented, http.StatusNotImplemented}))
	})
})

var _ = Describe("agent delivery over HTTP", func() {
	It("delivers a batch of threads in one message, refuses a stale or empty batch, and lists an ended managed session as resumable", func(ctx SpecContext) {
		db := dbtest.ForGinkgo(dbtest.Options{Name: "comments_delivery_http"})
		Expect(database.Migrate(ctx, db.DSN())).To(Succeed())
		store, err := postgres.New(db.Gorm())
		Expect(err).NotTo(HaveOccurred())
		dir := GinkgoT().TempDir()
		bridge := &agentstest.Bridge{Dir: dir, Sessions: []captaincli.SessionRecord{{ID: "live-1", Source: "claude", Model: "claude-opus", Slug: "review-pass", Live: &captaincli.SessionLiveWire{CWD: dir}}}}
		delivery, err := agents.Delivery(agents.Options{Bridge: bridge, Groups: store, Workdir: func(context.Context, comments.Target) (string, error) { return dir, nil }})
		Expect(err).NotTo(HaveOccurred())
		a := access{tenant: "tenant-deliver"}
		svc, err := comments.New(comments.Options{Store: store, Access: a, Targets: a, Workflows: []comments.Workflow{comments.HumanReview()}, Delivery: delivery})
		Expect(err).NotTo(HaveOccurred())
		router := route.NewRouter(nil)
		api.RegisterDeliveryRoutes(router, func(context.Context) (*comments.Service, error) { return svc, nil })
		const scope = "git://repo"
		call := func(method, path string, body any, out any) int {
			GinkgoHelper()
			data, err := json.Marshal(body)
			Expect(err).NotTo(HaveOccurred())
			w := httptest.NewRecorder()
			router.ServeHTTP(w, httptest.NewRequest(method, path, bytes.NewReader(data)))
			if out != nil && w.Code == http.StatusOK {
				Expect(json.Unmarshal(w.Body.Bytes(), out)).To(Succeed(), w.Body.String())
			}
			return w.Code
		}
		create := func(body string) comments.Comment {
			c, err := svc.Create(ctx, comments.CreateRequest{Target: comments.Target{URI: scope + "/widget.go", ScopeURI: scope, Kind: "code"}, Body: body, WorkflowID: "human-review", RequestID: "create-" + body})
			Expect(err).NotTo(HaveOccurred())
			return c
		}
		first, second := create("Check the error"), create("Name the constant")
		batch := func(requestID string, threads ...comments.Comment) map[string]any {
			items := []comments.DeliverThread{}
			for _, c := range threads {
				items = append(items, comments.DeliverThread{ID: c.ID, ExpectedVersion: c.Version})
			}
			return map[string]any{"requestId": requestID, "target": comments.DeliverTarget{Kind: "session", SessionID: "live-1"}, "threads": items}
		}
		var delivered []struct {
			ID      string `json:"id"`
			Content struct {
				AgentRun agents.AgentRun `json:"agentRun"`
			} `json:"content"`
		}
		Expect(call(http.MethodPost, "/api/v1/comments/deliver", batch("batch-http", first, second), &delivered)).To(Equal(http.StatusOK))
		runs := []any{}
		for _, c := range delivered {
			runs = append(runs, []string{c.ID, c.Content.AgentRun.SessionID, c.Content.AgentRun.State})
		}
		sent, _ := bridge.Snapshot()
		Expect([]any{runs, len(sent)}).To(Equal([]any{[]any{[]string{first.ID, "live-1", "sent"}, []string{second.ID, "live-1", "sent"}}, 1}))
		Expect([]int{
			call(http.MethodPost, "/api/v1/comments/deliver", batch("batch-http-stale", first, second), nil),
			call(http.MethodPost, "/api/v1/comments/deliver", batch("batch-http-empty"), nil),
		}).To(Equal([]int{http.StatusConflict, http.StatusBadRequest}))

		third := create("Log the retry")
		run := comments.DeliverTarget{Kind: "run", Prompt: "editor-comments"}
		Expect(call(http.MethodPost, "/api/v1/comments/"+third.ID+"/deliver", map[string]any{"expectedVersion": third.Version, "requestId": "run-http", "target": run}, nil)).To(Equal(http.StatusOK))
		Expect(bridge.End("managed-1")).To(Succeed())
		var sessions []comments.AgentSession
		Expect(call(http.MethodGet, "/api/v1/comments/agent-sessions?scope="+url.QueryEscape(scope), nil, &sessions)).To(Equal(http.StatusOK))
		Expect(sessions).To(Equal([]comments.AgentSession{
			{ID: "live-1", Source: "claude", Title: "review-pass", Model: "claude-opus", CWD: dir},
			{ID: "managed-1", Title: "Log the retry", CWD: dir, Status: "ended", Managed: true, Resumable: true},
		}))
		var mentionables []comments.Mentionable
		Expect(call(http.MethodGet, "/api/v1/comments/mentionables?scope="+url.QueryEscape(scope)+"&target="+url.QueryEscape(scope+"/widget.go"), nil, &mentionables)).To(Equal(http.StatusOK))
		Expect(mentionables).To(ContainElements(
			comments.Mentionable{ID: "runtime:claude", Name: "claude", Kind: "agent"},
			comments.Mentionable{ID: "session:live-1", Name: "review-pass", Kind: "agent"},
			comments.Mentionable{ID: "session:managed-1", Name: "managed-1", Kind: "agent"},
		))
		Expect(mentionables).NotTo(ContainElement(HaveField("ID", comments.MentionAgent)), "a host without comment agent rules has no @agent")
	})
})
