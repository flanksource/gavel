package ui

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	cmuxprov "github.com/flanksource/captain/pkg/ai/provider/cmux"
	captaindb "github.com/flanksource/captain/pkg/database"
	"github.com/flanksource/gavel/todos/run"
	"github.com/flanksource/gavel/todos/types"
)

// todoAnswerPayload answers the questions blocking an ask todo; the agent's
// prior session is resumed with the answer as the next user turn.
type todoAnswerPayload struct {
	Dir      string         `json:"dir,omitempty"`
	Ref      string         `json:"ref"`
	Answer   string         `json:"answer"`
	Answers  map[string]any `json:"answers,omitempty"`
	Rejected bool           `json:"rejected,omitempty"`
	// Optional run knobs for the resumed turn (model/mode/effort/timeout);
	// omitted fields keep the defaults derived from the todo.
	Options *todoRunPayload `json:"options,omitempty"`
}

// todoAnswerResponse is the todo after an answer or a revision. Status is
// "resumed"/"revising" when the continuation started and "failed" when the
// transition committed but the run did not, with Error saying why.
type todoAnswerResponse struct {
	Todo      todoSummary `json:"todo"`
	SessionID string      `json:"sessionId,omitempty"`
	Status    string      `json:"status"`
	Error     string      `json:"error,omitempty"`
}

func (s *Server) handleTodoAnswer(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var payload todoAnswerPayload
	if err := decodeTodoRequest(r, &payload); err != nil {
		writeTodoError(w, http.StatusBadRequest, err)
		return
	}
	answer, err := answerText(payload)
	if err != nil {
		writeTodoError(w, http.StatusBadRequest, err)
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
	if run.PriorSessionID(todo) == "" {
		writeTodoError(w, http.StatusConflict, fmt.Errorf("todo has no recorded agent session to resume"))
		return
	}

	activeRun, err := run.PriorRun(r.Context(), provider, todo)
	if err != nil {
		writeTodoError(w, http.StatusInternalServerError, err)
		return
	}
	// A workspace with no Captain handle has no approval table, so no approval
	// was ever brokered for this run and there is nothing to be blocked on. Any
	// other failure is a store that exists and could not be read — and a
	// liveness check that cannot read it cannot tell a parked agent from a
	// finished one, so it must not let a second agent through. The interface
	// stays nil rather than wrapping a nil pointer, which would read as present
	// and then panic.
	var approvals approvalStore
	store, storeErr := todoApprovalStore(r.Context(), source.Dir)
	switch {
	case storeErr == nil:
		approvals = store
	case !errors.Is(storeErr, errNoCaptainDatabase):
		writeTodoError(w, http.StatusInternalServerError, fmt.Errorf("check the run's pending approvals: %w", storeErr))
		return
	}
	if status, err := answerLivenessError(r.Context(), approvals, activeRun); err != nil {
		writeTodoError(w, status, err)
		return
	}
	step, err := activeStepFor(todo)
	if err != nil {
		writeTodoError(w, http.StatusConflict, err)
		return
	}
	override, err := continuationOverride(payload.Options)
	if err != nil {
		writeTodoError(w, http.StatusBadRequest, err)
		return
	}

	// A resumed turn continues the turn that asked: the same step, on the spec
	// that turn was dispatched with and the runtime it actually resolved —
	// without which the payload defaults would hand a codex session to
	// claude --resume.
	req, err := s.continuationRequest(run.Continuation{
		Dir: source.Dir, Provider: provider, Todo: todo, Prior: activeRun,
		Override: override, Step: step, Resume: true, Message: answer,
	})
	if err != nil {
		writeTodoError(w, http.StatusBadRequest, err)
		return
	}
	// Pre-flight the resolution before recording the answer, so a resume that
	// cannot start fails as a 4xx the answer box renders instead of leaving the
	// user with a recorded answer and a todo that never moves. The fold is
	// handed to the dispatch so the run is exactly the one that was checked.
	prepared, err := run.Resolve(r.Context(), req)
	if err != nil {
		writeTodoError(w, http.StatusBadRequest, err)
		return
	}
	req.Prepared = prepared

	// Record the answer on the todo so the resumed prompt and the timeline see it.
	commentLabel := "**Answer:** "
	if payload.Rejected {
		commentLabel = "**Rejected question:** "
	}
	if err := provider.Comment(r.Context(), todo, commentLabel+answer); err != nil {
		writeTodoError(w, http.StatusInternalServerError, err)
		return
	}

	if _, err := run.Start(req); err != nil {
		// The answer is recorded and the todo still asks: report both, so the
		// client can show the failure without re-recording the answer.
		sum, derr := todoDetail(r.Context(), provider, source.Dir, todo)
		if derr != nil {
			writeTodoError(w, http.StatusInternalServerError, derr)
			return
		}
		writeTodoJSON(w, continuationFailureStatus(err), todoAnswerResponse{
			Todo: sum, SessionID: run.PriorSessionID(todo),
			Status: "failed", Error: err.Error(),
		})
		return
	}
	// Snapshot after dispatch: the run goroutine owns the todo from here, and
	// the answered todo is leaving ask whatever the agent reports next.
	answered := *todo
	answered.Status = types.StatusInProgress
	sum, err := todoDetail(r.Context(), provider, source.Dir, &answered)
	if err != nil {
		writeTodoError(w, http.StatusInternalServerError, err)
		return
	}
	writeTodoJSON(w, http.StatusOK, todoAnswerResponse{
		Todo: sum, SessionID: run.PriorSessionID(todo), Status: "resumed",
	})
}

// answerText is the answer as the next user turn: the free-text answer, else
// the structured answers encoded, else the rejection notice.
func answerText(payload todoAnswerPayload) (string, error) {
	answer := strings.TrimSpace(payload.Answer)
	if answer == "" && len(payload.Answers) > 0 {
		encoded, err := json.Marshal(map[string]any{"answers": payload.Answers})
		if err != nil {
			return "", fmt.Errorf("invalid answers: %w", err)
		}
		answer = string(encoded)
	}
	if answer == "" && payload.Rejected {
		answer = "The pending question was rejected. Continue without that answer or explain what is required."
	}
	if answer == "" {
		return "", fmt.Errorf("answer is required")
	}
	return answer, nil
}

// answerLivenessError rejects an answer that would spawn a second agent against
// a session a live process still owns. An ask outcome is a finished turn — the
// agent exited and only the prompt run parks at waiting — so resuming it means
// dispatching a fresh turn.
//
// Two other states are live. A running prompt run means the agent is still
// executing. A waiting one with an unanswered tool request means the agent is
// alive and parked on the approval broker, which is not the same waiting as an
// ask: the answer belongs to the approval endpoint, not to a new run. The two
// are told apart by the durable request itself rather than by state alone,
// because a tool approval and an ask both park the run at waiting.
func answerLivenessError(ctx context.Context, store approvalStore, run *captaindb.PromptRun) (int, error) {
	if run == nil {
		return http.StatusOK, nil
	}
	if run.State == captaindb.PromptRunStateRunning {
		return http.StatusConflict, errors.New("todo run is still active; stop it before answering")
	}
	if run.State != captaindb.PromptRunStateWaiting || store == nil {
		return http.StatusOK, nil
	}
	pending, err := pendingApprovals(ctx, store, run.SessionID, &run.ID)
	if err != nil {
		return http.StatusInternalServerError, err
	}
	if len(pending) > 0 {
		return http.StatusConflict, fmt.Errorf(
			"agent is live and waiting on approval for %s; answer it via /api/todos/session/approve", pending[0].Tool)
	}
	return http.StatusOK, nil
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
