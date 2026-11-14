package comments

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/flanksource/commons/logger"
	"github.com/google/uuid"
)

// Route delivers new human-authored roots under TargetPrefix to Target automatically.
type Route struct {
	ID           string        `json:"id"`
	TenantID     string        `json:"-"`
	ScopeURI     string        `json:"scopeUri"`
	TargetPrefix string        `json:"targetPrefix"`
	WorkflowID   string        `json:"workflowId"`
	Target       DeliverTarget `json:"target"`
	Enabled      bool          `json:"enabled"`
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
	return Actor{Kind: "agent", ID: "route:" + r.ID, Label: label}
}

type RouteStore interface {
	Routes(ctx context.Context, tenant, scope string) ([]Route, error)
	Route(ctx context.Context, tenant, id string) (Route, error)
	SaveRoute(context.Context, Route) (Route, error)
	DeleteRoute(ctx context.Context, tenant, id string) error
}

// MatchRoute is the route with the longest TargetPrefix of uri for workflow.
// A disabled best match opts its subtree out: nothing is delivered.
func MatchRoute(routes []Route, uri, workflow string) (Route, bool) {
	var best Route
	found := false
	for _, r := range routes {
		if r.WorkflowID == workflow && under(uri, r.TargetPrefix) && (!found || len(r.TargetPrefix) > len(best.TargetPrefix)) {
			best, found = r, true
		}
	}
	return best, found && best.Enabled
}

// under is whether uri is prefix or lies beneath it: prefix .../page/team covers
// .../page/team/members but not its sibling .../page/team-x.
func under(uri, prefix string) bool {
	rest, ok := strings.CutPrefix(uri, prefix)
	return ok && (rest == "" || strings.HasSuffix(prefix, "/") || strings.HasPrefix(rest, "/"))
}

func (s *Service) requireRoutes() error {
	if s.routes == nil {
		return fmt.Errorf("%w: comment routes are not enabled on this comments host", ErrUnsupported)
	}
	return nil
}

// routePrincipal is the human allowed to read or change routes in scope.
func (s *Service) routePrincipal(ctx context.Context, scope string) (Principal, error) {
	if err := s.requireRoutes(); err != nil {
		return Principal{}, err
	}
	p, err := s.principal(ctx)
	if err != nil {
		return p, err
	}
	if p.Actor.Kind != "user" {
		return p, fmt.Errorf("%w: only a human manages comment routes", ErrForbidden)
	}
	if scope == "" {
		return p, fmt.Errorf("%w: scope is required", ErrInvalid)
	}
	return p, s.access.Authorize(ctx, p, "routes", Target{ScopeURI: scope})
}

func (s *Service) Routes(ctx context.Context, scope string) ([]Route, error) {
	p, err := s.routePrincipal(ctx, scope)
	if err != nil {
		return nil, err
	}
	return s.routes.Routes(ctx, p.TenantID, scope)
}

// SaveRoute creates a route (empty ID) or replaces one in the same tenant.
func (s *Service) SaveRoute(ctx context.Context, r Route) (Route, error) {
	p, err := s.routePrincipal(ctx, r.ScopeURI)
	if err != nil {
		return Route{}, err
	}
	if _, ok := s.workflows[r.WorkflowID]; !ok || !under(r.TargetPrefix, r.ScopeURI) {
		return Route{}, fmt.Errorf("%w: route needs a registered workflow and a targetPrefix inside its scope", ErrInvalid)
	}
	if err = s.checkTarget(r.Target); err != nil {
		return Route{}, err
	}
	if r.ID == "" {
		r.ID = uuid.NewString()
	} else if existing, err := s.routes.Route(ctx, p.TenantID, r.ID); err != nil {
		return Route{}, err
	} else if _, err = s.routePrincipal(ctx, existing.ScopeURI); err != nil {
		return Route{}, err
	}
	r.TenantID = p.TenantID
	return s.routes.SaveRoute(ctx, r)
}

func (s *Service) DeleteRoute(ctx context.Context, id string) error {
	if err := s.requireRoutes(); err != nil {
		return err
	}
	p, err := s.principal(ctx)
	if err != nil {
		return err
	}
	r, err := s.routes.Route(ctx, p.TenantID, id)
	if err != nil {
		return err
	}
	if _, err = s.routePrincipal(ctx, r.ScopeURI); err != nil {
		return err
	}
	return s.routes.DeleteRoute(ctx, p.TenantID, id)
}

// routeCreated delivers a new human-authored root through its matching route in
// the background: the create never waits, and failures land in its agentRun.
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

// deliverRoute delivers c through its matching route, queued against the
// thread's latest version: the creator may already have changed it (an
// attachment upload) by the time the route is looked up.
func (s *Service) deliverRoute(ctx context.Context, c Comment) (DeliverTarget, error) {
	routes, err := s.routes.Routes(ctx, c.TenantID, c.Target.ScopeURI)
	if err != nil {
		return DeliverTarget{}, fmt.Errorf("looking up the routes of scope %s: %w", c.Target.ScopeURI, err)
	}
	r, ok := MatchRoute(routes, c.Target.URI, c.WorkflowID)
	if !ok {
		return r.Target, nil
	}
	if r.Assign {
		agent := RouteAgent(r)
		if _, err = s.WithLatest(ctx, c.ID, func(latest Comment) (Comment, error) {
			return s.Act(ctx, ActionRequest{ID: c.ID, ExpectedVersion: latest.Version, Action: "assign", Assignee: &agent})
		}); err != nil {
			return r.Target, fmt.Errorf("assigning comment %s to route %s: %w", c.ID, r.ID, err)
		}
	}
	_, err = s.WithLatest(ctx, c.ID, func(latest Comment) (Comment, error) {
		return s.Deliver(ctx, DeliverRequest{CommentID: c.ID, ExpectedVersion: latest.Version, RequestID: "route:" + c.ID, Target: r.Target})
	})
	return r.Target, err
}
