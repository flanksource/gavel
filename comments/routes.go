package comments

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/flanksource/commons/logger"
)

// Route is how a new human-authored root is delivered automatically: the
// target it is handed to, as the host's configuration resolves it for the
// root's target.
type Route struct {
	// Source names the configuration the route was resolved from (a settings
	// layer), and is the route agent's identity.
	Source string        `json:"source"`
	Target DeliverTarget `json:"target"`
	// Assign makes RouteAgent the thread's assignee before the delivery starts,
	// until the session the delivery reaches claims the thread as its own agent.
	Assign bool `json:"assign,omitempty"`
}

// RouteAgent is the actor an Assign route makes a thread's assignee while its
// delivery has not yet reached a session.
func RouteAgent(r Route) Actor {
	label := r.Target.Prompt
	if r.Target.Kind == "session" {
		label = r.Target.SessionID
	}
	return Actor{Kind: "agent", ID: "route:" + r.Source, Label: label}
}

// RouteResolver is the host's automatic delivery configuration.
type RouteResolver interface {
	// RouteFor is the route new root c is delivered through; ok is false when
	// none applies to its target (unconfigured, or opted out).
	RouteFor(ctx context.Context, c Comment) (r Route, ok bool, err error)
}

// AgentRules is a RouteResolver whose configuration names a target's comment
// agent rule whether or not it routes new threads automatically: every
// delivery of the target's threads inherits the rule's edits policy, and @agent
// addresses it.
type AgentRules interface {
	// AgentRule is t's rule as the route it would deliver through; ok is false
	// when the host's configuration does not cover t or configures no agent for it.
	AgentRule(ctx context.Context, t Target) (r Route, ok bool, err error)
}

// CommentAgents is a RouteResolver whose routes come from configuration the
// host can also read and save per target: the comment agent a target's threads
// are handed to. Payloads are the host's own and opaque here.
type CommentAgents interface {
	// CommentAgent is the agent configuration of t; ErrInvalid when the host's
	// configuration does not cover t.
	CommentAgent(ctx context.Context, t Target) (any, error)
	// SaveCommentAgent saves req and answers t's agent configuration as CommentAgent does.
	SaveCommentAgent(ctx context.Context, req SaveCommentAgentRequest) (any, error)
}

// SaveCommentAgentRequest patches the comment agent at Layer of Target's
// configuration: the layer's default, or with Page its entry for the target.
// Agent is a JSON merge patch (RFC 7386) the host applies to the entry the
// layer already holds, under its save lock: a present field replaces that
// field, a null one is removed so it inherits, and a JSON null Agent removes
// the whole entry.
type SaveCommentAgentRequest struct {
	Target Target          `json:"target"`
	Layer  json.RawMessage `json:"layer"`
	Page   bool            `json:"page"`
	Agent  json.RawMessage `json:"agent"`
}

func (s *Service) commentAgents(ctx context.Context, action string, t Target) (CommentAgents, Principal, error) {
	agents, ok := s.routes.(CommentAgents)
	if !ok {
		return nil, Principal{}, fmt.Errorf("%w: comment agent settings are not enabled on this comments host", ErrUnsupported)
	}
	if t.ScopeURI == "" || t.URI == "" {
		return nil, Principal{}, fmt.Errorf("%w: a comment agent target needs scopeUri and uri", ErrInvalid)
	}
	p, err := s.principal(ctx)
	if err != nil {
		return nil, p, err
	}
	return agents, p, s.access.Authorize(ctx, p, action, Target{ScopeURI: t.ScopeURI, URI: t.URI})
}

// CommentAgent is the comment agent configuration of t, as the host's routes resolve it.
func (s *Service) CommentAgent(ctx context.Context, t Target) (any, error) {
	agents, _, err := s.commentAgents(ctx, "comment-agent", t)
	if err != nil {
		return nil, err
	}
	return agents.CommentAgent(ctx, t)
}

// SaveCommentAgent saves the comment agent req names; only a person configures one.
func (s *Service) SaveCommentAgent(ctx context.Context, req SaveCommentAgentRequest) (any, error) {
	agents, p, err := s.commentAgents(ctx, "save-comment-agent", req.Target)
	if err != nil {
		return nil, err
	}
	if p.Actor.Kind != "user" {
		return nil, fmt.Errorf("%w: only a person configures comment agents", ErrForbidden)
	}
	if len(req.Layer) == 0 || len(req.Agent) == 0 {
		return nil, fmt.Errorf("%w: a comment agent save needs layer and agent (null removes the entry)", ErrInvalid)
	}
	return agents.SaveCommentAgent(ctx, req)
}

