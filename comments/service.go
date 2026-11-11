package comments

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/url"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Service struct {
	store           Store
	access          Access
	targets         Targets
	workflows       map[string]Workflow
	providers       map[string]Provider
	blobs           Blobs
	attachmentLimit int64
	routes          RouteStore
	delivery        Delivery
}

func New(options Options) (*Service, error) {
	if options.Store == nil || options.Access == nil || options.Targets == nil {
		return nil, fmt.Errorf("%w: store, access and targets are required", ErrInvalid)
	}
	s := &Service{store: options.Store, access: options.Access, targets: options.Targets, workflows: map[string]Workflow{}, providers: map[string]Provider{}}
	if options.Blobs != nil && options.AttachmentLimit <= 0 {
		return nil, fmt.Errorf("%w: explicit positive attachment limit is required", ErrInvalid)
	}
	s.blobs, s.attachmentLimit = options.Blobs, options.AttachmentLimit
	if options.Routes != nil && options.Delivery == nil {
		return nil, fmt.Errorf("%w: comment routes require agent delivery", ErrInvalid)
	}
	s.routes = options.Routes
	if options.Delivery != nil {
		if s.delivery = options.Delivery(s); s.delivery == nil {
			return nil, fmt.Errorf("%w: delivery constructor returned nil", ErrInvalid)
		}
	}
	for _, p := range options.Providers {
		if p == nil || p.Name() == "" {
			return nil, ErrInvalid
		}
		if _, ok := s.providers[p.Name()]; ok {
			return nil, fmt.Errorf("%w: duplicate provider", ErrInvalid)
		}
		s.providers[p.Name()] = p
	}
	for _, w := range options.Workflows {
		if w.ID == "" || w.Statuses[w.Initial] != Active {
			return nil, fmt.Errorf("%w: workflow %s needs an active initial state", ErrInvalid, w.ID)
		}
		if _, ok := s.workflows[w.ID]; ok {
			return nil, fmt.Errorf("%w: duplicate workflow %s", ErrInvalid, w.ID)
		}
		for _, p := range w.Statuses {
			if p != Active && p != AwaitingOutcome && p != Final {
				return nil, fmt.Errorf("%w: workflow phase %s", ErrInvalid, p)
			}
		}
		for action, t := range w.Actions {
			if slices.Contains([]string{"reply", "assign", "update", "metadata", "retarget", "delete", "restore", "edit-reply", "delete-reply"}, action) || len(t.From) == 0 {
				return nil, fmt.Errorf("%w: invalid workflow action %s", ErrInvalid, action)
			}
			for _, from := range t.From {
				if _, ok := w.Statuses[from]; !ok {
					return nil, fmt.Errorf("%w: unknown workflow source %s", ErrInvalid, from)
				}
			}
			if _, ok := w.Statuses[t.To]; !ok {
				return nil, fmt.Errorf("%w: workflow destination %s", ErrInvalid, t.To)
			}
		}
		s.workflows[w.ID] = copyWorkflow(w)
	}
	if len(s.workflows) == 0 {
		return nil, fmt.Errorf("%w: workflows are required", ErrInvalid)
	}
	if len(s.providers) > 0 {
		for _, id := range []string{HumanReview().ID, NativePR().ID} {
			if _, ok := s.workflows[id]; !ok {
				return nil, fmt.Errorf("%w: native providers require workflow %s", ErrInvalid, id)
			}
		}
	}
	return s, nil
}
func (s *Service) principal(ctx context.Context) (Principal, error) {
	p, err := s.access.Principal(ctx)
	if err != nil {
		return p, err
	}
	if p.TenantID == "" || p.Actor.ID == "" || (p.Actor.Kind != "user" && p.Actor.Kind != "agent" && p.Actor.Kind != "system") {
		return p, fmt.Errorf("%w: authenticated tenant and actor required", ErrForbidden)
	}
	return p, nil
}

type CreateRequest struct {
	Target     Target          `json:"target"`
	Body       string          `json:"bodyMarkdown"`
	WorkflowID string          `json:"workflowId"`
	RequestID  string          `json:"requestId"`
	Content    json.RawMessage `json:"content,omitempty"`
	Links      []Link          `json:"links,omitempty"`
}

func (s *Service) Create(ctx context.Context, req CreateRequest) (Comment, error) {
	p, err := s.principal(ctx)
	if err != nil {
		return Comment{}, err
	}
	w, ok := s.workflows[req.WorkflowID]
	if !ok || strings.TrimSpace(req.Body) == "" || req.RequestID == "" {
		return Comment{}, fmt.Errorf("%w: body, request ID and known workflow required", ErrInvalid)
	}
	if err = s.access.Authorize(ctx, p, "create", req.Target); err != nil {
		return Comment{}, err
	}
	target, err := s.targets.Resolve(ctx, p, req.Target)
	if err != nil {
		return Comment{}, err
	}
	uri, err := url.Parse(target.URI)
	if err != nil || uri.Scheme == "" || target.ScopeURI == "" || target.Kind == "" {
		return Comment{}, fmt.Errorf("%w: resolved target needs URI, scope and kind", ErrInvalid)
	}
	if err = s.access.Authorize(ctx, p, "create", target); err != nil {
		return Comment{}, err
	}
	if len(req.Content) > 0 && !json.Valid(req.Content) {
		return Comment{}, fmt.Errorf("%w: malformed content", ErrInvalid)
	}
	identity := p.Actor
	identity.Label = ""
	data, err := json.Marshal(struct {
		Request CreateRequest
		Actor   Actor
	}{req, identity})
	if err != nil {
		return Comment{}, err
	}
	hash := sha256.Sum256(data)
	now := time.Now().UTC()
	c := Comment{ID: uuid.NewString(), TenantID: p.TenantID, Target: target, Body: req.Body, Author: p.Actor, WorkflowID: w.ID, Status: w.Initial, Phase: Active, Content: req.Content, Links: req.Links, Version: 1, CreatedAt: now, UpdatedAt: now}
	created, err := s.store.Create(ctx, c, req.RequestID, hex.EncodeToString(hash[:]), Event{CommentID: c.ID, Kind: "created", Actor: p.Actor, At: now})
	if err != nil {
		return Comment{}, err
	}
	if created.DeletedAt != nil {
		return Comment{}, ErrConflict
	}
	s.routeCreated(ctx, created)
	return created, nil
}
func (s *Service) Get(ctx context.Context, id string) (Comment, error) {
	p, err := s.principal(ctx)
	if err != nil {
		return Comment{}, err
	}
	c, err := s.store.Get(ctx, p.TenantID, id)
	if err != nil {
		return c, err
	}
	if c.DeletedAt != nil {
		return Comment{}, ErrNotFound
	}
	return c, s.access.Authorize(ctx, p, "read", c.Target)
}
func (s *Service) List(ctx context.Context, f Filter) ([]Comment, error) {
	page, err := s.Page(ctx, f)
	return page.Items, err
}
