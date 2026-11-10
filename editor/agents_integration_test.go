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
			// Captain's render, without the process database the default opens first.
			Render: captaincli.RenderPrompt,
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
		Expect([]any{claimed.Status, claimed.Assignee.Kind, content.AgentRun.State, content.AgentRun.SessionID, launched[0].Options.PresetsSet}).
			To(Equal([]any{"in_progress", "agent", "sent", "managed-session-1", false}))
	})

	It("lists the project's own prompts beside the library's, with the presets and host variables a run takes", func() {
		home := GinkgoT().TempDir()
		GinkgoT().Setenv("HOME", home)
		dir := mustTarget(doc).Dir
		prompts := filepath.Join(dir, ".captain", "prompts")
		Expect(os.MkdirAll(prompts, 0o755)).To(Succeed())
		DeferCleanup(os.RemoveAll, filepath.Join(dir, ".captain"))
		Expect(os.WriteFile(filepath.Join(prompts, "x-review.prompt"), []byte("---\nname: x-review\ndescription: Review the thread\nmodel: claude\nmode: agent\n---\n{{{message}}}\n"), 0o600)).To(Succeed())

		response := call("GET", "/api/file/"+doc.ID+"/agent-run/context", "", nil)
		Expect(response.Code).To(Equal(200), response.Body.String())
		var runContext struct {
			Prompts       []captaincli.PromptSummary `json:"prompts"`
			Presets       json.RawMessage            `json:"presets"`
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
		Expect(string(runContext.Presets)).To(HavePrefix("["))
		Expect([]any{runContext.DefaultPrompt, runContext.HostVariables, runContext.Dir}).
			To(Equal([]any{"editor-comments", []string{"dir", "scope", "target", "actor", "gavelExecutable", "message"}, dir}))
	})

	var toolThread comments.Comment
	It("runs the chosen prompt with its presets and variables, the host's variables winning, and gives it the thread tools", func(ctx SpecContext) {
		launched = nil
		place := mustTarget(doc)
		var err error
		toolThread, err = svc.Create(ctx, comments.CreateRequest{Target: comments.Target{ScopeURI: place.Scope, URI: place.URI, Kind: "document", Revision: doc.Revision, Selector: selector}, Body: "Shorten the title", WorkflowID: "human-review", RequestID: "run-chosen"})
		Expect(err).NotTo(HaveOccurred())
		body := `{"prompt":"x-review","presets":["fast","strict"],"variables":{"dir":"/elsewhere","focus":"intro"}}`
		response := call("POST", "/api/file/"+doc.ID+"/comments/"+toolThread.ID+"/run", body, nil)
		Expect(response.Code).To(Equal(200), response.Body.String())

		Expect(launched).To(HaveLen(1))
		var vars map[string]any
		Expect(json.Unmarshal([]byte(launched[0].Options.Vars), &vars)).To(Succeed())
		Expect([]any{launched[0].Reference, launched[0].Options.Presets, launched[0].Options.PresetsSet, toolNames(launched[0].Tools), vars["dir"], vars["focus"]}).
			To(Equal([]any{"x-review", []string{"fast", "strict"}, true, []string{"list_threads", "reply_to_thread", "resolve_thread"}, place.Dir, "intro"}))
	})

	It("replies to and resolves its thread as the agent through the thread tools, and the agent cannot close it", func(ctx SpecContext) {
		tools := map[string]api.ToolDefinition{}
		for _, tool := range launched[0].Tools {
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
		launched = nil
		place := mustTarget(doc)
		thread, err := svc.Create(ctx, comments.CreateRequest{Target: comments.Target{ScopeURI: place.Scope, URI: place.URI, Kind: "document", Revision: doc.Revision, Selector: selector}, Body: "Fix the typo", WorkflowID: "human-review", RequestID: "run-preview"})
		Expect(err).NotTo(HaveOccurred())
		response := call("POST", "/api/file/"+doc.ID+"/comments/"+thread.ID+"/run/preview", `{}`, nil)
		Expect(response.Code).To(Equal(200), response.Body.String())
		var preview struct {
			Model, Mode string
			Resolution  json.RawMessage
			Tools       []struct{ Name, Source string }
		}
		Expect(json.Unmarshal(response.Body.Bytes(), &preview)).To(Succeed())
		Expect([]any{preview.Mode, preview.Tools}).To(Equal([]any{"agent", []struct{ Name, Source string }{
			{"list_threads", "thread"}, {"reply_to_thread", "thread"}, {"resolve_thread", "thread"},
		}}))
		Expect(preview.Model).NotTo(BeEmpty())
		Expect(string(preview.Resolution)).To(ContainSubstring(`"spec"`))

		unclaimed, err := svc.Get(ctx, thread.ID)
		Expect(err).NotTo(HaveOccurred())
		Expect([]any{unclaimed.Status, unclaimed.Assignee, launched}).To(Equal([]any{"open", (*comments.Actor)(nil), []captaincli.PromptInvocation(nil)}))

		missing := call("POST", "/api/file/"+doc.ID+"/comments/"+thread.ID+"/run/preview", `{"prompt":"no-such-prompt"}`, nil)
		Expect(missing.Code).To(Equal(http.StatusBadRequest), missing.Body.String())
		toolless := call("POST", "/api/file/"+doc.ID+"/comments/"+thread.ID+"/run/preview", `{"spec":{"model":"gemini-2.5-pro","mode":"cli"}}`, nil)
		Expect([]any{toolless.Code, toolless.Body.String()}).To(ConsistOf(http.StatusBadRequest, ContainSubstring("cannot expose caller tools")))
	})

	It("layers the dialog's runtime over the prompt and appends its instructions to the thread", func(ctx SpecContext) {
		launched = nil
		place := mustTarget(doc)
		thread, err := svc.Create(ctx, comments.CreateRequest{Target: comments.Target{ScopeURI: place.Scope, URI: place.URI, Kind: "document", Revision: doc.Revision, Selector: selector}, Body: "Tighten the intro", WorkflowID: "human-review", RequestID: "run-runtime"})
		Expect(err).NotTo(HaveOccurred())
		model, instructions := "claude-opus-4-6", "Keep the heading"
		body := `{"spec":{"model":"` + model + `","mode":"cli","permissions":{"mode":"plan"}},"instructions":"` + instructions + `"}`
		response := call("POST", "/api/file/"+doc.ID+"/comments/"+thread.ID+"/run", body, nil)
		Expect(response.Code).To(Equal(200), response.Body.String())

		Expect(launched).To(HaveLen(1))
		spec := launched[0].Options.Spec
		Expect(spec).NotTo(BeNil())
		Expect([]any{spec.Model.Name, spec.Model.Mode, spec.Permissions.Mode}).To(Equal([]any{model, api.ModeCLI, api.PermissionPlan}))
		var vars map[string]string
		Expect(json.Unmarshal([]byte(launched[0].Options.Vars), &vars)).To(Succeed())
		Expect(vars["message"]).To(ContainSubstring("## Additional instructions\n\n" + instructions))
	})

	It("refuses a run request whose body is not a run request", func(ctx SpecContext) {
		launched = nil
		place := mustTarget(doc)
		thread, err := svc.Create(ctx, comments.CreateRequest{Target: comments.Target{ScopeURI: place.Scope, URI: place.URI, Kind: "document", Revision: doc.Revision, Selector: selector}, Body: "Fix it", WorkflowID: "human-review", RequestID: "run-invalid"})
		Expect(err).NotTo(HaveOccurred())
		response := call("POST", "/api/file/"+doc.ID+"/comments/"+thread.ID+"/run", `{"spec":"opus"}`, nil)
		Expect([]any{response.Code, launched}).To(Equal([]any{400, []captaincli.PromptInvocation(nil)}))
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
