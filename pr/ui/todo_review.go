package ui

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	cmuxprov "github.com/flanksource/captain/pkg/ai/provider/cmux"
	captaindb "github.com/flanksource/captain/pkg/database"
	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/drivers"
	"github.com/flanksource/gavel/todos/types"
)

const todoReviewActor = "gavel-ui"

// Plan-review actions: approving a reviewed plan (optionally chaining straight
// into the implementing run) and answering an agent's blocking questions by
// resuming its session. Both are domain transitions, not bare status writes —
// they validate the todo is actually in the expected state so a stale dashboard
// gets a 409 instead of silently clobbering a concurrent change.

// todoApprovePayload approves the plan of a todo in review. Run chains the
// implementing run immediately; Options (optional) carries the run knobs the
// dialog would send to /api/todos/run.
type todoApprovePayload struct {
	Dir     string          `json:"dir,omitempty"`
	Ref     string          `json:"ref"`
	Run     bool            `json:"run,omitempty"`
	Options *todoRunPayload `json:"options,omitempty"`
}

type todoApproveResponse struct {
	Todo todoSummary      `json:"todo"`
	Run  *todoRunResponse `json:"run,omitempty"`
}

func (s *Server) handleTodoPlanApprove(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var payload todoApprovePayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeTodoError(w, http.StatusBadRequest, fmt.Errorf("invalid json"))
		return
	}
	provider, source, todo, status, err := s.loadTodoForWrite(r, payload.Dir, payload.Ref)
	if err != nil {
		writeTodoError(w, status, err)
		return
	}
	if todo.Status != types.StatusReview {
		writeTodoError(w, http.StatusConflict, fmt.Errorf("todo is not awaiting plan review (status: %s)", todo.Status))
		return
	}

	reviewer, err := requirePlanReviewProvider(provider)
	if err != nil {
		writeTodoError(w, http.StatusInternalServerError, err)
		return
	}
	// Read the plan's prompt run before approving: the approval retires it, and it
	// is the record of the runtime the implement run continues.
	planRun, err := activeTodoPromptRun(r.Context(), provider, todo)
	if err != nil {
		writeTodoError(w, http.StatusInternalServerError, err)
		return
	}
	todo, err = reviewer.ApprovePlan(r.Context(), todo, todoReviewActor, "")
	if err != nil {
		writeTodoError(w, http.StatusInternalServerError, err)
		return
	}

	resp := todoApproveResponse{Todo: summarizeTodo(todo, true)}
	if payload.Run {
		override := todoRunPayload{}
		if payload.Options != nil {
			override = *payload.Options
		}
		// The approved plan executes as an implement run, whatever the dialog sent.
		opts, err := continueRun(continuation{
			Dir: source.Dir, Todos: []*types.TODO{todo}, Prior: planRun,
			Override: override, Mode: types.ModeRun,
		})
		if err != nil {
			writeTodoError(w, http.StatusBadRequest, err)
			return
		}
		req := todoRunRequest{Provider: provider, Registry: &s.todoRuns, Todos: []*types.TODO{todo}, Source: source, Backend: todos.ProviderDB, Options: opts}
		if err := startTodoRun(req); err != nil {
			writeTodoError(w, http.StatusBadRequest, err)
			return
		}
		resp.Run = &todoRunResponse{
			Status:    "started",
			Ref:       todos.TODOReference(todo),
			Count:     1,
			Dir:       source.Dir,
			Agent:     opts.agent(),
			Mode:      opts.legacyMode(),
			Driver:    opts.Driver,
			Backend:   string(opts.Spec.Backend),
			Model:     opts.Spec.Name,
			Effort:    string(opts.Spec.Effort),
			RunMode:   string(opts.RunMode),
			SessionID: opts.Spec.SessionID,
			Timeout:   opts.Timeout.String(),
			Commit:    specCommit(opts.Spec) && !specDryRun(opts.Spec),
			Message:   "Approved plan — implementing",
		}
	}
	json.NewEncoder(w).Encode(resp) //nolint:errcheck
}

// todoRejectPayload rejects a reviewed plan: the todo returns to pending and its
// recorded plan pointer is cleared, so a later run re-plans from scratch rather
// than following the discarded plan.
type todoRejectPayload struct {
	Dir string `json:"dir,omitempty"`
	Ref string `json:"ref"`
}

