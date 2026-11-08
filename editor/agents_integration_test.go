package editor

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"

	"github.com/flanksource/captain/pkg/api"
	captaincli "github.com/flanksource/captain/pkg/cli"
	"github.com/flanksource/captain/pkg/session/deliver"
	"github.com/flanksource/commons-db/dbtest"
	"github.com/flanksource/gavel/comments"
	commentapi "github.com/flanksource/gavel/comments/clicky"
	"github.com/flanksource/gavel/comments/postgres"
	"github.com/flanksource/gavel/internal/database"
	ui "github.com/flanksource/gavel/pr/ui"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Sending an editor comment thread to an agent session", Ordered, func() {
	const sessionID = "claude-session-1"
	var (
		handler   http.Handler
		doc       Document
		root      comments.Comment
		delivered []string
		cli       *comments.Service
		svc       *comments.Service
		launched  []captaincli.PromptInvocation
		selector  json.RawMessage
		agent     = comments.Actor{Kind: "agent", ID: "claude:" + sessionID, Label: "claude-opus"}
	)
	call := func(method, path, body string, actor *comments.Actor) *httptest.ResponseRecorder {
		req := httptest.NewRequest(method, path, strings.NewReader(body))
		req.RemoteAddr = "127.0.0.1:40000"
		req.Header.Set("Content-Type", "application/json")
		if actor != nil {
			req.Header.Set(commentapi.ActorHeader, commentapi.EncodeActor(*actor))
		}
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, req)
		return response
	}
	current := func() comments.Comment {
		response := call("GET", "/api/v1/comments/"+root.ID, "", nil)
		Expect(response.Code).To(Equal(200), response.Body.String())
		var c comments.Comment
		Expect(json.Unmarshal(response.Body.Bytes(), &c)).To(Succeed())
		return c
	}

	BeforeAll(func(ctx SpecContext) {
		db := dbtest.ForGinkgo(dbtest.Options{Name: "editor_agent_send"})
		Expect(database.Migrate(ctx, db.DSN())).To(Succeed())
		dir := GinkgoT().TempDir()
		useProject(dir)
		path := filepath.Join(dir, "a.yaml")
		Expect(os.WriteFile(path, []byte("name: x\n"), 0600)).To(Succeed())
		files, err := NewStore([]string{path})
		Expect(err).NotTo(HaveOccurred())
		doc, err = files.Read(files.Files()[0].ID)
		Expect(err).NotTo(HaveOccurred())
		storage, err := postgres.New(db.Gorm())
		Expect(err).NotTo(HaveOccurred())
		host := commentHost{store: files}
		svc, err = comments.New(comments.Options{Store: storage, Access: host, Targets: host, Workflows: []comments.Workflow{comments.HumanReview()}, Blobs: storage, AttachmentLimit: 1024})
		Expect(err).NotTo(HaveOccurred())
		// The service `gavel comments` runs: pr/ui's project host over the same tables.
		cli, err = comments.New(comments.Options{Store: storage, Access: ui.CommentAccess{}, Targets: ui.CommentAccess{}, Workflows: []comments.Workflow{comments.HumanReview()}, Blobs: storage, AttachmentLimit: 1024})
		Expect(err).NotTo(HaveOccurred())
		selector, err = json.Marshal(CommentSelector{Kind: "document", View: "source", Version: 1})
		Expect(err).NotTo(HaveOccurred())
		place := mustTarget(doc)
		root, err = svc.Create(ctx, comments.CreateRequest{Target: comments.Target{ScopeURI: place.Scope, URI: place.URI, Kind: "document", Revision: doc.Revision, Selector: selector}, Body: "Rename x to y", WorkflowID: "human-review", RequestID: "send-1", Content: json.RawMessage(`{"screenshot":{"status":"unavailable"}}`)})
		Expect(err).NotTo(HaveOccurred())
		server := &Server{Store: files, CommentService: func(context.Context) (*comments.Service, error) { return svc, nil }, Agents: AgentBridge{
			Live: func(context.Context) ([]captaincli.SessionRecord, error) {
				return []captaincli.SessionRecord{{ID: sessionID, Source: "claude", Model: "claude-opus", Slug: "rename-pass", Live: &captaincli.SessionLiveWire{PID: 42, CWD: dir}}}, nil
			},
			Deliver: func(_ context.Context, id, text string) (deliver.Result, error) {
				delivered = append(delivered, id+"\n"+text)
				return deliver.Result{Channel: deliver.ChannelClaudeSocket, Status: deliver.StatusSent}, nil
			},
			Launch: func(_ context.Context, invocation captaincli.PromptInvocation) (captaincli.PromptRunResult, error) {
				launched = append(launched, invocation)
				invocation.OnEvent(0, api.Event{Kind: api.EventSystem, SessionID: "managed-session-1"})
				return captaincli.PromptRunResult{RunID: "run-1"}, nil
			},
		}}
		// The entity resolver is a fallback: Handler serves every request through its own resolver.
		Expect(commentapi.Register(func(context.Context) (*comments.Service, error) { return svc, nil })).To(Succeed())
		you, err := commentapi.LocalUser()
		Expect(err).NotTo(HaveOccurred())
		server.Comments = commentapi.LocalActor(commentapi.Handler(server.OpenComments), you)
		handler = server.Handler()
	})

	It("lists the live session working on the edited file", func() {
		response := call("GET", "/api/agent-sessions", "", nil)
		Expect(response.Code).To(Equal(200), response.Body.String())
		var sessions []AgentSession
		Expect(json.Unmarshal(response.Body.Bytes(), &sessions)).To(Succeed())
		Expect(sessions).To(HaveLen(1))
		Expect([]string{sessions[0].ID, sessions[0].Title}).To(Equal([]string{sessionID, "rename-pass"}))
	})

	It("assigns and acknowledges the thread as the agent, delivers it, and records the run beside the existing content", func() {
		response := call("POST", "/api/file/"+doc.ID+"/comments/"+root.ID+"/send", `{"sessionId":"`+sessionID+`"}`, nil)
		Expect(response.Code).To(Equal(200), response.Body.String())
		c := current()
		var content struct {
			Screenshot map[string]string `json:"screenshot"`
			AgentRun   AgentRun          `json:"agentRun"`
		}
		Expect(json.Unmarshal(c.Content, &content)).To(Succeed())
		Expect([]any{c.Status, *c.Assignee, content.Screenshot["status"], content.AgentRun.State, content.AgentRun.Target}).
			To(Equal([]any{"in_progress", agent, "unavailable", "sent", AgentRunTarget{Label: "rename-pass", Channel: "claude-socket"}}))
		Expect(delivered).To(HaveLen(1))
		Expect(delivered[0]).To(HavePrefix(sessionID + "\n"))
		Expect(delivered[0]).To(ContainSubstring("export GAVEL_COMMENT_ACTOR=" + commentapi.EncodeActor(agent)))
		Expect(delivered[0]).To(ContainSubstring("gavel comments reply " + root.ID + " --expectedVersion"))
		Expect(delivered[0]).To(ContainSubstring("gavel comments resolve " + root.ID + " --expectedVersion"))
		Expect(delivered[0]).NotTo(ContainSubstring("curl"))
		Expect(delivered[0]).To(ContainSubstring("Rename x to y"))
	})

	It("refuses a session that is not live", func() {
		response := call("POST", "/api/file/"+doc.ID+"/comments/"+root.ID+"/send", `{"sessionId":"gone"}`, nil)
		Expect(response.Code).To(Equal(http.StatusConflict), response.Body.String())
	})

	It("posts the agent's reply as the agent and refuses the agent closing the thread", func() {
		reply := call("POST", "/api/v1/comments/"+root.ID+"/reply", `{"expectedVersion":`+jsonInt(current().Version)+`,"bodyMarkdown":"Renamed and linted"}`, &agent)
		Expect(reply.Code).To(Equal(200), reply.Body.String())
		closing := call("POST", "/api/v1/comments/"+root.ID+"/close", `{"expectedVersion":`+jsonInt(current().Version)+`}`, &agent)
		Expect(closing.Code).NotTo(Equal(200), closing.Body.String())
		c := current()
		last := c.Messages[len(c.Messages)-1]
		Expect([]any{last.Author, last.Body, c.Status}).To(Equal([]any{agent, "Renamed and linted", "in_progress"}))
	})

	It("lets gavel comments reply to and resolve the editor thread as the agent the env asserts", func(ctx SpecContext) {
		asAgent := commentapi.WithActor(ctx, agent)
		threads, err := cli.List(asAgent, comments.Filter{ScopeURI: mustTarget(doc).Scope, Limit: 10})
		Expect(err).NotTo(HaveOccurred())
		Expect(threads).To(HaveLen(1))
		c, err := cli.Act(asAgent, comments.ActionRequest{ID: root.ID, ExpectedVersion: threads[0].Version, Action: "reply", Body: "Done via the CLI"})
		Expect(err).NotTo(HaveOccurred())
		c, err = cli.Act(asAgent, comments.ActionRequest{ID: root.ID, ExpectedVersion: c.Version, Action: "resolve", Summary: "renamed"})
		Expect(err).NotTo(HaveOccurred())
		last := c.Messages[len(c.Messages)-1]
		Expect([]any{last.Author, c.Status, c.Resolution.Actor}).To(Equal([]any{agent, "resolved", agent}))
	})

	It("starts a managed session for a thread from the editor-comments prompt and records the session it reports", func(ctx SpecContext) {
		place := mustTarget(doc)
		thread, err := svc.Create(ctx, comments.CreateRequest{Target: comments.Target{ScopeURI: place.Scope, URI: place.URI, Kind: "document", Revision: doc.Revision, Selector: selector}, Body: "Add a summary", WorkflowID: "human-review", RequestID: "run-1"})
		Expect(err).NotTo(HaveOccurred())
		response := call("POST", "/api/file/"+doc.ID+"/comments/"+thread.ID+"/run", `{}`, nil)
		Expect(response.Code).To(Equal(200), response.Body.String())

		Expect(launched).To(HaveLen(1))
		var vars map[string]string
		Expect(json.Unmarshal([]byte(launched[0].Options.Vars), &vars)).To(Succeed())
		claimed, err := svc.Get(ctx, thread.ID)
		Expect(err).NotTo(HaveOccurred())
		Expect([]any{launched[0].Reference, vars["dir"], vars["scope"], vars["target"], vars["actor"]}).
			To(Equal([]any{"editor-comments", place.Dir, place.Scope, place.URI, commentapi.EncodeActor(*claimed.Assignee)}))
		Expect(vars["message"]).To(ContainSubstring("Add a summary"))
		var content struct {
			AgentRun AgentRun `json:"agentRun"`
		}
		Expect(json.Unmarshal(claimed.Content, &content)).To(Succeed())
		Expect([]any{claimed.Status, claimed.Assignee.Kind, content.AgentRun.State, content.AgentRun.SessionID}).
			To(Equal([]any{"in_progress", "agent", "sent", "managed-session-1"}))
	})
})

func jsonInt(v int64) string {
	data, _ := json.Marshal(v)
	return string(data)
}
