package ui

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/types"
)

// Plan-review actions: approving a reviewed plan (optionally chaining straight
// into the implementing run) and answering an agent's blocking questions by
// resuming its session. Both are domain transitions, not bare status writes —
// they validate the todo is actually in the expected state so a stale dashboard
// gets a 409 instead of silently clobbering a concurrent change.

// todoApprovePayload approves the plan of a todo in review. Run chains the
// implementing run immediately; Options (optional) carries the run knobs the
// dialog would send to /api/todos/run.
type todoApprovePayload struct {
	Provider string          `json:"provider,omitempty"`
	Dir      string          `json:"dir,omitempty"`
	Ref      string          `json:"ref"`
	Run      bool            `json:"run,omitempty"`
	Options  *todoRunPayload `json:"options,omitempty"`
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
	provider, source, todo, status, err := s.loadTodoForWrite(r, payload.Provider, payload.Dir, payload.Ref)
	if err != nil {
		writeTodoError(w, status, err)
		return
	}
	if todo.Status != types.StatusReview {
		writeTodoError(w, http.StatusConflict, fmt.Errorf("todo is not awaiting plan review (status: %s)", todo.Status))
		return
	}

	pending := types.StatusPending
	todo.Status = pending
	if err := provider.UpdateState(r.Context(), todo, todos.StateUpdate{Status: &pending}); err != nil {
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
		req := todoRunRequest{Provider: provider, Todos: []*types.TODO{todo}, Source: source, Backend: source.Provider, Options: opts}
		if err := startTodoRun(req); err != nil {
			writeTodoError(w, http.StatusBadRequest, err)
			return
		}
		resp.Run = &todoRunResponse{
			Status:    "started",
			Ref:       todos.TODOReference(todo),
			Count:     1,
			Dir:       source.Dir,
			Provider:  source.Provider,
			Agent:     opts.Agent,
			Mode:      opts.Mode,
			Driver:    opts.Driver,
			Backend:   opts.Backend,
			Model:     opts.Model,
			Effort:    opts.Effort,
			RunMode:   string(opts.RunMode),
			SessionID: opts.SessionID,
			Timeout:   opts.Timeout.String(),
			Commit:    opts.Commit,
			Message:   "Approved plan — implementing",
		}
	}
	json.NewEncoder(w).Encode(resp) //nolint:errcheck
}

// todoAnswerPayload answers the questions blocking an ask todo; the agent's
// prior session is resumed with the answer as the next user turn.
type todoAnswerPayload struct {
	Provider string `json:"provider,omitempty"`
	Dir      string `json:"dir,omitempty"`
	Ref      string `json:"ref"`
	Answer   string `json:"answer"`
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
	if answer == "" {
		writeTodoError(w, http.StatusBadRequest, fmt.Errorf("answer is required"))
		return
	}
	provider, source, todo, status, err := s.loadTodoForWrite(r, payload.Provider, payload.Dir, payload.Ref)
	if err != nil {
		writeTodoError(w, status, err)
		return
	}
	if todo.Status != types.StatusAsk {
		writeTodoError(w, http.StatusConflict, fmt.Errorf("todo is not awaiting an answer (status: %s)", todo.Status))
		return
	}
	if todo.LLM == nil || todo.LLM.SessionId == "" {
		writeTodoError(w, http.StatusConflict, fmt.Errorf("todo has no recorded agent session to resume"))
		return
	}

	// Record the answer on the todo so the resumed prompt and the timeline see it.
	if err := provider.Comment(r.Context(), todo, "**Answer:** "+answer); err != nil {
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
	req := todoRunRequest{Provider: provider, Todos: []*types.TODO{todo}, Source: source, Backend: source.Provider, Options: opts}
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

// defaultStartTodoAnswer resumes the todo's session with the answer in a
// background goroutine, mirroring defaultStartTodoRun's lifecycle (timeout,
// executor construction, commit tail).
func defaultStartTodoAnswer(req todoRunRequest, answer string) error {
	executor, sessionID, err := newTodoRunExecutor(req)
	if err != nil {
		return err
	}
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), req.Options.Timeout)
		defer cancel()

		execCtx := todos.NewExecutorContext(ctx, logger.StandardLogger(), nil)
		runner := todos.NewTODOExecutor(req.Source.Dir, executor, sessionID, req.Provider)
		runner.SetMode(req.Options.RunMode)
		results, runErr := runner.Resume(execCtx, req.Todos, answer)
		if runErr != nil {
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
