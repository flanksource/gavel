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

// SaveCommentAgentRequest makes Agent the comment agent at Layer of Target's
// configuration: the layer's default, or with Page its entry for the target.
// A JSON null Agent removes that entry.
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

// routeCreated delivers a new human-authored root through its route in the
// background: the create never waits, and failures land in its agentRun.
func (s *Service) routeCreated(ctx context.Context, c Comment) {
	if s.routes == nil || c.Author.Kind != "user" || c.Version != 1 {
		return
	}
	go func() {
		ctx := context.WithoutCancel(ctx)
		target, err := s.deliverRoute(ctx, c)
		if err == nil {
			return
		}
		// An executed delivery records its own failure; one that never ran is recorded here.
		if !errors.Is(err, ErrDelivery) {
			err = errors.Join(err, s.delivery.RecordFailure(ctx, c.ID, target, err))
		}
		// No request is left to answer: the create has already responded.
		logger.Errorf("routing comment %s to an agent: %v", c.ID, err)
	}()
}

// deliverRoute delivers c through its route, queued against the thread's
// latest version: the creator may already have changed it (an attachment
// upload) by the time the route is resolved.
func (s *Service) deliverRoute(ctx context.Context, c Comment) (DeliverTarget, error) {
	r, ok, err := s.routes.RouteFor(ctx, c)
	if err != nil {
		return DeliverTarget{}, fmt.Errorf("resolving the route of %s: %w", c.Target.URI, err)
	}
	if !ok {
		return r.Target, nil
	}
	if r.Assign {
		agent := RouteAgent(r)
		if _, err = s.WithLatest(ctx, c.ID, func(latest Comment) (Comment, error) {
			return s.Act(ctx, ActionRequest{ID: c.ID, ExpectedVersion: latest.Version, Action: "assign", Assignee: &agent})
		}); err != nil {
			return r.Target, fmt.Errorf("assigning comment %s to route %s: %w", c.ID, r.Source, err)
		}
	}
	_, err = s.WithLatest(ctx, c.ID, func(latest Comment) (Comment, error) {
		return s.Deliver(ctx, DeliverRequest{CommentID: c.ID, ExpectedVersion: latest.Version, RequestID: "route:" + c.ID, Target: r.Target})
	})
	return r.Target, err
}