func (s *Server) handleTodoPlanReject(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var payload todoRejectPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeTodoError(w, http.StatusBadRequest, fmt.Errorf("invalid json"))
		return
	}
	provider, _, todo, status, err := s.loadTodoForWrite(r, payload.Dir, payload.Ref)
	if err != nil {
		writeTodoError(w, status, err)
		return
	}
	if todo.Status != types.StatusReview {
		writeTodoError(w, http.StatusConflict, fmt.Errorf("todo is not awaiting plan review (status: %s)", todo.Status))
		return
	}
	reviewer, err := requirePlanReviewProvider(provider)
	if err != nil {
		writeTodoError(w, http.StatusInternalServerError, err)
		return
	}
	todo, err = reviewer.RejectPlan(r.Context(), todo, todoReviewActor, "")
	if err != nil {
		writeTodoError(w, http.StatusInternalServerError, err)
		return
	}
	json.NewEncoder(w).Encode(todoApproveResponse{Todo: summarizeTodo(todo, true)}) //nolint:errcheck
}

// todoRevisePayload asks the agent to revise a reviewed plan with the reviewer's
// feedback: the plan session resumes, the agent updates its native plan file,
// and the todo returns to review with the revised plan.
type todoRevisePayload struct {
	Dir      string          `json:"dir,omitempty"`
	Ref      string          `json:"ref"`
	Feedback string          `json:"feedback"`
	Options  *todoRunPayload `json:"options,omitempty"`
}

func (s *Server) handleTodoPlanRevise(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var payload todoRevisePayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeTodoError(w, http.StatusBadRequest, fmt.Errorf("invalid json"))
		return
	}
	feedback := strings.TrimSpace(payload.Feedback)
	if feedback == "" {
		writeTodoError(w, http.StatusBadRequest, fmt.Errorf("feedback is required"))
		return
	}
	provider, source, todo, status, err := s.loadTodoForWrite(r, payload.Dir, payload.Ref)
	if err != nil {
		writeTodoError(w, status, err)
		return
	}
	if todo.Status != types.StatusReview {
		writeTodoError(w, http.StatusConflict, fmt.Errorf("todo is not awaiting plan review (status: %s)", todo.Status))
		return
	}
	reviewer, err := requirePlanReviewProvider(provider)
	if err != nil {
		writeTodoError(w, http.StatusInternalServerError, err)
		return
	}
	// Read the plan's prompt run before the revision transition retires it.
	planRun, err := activeTodoPromptRun(r.Context(), provider, todo)
	if err != nil {
		writeTodoError(w, http.StatusInternalServerError, err)
		return
	}
	// The native review operation records the feedback and revision-requested
	// state together; a separate provider comment would duplicate the timeline.
	todo, err = reviewer.RequestPlanRevision(r.Context(), todo, todoReviewActor, feedback)
	if err != nil {
		writeTodoError(w, http.StatusInternalServerError, err)
		return
	}

	override := todoRunPayload{}
	if payload.Options != nil {
		override = *payload.Options
	}
	sessionID := ""
	if todo.LLM != nil {
		sessionID = todo.LLM.SessionId
	}
	// Revise re-enters plan mode: it continues the plan run's own configuration,
	// resuming its session when there is one to resume.
	opts, err := continueRun(continuation{
		Dir: source.Dir, Todos: []*types.TODO{todo}, Prior: planRun,
		Override: override, Mode: types.ModePlan, Resume: sessionID != "",
	})
	if err != nil {
		writeTodoError(w, http.StatusBadRequest, err)
		return
	}
	req := todoRunRequest{Provider: provider, Registry: &s.todoRuns, Todos: []*types.TODO{todo}, Source: source, Backend: todos.ProviderDB, Options: opts}
	if opts.Resume {
		err = startTodoAnswer(req, feedback)
	} else {
		todo.Prompt = "Revise the existing plan using this reviewer feedback:\n\n" + feedback
		err = startTodoRun(req)
	}
	if err != nil {
		writeTodoError(w, http.StatusBadRequest, err)
		return
	}
	json.NewEncoder(w).Encode(todoAnswerResponse{ //nolint:errcheck
		Todo:      summarizeTodo(todo, true),
		SessionID: sessionID,
		Status:    "revising",
	})
}

func requirePlanReviewProvider(provider todos.Provider) (todos.PlanReviewProvider, error) {
	reviewer, ok := provider.(todos.PlanReviewProvider)
	if !ok {
		return nil, fmt.Errorf("PostgreSQL TODO runtime does not support durable plan review")
	}
	return reviewer, nil
}

// todoAnswerPayload answers the questions blocking an ask todo; the agent's
// prior session is resumed with the answer as the next user turn.
type todoAnswerPayload struct {
	Dir      string         `json:"dir,omitempty"`
	Ref      string         `json:"ref"`
	Answer   string         `json:"answer"`
	Answers  map[string]any `json:"answers,omitempty"`
	Rejected bool           `json:"rejected,omitempty"`
	// Optional run knobs for the resumed turn (model/backend/effort/timeout);
	// omitted fields keep the defaults derived from the todo.
	Options *todoRunPayload `json:"options,omitempty"`
}

