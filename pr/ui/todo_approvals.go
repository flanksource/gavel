package ui

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/flanksource/captain/pkg/ai/approval"
	"github.com/flanksource/captain/pkg/api"
	captaindb "github.com/flanksource/captain/pkg/database"
	"github.com/flanksource/gavel/todos"
	"github.com/google/uuid"
)

// approvalStore is the durable approval seam: captain's `captain_turn_requests`
// table, addressed by the session and prompt run Captain admitted for a run.
//
// It replaced a process-wide in-memory registry. That registry could only ever
// answer a run started by the same process, held one pending request per
// session, and lost every outstanding approval on restart — so a dashboard that
// reconnected to a live run had nothing to show and no way to unblock it.
type approvalStore interface {
	ListTurnRequests(context.Context, captaindb.TurnRequestFilter) ([]captaindb.TurnRequest, error)
	GetTurnRequest(context.Context, uuid.UUID) (*captaindb.TurnRequest, error)
	ResolveToolApprovalRequest(context.Context, captaindb.ResolveToolApprovalRequestInput) (*captaindb.TurnRequest, error)
	CancelPendingTurnRequests(ctx context.Context, sessionID, promptRunID uuid.UUID, reason string) error
	GetPromptRun(context.Context, uuid.UUID) (*captaindb.PromptRun, error)
	UpdatePromptRun(context.Context, captaindb.UpdatePromptRunInput) (*captaindb.PromptRun, error)
}

// todoApprovalStore resolves the workspace's captain handle. A provider with no
// captain handle cannot broker approvals at all, which is reported rather than
// degraded: a run configured to ask, dispatched with nothing able to answer,
// blocks until its timeout.
//
// It is a var so a test can hand the handlers a Captain database of its own
// rather than standing up a whole workspace to reach the same one.
var todoApprovalStore = openTodoApprovalStore

// errNoCaptainDatabase marks a workspace whose TODO runtime keeps no Captain
// handle. It is the one failure a reader may take as "no approval was ever
// brokered here"; every other error from todoApprovalStore is a store that
// exists and could not be reached, which says nothing about what it holds.
var errNoCaptainDatabase = errors.New("no Captain database")

func openTodoApprovalStore(ctx context.Context, dir string) (*captaindb.DB, error) {
	provider, err := openTodoProvider(ctx, dir)
	if err != nil {
		return nil, err
	}
	native, ok := provider.(captainSessionProvider)
	if !ok || native.Captain() == nil {
		return nil, fmt.Errorf("%w: the TODO runtime for %s cannot record tool approvals", errNoCaptainDatabase, dir)
	}
	return native.Captain(), nil
}

// todoApprovalBroker is the dashboard's ApprovalBroker: every tool call the
// run's permission mode does not pre-approve becomes a durable row a person can
// answer from the TODO's session view.
//
// It is built per execution because the rows are keyed on the session and
// prompt run Captain admits, neither of which exists when the executor is
// constructed.
func todoApprovalBroker(dir string) todos.ApprovalBroker {
	return func(ctx *todos.ExecutorContext) (api.PermissionFunc, error) {
		store, err := todoApprovalStore(ctx, dir)
		if err != nil {
			return nil, err
		}
		sessionID := todos.CaptainSessionFromContext(ctx)
		promptRunID := todos.PromptRunFromContext(ctx)
		if sessionID == uuid.Nil || promptRunID == uuid.Nil {
			return nil, fmt.Errorf("tool approvals need Captain's admitted session and prompt run; got session %s run %s", sessionID, promptRunID)
		}
		broker := &approval.Broker{
			DB:          store,
			SessionID:   sessionID,
			PromptRunID: promptRunID,
			RequestedBy: "gavel-dashboard",
			// A provider approval suspends the run and may be answered long after
			// the turn that raised it; captain's own provider timeout is the one
			// the dashboard inherits rather than a second number to keep in step.
			Timeout:   approval.ProviderTimeout,
			Notify:    notifyApproval(ctx),
			OnWaiting: setPromptRunState(store, promptRunID, captaindb.PromptRunStateWaiting),
			OnRunning: setPromptRunState(store, promptRunID, captaindb.PromptRunStateRunning),
		}
		if err := broker.Validate(); err != nil {
			return nil, err
		}
		return broker.CanUseTool, nil
	}
}

// notifyApproval surfaces a pending request on the run's own narration, which is
// what the TODO session view already renders — the same frame the provider's
// EventPermission produces, so an approval reads identically wherever it came
// from. The durable row is what the dashboard acts on; this is how it learns
// there is one without waiting for its next poll.
func notifyApproval(ctx *todos.ExecutorContext) func(context.Context, api.Event) error {
	return func(_ context.Context, event api.Event) error {
		detail := map[string]any{"tool": event.Tool, "approvalId": event.ApprovalID, "input": event.Input}
		ctx.GetTranscript().AddExecutorMessage("awaiting approval: "+event.Tool, todos.EntryAction, detail)
		ctx.Notify(todos.Notification{
			Type:    todos.NotifyApproval,
			Message: event.Tool,
			Data:    detail,
		})
		return nil
	}
}

