package comments

import (
	"context"
	"fmt"
)

// CreateDeliveryChoice supplies a draft's picked mention and the session to use
// when neither a picked nor a written mention addresses another destination.
type CreateDeliveryChoice struct {
	DefaultSessionID string `json:"defaultSessionId,omitempty"`
	MentionID        string `json:"mentionId,omitempty"`
}

type CreateDeliveryDraft struct {
	Target   Target                `json:"target"`
	Body     string                `json:"bodyMarkdown,omitempty"`
	Deliver  *DeliverTarget        `json:"deliver,omitempty"`
	Delivery *CreateDeliveryChoice `json:"delivery,omitempty"`
}

type CreateDeliveryPreview struct {
	Route   *Route        `json:"route,omitempty"`
	Session *AgentSession `json:"session,omitempty"`
	Mention bool          `json:"mention,omitempty"`
}

const rememberedSessionSource = "remembered-session"

func (s *Service) draftRoute(ctx context.Context, p Principal, draft CreateDeliveryDraft) (*Route, error) {
	if draft.Deliver != nil {
		if draft.Delivery != nil {
			return nil, fmt.Errorf("%w: deliver and delivery cannot be combined", ErrInvalid)
		}
		return s.normalizeRoute(Route{Source: "deliver", Target: *draft.Deliver})
	}
	choice := draft.Delivery
	if choice != nil && (choice.MentionID != "" || choice.DefaultSessionID != "") {
		if p.Actor.Kind != "user" {
			return nil, fmt.Errorf("%w: only a person can choose a draft delivery target", ErrForbidden)
		}
		if err := s.requireDelivery(); err != nil {
			return nil, err
		}
		if choice.MentionID != "" {
			return s.pickedMentionRoute(ctx, p, draft.Target, choice.MentionID)
		}
	}
	route, err := s.mentionRoute(ctx, p, draft.Target, draft.Body)
	if route != nil || err != nil || choice == nil || choice.DefaultSessionID == "" {
		return route, err
	}
	target := DeliverTarget{Kind: "session", SessionID: choice.DefaultSessionID}
	return s.normalizeRoute(Route{Source: rememberedSessionSource, Target: target})
}

func (s *Service) normalizeRoute(route Route) (*Route, error) {
	target, err := s.normalizeDeliveryTarget(route.Target)
	route.Target = target
	return &route, err
}

func (s *Service) pickedMentionRoute(ctx context.Context, p Principal, target Target, id string) (*Route, error) {
	mentionables, err := s.mentionables(ctx, p, target, true)
	if err != nil {
		return nil, err
	}
	for _, mention := range mentionables {
		if mention.ID == id {
			return s.normalizeRoute(mention.Route)
		}
	}
	return nil, fmt.Errorf("%w: delivery mention %q is no longer available", ErrInvalid, id)
}

// PreviewCreateDelivery resolves delivery for a human-review draft without creating a thread.
func (s *Service) PreviewCreateDelivery(ctx context.Context, draft CreateDeliveryDraft) (CreateDeliveryPreview, error) {
	p, err := s.principal(ctx)
	if err != nil {
		return CreateDeliveryPreview{}, err
	}
	draft.Target, err = s.resolveCreateTarget(ctx, p, draft.Target)
	if err != nil {
		return CreateDeliveryPreview{}, err
	}
	route, err := s.draftRoute(ctx, p, draft)
	mention := route != nil && route.Source != rememberedSessionSource && route.Source != "deliver"
	if err == nil && route == nil {
		route, err = s.automaticRoute(ctx, Comment{Target: draft.Target, Author: p.Actor, WorkflowID: HumanReview().ID})
	}
	if err != nil || route == nil {
		return CreateDeliveryPreview{}, err
	}
	route.Target, err = s.withRuleEdits(ctx, draft.Target, route.Target)
	if err != nil {
		return CreateDeliveryPreview{}, err
	}
	session, err := s.previewSession(ctx, draft.Target.ScopeURI, route.Target)
	return CreateDeliveryPreview{Route: route, Session: session, Mention: mention}, err
}

func (s *Service) previewSession(ctx context.Context, scope string, target DeliverTarget) (*AgentSession, error) {
	if target.Fresh {
		return nil, nil
	}
	query := SessionQuery{ScopeURI: scope, All: true}
	if target.Kind == "run" {
		query.GroupKey = target.EffectiveGroupKey()
	} else {
		query.SessionID = target.SessionID
	}
	sessions, err := s.AgentSessions(ctx, query)
	if err != nil {
		return nil, err
	}
	for _, session := range sessions {
		if query.GroupKey != "" || session.ID == target.SessionID {
			return &session, nil
		}
	}
	return nil, nil
}