type todoAnswerResponse struct {
	Todo      todoSummary `json:"todo"`
	SessionID string      `json:"sessionId,omitempty"`
	Status    string      `json:"status"`
}

// startTodoAnswer dispatches the async resume; a var so handler tests can stub it.
var startTodoAnswer = defaultStartTodoAnswer

func (s *Server) handleTodoAnswer(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var payload todoAnswerPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeTodoError(w, http.StatusBadRequest, fmt.Errorf("invalid json"))
		return
	}
	answer := strings.TrimSpace(payload.Answer)
	if answer == "" && len(payload.Answers) > 0 {
		encoded, err := json.Marshal(map[string]any{"answers": payload.Answers})
		if err != nil {
			writeTodoError(w, http.StatusBadRequest, fmt.Errorf("invalid answers: %w", err))
			return
		}
		answer = string(encoded)
	}
	if answer == "" && payload.Rejected {
		answer = "The pending question was rejected. Continue without that answer or explain what is required."
	}
	if answer == "" {
		writeTodoError(w, http.StatusBadRequest, fmt.Errorf("answer is required"))
		return
	}
	provider, source, todo, status, err := s.loadTodoForWrite(r, payload.Dir, payload.Ref)
	if err != nil {
		writeTodoError(w, status, err)
		return
	}
	if todo.Status != types.StatusAsk && !zombieAskSession(source.Dir, todo) {
		writeTodoError(w, http.StatusConflict, fmt.Errorf("todo is not awaiting an answer (status: %s)", todo.Status))
		return
	}
	if todo.LLM == nil || todo.LLM.SessionId == "" {
		writeTodoError(w, http.StatusConflict, fmt.Errorf("todo has no recorded agent session to resume"))
		return
	}

	activeRun, err := activeTodoPromptRun(r.Context(), provider, todo)
	if err != nil {
		writeTodoError(w, http.StatusInternalServerError, err)
		return
	}
	if status, err := answerLivenessError(activeRun, todo.LLM.SessionId); err != nil {
		writeTodoError(w, status, err)
		return
	}

	// A resumed turn continues the turn that asked: it stays in that mode, on the
	// spec that turn was dispatched with and the runtime it actually resolved —
	// without which the payload defaults would hand a codex session to
	// claude --resume.
	override := todoRunPayload{}
	if payload.Options != nil {
		override = *payload.Options
	}
	opts, err := continueRun(continuation{
		Dir: source.Dir, Todos: []*types.TODO{todo}, Prior: activeRun,
		Override: override, Mode: todo.RunMode, Resume: true,
	})
	if err != nil {
		writeTodoError(w, http.StatusBadRequest, err)
		return
	}
	req := todoRunRequest{Provider: provider, Registry: &s.todoRuns, Todos: []*types.TODO{todo}, Source: source, Backend: todos.ProviderDB, Options: opts}
	// Pre-flight the executor before recording the answer, so a resume that
	// cannot start fails as a 4xx the answer box renders instead of leaving the
	// user with a recorded answer and a todo that never moves.
	if _, _, err := newTodoRunExecutorContext(r.Context(), req); err != nil {
		writeTodoError(w, http.StatusBadRequest, err)
		return
	}

	// Record the answer on the todo so the resumed prompt and the timeline see it.
	commentLabel := "**Answer:** "
	if payload.Rejected {
		commentLabel = "**Rejected question:** "
	}
	if err := provider.Comment(r.Context(), todo, commentLabel+answer); err != nil {
		writeTodoError(w, http.StatusInternalServerError, err)
		return
	}

	// Snapshot before dispatch: the run goroutine owns the todo from here, and
	// the answered todo is leaving ask whatever the agent reports next.
	todo.Status = types.StatusInProgress
	resp := todoAnswerResponse{
		Todo:      summarizeTodo(todo, true),
		SessionID: todo.LLM.SessionId,
		Status:    "resumed",
	}
	if err := startTodoAnswer(req, answer); err != nil {
		writeTodoError(w, http.StatusBadRequest, err)
		return
	}
	json.NewEncoder(w).Encode(resp) //nolint:errcheck
}

// activePromptRunProvider exposes the prompt run backing a todo's current
// attempt. The native PostgreSQL runtime implements it; a provider that keeps no
// run history simply has none to report.
type activePromptRunProvider interface {
	ActivePromptRun(context.Context, *types.TODO) (*captaindb.PromptRun, error)
}

// activeTodoPromptRun returns the prompt run currently attached to the todo, or
// nil when there is none; every other failure is real.
func activeTodoPromptRun(ctx context.Context, provider todos.Provider, todo *types.TODO) (*captaindb.PromptRun, error) {
	runs, ok := provider.(activePromptRunProvider)
	if !ok {
		return nil, nil
	}
	return runs.ActivePromptRun(ctx, todo)
}