// setPromptRunState brackets the wait with the run's own state. It is not
// bookkeeping: captain only resolves a credential-less approval while its prompt
// run is `waiting`, so a broker whose OnWaiting did nothing would record a
// request nobody could ever answer.
//
// It never resurrects a finished run. The broker's OnRunning fires on every
// exit from the wait — including the one where the run was stopped underneath
// it: the stop cancelled the request, ReclaimRun wrote the terminal state, and
// the broker woke to find its request gone. Writing `running` then would leave a
// dead run the dashboard offers to stop forever. A concurrent write to the run
// row between the read and the update is retried once on the fresh version; a
// second conflict is reported rather than fought over.
func setPromptRunState(store approvalStore, promptRunID uuid.UUID, state captaindb.PromptRunState) func(context.Context) error {
	return func(ctx context.Context) error {
		for attempt := 0; ; attempt++ {
			run, err := store.GetPromptRun(ctx, promptRunID)
			if err != nil {
				return err
			}
			if run.State == state {
				return nil
			}
			if terminalPromptRunState(run.State) {
				return fmt.Errorf("prompt run %s is %s and cannot move to %s", promptRunID, run.State, state)
			}
			_, err = store.UpdatePromptRun(ctx, captaindb.UpdatePromptRunInput{
				ID:              promptRunID,
				ExpectedVersion: run.Version,
				State:           &state,
			})
			if err == nil || attempt > 0 || !errors.Is(err, captaindb.ErrPromptRunConflict) {
				return err
			}
		}
	}
}

// terminalPromptRunState reports whether a prompt run has finished for good.
func terminalPromptRunState(state captaindb.PromptRunState) bool {
	switch state {
	case captaindb.PromptRunStateSucceeded, captaindb.PromptRunStateFailed, captaindb.PromptRunStateCancelled:
		return true
	}
	return false
}

// pendingApprovals are the run's unanswered tool requests, oldest first. The
// filter is the durable identity, so a dashboard that reconnected — or one
// running in a different process from the run — sees exactly what is
// outstanding.
func pendingApprovals(ctx context.Context, store approvalStore, sessionID uuid.UUID, promptRunID *uuid.UUID) ([]todoApproval, error) {
	requests, err := store.ListTurnRequests(ctx, captaindb.TurnRequestFilter{SessionID: sessionID, PromptRunID: promptRunID})
	if err != nil {
		return nil, err
	}
	var pending []todoApproval
	for _, request := range requests {
		if request.State != captaindb.TurnRequestStatePending {
			continue
		}
		pending = append(pending, todoApprovalOf(request))
	}
	return pending, nil
}

// todoApproval is one pending tool request as the dashboard reads it. ID is the
// durable approval id the client must send back — a session id is no longer
// enough to name a request, because a run can have more than one outstanding.
type todoApproval struct {
	ID        string         `json:"approvalId"`
	SessionID string         `json:"sessionId"`
	Tool      string         `json:"tool"`
	Input     map[string]any `json:"input,omitempty"`
	CreatedAt string         `json:"createdAt,omitempty"`
}

func todoApprovalOf(request captaindb.TurnRequest) todoApproval {
	approvalRow := todoApproval{
		ID:        request.ID.String(),
		SessionID: request.SessionID.String(),
		CreatedAt: request.CreatedAt.UTC().Format("2006-01-02T15:04:05Z07:00"),
	}
	if tool, ok := request.Request["tool"].(string); ok {
		approvalRow.Tool = tool
	}
	if input, ok := request.Request["input"].(map[string]any); ok {
		approvalRow.Input = input
	}
	return approvalRow
}

// todoApprovalAction is what the dashboard's buttons ask for.
type todoApprovalAction string

const (
	// approvalApprove runs the tool as requested.
	approvalApprove todoApprovalAction = "approve"
	// approvalDeny refuses it; Message is fed back to the agent as the reason.
	approvalDeny todoApprovalAction = "deny"
	// approvalRespond runs the tool with the operator's edited input — an answer,
	// not a veto, for the case where the call is right but its arguments are not.
	approvalRespond todoApprovalAction = "respond"
)

func parseApprovalAction(value string) (todoApprovalAction, error) {
	switch todoApprovalAction(strings.TrimSpace(value)) {
	case approvalApprove:
		return approvalApprove, nil
	case approvalDeny:
		return approvalDeny, nil
	case approvalRespond:
		return approvalRespond, nil
	}
	return "", fmt.Errorf("invalid approval action %q (valid: approve, deny, respond)", value)
}

// resolveApproval answers one durable request. `respond` implies approval — it
// is "run it, with this input instead" — so only `deny` refuses.
func resolveApproval(
	ctx context.Context,
	store approvalStore,
	sessionID, requestID uuid.UUID,
	action todoApprovalAction,
	message string,
	input map[string]any,
) (*captaindb.TurnRequest, error) {
	if action == approvalRespond && len(input) == 0 {
		return nil, fmt.Errorf("respond needs the replacement tool input; send approve to run the call unchanged")
	}
	return store.ResolveToolApprovalRequest(ctx, captaindb.ResolveToolApprovalRequestInput{
		SessionID:  sessionID,
		RequestID:  requestID,
		Approved:   action != approvalDeny,
		ResolvedBy: "gavel-dashboard",
		Reason:     strings.TrimSpace(message),
		// ExpectedTurnID stays nil: a provider approval belongs to a prompt run,
		// never to an aichat turn, so there is no turn to match it against.
		UpdatedInput: input,
	})
}