// routeCreated delivers a new root in the background through explicit (its
// deliver target or its mention), or, for a human-authored root without one,
// through its route: the create never waits, and failures land in its agentRun.
func (s *Service) routeCreated(ctx context.Context, c Comment, explicit *Route) {
	if explicit == nil && (s.routes == nil || c.Author.Kind != "user") {
		return
	}
	s.deliverInBackground(ctx, c, func(ctx context.Context) (DeliverTarget, error) {
		return s.deliverRoute(ctx, c, explicit)
	})
}

// deliverInBackground runs deliver for thread c once the request that changed
// it has answered; failures land in its agentRun.
func (s *Service) deliverInBackground(ctx context.Context, c Comment, deliver func(context.Context) (DeliverTarget, error)) {
	go func() {
		ctx := context.WithoutCancel(ctx)
		target, err := deliver(ctx)
		if err == nil {
			return
		}
		// An executed delivery records its own failure; one that never ran is recorded here.
		if !errors.Is(err, ErrDelivery) {
			err = errors.Join(err, s.recordFailure(ctx, c, target, err))
		}
		// No request is left to answer: it has already responded.
		logger.Errorf("delivering comment %s to an agent: %v", c.ID, err)
	}()
}

// recordFailure records on c a delivery to target that failed before its
// executor ran, under the edits policy the delivery would have had. When that
// policy cannot be resolved (the rule lookup is what failed) the failure is
// still recorded, with the lookup error joined to cause.
func (s *Service) recordFailure(ctx context.Context, c Comment, target DeliverTarget, cause error) error {
	resolved, err := s.withRuleEdits(ctx, c.Target, target)
	if err != nil {
		cause = errors.Join(cause, err)
	}
	return errors.Join(err, s.delivery.RecordFailure(ctx, c.ID, resolved, cause))
}

// deliverRoute delivers c through explicit, or else through its route, queued
// against the thread's latest version: the creator may already have changed it
// (an attachment upload) by the time the delivery starts. An explicit route
// replaces the automatic one entirely, including its assignment.
func (s *Service) deliverRoute(ctx context.Context, c Comment, explicit *Route) (DeliverTarget, error) {
	if explicit != nil {
		return explicit.Target, s.deliverThrough(ctx, c.ID, "create:"+c.ID, *explicit)
	}
	r, err := s.automaticRoute(ctx, c)
	if err != nil || r == nil {
		return DeliverTarget{}, err
	}
	return r.Target, s.deliverThrough(ctx, c.ID, "route:"+c.ID, *r)
}

func (s *Service) automaticRoute(ctx context.Context, c Comment) (*Route, error) {
	if s.routes == nil || c.Author.Kind != "user" {
		return nil, nil
	}
	route, ok, err := s.routes.RouteFor(ctx, c)
	if err != nil {
		return nil, fmt.Errorf("resolving the route of %s: %w", c.Target.URI, err)
	}
	if !ok {
		return nil, nil
	}
	return &route, nil
}

// deliverThrough delivers thread id through r: an Assign route makes its
// agent the thread's assignee before the delivery starts.
func (s *Service) deliverThrough(ctx context.Context, id, requestID string, r Route) error {
	if r.Assign {
		agent := RouteAgent(r)
		if _, err := s.WithLatest(ctx, id, func(latest Comment) (Comment, error) {
			return s.Act(ctx, ActionRequest{ID: id, ExpectedVersion: latest.Version, Action: "assign", Assignee: &agent})
		}); err != nil {
			return fmt.Errorf("assigning comment %s to route %s: %w", id, r.Source, err)
		}
	}
	return s.deliverLatest(ctx, id, requestID, r.Target)
}

func (s *Service) deliverLatest(ctx context.Context, id, requestID string, target DeliverTarget) error {
	_, err := s.WithLatest(ctx, id, func(latest Comment) (Comment, error) {
		return s.Deliver(ctx, DeliverRequest{CommentID: id, ExpectedVersion: latest.Version, RequestID: requestID, Target: target})
	})
	return err
}
