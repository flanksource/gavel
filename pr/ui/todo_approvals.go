package ui

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/flanksource/captain/pkg/ai/approval"
	"github.com/flanksource/captain/pkg/api"
	captaindb "github.com/flanksource/captain/pkg/database"
	"github.com/google/uuid"
)

// approvalStore is what the dashboard reads of captain's durable approvals:
// `captain_turn_requests` rows, addressed by the session and prompt run Captain
// admitted for a run. It is read-only by construction — every write (raising,
// answering, withdrawing an approval, and the run's waiting posture) goes
// through captain's approval package, never through this seam.
type approvalStore interface {
	ListTurnRequests(context.Context, captaindb.TurnRequestFilter) ([]captaindb.TurnRequest, error)
	GetTurnRequest(context.Context, uuid.UUID) (*captaindb.TurnRequest, error)
}

// todoApprovalStore resolves the workspace's Captain handle for listing and
// answering durable approvals. Captain binds the run's broker itself.
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
		approvalRow, err := todoApprovalOf(request)
		if err != nil {
			return nil, err
		}
		pending = append(pending, approvalRow)
	}
	return pending, nil
}

// todoApproval is one pending request as the dashboard reads it. ID is the
// durable approval id the client must send back — a session id is no longer
// enough to name a request, because a run can have more than one outstanding.
//
// Tool and Input are what every reader has always seen. Kind and Request add
// the typed request: Request is the stored document as captain wrote it, so the
// dashboard renders each kind's payload (command, filesystem, network,
// permissions, questions, elicitation) and its offered scopes without Gavel
// re-modelling them.
type todoApproval struct {
	ID        string           `json:"approvalId"`
	SessionID string           `json:"sessionId"`
	Tool      string           `json:"tool"`
	Input     map[string]any   `json:"input,omitempty"`
	Kind      api.ApprovalKind `json:"kind"`
	Request   map[string]any   `json:"request"`
	CreatedAt string           `json:"createdAt,omitempty"`
}

func todoApprovalOf(request captaindb.TurnRequest) (todoApproval, error) {
	kind, err := storedApprovalKind(request)
	if err != nil {
		return todoApproval{}, err
	}
	if kind == "" {
		kind = api.ApprovalKindTool
	}
	approvalRow := todoApproval{
		ID:        request.ID.String(),
		SessionID: request.SessionID.String(),
		Kind:      kind,
		Request:   request.Request,
		CreatedAt: request.CreatedAt.UTC().Format("2006-01-02T15:04:05Z07:00"),
	}
	if tool, ok := request.Request["tool"].(string); ok {
		approvalRow.Tool = tool
	}
	if input, ok := request.Request["input"].(map[string]any); ok {
		approvalRow.Input = input
	}
	return approvalRow, nil
}

// storedApprovalKind is the kind a row was raised with, or empty for a row
// written before kinds existed — which only ever held a tool approval.
func storedApprovalKind(request captaindb.TurnRequest) (api.ApprovalKind, error) {
	raw, present := request.Request["kind"]
	if !present {
		return "", nil
	}
	kind, ok := raw.(string)
	if !ok || !api.ApprovalKind(kind).Valid() {
		return "", fmt.Errorf("approval %s has invalid kind %v", request.ID, raw)
	}
	return api.ApprovalKind(kind), nil
}

// todoApprovalAction is what the dashboard's buttons ask for.
type todoApprovalAction string

const (
	// approvalApprove runs the tool as requested.
	approvalApprove todoApprovalAction = "approve"
	// approvalDeny refuses it; Message is fed back to the agent as the reason.
	approvalDeny todoApprovalAction = "deny"
	// approvalRespond answers the request with content: the operator's edited
	// tool input, a question's answers or a form's content (all as Input), or a
	// permissions request's granted subset (as Grants).
	approvalRespond todoApprovalAction = "respond"
	// approvalCancel refuses the request and interrupts the agent's turn;
	// Message is fed back as the reason.
	approvalCancel todoApprovalAction = "cancel"
)

func parseApprovalAction(value string) (todoApprovalAction, error) {
	switch action := todoApprovalAction(strings.TrimSpace(value)); action {
	case approvalApprove, approvalDeny, approvalRespond, approvalCancel:
		return action, nil
	}
	return "", fmt.Errorf("invalid approval action %q (valid: approve, deny, respond, cancel)", value)
}

// approvalAnswer is one dashboard decision on one durable request. A nil
// SessionID lets captain take the session off the row.
type approvalAnswer struct {
	SessionID, RequestID uuid.UUID
	Action               todoApprovalAction
	Message              string
	Input                map[string]any
	Scope                api.ApprovalScope
	Grants               *api.NativeSandboxPolicy
}

// resolveApproval answers one durable request through captain's approval.Resolve,
// which validates the decision against the stored request and refuses one it
// cannot take with approval.ErrInvalidResolution. `respond` implies approval —
// it is "run it, with this content" — so only `deny` and `cancel` refuse.
//
// The run's posture is not touched here: the broker waiting on the request reads
// the answer and releases the run itself.
func resolveApproval(ctx context.Context, db *captaindb.DB, answer approvalAnswer) (*captaindb.TurnRequest, error) {
	if answer.Action == approvalRespond && len(answer.Input) == 0 && answer.Grants == nil {
		return nil, fmt.Errorf("respond needs the replacement tool input, answers, form content or grants; send approve to run the call unchanged")
	}
	request, err := db.GetTurnRequest(ctx, answer.RequestID)
	if err != nil {
		return nil, err
	}
	// A question with nothing substituted leaves the agent waiting on a prompt no
	// one can answer. The answers themselves are shaped and validated by the
	// provider that asked (captain's api.AnswersForQuestions), which is the only
	// side that knows how its host keys them.
	isQuestion, err := isQuestionRequest(*request)
	if err != nil {
		return nil, err
	}
	if isQuestion && answer.Action == approvalApprove {
		return nil, fmt.Errorf("a question needs answers; send respond or deny")
	}
	return approval.Resolve(ctx, db, approval.ResolveInput{
		RequestID:    answer.RequestID,
		SessionID:    answer.SessionID,
		Approved:     answer.Action == approvalApprove || answer.Action == approvalRespond,
		ResolvedBy:   "gavel-dashboard",
		Reason:       strings.TrimSpace(answer.Message),
		UpdatedInput: answer.Input,
		Interrupt:    answer.Action == approvalCancel,
		Scope:        answer.Scope,
		Grants:       answer.Grants,
	})
}

// isQuestionRequest reports a request that asks the operator a question. A row
// written before kinds existed names its question by Claude's tool alone.
func isQuestionRequest(request captaindb.TurnRequest) (bool, error) {
	kind, err := storedApprovalKind(request)
	if err != nil {
		return false, err
	}
	if kind == "" {
		return request.Request["tool"] == "AskUserQuestion", nil
	}
	return kind == api.ApprovalKindQuestion, nil
}
