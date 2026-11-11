package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	captaincli "github.com/flanksource/captain/pkg/cli"
	"github.com/flanksource/commons-db/dbtest"
	"github.com/flanksource/gavel/comments"
	"github.com/flanksource/gavel/comments/agents"
	"github.com/flanksource/gavel/comments/agents/agentstest"
	"github.com/flanksource/gavel/internal/database"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestTesthost(t *testing.T) { RegisterFailHandler(Fail); RunSpecs(t, "Comments testhost") }

const scope = "gavel://project/acme"

var _ = Describe("Agent delivery and routes over HTTP", Ordered, func() {
	var (
		handler http.Handler
		bridge  *agentstest.Bridge
	)
	call := func(method, path string, body any) *httptest.ResponseRecorder {
		var data []byte
		if body != nil {
			var err error
			data, err = json.Marshal(body)
			Expect(err).NotTo(HaveOccurred())
		}
		r := httptest.NewRequest(method, path, bytes.NewReader(data))
		r.Header.Set("Content-Type", "application/json")
		r.Header.Set("X-Gavel-Comment-Scope", scope)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)
		return w
	}
	decode := func(w *httptest.ResponseRecorder, out any) {
		Expect(json.Unmarshal(w.Body.Bytes(), out)).To(Succeed(), w.Body.String())
	}
	create := func(page string) comments.Comment {
		w := call("POST", "/api/v1/comments", comments.CreateRequest{Target: comments.Target{URI: scope + "/pages/" + page, ScopeURI: scope, Kind: "web-page"}, Body: "Fix " + page, WorkflowID: "human-review", RequestID: "create-" + page})
		Expect(w.Code).To(Equal(200), w.Body.String())
		var c comments.Comment
		decode(w, &c)
		return c
	}
	agentRunOf := func(id string) agents.AgentRun {
		w := call("GET", "/api/v1/comments/"+id, nil)
		Expect(w.Code).To(Equal(200), w.Body.String())
		var c struct {
			Content struct {
				AgentRun agents.AgentRun `json:"agentRun"`
			} `json:"content"`
		}
		decode(w, &c)
		return c.Content.AgentRun
	}
	sessionTarget := map[string]string{"kind": "session", "sessionId": "fake-session"}

	BeforeAll(func(ctx SpecContext) {
		db := dbtest.ForGinkgo(dbtest.Options{Name: "comments_testhost"})
		Expect(database.Migrate(ctx, db.DSN())).To(Succeed())
		dir := GinkgoT().TempDir()
		bridge = &agentstest.Bridge{Dir: dir, Sessions: []captaincli.SessionRecord{{ID: "fake-session", Source: "claude", Model: "claude-test", Slug: "fake-pass", Live: &captaincli.SessionLiveWire{CWD: dir}}}}
		var err error
		handler, err = newHandler(db.Gorm(), bridge, dir, "http://testhost.example.test")
		Expect(err).NotTo(HaveOccurred())
	})

	It("lists live agent sessions without colliding with comment ids", func() {
		w := call("GET", "/api/v1/comments/agent-sessions?all=1", nil)
		Expect(w.Code).To(Equal(200), w.Body.String())
		var sessions []comments.AgentSession
		decode(w, &sessions)
		Expect(sessions).To(Equal([]comments.AgentSession{{ID: "fake-session", Source: "claude", Title: "fake-pass", Model: "claude-test", CWD: bridge.Dir}}))
	})

	It("delivers a thread, answers a stale version 409 and a failed delivery 502 with the error recorded", func() {
		c := create("home")
		w := call("POST", "/api/v1/comments/"+c.ID+"/deliver", map[string]any{"expectedVersion": c.Version, "requestId": "deliver-home", "target": sessionTarget})
		Expect(w.Code).To(Equal(200), w.Body.String())
		var delivered struct {
			Version int64 `json:"version"`
			Content struct {
				AgentRun agents.AgentRun `json:"agentRun"`
			} `json:"content"`
		}
		decode(w, &delivered)
		Expect([]string{delivered.Content.AgentRun.SessionID, delivered.Content.AgentRun.State, delivered.Content.AgentRun.Target.Channel}).To(Equal([]string{"fake-session", "sent", "claude-socket"}))

		w = call("POST", "/api/v1/comments/"+c.ID+"/deliver", map[string]any{"expectedVersion": c.Version, "requestId": "deliver-home-stale", "target": sessionTarget})
		Expect(w.Code).To(Equal(http.StatusConflict), w.Body.String())

		bridge.FailDeliver = "socket closed"
		DeferCleanup(func() { bridge.FailDeliver = "" })
		w = call("POST", "/api/v1/comments/"+c.ID+"/deliver", map[string]any{"expectedVersion": delivered.Version, "requestId": "deliver-home-fail", "target": sessionTarget})
		Expect(w.Code).To(Equal(http.StatusBadGateway), w.Body.String())
		Expect(w.Body.String()).To(ContainSubstring("socket closed"))
		run := agentRunOf(c.ID)
		Expect([]string{run.State, run.SessionID}).To(Equal([]string{"error", "fake-session"}))

		bridge.FailDeliver = ""
		w = call("GET", "/api/v1/comments/"+c.ID, nil)
		var failed comments.Comment
		decode(w, &failed)
		w = call("POST", "/api/v1/comments/"+c.ID+"/deliver", map[string]any{"expectedVersion": failed.Version, "requestId": "deliver-home-retry", "target": map[string]any{"kind": "run", "prompt": "editor-comments", "fresh": true}})
		Expect(w.Code).To(Equal(200), w.Body.String())
		run = agentRunOf(c.ID)
		Expect([]string{run.State, run.Error, run.SessionID}).To(Equal([]string{"sent", "", "managed-1"}))
	})

	It("answers fresh on a session target 400 naming the fresh option", func() {
		c := create("fresh-session")
		w := call("POST", "/api/v1/comments/"+c.ID+"/deliver", map[string]any{"expectedVersion": c.Version, "requestId": "deliver-fresh-session", "target": map[string]any{"kind": "session", "sessionId": "fake-session", "fresh": true}})
		Expect(w.Code).To(Equal(http.StatusBadRequest), w.Body.String())
		Expect(w.Body.String()).To(ContainSubstring("fresh?"))
	})

	It("saves, lists and deletes routes, and a saved route delivers a new root without delaying its create", func() {
		w := call("PUT", "/api/v1/comments/routes", map[string]any{"scopeUri": scope, "targetPrefix": scope + "/pages/", "workflowId": "human-review", "target": sessionTarget, "enabled": true})
		Expect(w.Code).To(Equal(200), w.Body.String())
		var saved comments.Route
		decode(w, &saved)
		Expect(saved.ID).NotTo(BeEmpty())

		routed := create("pricing")
		Expect(routed.Version).To(Equal(int64(1)))
		// The playground attaches its screenshot right after the create, with the create's version.
		upload := httptest.NewRequest("POST", "/api/v1/comments/"+routed.ID+"/attachments?name=screenshot.png&expectedVersion=1", bytes.NewReader([]byte("png")))
		upload.Header.Set("Content-Type", "image/png")
		upload.Header.Set("X-Gavel-Comment-Scope", scope)
		w = httptest.NewRecorder()
		handler.ServeHTTP(w, upload)
		Expect(w.Code).To(Equal(200), w.Body.String())
		Eventually(func() string { return agentRunOf(routed.ID).State }).WithTimeout(10 * time.Second).Should(Equal("sent"))

		w = call("GET", "/api/v1/comments/routes?scope="+url.QueryEscape(scope), nil)
		Expect(w.Code).To(Equal(200), w.Body.String())
		var routes []comments.Route
		decode(w, &routes)
		Expect(routes).To(Equal([]comments.Route{saved}))

		w = call("DELETE", "/api/v1/comments/routes/"+saved.ID, nil)
		Expect(w.Code).To(Equal(http.StatusNoContent), w.Body.String())
		w = call("GET", "/api/v1/comments/routes?scope="+url.QueryEscape(scope), nil)
		decode(w, &routes)
		Expect(routes).To(BeEmpty())
	})
})
