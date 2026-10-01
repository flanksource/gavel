package ui

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/flanksource/captain/pkg/ai/approval"
	"github.com/flanksource/captain/pkg/api"
	captaindb "github.com/flanksource/captain/pkg/database"
	"github.com/flanksource/captain/pkg/promptrun"
	"github.com/flanksource/commons-db/dbtest"
	"github.com/flanksource/gavel/github"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// The approval surface is durable: a request is a `captain_turn_requests` row,
// so it survives the process that raised it and can be answered by a dashboard
// that reconnected. That is the whole reason these specs need a real database —
// the behaviour under test IS the row and the state machine around it.
var _ = Describe("todo session approvals", Ordered, func() {
	var (
		ctx     context.Context
		db      *captaindb.DB
		server  *Server
		session uuid.UUID
		promptR uuid.UUID
	)

	BeforeAll(func() {
		ctx = context.Background()
		handle := dbtest.ForGinkgo(dbtest.Options{Name: "gavel_todo_approvals"})
		var err error
		db, err = captaindb.Open(ctx, captaindb.WithDSN(handle.DSN()), captaindb.WithMigrations())
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(func() { Expect(db.Close()).To(Succeed()) })
	})

	BeforeEach(func() {
		server = &Server{ghOpts: github.Options{WorkDir: GinkgoT().TempDir()}}
		prev := todoApprovalStore
		todoApprovalStore = func(context.Context, string) (*captaindb.DB, error) { return db, nil }
		DeferCleanup(func() { todoApprovalStore = prev })

		created, err := db.CreateOrGetSession(ctx, captaindb.CreateSessionInput{
			ID: uuid.New(), Source: "gavel", Provider: "anthropic",
		})
		Expect(err).NotTo(HaveOccurred())
		session = created.ID
		run, err := db.CreatePromptRun(ctx, captaindb.CreatePromptRunInput{SessionID: session})
		Expect(err).NotTo(HaveOccurred())
		promptR = run.ID
	})

	// pending records a request the way captain's broker does: the row, then the
	// hold that parks the run in waiting — the only state a credential-less
	// approval resolves in.
	pending := func(tool string, input map[string]any) uuid.UUID {
		GinkgoHelper()
		request, err := db.CreateToolApprovalRequest(ctx, captaindb.CreateToolApprovalRequestInput{
			SessionID: session, PromptRunID: promptR,
			ToolCallID: uuid.NewString(), Tool: tool, Input: input,
			RequestedBy: "provider", ExpiresAt: time.Now().Add(time.Hour),
		})
		Expect(err).NotTo(HaveOccurred())
		_, _, err = db.HoldPromptRunForApprovals(ctx, promptR)
		Expect(err).NotTo(HaveOccurred())
		return request.ID
	}

	// pendingTyped records a typed request the way captain's broker stores one:
	// its kind and payload beside the tool and input older readers use.
	pendingTyped := func(request api.ApprovalRequest) uuid.UUID {
		GinkgoHelper()
		request.ToolUseID = uuid.NewString()
		row, err := db.CreateToolApprovalRequest(ctx, captaindb.CreateToolApprovalRequestInput{
			SessionID: session, PromptRunID: promptR,
			ToolCallID: request.ToolUseID, Tool: request.Tool, Input: request.Input,
			RequestedBy: "provider", ExpiresAt: time.Now().Add(time.Hour), Approval: &request,
		})
		Expect(err).NotTo(HaveOccurred())
		_, _, err = db.HoldPromptRunForApprovals(ctx, promptR)
		Expect(err).NotTo(HaveOccurred())
		return row.ID
	}

	commandRequest := func() api.ApprovalRequest {
		return api.ApprovalRequest{
			Tool: "exec_command", Input: map[string]any{"command": "make build"},
			Kind: api.ApprovalKindCommand, Interruptible: true,
			SupportedScopes: []api.ApprovalScope{api.ApprovalScopeRequest, api.ApprovalScopeSession},
			Command:         &api.CommandApproval{Command: "make build", Cwd: "/work/acme"},
		}
	}

	approveWithQuery := func(query string, body map[string]any) *httptest.ResponseRecorder {
		GinkgoHelper()
		raw, err := json.Marshal(body)
		Expect(err).NotTo(HaveOccurred())
		rec := httptest.NewRecorder()
		server.handleTodoSessionApprove(rec, httptest.NewRequest(
			http.MethodPost, "/api/todos/session/approve"+query, bytes.NewReader(raw)))
		return rec
	}

	approve := func(body map[string]any) *httptest.ResponseRecorder {
		GinkgoHelper()
		return approveWithQuery("?sessionId="+session.String(), body)
	}

	listPending := func() []todoApproval {
		GinkgoHelper()
		rec := httptest.NewRecorder()
		server.handleTodoSessionApprovals(rec, httptest.NewRequest(http.MethodGet,
			"/api/todos/session/approvals?sessionId="+session.String()+"&promptRunId="+promptR.String(), nil))
		Expect(rec.Code).To(Equal(http.StatusOK), rec.Body.String())
		var body struct {
			Approvals []todoApproval `json:"approvals"`
		}
		Expect(json.Unmarshal(rec.Body.Bytes(), &body)).To(Succeed())
		return body.Approvals
	}

	It("lists the run's unanswered requests with the tool and its input", func() {
		id := pending("Bash", map[string]any{"command": "ls"})

		approvals := listPending()

		Expect(approvals).To(HaveLen(1))
		Expect(approvals[0].ID).To(Equal(id.String()))
		Expect(approvals[0].Tool).To(Equal("Bash"))
		Expect(approvals[0].Input).To(HaveKeyWithValue("command", "ls"))
	})

	// The dashboard labels and answers a request by its kind, so it reads the
	// whole stored document — the typed payload travels beside tool and input.
	It("lists a command request with its kind and the stored request document", func() {
		id := pendingTyped(commandRequest())

		approvals := listPending()

		Expect(approvals).To(HaveLen(1))
		Expect(approvals[0].ID).To(Equal(id.String()))
		Expect(approvals[0].Kind).To(Equal(api.ApprovalKindCommand))
		Expect(approvals[0].Request).To(Equal(map[string]any{
			"tool": "exec_command", "input": map[string]any{"command": "make build"},
			"kind": "command", "interruptible": true, "supportedScopes": []any{"request", "session"},
			"command": map[string]any{"command": "make build", "cwd": "/work/acme"},
		}))
	})

	// A row written before kinds existed only ever held a tool approval.
	It("lists a kindless row as a tool request", func() {
		pending("Bash", map[string]any{"command": "ls"})

		approvals := listPending()

		Expect(approvals).To(HaveLen(1))
		Expect(approvals[0].Kind).To(Equal(api.ApprovalKindTool))
		Expect(approvals[0].Request).To(Equal(map[string]any{"tool": "Bash", "input": map[string]any{"command": "ls"}}))
	})

	// A notification carries the approval id and nothing else, so a client may
	// answer it without knowing the session; the row names its own.
	It("answers an approval addressed by its id alone", func() {
		id := pending("Bash", map[string]any{"command": "ls"})

		rec := approveWithQuery("", map[string]any{"approvalId": id.String(), "action": "approve"})

		Expect(rec.Code).To(Equal(http.StatusOK), rec.Body.String())
		Expect(stateOf(ctx, db, id)).To(Equal(captaindb.TurnRequestStateApproved))
	})

	It("rejects a sessionId that is not a UUID rather than ignoring it", func() {
		id := pending("Bash", nil)
		const malformedSession = "not-a-session"

		rec := approveWithQuery("?sessionId="+malformedSession, map[string]any{"approvalId": id.String(), "action": "approve"})

		Expect(rec.Code).To(Equal(http.StatusBadRequest), rec.Body.String())
		Expect(rec.Body.String()).To(ContainSubstring(malformedSession))
		Expect(stateOf(ctx, db, id)).To(Equal(captaindb.TurnRequestStatePending))
	})

	It("reports an approval that does not exist as not found", func() {
		rec := approve(map[string]any{"approvalId": uuid.NewString(), "action": "approve"})

		Expect(rec.Code).To(Equal(http.StatusNotFound), rec.Body.String())
	})

	It("does not find an approval under a session it does not belong to", func() {
		id := pending("Bash", nil)

		rec := approveWithQuery("?sessionId="+uuid.NewString(), map[string]any{"approvalId": id.String(), "action": "approve"})

		Expect(rec.Code).To(Equal(http.StatusNotFound), rec.Body.String())
		Expect(stateOf(ctx, db, id)).To(Equal(captaindb.TurnRequestStatePending))
	})

	It("approves a request, so the broker reads an allow decision", func() {
		id := pending("Bash", map[string]any{"command": "ls"})

		rec := approve(map[string]any{"approvalId": id.String(), "action": "approve"})

		Expect(rec.Code).To(Equal(http.StatusOK), rec.Body.String())
		Expect(stateOf(ctx, db, id)).To(Equal(captaindb.TurnRequestStateApproved))
	})

	It("denies a request and carries the reason back to the agent", func() {
		id := pending("Bash", map[string]any{"command": "rm -rf /"})

		rec := approve(map[string]any{"approvalId": id.String(), "action": "deny", "message": "never that"})

		Expect(rec.Code).To(Equal(http.StatusOK), rec.Body.String())
		request, err := db.GetTurnRequest(ctx, id)
		Expect(err).NotTo(HaveOccurred())
		Expect(request.State).To(Equal(captaindb.TurnRequestStateDenied))
		Expect(request.Reason).To(Equal("never that"))
	})

	It("responds with replacement input, which runs the call rather than refusing it", func() {
		id := pending("Bash", map[string]any{"command": "ls /"})

		rec := approve(map[string]any{
			"approvalId": id.String(), "action": "respond",
			"input": map[string]any{"command": "ls ."},
		})

		Expect(rec.Code).To(Equal(http.StatusOK), rec.Body.String())
		request, err := db.GetTurnRequest(ctx, id)
		Expect(err).NotTo(HaveOccurred())
		Expect(request.State).To(Equal(captaindb.TurnRequestStateApproved))
		Expect(request.Response).To(HaveKey("updatedInput"))
	})

	It("refuses a respond that names no replacement input", func() {
		id := pending("Bash", map[string]any{"command": "ls"})

		rec := approve(map[string]any{"approvalId": id.String(), "action": "respond"})

		Expect(rec.Code).To(Equal(http.StatusConflict), rec.Body.String())
		Expect(rec.Body.String()).To(ContainSubstring("replacement tool input"))
		Expect(stateOf(ctx, db, id)).To(Equal(captaindb.TurnRequestStatePending))
	})

	// Claude's AskUserQuestion gives its questions no ids, so the dashboard keys
	// its answers by position. The shaping is the asking provider's job — gavel
	// stores the answers as sent and never second-guesses their shape.
	It("stores the answers to an id-less question verbatim", func() {
		questions := []any{
			map[string]any{"header": "Scope", "question": "How far should this land?", "multiSelect": false},
			map[string]any{"header": "Surface", "question": "Which surface?", "multiSelect": false},
		}
		id := pending("AskUserQuestion", map[string]any{"questions": questions})
		answers := map[string]any{"1": "Phases 0+1 only", "2": "Not applicable / defer"}

		rec := approve(map[string]any{"approvalId": id.String(), "action": "respond", "input": map[string]any{"questions": questions, "answers": answers}})

		Expect(rec.Code).To(Equal(http.StatusOK), rec.Body.String())
		Expect(stateOf(ctx, db, id)).To(Equal(captaindb.TurnRequestStateApproved))
		request, err := db.GetTurnRequest(ctx, id)
		Expect(err).NotTo(HaveOccurred())
		Expect(request.Response["updatedInput"]).To(HaveKeyWithValue("answers", answers))
	})

	It("does not approve a question without an answer", func() {
		id := pending("AskUserQuestion", map[string]any{"questions": []any{map[string]any{"question": "Where?"}}})
		rec := approve(map[string]any{"approvalId": id.String(), "action": "approve"})
		Expect(rec.Code).To(Equal(http.StatusConflict), rec.Body.String())
		Expect(stateOf(ctx, db, id)).To(Equal(captaindb.TurnRequestStatePending))
	})

	// The guard follows the request's kind, not the tool name a provider chose.
	It("does not approve a question request without an answer, whatever its tool is called", func() {
		id := pendingTyped(api.ApprovalRequest{
			Tool: "request_user_input", Input: map[string]any{}, Kind: api.ApprovalKindQuestion,
			Questions: []api.TerminalQuestion{{ID: "db", Text: "Which database?"}},
		})

		rec := approve(map[string]any{"approvalId": id.String(), "action": "approve"})

		Expect(rec.Code).To(Equal(http.StatusConflict), rec.Body.String())
		Expect(stateOf(ctx, db, id)).To(Equal(captaindb.TurnRequestStatePending))
	})

	It("cancels a request, so the waiting provider is denied and told to interrupt the turn", func() {
		type outcome struct {
			decision api.ApprovalDecision
			err      error
		}
		raised := make(chan string, 1)
		broker := &approval.Broker{
			DB: db, SessionID: session, PromptRunID: promptR, RequestedBy: "provider",
			Timeout: time.Hour, Poll: 10 * time.Millisecond,
			Notify: func(_ context.Context, event api.Event) error {
				if event.Reason == "" {
					raised <- event.ApprovalID
				}
				return nil
			},
		}
		outcomes := make(chan outcome, 1)
		go func() {
			defer GinkgoRecover()
			request := commandRequest()
			request.ToolUseID = uuid.NewString()
			decision, err := broker.OnApproval(ctx, request)
			outcomes <- outcome{decision: decision, err: err}
		}()
		var approvalID string
		Eventually(raised, 5*time.Second).Should(Receive(&approvalID))
		const reason = "stop and rethink the build"

		rec := approve(map[string]any{"approvalId": approvalID, "action": "cancel", "message": reason})

		Expect(rec.Code).To(Equal(http.StatusOK), rec.Body.String())
		var got outcome
		Eventually(outcomes, 5*time.Second).Should(Receive(&got))
		Expect(got.err).NotTo(HaveOccurred())
		Expect(got.decision).To(Equal(api.ApprovalDecision{Allow: false, Message: reason, Interrupt: true}))
	})

	It("approves a request for the rest of the session when the request offers that scope", func() {
		id := pendingTyped(commandRequest())

		rec := approve(map[string]any{"approvalId": id.String(), "action": "approve", "scope": "session"})

		Expect(rec.Code).To(Equal(http.StatusOK), rec.Body.String())
		request, err := db.GetTurnRequest(ctx, id)
		Expect(err).NotTo(HaveOccurred())
		Expect(request.State).To(Equal(captaindb.TurnRequestStateApproved))
		Expect(request.Response).To(Equal(map[string]any{"scope": "session"}))
	})

	It("responds to a permissions request with the granted subset", func() {
		id := pendingTyped(api.ApprovalRequest{
			Tool: "request_permissions", Input: map[string]any{}, Kind: api.ApprovalKindPermissions,
			SupportedScopes: []api.ApprovalScope{api.ApprovalScopeTurn, api.ApprovalScopeSession},
			Permissions: &api.NativeSandboxPolicy{
				Filesystem: &api.SandboxFilesystemPolicy{WritableRoots: []string{"/work/acme", "/work/shared"}},
			},
		})
		grants := map[string]any{"filesystem": map[string]any{"writableRoots": []any{"/work/acme"}}}

		rec := approve(map[string]any{"approvalId": id.String(), "action": "respond", "scope": "turn", "grants": grants})

		Expect(rec.Code).To(Equal(http.StatusOK), rec.Body.String())
		request, err := db.GetTurnRequest(ctx, id)
		Expect(err).NotTo(HaveOccurred())
		Expect(request.State).To(Equal(captaindb.TurnRequestStateApproved))
		Expect(request.Response).To(Equal(map[string]any{"scope": "turn", "grants": grants}))
	})

	// Captain validates the decision against the request before the row goes
	// terminal; a decision the request cannot take is the client's mistake.
	DescribeTable("refuses a decision the request cannot take as a bad request, leaving it pending",
		func(body map[string]any) {
			id := pending("Bash", map[string]any{"command": "ls"})
			body["approvalId"] = id.String()

			rec := approve(body)

			Expect(rec.Code).To(Equal(http.StatusBadRequest), rec.Body.String())
			Expect(stateOf(ctx, db, id)).To(Equal(captaindb.TurnRequestStatePending))
		},
		Entry("a scope the request does not offer", map[string]any{"action": "approve", "scope": "session"}),
		Entry("a cancel on a request that cannot interrupt", map[string]any{"action": "cancel"}),
		Entry("grants on a request that is not for permissions", map[string]any{
			"action": "respond", "grants": map[string]any{"network": map[string]any{"allowedDomains": []any{"example.com"}}},
		}),
	)

	It("rejects an unknown action rather than guessing", func() {
		id := pending("Bash", nil)

		rec := approve(map[string]any{"approvalId": id.String(), "action": "maybe"})

		Expect(rec.Code).To(Equal(http.StatusBadRequest), rec.Body.String())
		Expect(rec.Body.String()).To(ContainSubstring("approve, deny, respond, cancel"))
	})

	It("rejects a decision that names no approval", func() {
		rec := approve(map[string]any{"action": "approve"})

		Expect(rec.Code).To(Equal(http.StatusBadRequest), rec.Body.String())
		Expect(rec.Body.String()).To(ContainSubstring("approvalId"))
	})

	It("reports a conflict for an approval that is already answered", func() {
		id := pending("Bash", nil)
		Expect(approve(map[string]any{"approvalId": id.String(), "action": "approve"}).Code).To(Equal(http.StatusOK))

		rec := approve(map[string]any{"approvalId": id.String(), "action": "deny"})

		Expect(rec.Code).To(Equal(http.StatusConflict), rec.Body.String())
	})

	// A stop ends the run through captain's promptrun.Cancel, which withdraws the
	// run's approvals with it. The dashboard must then neither offer them nor
	// accept an answer that would unblock a broker no longer there to hear it.
	It("neither lists nor answers the approvals of a cancelled run", func() {
		id := pending("Bash", nil)

		_, err := promptrun.Cancel(ctx, db, promptR, "run stopped")
		Expect(err).NotTo(HaveOccurred())

		Expect(listPending()).To(BeEmpty())
		rec := approve(map[string]any{"approvalId": id.String(), "action": "approve"})
		Expect(rec.Code).To(Equal(http.StatusConflict), rec.Body.String())
		Expect(stateOf(ctx, db, id)).To(Equal(captaindb.TurnRequestStateCancelled))
	})
})

func stateOf(ctx context.Context, db *captaindb.DB, id uuid.UUID) captaindb.TurnRequestState {
	GinkgoHelper()
	request, err := db.GetTurnRequest(ctx, id)
	Expect(err).NotTo(HaveOccurred())
	return request.State
}
