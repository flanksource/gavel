package ui

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/run"
	"github.com/flanksource/gavel/todos/types"
)

const todoReviewActor = "gavel-ui"

// Plan-review actions: approving a reviewed plan (optionally chaining straight
// into the implementing run), revising it, and answering an agent's blocking
// questions by resuming its session. All are domain transitions, not bare
// status writes — they validate the todo is actually in the expected state so
// a stale dashboard gets a 409 instead of silently clobbering a concurrent
// change — and every one of them dispatches through the same run seam as
// /api/todos/run: a continuation is a lifecycle step run like any other.
//
// Each builds and validates the run it will chain BEFORE it commits the
// transition, so a request the run seam refuses is a 400 that changed nothing.
// A run that fails to start after the transition committed is reported with the
// todo as it now is plus the error: a bare error would tell the client nothing
// happened, and its retry would then meet a todo no longer in review.

// todoApprovePayload approves the plan of a todo in review. Run chains the
// implementing run immediately; Options (optional) carries the run knobs the
// dialog would send to /api/todos/run.
type todoApprovePayload struct {
	Dir     string          `json:"dir,omitempty"`
	Ref     string          `json:"ref"`
	Run     bool            `json:"run,omitempty"`
	Options *todoRunPayload `json:"options,omitempty"`
}

// todoApproveResponse is the todo after the transition. Error is set only when
// the transition committed but the chained run did not start.
type todoApproveResponse struct {
	Todo  todoSummary      `json:"todo"`
	Run   *todoRunResponse `json:"run,omitempty"`
	Error string           `json:"error,omitempty"`
}

