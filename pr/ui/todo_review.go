package ui

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	cmuxprov "github.com/flanksource/captain/pkg/ai/provider/cmux"
	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/todos"
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
	todo, err = reviewer.ApprovePlan(r.Context(), todo, todoReviewActor, "")
	if err != nil {
		writeTodoError(w, http.StatusInternalServerError, err)
		return
	}

	resp := todoApproveResponse{Todo: summarizeTodo(todo, true)}
	if payload.Run {
		runPayload := todoRunPayload{}
		if payload.Options != nil {
			runPayload = *payload.Options
		}
		// The approved plan executes as an implement run, whatever the dialog sent.
		runPayload.RunMode = string(types.ModeRun)
		runPayload.Plan = false
		opts, err := normalizeTodoRunOptions(runPayload)
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
			Agent:     opts.Agent,
			Mode:      opts.Mode,
			Driver:    opts.Driver,
			Backend:   string(opts.Spec.Backend),
			Model:     opts.Spec.Name,
			Effort:    string(opts.Spec.Effort),
			RunMode:   string(opts.RunMode),
			SessionID: opts.Spec.SessionID,
			Timeout:   opts.timeout().String(),
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
	// The native review operation records the feedback and revision-requested
	// state together; a separate provider comment would duplicate the timeline.
	todo, err = reviewer.RequestPlanRevision(r.Context(), todo, todoReviewActor, feedback)
	if err != nil {
		writeTodoError(w, http.StatusInternalServerError, err)
		return
	}

	runPayload := todoRunPayload{}
	if payload.Options != nil {
		runPayload = *payload.Options
	}
	sessionID := ""
	if todo.LLM != nil {
		sessionID = todo.LLM.SessionId
	}
	runPayload.Resume = sessionID != ""
	runPayload.RunMode = string(types.ModePlan) // revise re-enters plan mode on the same session
	runPayload.Plan = true
	opts, err := normalizeTodoRunOptions(runPayload)
	if err != nil {
		writeTodoError(w, http.StatusBadRequest, err)
		return
	}
	req := todoRunRequest{Provider: provider, Registry: &s.todoRuns, Todos: []*types.TODO{todo}, Source: source, Backend: todos.ProviderDB, Options: opts}
	if runPayload.Resume {
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

	// Record the answer on the todo so the resumed prompt and the timeline see it.
	commentLabel := "**Answer:** "
	if payload.Rejected {
		commentLabel = "**Rejected question:** "
	}
	if err := provider.Comment(r.Context(), todo, commentLabel+answer); err != nil {
		writeTodoError(w, http.StatusInternalServerError, err)
		return
	}

	runPayload := todoRunPayload{}
	if payload.Options != nil {
		runPayload = *payload.Options
	}
	runPayload.Resume = true
	runPayload.RunMode = string(todo.RunMode) // continue in the mode that asked
	runPayload.Plan = false
	opts, err := normalizeTodoRunOptions(runPayload)
	if err != nil {
		writeTodoError(w, http.StatusBadRequest, err)
		return
	}
	req := todoRunRequest{Provider: provider, Registry: &s.todoRuns, Todos: []*types.TODO{todo}, Source: source, Backend: todos.ProviderDB, Options: opts}
	if err := startTodoAnswer(req, answer); err != nil {
		writeTodoError(w, http.StatusBadRequest, err)
		return
	}
	json.NewEncoder(w).Encode(todoAnswerResponse{ //nolint:errcheck
		Todo:      summarizeTodo(todo, true),
		SessionID: todo.LLM.SessionId,
		Status:    "resumed",
	})
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
	ctx, timeoutCancel := context.WithTimeout(context.Background(), req.Options.timeout())
	runCtx, stop := context.WithCancelCause(ctx)
	cleanup, err := req.Registry.register(todoRunIssueIDs(req.Todos), strings.HasPrefix(executor.Name(), "headless-"), stop)
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
		if runErr != nil && (len(results) == 0 || results[0] == nil || !results[0].Cancelled) {
			logger.Warnf("todo answer %s failed: %v", todoRunLabel(req.Todos), runErr)
		}
		var result *todos.ExecutionResult
		if len(results) > 0 {
			result = results[0]
		}
		maybeCommitAfterRun(req, result)
	}()
	return nil
}
