package comments

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
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
	routes          RouteResolver
	delivery        Delivery
	locate          func(context.Context, string) (Location, error)
	proposals       Proposals
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
	s.routes, s.locate, s.proposals = options.Routes, options.Locate, options.Proposals
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
			if slices.Contains([]string{"reply", "assign", "update", "metadata", "retarget", "delete", "restore", "edit-reply", "delete-reply", ActionAcceptProposal, ActionRejectProposal, ActionRequestRevision}, action) || len(t.From) == 0 {
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
	// Deliver hands the new root to this target instead of the route its
	// target resolves to; the create fails when the target is invalid. Without
	// it, a person's body that mentions one agent (an @name of Mentionables) is
	// delivered there; two different agents fail the create.
	Deliver *DeliverTarget `json:"deliver,omitempty"`
	// Attachments are stored on the new root, in the same transaction, before
	// it is routed or delivered; each is checked (name, media type, size up to
	// the host's attachment limit) before anything is written.
	Attachments []CreateAttachment `json:"attachments,omitempty"`
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
	route, err := s.createRoute(ctx, p, target, req)
	if err != nil {
		return Comment{}, err
	}
	files, err := s.createAttachments(p, req.Attachments)
	if err != nil {
		return Comment{}, err
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
	created, fresh, err := s.createRoot(ctx, p, newRoot{comment: c, requestID: req.RequestID, hash: hex.EncodeToString(hash[:]), files: files})
	if err != nil {
		return Comment{}, err
	}
	if created.DeletedAt != nil {
		return Comment{}, ErrConflict
	}
	if fresh {
		s.routeCreated(ctx, created, route)
	}
	return created, nil
}

// createRoute is the route req's root is delivered through instead of its
// automatic one: its explicit deliver target, else its body's agent mention.
func (s *Service) createRoute(ctx context.Context, p Principal, target Target, req CreateRequest) (*Route, error) {
	if req.Deliver == nil {
		return s.mentionRoute(ctx, p, target, req.Body)
	}
	return &Route{Source: "deliver", Target: *req.Deliver}, s.checkTarget(*req.Deliver)
}

// createFile is one of a create's attachments, checked and described.
type createFile struct {
	Attachment
	data []byte
}

func (s *Service) createAttachments(p Principal, files []CreateAttachment) ([]createFile, error) {
	if len(files) > MaxCreateAttachments {
		return nil, fmt.Errorf("%w: a create carries at most %d attachments, not %d", ErrInvalid, MaxCreateAttachments, len(files))
	}
	out := make([]createFile, 0, len(files))
	for _, f := range files {
		a, err := s.newAttachment(p, "", f)
		if err != nil {
			return nil, err
		}
		out = append(out, createFile{Attachment: a, data: f.Data})
	}
	return out, nil
}

// newRoot is a root to store: the comment, its create's idempotency key and
// request hash, and the files it carries.
type newRoot struct {
	comment         Comment
	requestID, hash string
	files           []createFile
}

// createRoot stores the root and, when it is new, its files, in one
// transaction; fresh is false for a replayed create, which stores nothing and
// is not routed again.
func (s *Service) createRoot(ctx context.Context, p Principal, root newRoot) (Comment, bool, error) {
	c, key, hash, files := root.comment, root.requestID, root.hash, root.files
	event := Event{CommentID: c.ID, Kind: "created", Actor: p.Actor, At: c.CreatedAt}
	if len(files) == 0 {
		created, err := s.store.Create(ctx, c, key, hash, event)
		return created, err == nil && created.ID == c.ID, err
	}
	store, ok := s.store.(TransactionStore)
	if !ok {
		return Comment{}, false, fmt.Errorf("%w: this comment store cannot create a thread with attachments", ErrUnsupported)
	}
	var created Comment
	var stored []string
	err := store.Transaction(ctx, func(tx Store) error {
		var err error
		if created, err = tx.Create(ctx, c, key, hash, event); err != nil || created.ID != c.ID {
			return err
		}
		for _, f := range files {
			if err = s.blobs.Put(ctx, p.TenantID, f.ID, f.data); err != nil {
				return err
			}
			stored = append(stored, f.ID)
			if created, err = tx.Mutate(ctx, p.TenantID, c.ID, created.Version, attachTo(p, f.Attachment)); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		for _, id := range stored {
			err = errors.Join(err, s.blobs.Delete(context.WithoutCancel(ctx), p.TenantID, id))
		}
		return Comment{}, false, err
	}
	return created, created.ID == c.ID, nil
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