func (s *Server) handleTodoPlanApprove(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var payload todoApprovePayload
	if err := decodeTodoRequest(r, &payload); err != nil {
		writeTodoError(w, http.StatusBadRequest, err)
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
	planRun, err := run.PriorRun(r.Context(), provider, todo)
	if err != nil {
		writeTodoError(w, http.StatusInternalServerError, err)
		return
	}
	var chained *run.Request
	if payload.Run {
		override, err := continuationOverride(payload.Options)
		if err != nil {
			writeTodoError(w, http.StatusBadRequest, err)
			return
		}
		// The approved plan executes as an implement run, whatever the dialog sent,
		// and as a fresh turn: the plan reaches it through the provider, not
		// through the planning conversation.
		req, err := s.continuationRequest(run.Continuation{
			Dir: source.Dir, Provider: provider, Todo: todo, Prior: planRun,
			Override: override, Step: string(types.RunPhase),
		})
		if err != nil {
			writeTodoError(w, http.StatusBadRequest, err)
			return
		}
		chained = &req
	}

	todo, err = reviewer.ApprovePlan(r.Context(), todo, todoReviewActor, "")
	if err != nil {
		writeTodoError(w, http.StatusInternalServerError, err)
		return
	}
	sum, err := todoDetail(r.Context(), provider, source.Dir, todo)
	if err != nil {
		writeTodoError(w, http.StatusInternalServerError, err)
		return
	}
	resp := todoApproveResponse{Todo: sum}
	if chained != nil {
		chained.Todo = todo
		started, err := s.startContinuation(r.Context(), *chained)
		if err != nil {
			resp.Error = err.Error()
			writeTodoJSON(w, continuationFailureStatus(err), resp)
			return
		}
		started.Message = "Approved plan — implementing"
		resp.Run = &started
	}
	writeTodoJSON(w, http.StatusOK, resp)
}

// todoRejectPayload rejects a reviewed plan: the todo returns to pending and the
// plan stays selected as rejected, so no run can follow it until a new revision
// or an approval supersedes that decision.
type todoRejectPayload struct {
	Dir string `json:"dir,omitempty"`
	Ref string `json:"ref"`
}

func (s *Server) handleTodoPlanReject(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var payload todoRejectPayload
	if err := decodeTodoRequest(r, &payload); err != nil {
		writeTodoError(w, http.StatusBadRequest, err)
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
	todo, err = reviewer.RejectPlan(r.Context(), todo, todoReviewActor, "")
	if err != nil {
		writeTodoError(w, http.StatusInternalServerError, err)
		return
	}
	sum, err := todoDetail(r.Context(), provider, source.Dir, todo)
	if err != nil {
		writeTodoError(w, http.StatusInternalServerError, err)
		return
	}
	writeTodoJSON(w, http.StatusOK, todoApproveResponse{Todo: sum})
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
	if err := decodeTodoRequest(r, &payload); err != nil {
		writeTodoError(w, http.StatusBadRequest, err)
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
	planRun, err := run.PriorRun(r.Context(), provider, todo)
	if err != nil {
		writeTodoError(w, http.StatusInternalServerError, err)
		return
	}
	override, err := continuationOverride(payload.Options)
	if err != nil {
		writeTodoError(w, http.StatusBadRequest, err)
		return
	}
	// Revise re-enters the plan step: it continues the plan run's own
	// configuration, resuming its session with the feedback as the next turn
	// when there is one, and otherwise planning afresh with the feedback in the
	// todo's own prompt.
	c := run.Continuation{
		Dir: source.Dir, Provider: provider, Todo: todo, Prior: planRun,
		Override: override, Step: string(types.PlanPhase),
	}
	resume := run.PriorSessionID(todo) != ""
	if resume {
		c.Resume, c.Message = true, feedback
	}
	req, err := s.continuationRequest(c)
	if err != nil {
		writeTodoError(w, http.StatusBadRequest, err)
		return
	}

	// The native review operation records the feedback and revision-requested
	// state together; a separate provider comment would duplicate the timeline.
	todo, err = reviewer.RequestPlanRevision(r.Context(), todo, todoReviewActor, feedback)
	if err != nil {
		writeTodoError(w, http.StatusInternalServerError, err)
		return
	}
	if !resume {
		todo.Prompt = "Revise the existing plan using this reviewer feedback:\n\n" + feedback
	}
	req.Todo = todo
	sum, err := todoDetail(r.Context(), provider, source.Dir, todo)
	if err != nil {
		writeTodoError(w, http.StatusInternalServerError, err)
		return
	}
	resp := todoAnswerResponse{Todo: sum, Status: "revising"}
	started, err := s.startContinuation(r.Context(), req)
	if err != nil {
		resp.Status, resp.Error = "failed", err.Error()
		writeTodoJSON(w, continuationFailureStatus(err), resp)
		return
	}
	resp.SessionID = started.SessionID
	writeTodoJSON(w, http.StatusOK, resp)
}

func requirePlanReviewProvider(provider todos.Provider) (todos.PlanReviewProvider, error) {
	reviewer, ok := provider.(todos.PlanReviewProvider)
	if !ok {
		return nil, fmt.Errorf("PostgreSQL TODO runtime does not support durable plan review")
	}
	return reviewer, nil
}

// continuationOverride validates the run knobs a review action's dialog sent —
// the request layer the continuation's inheritance sits below. A missing
// options object is an empty layer.
func continuationOverride(payload *todoRunPayload) (run.Options, error) {
	if payload == nil {
		return buildTodoRunOptions(todoRunPayload{}, nil)
	}
	return buildTodoRunOptions(*payload, nil)
}

// continuationRequest is the run request a continuation dispatches: resolved
// through run.Continue, brokered by the dashboard like every run it starts.
func (s *Server) continuationRequest(c run.Continuation) (run.Request, error) {
	opts, err := run.Continue(c)
	if err != nil {
		return run.Request{}, err
	}
	return run.Request{
		Provider: c.Provider, Registry: todoRuns(), Todo: c.Todo, Dir: c.Dir,
		Options: opts, Broker: todoApprovalBroker(c.Dir),
	}, nil
}

// startContinuation resolves and starts a continuation, answering with the same
// description of the run /api/todos/run gives. The fold it reports is the fold
// it dispatches.
func (s *Server) startContinuation(ctx context.Context, req run.Request) (todoRunResponse, error) {
	prepared, err := run.Resolve(ctx, req)
	if err != nil {
		return todoRunResponse{}, err
	}
	req.Prepared = prepared
	started, err := run.Start(req)
	if err != nil {
		return todoRunResponse{}, err
	}
	resp := todoRunResponseFor(todoSource{Dir: req.Dir}, []*types.TODO{req.Todo}, req.Options, prepared)
	resp.Status, resp.SessionID, resp.Message = started.Status, started.SessionID, started.Message
	return resp, nil
}
