package todos

import (
	"context"
	"fmt"
	"sync"
)

// ApprovalRequest is a tool-permission request a driver surfaces for human
// review when an agent wants to use a tool that is not pre-approved. It is
// driver-agnostic: the cmux driver detects it on the terminal surface, the
// sdk/headless drivers receive it over the stream-json control protocol.
type ApprovalRequest struct {
	SessionID string         `json:"sessionId"`
	ToolUseID string         `json:"toolUseId,omitempty"`
	Tool      string         `json:"tool"`
	Input     map[string]any `json:"input,omitempty"`
}

// ApprovalDecision is the human's answer to an ApprovalRequest.
type ApprovalDecision struct {
	// Allow runs the tool; otherwise it is denied and Message is fed back to the
	// agent as the reason.
	Allow   bool   `json:"allow"`
	Message string `json:"message,omitempty"`
	// UpdatedInput optionally replaces the tool input on allow (drivers that
	// support the control protocol forward it; others ignore it).
	UpdatedInput map[string]any `json:"updatedInput,omitempty"`
}

type pendingApproval struct {
	req     ApprovalRequest
	decided chan ApprovalDecision
}

// ApprovalRegistry brokers tool-permission requests between a running driver
// (the producer, which Awaits a decision) and the dashboard (the consumer,
// which Resolves it). One request is pending per session at a time.
type ApprovalRegistry struct {
	mu      sync.Mutex
	pending map[string]*pendingApproval
}

// Await registers req as the session's pending approval and blocks until the
// dashboard resolves it or ctx is cancelled. The pending entry is always
// removed before returning.
func (r *ApprovalRegistry) Await(ctx context.Context, req ApprovalRequest) (ApprovalDecision, error) {
	if req.SessionID == "" {
		return ApprovalDecision{}, fmt.Errorf("approval request requires a session id")
	}
	p := &pendingApproval{req: req, decided: make(chan ApprovalDecision, 1)}

	r.mu.Lock()
	if r.pending == nil {
		r.pending = make(map[string]*pendingApproval)
	}
	r.pending[req.SessionID] = p
	r.mu.Unlock()

	defer func() {
		r.mu.Lock()
		if r.pending[req.SessionID] == p {
			delete(r.pending, req.SessionID)
		}
		r.mu.Unlock()
	}()

	select {
	case <-ctx.Done():
		return ApprovalDecision{}, ctx.Err()
	case d := <-p.decided:
		return d, nil
	}
}

// Pending returns the session's outstanding approval request, if any.
func (r *ApprovalRegistry) Pending(sessionID string) (ApprovalRequest, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if p, ok := r.pending[sessionID]; ok {
		return p.req, true
	}
	return ApprovalRequest{}, false
}

// Resolve delivers a decision to the session's pending approval, unblocking the
// awaiting driver. It errors when there is nothing pending for the session.
func (r *ApprovalRegistry) Resolve(sessionID string, decision ApprovalDecision) error {
	r.mu.Lock()
	p, ok := r.pending[sessionID]
	r.mu.Unlock()
	if !ok {
		return fmt.Errorf("no pending approval for session %s", sessionID)
	}
	select {
	case p.decided <- decision:
		return nil
	default:
		return fmt.Errorf("approval for session %s was already resolved", sessionID)
	}
}

var defaultApprovals = &ApprovalRegistry{}

// GlobalApprovals is the process-wide approval registry. The dashboard and the
// running drivers share it the same way they share the session-stats cache.
func GlobalApprovals() *ApprovalRegistry { return defaultApprovals }
