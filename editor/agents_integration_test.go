package editor

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/flanksource/captain/pkg/api"
	captaincli "github.com/flanksource/captain/pkg/cli"
	"github.com/flanksource/commons-db/dbtest"
	"github.com/flanksource/gavel/comments"
	"github.com/flanksource/gavel/comments/agents"
	"github.com/flanksource/gavel/comments/agents/agentstest"
	commentapi "github.com/flanksource/gavel/comments/clicky"
	"github.com/flanksource/gavel/comments/postgres"
	"github.com/flanksource/gavel/internal/database"
	"github.com/flanksource/gavel/internal/safefs"
	ui "github.com/flanksource/gavel/pr/ui"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Delivering an editor comment thread through the shared comments routes", Ordered, func() {
	const sessionID = "claude-session-1"
	var (
		handler  http.Handler
		doc      Document
		root     comments.Comment
		bridge   *agentstest.Bridge
		cli      *comments.Service
		svc      *comments.Service
		selector json.RawMessage
		agent    = comments.Actor{Kind: "agent", ID: "claude:" + sessionID, Label: "claude-opus"}
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
	get := func(id string) comments.Comment {
		response := call("GET", "/api/v1/comments/"+id, "", nil)
		Expect(response.Code).To(Equal(200), response.Body.String())
		var c comments.Comment
		Expect(json.Unmarshal(response.Body.Bytes(), &c)).To(Succeed())
		return c
	}
	thread := func(ctx context.Context, body string) comments.Comment {
		place := mustTarget(doc)
		c, err := svc.Create(ctx, comments.CreateRequest{Target: comments.Target{ScopeURI: place.Scope, URI: place.URI, Kind: "document", Revision: doc.Revision, Selector: selector}, Body: body, WorkflowID: "human-review", RequestID: body})
		Expect(err).NotTo(HaveOccurred())
		return c
	}
	// deliverRun is the editor dialog's request: the chosen prompt plus its run options.
	deliverRun := func(c comments.Comment, prompt, runOptions string) *httptest.ResponseRecorder {
		body := `{"expectedVersion":` + jsonInt(c.Version) + `,"requestId":"run-` + c.ID + `","target":{"kind":"run","prompt":"` + prompt + `","runOptions":` + runOptions + `}}`
		return call("POST", "/api/v1/comments/"+c.ID+"/deliver", body, nil)
	}
	launches := func() []captaincli.PromptInvocation {
		_, launched := bridge.Snapshot()
		return launched
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
		bridge = &agentstest.Bridge{Dir: mustTarget(doc).Dir, Sessions: []captaincli.SessionRecord{{ID: sessionID, Source: "claude", Model: "claude-opus", Slug: "rename-pass", Live: &captaincli.SessionLiveWire{PID: 42, CWD: dir}}}}
		delivery, err := agents.Delivery(agents.Options{Bridge: bridge, Groups: storage, Workdir: ui.CommentWorkdir})
		Expect(err).NotTo(HaveOccurred())
		host := commentHost{store: files}
		svc, err = comments.New(comments.Options{Store: storage, Access: host, Targets: host, Workflows: []comments.Workflow{comments.HumanReview()}, Blobs: storage, AttachmentLimit: 1024, Routes: storage, Delivery: delivery})
		Expect(err).NotTo(HaveOccurred())
		// The service `gavel comments` runs: pr/ui's project host over the same tables.
		cli, err = comments.New(comments.Options{Store: storage, Access: ui.CommentAccess{}, Targets: ui.CommentAccess{}, Workflows: []comments.Workflow{comments.HumanReview()}, Blobs: storage, AttachmentLimit: 1024})
		Expect(err).NotTo(HaveOccurred())
		selector, err = json.Marshal(CommentSelector{Kind: "document", View: "source", Version: 1})
		Expect(err).NotTo(HaveOccurred())
		place := mustTarget(doc)
		root, err = svc.Create(ctx, comments.CreateRequest{Target: comments.Target{ScopeURI: place.Scope, URI: place.URI, Kind: "document", Revision: doc.Revision, Selector: selector}, Body: "Rename x to y", WorkflowID: "human-review", RequestID: "send-1", Content: json.RawMessage(`{"screenshot":{"status":"unavailable"}}`)})
		Expect(err).NotTo(HaveOccurred())
		server := &Server{Store: files, CommentService: func(context.Context) (*comments.Service, error) { return svc, nil }, Agents: bridge}
		// The entity resolver is a fallback: Handler serves every request through its own resolver.
		Expect(commentapi.Register(func(context.Context) (*comments.Service, error) { return svc, nil })).To(Succeed())
		you, err := commentapi.LocalUser()
		Expect(err).NotTo(HaveOccurred())
		server.Comments = commentapi.LocalActor(commentapi.Handler(server.OpenComments), you)
		handler = server.Handler()
	})

	It("lists the live session working in the file's project", func() {
		response := call("GET", "/api/v1/comments/agent-sessions?scope="+url.QueryEscape(mustTarget(doc).Scope), "", nil)
		Expect(response.Code).To(Equal(200), response.Body.String())
		var sessions []comments.AgentSession
		Expect(json.Unmarshal(response.Body.Bytes(), &sessions)).To(Succeed())
		Expect(sessions).To(HaveLen(1))
		Expect([]string{sessions[0].ID, sessions[0].Title}).To(Equal([]string{sessionID, "rename-pass"}))
	})

	It("sends a thread to a session as its agent and records the run beside the existing content", func() {
		response := call("POST", "/api/v1/comments/"+root.ID+"/deliver", `{"expectedVersion":1,"requestId":"send-root","target":{"kind":"session","sessionId":"`+sessionID+`"}}`, nil)
		Expect(response.Code).To(Equal(200), response.Body.String())
		c := get(root.ID)
		var content struct {
			Screenshot map[string]string `json:"screenshot"`
			AgentRun   agents.AgentRun   `json:"agentRun"`
		}
		Expect(json.Unmarshal(c.Content, &content)).To(Succeed())
		Expect([]any{c.Status, *c.Assignee, content.Screenshot["status"], content.AgentRun.State, content.AgentRun.Target}).
			To(Equal([]any{"in_progress", agent, "unavailable", "sent", agents.AgentRunTarget{Label: "rename-pass", Channel: "claude-socket"}}))
		delivered, _ := bridge.Snapshot()
		Expect(delivered).To(HaveLen(1))
		Expect(delivered[0].SessionID).To(Equal(sessionID))
		Expect(delivered[0].Text).To(ContainSubstring("on a.yaml"))
		Expect(delivered[0].Text).To(ContainSubstring("gavel comments reply " + root.ID + " --expectedVersion"))
		Expect(delivered[0].Text).To(ContainSubstring("Rename x to y"))
	})

	It("posts the agent's reply as the agent and refuses the agent closing the thread", func() {
		reply := call("POST", "/api/v1/comments/"+root.ID+"/reply", `{"expectedVersion":`+jsonInt(get(root.ID).Version)+`,"bodyMarkdown":"Renamed and linted"}`, &agent)
		Expect(reply.Code).To(Equal(200), reply.Body.String())
		closing := call("POST", "/api/v1/comments/"+root.ID+"/close", `{"expectedVersion":`+jsonInt(get(root.ID).Version)+`}`, &agent)
		Expect(closing.Code).NotTo(Equal(200), closing.Body.String())
		c := get(root.ID)
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

	It("lists the project's own prompts beside the library's, with the presets and host variables a run takes", func() {
		home := GinkgoT().TempDir()
		GinkgoT().Setenv("HOME", home)
		dir := mustTarget(doc).Dir
		prompts := filepath.Join(dir, ".captain", "prompts")
		Expect(os.MkdirAll(prompts, 0o755)).To(Succeed())
		DeferCleanup(safefs.SafeRemove, filepath.Join(dir, ".captain"), safefs.RemoveContext{AllowedRoot: dir})
		Expect(os.WriteFile(filepath.Join(prompts, "x-review.prompt"), []byte("---\nname: x-review\ndescription: Review the thread\nmodel: claude\nmode: agent\n---\n{{{message}}}\n"), 0o600)).To(Succeed())

		response := call("GET", "/api/file/"+doc.ID+"/agent-run/context", "", nil)
		Expect(response.Code).To(Equal(200), response.Body.String())
		var runContext struct {
			Prompts       []captaincli.PromptSummary `json:"prompts"`
			DefaultPrompt string                     `json:"defaultPrompt"`
			HostVariables []string                   `json:"hostVariables"`
			Dir           string                     `json:"dir"`
		}
		Expect(json.Unmarshal(response.Body.Bytes(), &runContext)).To(Succeed())
		names := []string{}
		for _, prompt := range runContext.Prompts {
			names = append(names, prompt.Name)
		}
		Expect(names).To(ContainElements("x-review", "editor-comments"))
		Expect([]any{runContext.DefaultPrompt, runContext.HostVariables, runContext.Dir}).
			To(Equal([]any{"editor-comments", []string{"dir", "scope", "target", "actor", "gavelExecutable", "message"}, dir}))
	})

	var toolThread comments.Comment
	It("runs the chosen prompt with its presets and variables, the host's variables winning, and gives it the thread tools", func(ctx SpecContext) {
		toolThread = thread(ctx, "Shorten the title")
		response := deliverRun(toolThread, "x-review", `{"presets":["fast","strict"],"variables":{"dir":"/elsewhere","focus":"intro"},"instructions":"Keep the heading"}`)
		Expect(response.Code).To(Equal(200), response.Body.String())

		launched := launches()
		Expect(launched).To(HaveLen(1))
		place := mustTarget(doc)
		var vars map[string]any
		Expect(json.Unmarshal([]byte(launched[0].Options.Vars), &vars)).To(Succeed())
		Expect([]any{launched[0].Reference, launched[0].Options.Presets, launched[0].Options.PresetsSet, toolNames(launched[0].Tools), vars["dir"], vars["scope"], vars["target"], vars["focus"]}).
			To(Equal([]any{"x-review", []string{"fast", "strict"}, true, []string{"list_threads", "reply_to_thread", "resolve_thread"}, place.Dir, place.Scope, place.URI, "intro"}))
		Expect(vars["message"]).To(ContainSubstring("## Additional instructions\n\nKeep the heading"))
		var content struct {
			AgentRun agents.AgentRun `json:"agentRun"`
		}
		claimed := get(toolThread.ID)
		Expect(json.Unmarshal(claimed.Content, &content)).To(Succeed())
		Expect([]any{claimed.Status, claimed.Assignee.Kind, content.AgentRun.State, content.AgentRun.SessionID}).To(Equal([]any{"in_progress", "agent", "sent", "managed-1"}))
	})

	It("replies to and resolves its thread as the agent through the thread tools, and the agent cannot close it", func(ctx SpecContext) {
		tools := map[string]api.ToolDefinition{}
		for _, tool := range launches()[0].Tools {
			tools[tool.Name] = tool
		}
		listed, err := tools["list_threads"].Handler(ctx, map[string]any{})
		Expect(err).NotTo(HaveOccurred())
		Expect(listed).To(ContainElement(HaveField("ID", toolThread.ID)))
		_, err = tools["reply_to_thread"].Handler(ctx, map[string]any{"id": toolThread.ID, "body": "Shortened to four words"})
		Expect(err).NotTo(HaveOccurred())
		_, err = tools["resolve_thread"].Handler(ctx, map[string]any{"id": toolThread.ID, "summary": "title shortened"})
		Expect(err).NotTo(HaveOccurred())

		c, err := svc.Get(ctx, toolThread.ID)
		Expect(err).NotTo(HaveOccurred())
		last := c.Messages[len(c.Messages)-1]
		Expect([]any{last.Author, last.Body, c.Status, c.Resolution.Actor}).
			To(Equal([]any{*c.Assignee, "Shortened to four words", "resolved", *c.Assignee}))
		_, err = svc.Act(commentapi.WithActor(ctx, *c.Assignee), comments.ActionRequest{ID: c.ID, ExpectedVersion: c.Version, Action: "close"})
		Expect(err).To(HaveOccurred())
	})

	It("previews the run's resolved runtime and tools without claiming the thread or launching", func(ctx SpecContext) {
		c := thread(ctx, "Fix the typo")
		preview := func(target string) *httptest.ResponseRecorder {
			return call("POST", "/api/v1/comments/"+c.ID+"/deliver/preview", `{"target":`+target+`}`, nil)
		}
		response := preview(`{"kind":"run","prompt":"editor-comments"}`)
		Expect(response.Code).To(Equal(200), response.Body.String())
		var rendered agents.RunPreview
		Expect(json.Unmarshal(response.Body.Bytes(), &rendered)).To(Succeed())
		Expect([]any{rendered.Mode, rendered.Tools}).To(Equal([]any{"agent", []agents.PreviewTool{
			{Name: "list_threads", Description: "List the comment threads on this session's target, with their replies and current versions.", Source: "thread"},
			{Name: "reply_to_thread", Description: "Reply to a comment thread as this session's agent. Say what you changed and how you verified it.", Source: "thread"},
			{Name: "resolve_thread", Description: "Resolve a comment thread once its change is made, verified and replied to. A human then reviews it.", Source: "thread"},
		}}))
		unclaimed := get(c.ID)
		Expect([]any{unclaimed.Status, unclaimed.Assignee, len(launches())}).To(Equal([]any{"open", (*comments.Actor)(nil), 1}))

		missing := preview(`{"kind":"run","prompt":"no-such-prompt"}`)
		Expect(missing.Code).To(Equal(http.StatusBadRequest), missing.Body.String())
		toolless := preview(`{"kind":"run","prompt":"editor-comments","runOptions":{"spec":{"model":"gemini-2.5-pro","mode":"cli"}}}`)
		Expect([]any{toolless.Code, toolless.Body.String()}).To(ConsistOf(http.StatusBadRequest, ContainSubstring("cannot expose caller tools")))
	})

	It("refuses run options that are not a run request without claiming the thread", func(ctx SpecContext) {
		c := thread(ctx, "Fix it")
		response := deliverRun(c, "editor-comments", `{"spec":"opus"}`)
		Expect([]any{response.Code, len(launches()), get(c.ID).Version}).To(Equal([]any{400, 1, c.Version}))
	})
})

func toolNames(tools []api.ToolDefinition) []string {
	names := make([]string, 0, len(tools))
	for _, tool := range tools {
		names = append(names, tool.Name)
	}
	return names
}

func jsonInt(v int64) string {
	data, _ := json.Marshal(v)
	return string(data)
}
