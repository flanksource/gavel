package comments

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"
)

var (
	// ErrDelivery is an agent delivery that ran and failed; the failure is also
	// recorded on the comment (content.agentRun) by the host's Delivery.
	ErrDelivery = errors.New("agent delivery failed")
	// ErrUnsupported is a request the host did not opt into, e.g. agent delivery.
	ErrUnsupported = errors.New("comment host does not support this")
	// ErrNoWorkdir is a scope with no local working directory for an agent
	// session: its threads can start no managed session and mention none.
	ErrNoWorkdir = errors.New("scope has no local working directory")
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
	// Edits is EditsDirect or EditsPropose: the agent answers with a proposed
	// patch and must not edit the project's files. Empty inherits the policy of
	// the target's comment agent rule (AgentRules), EditsDirect when it sets none;
	// the service resolves it before a delivery reaches the host's Delivery.
	Edits string `json:"edits,omitempty"`
}

func (t DeliverTarget) Validate() error {
	if t.Edits != "" && t.Edits != EditsDirect && t.Edits != EditsPropose {
		return fmt.Errorf("%w: edits must be %q or %q, not %q", ErrInvalid, EditsDirect, EditsPropose, t.Edits)
	}
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

// AgentSession is a Claude or Codex session a thread can be delivered to: a
// live one, or a managed session the host started for the scope that has ended
// (Status "ended") and is resumed by delivering to it.
type AgentSession struct {
	ID     string `json:"id"`
	Source string `json:"source"`
	Title  string `json:"title"`
	Model  string `json:"model,omitempty"`
	CWD    string `json:"cwd,omitempty"`
	Status string `json:"status,omitempty"`
	InCmux bool   `json:"inCmux"`
	// Managed is a session the host launched for the scope's threads.
	Managed bool `json:"managed,omitempty"`
	// Resumable is an ended managed session a delivery resumes with its recorded prompt.
	Resumable bool `json:"resumable,omitempty"`

	Context *AgentSessionContext `json:"context,omitempty"`
}

// ManagedSession is a session a host's Delivery launched for a scope's threads,
// recorded with what it runs so it is listed after it ends and resumed with the
// same prompt and run options.
type ManagedSession struct {
	ID         string          `json:"id"`
	ScopeURI   string          `json:"scopeUri"`
	GroupKey   string          `json:"groupKey,omitempty"`
	Prompt     string          `json:"prompt"`
	RunOptions json.RawMessage `json:"runOptions,omitempty"`
	// Label titles the session once it has ended: the first line of the thread it started for.
	Label           string    `json:"label"`
	CWD             string    `json:"cwd"`
	Source          string    `json:"source,omitempty"`
	Model           string    `json:"model,omitempty"`
	StartedAt       time.Time `json:"startedAt"`
	LastDeliveredAt time.Time `json:"lastDeliveredAt"`
}

// AgentSessionContext is how much of the model's context window the session currently fills.
type AgentSessionContext struct {
	UsedTokens   int `json:"usedTokens"`
	WindowTokens int `json:"windowTokens"`
}

// SessionQuery narrows live sessions to those working in ScopeURI's directory,
// unless All; with a ScopeURI the scope's managed sessions are listed too.
// TenantID is the requesting principal's, set by Service.AgentSessions.
type SessionQuery struct {
	TenantID string
	ScopeURI string
	All      bool
}

// Delivery executes deliver-agent operations for a host that opted into agents.
type Delivery interface {
	Executor
	// Check rejects a target the host cannot run (e.g. malformed runOptions) before it is queued.
	Check(DeliverTarget) error
	Sessions(context.Context, SessionQuery) ([]AgentSession, error)
	// Mentionables are the agents, besides the rule's @agent, a comment on the
	// query's target addresses with an @name token, each with where it delivers:
	// the runtime families always, the scope's sessions only when the query asks.
	Mentionables(context.Context, MentionQuery) ([]Mentionable, error)
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

// DeliverBatchRequest hands several threads of one scope to Target in one
// delivery: one message, one session.
type DeliverBatchRequest struct {
	RequestID string          `json:"requestId"`
	Target    DeliverTarget   `json:"target"`
	Threads   []DeliverThread `json:"threads"`
}

// DeliverThread is one thread of a batch at the version the caller last read.
type DeliverThread struct {
	ID              string `json:"id"`
	ExpectedVersion int64  `json:"expectedVersion"`
}

// DeliverOperation is a deliver-agent operation's payload: its target and, for
// a batch, every thread delivered with it, the operation's own first.
type DeliverOperation struct {
	DeliverTarget
	Threads []string `json:"threads,omitempty"`
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
	if err := s.checkTarget(req.Target); err != nil {
		return Comment{}, err
	}
	c, err := s.Get(ctx, req.CommentID)
	if err != nil {
		return Comment{}, err
	}
	if req.Target, err = s.withRuleEdits(ctx, c.Target, req.Target); err != nil {
		return Comment{}, err
	}
	payload, err := json.Marshal(req.Target)
	if err != nil {
		return Comment{}, err
	}
	delivered, err := s.runDelivery(ctx, QueueRequest{CommentID: req.CommentID, ExpectedVersion: req.ExpectedVersion, Kind: DeliverAgent, RequestID: req.RequestID, Payload: payload}, []string{req.CommentID})
	if len(delivered) == 0 {
		return Comment{}, err
	}
	return delivered[0], err
}

// DeliverBatch hands req's threads, all of one scope, to req.Target in one
// delivery: one message into one session, recorded as every thread's
// agentRun. Unless the target sets an edits policy, every thread's rule must
// resolve to the same one (ErrInvalid otherwise). Every thread is checked
// before anything is queued, so one stale (ErrConflict) or closed thread fails
// the whole batch. It is one deliver-agent operation on the first thread,
// idempotent by RequestID as Deliver is.
func (s *Service) DeliverBatch(ctx context.Context, req DeliverBatchRequest) ([]Comment, error) {
	if err := s.checkTarget(req.Target); err != nil {
		return nil, err
	}
	ids, err := s.checkBatch(ctx, req)
	if err != nil {
		return nil, err
	}
	if req.Target, err = s.batchEdits(ctx, ids, req.Target); err != nil {
		return nil, err
	}
	payload, err := json.Marshal(DeliverOperation{DeliverTarget: req.Target, Threads: ids})
	if err != nil {
		return nil, err
	}
	return s.runDelivery(ctx, QueueRequest{CommentID: ids[0], ExpectedVersion: req.Threads[0].ExpectedVersion, Kind: DeliverAgent, RequestID: req.RequestID, Payload: payload}, ids)
}

// batchEdits is target with an empty edits policy resolved from each thread's
// rule; threads whose rules disagree need the target to choose one.
func (s *Service) batchEdits(ctx context.Context, ids []string, target DeliverTarget) (DeliverTarget, error) {
	if target.Edits != "" {
		return target, nil
	}
	policies := map[string][]string{}
	for _, id := range ids {
		c, err := s.Get(ctx, id)
		if err != nil {
			return target, err
		}
		resolved, err := s.withRuleEdits(ctx, c.Target, target)
		if err != nil {
			return target, err
		}
		policies[resolved.Edits] = append(policies[resolved.Edits], id)
	}
	if len(policies) > 1 {
		return target, fmt.Errorf("%w: the batch's threads inherit different edits policies %v; set the target's edits", ErrInvalid, policies)
	}
	for edits := range policies {
		target.Edits = edits
	}
	return target, nil
}

// runDelivery queues a deliver-agent operation, processes exactly it inline
// unless a replay finds it already run, and answers the delivered threads as
// they now are; a failed delivery answers them with ErrDelivery.
func (s *Service) runDelivery(ctx context.Context, queue QueueRequest, threads []string) ([]Comment, error) {
	op, err := s.Queue(ctx, queue)
	if err != nil {
		return nil, err
	}
	if op.State == "pending" {
		if op, err = s.Process(ctx, ProcessRequest{Executor: s.delivery, Lease: deliveryLease, OperationID: op.ID}); err != nil {
			return nil, fmt.Errorf("process delivery %s: %w", op.ID, err)
		}
	}
	delivered := make([]Comment, 0, len(threads))
	for _, id := range threads {
		c, err := s.Get(ctx, id)
		if err != nil {
			return nil, err
		}
		delivered = append(delivered, c)
	}
	if op.State != "done" {
		return delivered, fmt.Errorf("%w: %s", ErrDelivery, op.Error)
	}
	return delivered, nil
}

// checkBatch answers the batch's thread ids once each is an active thread of
// one scope the principal may deliver, at its expected version. A replayed
// request, whose operation is already queued, is left to Queue to match.
func (s *Service) checkBatch(ctx context.Context, req DeliverBatchRequest) ([]string, error) {
	if len(req.Threads) == 0 || req.RequestID == "" {
		return nil, fmt.Errorf("%w: a batch delivery needs a requestId and at least one thread", ErrInvalid)
	}
	ids := make([]string, 0, len(req.Threads))
	for _, t := range req.Threads {
		if slices.Contains(ids, t.ID) {
			return nil, fmt.Errorf("%w: thread %s is in the batch twice", ErrInvalid, t.ID)
		}
		ids = append(ids, t.ID)
	}
	store, ok := s.store.(OperationStore)
	if !ok {
		return nil, fmt.Errorf("comment store does not support operations")
	}
	p, err := s.principal(ctx)
	if err != nil {
		return nil, err
	}
	if _, err = store.OperationByRequest(ctx, p.TenantID, req.RequestID); !errors.Is(err, ErrNotFound) {
		return ids, err
	}
	scope := ""
	for _, t := range req.Threads {
		if scope, err = s.checkBatchThread(ctx, p, t, scope); err != nil {
			return nil, err
		}
	}
	return ids, nil
}

// checkBatchThread checks one thread of a batch, answering its scope, which
// must be scope unless scope is empty.
func (s *Service) checkBatchThread(ctx context.Context, p Principal, t DeliverThread, scope string) (string, error) {
	c, err := s.Get(ctx, t.ID)
	if err != nil {
		return "", fmt.Errorf("thread %s: %w", t.ID, err)
	}
	if err = s.access.Authorize(ctx, p, DeliverAgent, c.Target); err != nil {
		return "", fmt.Errorf("thread %s: %w", t.ID, err)
	}
	switch {
	case c.Version != t.ExpectedVersion:
		return "", fmt.Errorf("%w: thread %s is at version %d, not %d", ErrConflict, t.ID, c.Version, t.ExpectedVersion)
	case c.Phase != Active:
		return "", fmt.Errorf("%w: thread %s is %s; assigning requires a human and active comment", ErrInvalid, t.ID, c.Status)
	case scope != "" && c.Target.ScopeURI != scope:
		return "", fmt.Errorf("%w: thread %s is in %s, not the batch's scope %s", ErrInvalid, t.ID, c.Target.ScopeURI, scope)
	}
	return c.Target.ScopeURI, nil
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
	q.TenantID = p.TenantID
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
	if target, err = s.withRuleEdits(ctx, c.Target, target); err != nil {
		return nil, err
	}
	return s.delivery.Preview(ctx, c, target)
}
