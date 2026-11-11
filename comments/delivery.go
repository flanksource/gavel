package comments

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

var (
	// ErrDelivery is an agent delivery that ran and failed; the failure is also
	// recorded on the comment (content.agentRun) by the host's Delivery.
	ErrDelivery = errors.New("agent delivery failed")
	// ErrUnsupported is a request the host did not opt into, e.g. agent delivery.
	ErrUnsupported = errors.New("comment host does not support this")
)

// DeliverAgent is the operation kind that hands a thread to an agent.
const DeliverAgent = "deliver-agent"

// deliveryLease bounds one inline delivery: a live-session send or a session launch.
const deliveryLease = 5 * time.Minute

// DeliverTarget is where a thread goes: a live session, or a new managed run.
type DeliverTarget struct {
	Kind      string `json:"kind"`
	SessionID string `json:"sessionId,omitempty"`
	Prompt    string `json:"prompt,omitempty"`
	// RunOptions are the host Delivery's own run choices, decoded strictly by it.
	RunOptions json.RawMessage `json:"runOptions,omitempty"`
	// GroupKey reuses the live session recorded for it in the target's scope.
	GroupKey string `json:"groupKey,omitempty"`
	// Fresh always launches a new session, never reusing GroupKey's; the launched
	// session becomes the group's, so the group's later deliveries reuse it.
	Fresh bool `json:"fresh,omitempty"`
}

func (t DeliverTarget) Validate() error {
	switch {
	case t.Kind == "session" && strings.TrimSpace(t.SessionID) != "" && t.Prompt == "" && len(t.RunOptions) == 0 && t.GroupKey == "" && !t.Fresh:
		return nil
	case t.Kind == "run" && strings.TrimSpace(t.Prompt) != "" && t.SessionID == "":
		if len(t.RunOptions) > 0 && !json.Valid(t.RunOptions) {
			return fmt.Errorf("%w: runOptions must be JSON", ErrInvalid)
		}
		return nil
	}
	return fmt.Errorf("%w: target must be {kind:\"session\", sessionId} or {kind:\"run\", prompt, runOptions?, groupKey?, fresh?}", ErrInvalid)
}

// AgentSession is a live Claude or Codex session a thread can be delivered to.
type AgentSession struct {
	ID     string `json:"id"`
	Source string `json:"source"`
	Title  string `json:"title"`
	Model  string `json:"model,omitempty"`
	CWD    string `json:"cwd,omitempty"`
	Status string `json:"status,omitempty"`
	InCmux bool   `json:"inCmux"`
}

// SessionQuery narrows live sessions to those working in ScopeURI's directory, unless All.
type SessionQuery struct {
	ScopeURI string
	All      bool
}

// Delivery executes deliver-agent operations for a host that opted into agents.
type Delivery interface {
	Executor
	// Check rejects a target the host cannot run (e.g. malformed runOptions) before it is queued.
	Check(DeliverTarget) error
	Sessions(context.Context, SessionQuery) ([]AgentSession, error)
	// Preview renders what a run target would start for c without claiming or launching.
	Preview(context.Context, Comment, DeliverTarget) (json.RawMessage, error)
	// RecordFailure records on the thread an automatic delivery to target that
	// failed before its executor ran (a route lookup, or a queue that kept conflicting).
	RecordFailure(ctx context.Context, commentID string, target DeliverTarget, cause error) error
}

type DeliverRequest struct {
	CommentID       string        `json:"-"`
	ExpectedVersion int64         `json:"expectedVersion"`
	RequestID       string        `json:"requestId"`
	Target          DeliverTarget `json:"target"`
}

func (s *Service) requireDelivery() error {
	if s.delivery == nil {
		return fmt.Errorf("%w: agent delivery is not enabled on this comments host", ErrUnsupported)
	}
	return nil
}

func (s *Service) checkTarget(t DeliverTarget) error {
	if err := s.requireDelivery(); err != nil {
		return err
	}
	if err := t.Validate(); err != nil {
		return err
	}
	return s.delivery.Check(t)
}

// Deliver queues deliver-agent for one thread and processes exactly that
// operation inline. A failed delivery returns the updated comment with ErrDelivery.
func (s *Service) Deliver(ctx context.Context, req DeliverRequest) (Comment, error) {
	if err := s.requireDelivery(); err != nil {
		return Comment{}, err
	}
	if err := s.checkTarget(req.Target); err != nil {
		return Comment{}, err
	}
	payload, err := json.Marshal(req.Target)
	if err != nil {
		return Comment{}, err
	}
	op, err := s.Queue(ctx, QueueRequest{CommentID: req.CommentID, ExpectedVersion: req.ExpectedVersion, Kind: DeliverAgent, RequestID: req.RequestID, Payload: payload})
	if err != nil {
		return Comment{}, err
	}
	if op.State == "pending" {
		if op, err = s.Process(ctx, ProcessRequest{Executor: s.delivery, Lease: deliveryLease, OperationID: op.ID}); err != nil {
			return Comment{}, fmt.Errorf("process delivery %s: %w", op.ID, err)
		}
	}
	comment, err := s.Get(ctx, req.CommentID)
	if err != nil {
		return Comment{}, err
	}
	if op.State != "done" {
		return comment, fmt.Errorf("%w: %s", ErrDelivery, op.Error)
	}
	return comment, nil
}

// AgentSessions lists the live agent sessions a principal can deliver threads to.
func (s *Service) AgentSessions(ctx context.Context, q SessionQuery) ([]AgentSession, error) {
	if err := s.requireDelivery(); err != nil {
		return nil, err
	}
	p, err := s.principal(ctx)
	if err != nil {
		return nil, err
	}
	if q.ScopeURI != "" {
		if err = s.access.Authorize(ctx, p, "agent-sessions", Target{ScopeURI: q.ScopeURI}); err != nil {
			return nil, err
		}
	}
	return s.delivery.Sessions(ctx, q)
}

// PreviewDelivery renders a run target for a thread without claiming or launching it.
func (s *Service) PreviewDelivery(ctx context.Context, id string, target DeliverTarget) (json.RawMessage, error) {
	if err := s.requireDelivery(); err != nil {
		return nil, err
	}
	if err := s.checkTarget(target); err != nil {
		return nil, err
	}
	if target.Kind != "run" {
		return nil, fmt.Errorf("%w: only run targets have a preview", ErrInvalid)
	}
	c, err := s.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	return s.delivery.Preview(ctx, c, target)
}