// answerLivenessError rejects an answer that would spawn a second agent against
// a session a live process still owns. An ask outcome is a finished turn — the
// agent exited and only the prompt run parks at waiting — so resuming it means
// dispatching a fresh turn. A running prompt run is the opposite: the agent is
// alive, and if it is blocked on a tool permission the answer belongs to the
// approval endpoint, not to a new run.
func answerLivenessError(run *captaindb.PromptRun, sessionID string) (int, error) {
	if run == nil || run.State != captaindb.PromptRunStateRunning {
		return http.StatusOK, nil
	}
	if pending, ok := todos.GlobalApprovals().Pending(sessionID); ok {
		return http.StatusConflict, fmt.Errorf(
			"agent is live and waiting on approval for %s; answer it via /api/todos/session/approve", pending.Tool)
	}
	return http.StatusConflict, errors.New("todo run is still active; stop it before answering")
}

// zombieAskSession permits resuming a todo whose worker died after Claude wrote
// a structured AskUserQuestion but before Gavel persisted StatusAsk. The check is
// deliberately server-side: the recorded session must exist, be stopped, and
// have ask as its latest meaningful state.
func zombieAskSession(dir string, todo *types.TODO) bool {
	if todo == nil || todo.Status != types.StatusInProgress || todo.LLM == nil || todo.LLM.SessionId == "" {
		return false
	}
	path, err := cmuxprov.SessionLogPath(dir, todo.LLM.SessionId)
	if err != nil {
		return false
	}
	stats, err := cmuxprov.GlobalSessionStats().Get(todo.LLM.SessionId, path)
	return err == nil && stats.Found && stats.State == "ask" && !stats.InProgress
}

// defaultStartTodoAnswer resumes the todo's session with the answer in a
// background goroutine, mirroring defaultStartTodoRun's lifecycle (timeout,
// executor construction, commit tail).
func defaultStartTodoAnswer(req todoRunRequest, answer string) error {
	if req.Registry == nil {
		return errors.New("todo run registry is required")
	}
	executor, sessionID, err := newTodoRunExecutor(req)
	if err != nil {
		return err
	}
	ctx, timeoutCancel := context.WithTimeout(context.Background(), req.Options.Timeout)
	runCtx, stop := context.WithCancelCause(ctx)
	cleanup, err := req.Registry.register(todoRunIssueIDs(req.Todos), runIsStoppable(req.Options), stop)
	if err != nil {
		timeoutCancel()
		return err
	}
	go func() {
		defer timeoutCancel()
		defer cleanup()

		execCtx := todos.NewExecutorContext(runCtx, logger.StandardLogger(), nil)
		runner := todos.NewTODOExecutor(req.Source.Dir, executor, sessionID, req.Provider)
		runner.SetMode(req.Options.RunMode)
		results, runErr := runner.Resume(execCtx, req.Todos, answer)
		var result *todos.ExecutionResult
		if len(results) > 0 {
			result = results[0]
		}
		if runErr != nil && (result == nil || !result.Cancelled) {
			logger.Warnf("todo answer %s failed: %v", todoRunLabel(req.Todos), runErr)
			// The HTTP response is long gone: persist the failure so the timeline
			// shows why the answered todo stopped moving instead of leaving it
			// parked with an answer nobody acted on.
			recordAnswerFailure(req, runErr)
		}
	}()
	return nil
}

// runIsStoppable reports whether a run's driver can be cancelled mid-flight.
// Headless drivers own the agent process and honour context cancellation; a cmux
// run is driven through a detached surface that outlives the request.
func runIsStoppable(opts todoRunOptions) bool {
	kind, err := drivers.Parse(opts.Driver)
	return err == nil && kind != drivers.Cmux
}

// recordAnswerFailure marks an asynchronously-failed resume on the todo. Resume
// already persists a failed attempt once the agent has started; this covers the
// earlier failures (prepare, dispatch) that leave no attempt behind.
func recordAnswerFailure(req todoRunRequest, runErr error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	failed := types.StatusFailed
	for _, todo := range req.Todos {
		if todo == nil || todo.Status == failed {
			continue
		}
		if err := req.Provider.UpdateState(ctx, todo, todos.StateUpdate{Status: &failed}); err != nil {
			logger.Warnf("failed to record todo answer failure for %s: %v", todos.TODOReference(todo), err)
			continue
		}
		if err := req.Provider.Comment(ctx, todo, "**Resume failed:** "+runErr.Error()); err != nil {
			logger.Warnf("failed to comment todo answer failure for %s: %v", todos.TODOReference(todo), err)
		}
	}
}
