package ui

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"time"

	captaindb "github.com/flanksource/captain/pkg/database"
	commonsdb "github.com/flanksource/commons-db/db"
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
		if os.Getenv("GAVEL_DB_EMBEDDED_TEST") == "" {
			Skip("set GAVEL_DB_EMBEDDED_TEST=1 to run embedded-postgres approval tests")
		}
		ctx = context.Background()
		dsn, stop, err := commonsdb.StartEmbedded(commonsdb.EmbeddedConfig{
			DataDir:  filepath.Join(GinkgoT().TempDir(), "postgres"),
			Database: "gavel_todo_approvals",
		})
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(func() { Expect(stop()).To(Succeed()) })

		db, err = captaindb.Open(ctx, captaindb.WithDSN(dsn), captaindb.WithMigrations())
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
		// A credential-less approval only resolves while its prompt run is waiting,
		// which is exactly what the broker's OnWaiting sets before it blocks.
		markWaiting(ctx, db, promptR)
	})

	pending := func(tool string, input map[string]any) uuid.UUID {
		GinkgoHelper()
		request, err := db.CreateToolApprovalRequest(ctx, captaindb.CreateToolApprovalRequestInput{
			SessionID: session, PromptRunID: promptR,
			ToolCallID: uuid.NewString(), Tool: tool, Input: input,
			RequestedBy: "provider", ExpiresAt: time.Now().Add(time.Hour),
		})
		Expect(err).NotTo(HaveOccurred())
		return request.ID
	}

	approve := func(body map[string]any) *httptest.ResponseRecorder {
		GinkgoHelper()
		raw, err := json.Marshal(body)
		Expect(err).NotTo(HaveOccurred())
		rec := httptest.NewRecorder()
		server.handleTodoSessionApprove(rec, httptest.NewRequest(
			http.MethodPost, "/api/todos/session/approve?sessionId="+session.String(), bytes.NewReader(raw)))
		return rec
	}

	It("lists the run's unanswered requests with the tool and its input", func() {
		id := pending("Bash", map[string]any{"command": "ls"})

		rec := httptest.NewRecorder()
		server.handleTodoSessionApprovals(rec, httptest.NewRequest(http.MethodGet,
			"/api/todos/session/approvals?sessionId="+session.String()+"&promptRunId="+promptR.String(), nil))

		Expect(rec.Code).To(Equal(http.StatusOK), rec.Body.String())
		var body struct {
			Approvals []todoApproval `json:"approvals"`
		}
		Expect(json.Unmarshal(rec.Body.Bytes(), &body)).To(Succeed())
		Expect(body.Approvals).To(HaveLen(1))
		Expect(body.Approvals[0].ID).To(Equal(id.String()))
		Expect(body.Approvals[0].Tool).To(Equal("Bash"))
		Expect(body.Approvals[0].Input).To(HaveKeyWithValue("command", "ls"))
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

	It("rejects an unknown action rather than guessing", func() {
		id := pending("Bash", nil)

		rec := approve(map[string]any{"approvalId": id.String(), "action": "maybe"})

		Expect(rec.Code).To(Equal(http.StatusBadRequest), rec.Body.String())
		Expect(rec.Body.String()).To(ContainSubstring("approve, deny, respond"))
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

	It("drops a stopped run's requests from the pending queue", func() {
		id := pending("Bash", nil)

		Expect(db.CancelPendingTurnRequests(ctx, session, promptR, "run stopped")).To(Succeed())

		queue, err := pendingApprovals(ctx, db, session, &promptR)
		Expect(err).NotTo(HaveOccurred())
		Expect(queue).To(BeEmpty())
		Expect(stateOf(ctx, db, id)).To(Equal(captaindb.TurnRequestStateCancelled))
	})

	// A stop reclaims the run while its broker is still parked on the wait. The
	// broker then wakes and leaves the wait through OnRunning, which must find
	// the run finished and leave it so — not write `running` over the stop.
	It("leaves a stopped run stopped when the broker leaves the wait", func() {
		pending("Bash", nil)
		Expect(db.CancelPendingTurnRequests(ctx, session, promptR, "run stopped")).To(Succeed())
		current, err := db.GetPromptRun(ctx, promptR)
		Expect(err).NotTo(HaveOccurred())
		cancelled := captaindb.PromptRunStateCancelled
		_, err = db.UpdatePromptRun(ctx, captaindb.UpdatePromptRunInput{
			ID: promptR, ExpectedVersion: current.Version, State: &cancelled,
		})
		Expect(err).NotTo(HaveOccurred())

		err = setPromptRunState(db, promptR, captaindb.PromptRunStateRunning)(ctx)

		Expect(err).To(MatchError(ContainSubstring("cancelled")))
		after, err := db.GetPromptRun(ctx, promptR)
		Expect(err).NotTo(HaveOccurred())
		Expect(after.State).To(Equal(captaindb.PromptRunStateCancelled))
	})
})

func markWaiting(ctx context.Context, db *captaindb.DB, promptRunID uuid.UUID) {
	GinkgoHelper()
	run, err := db.GetPromptRun(ctx, promptRunID)
	Expect(err).NotTo(HaveOccurred())
	state := captaindb.PromptRunStateWaiting
	_, err = db.UpdatePromptRun(ctx, captaindb.UpdatePromptRunInput{
		ID: run.ID, ExpectedVersion: run.Version, State: &state,
	})
	Expect(err).NotTo(HaveOccurred())
}

func stateOf(ctx context.Context, db *captaindb.DB, id uuid.UUID) captaindb.TurnRequestState {
	GinkgoHelper()
	request, err := db.GetTurnRequest(ctx, id)
	Expect(err).NotTo(HaveOccurred())
	return request.State
}
