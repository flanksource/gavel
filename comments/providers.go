package comments

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type Binding struct {
	Provider   string            `json:"provider"`
	RemoteKey  string            `json:"remoteKey"`
	ThreadID   string            `json:"threadId,omitempty"`
	RootID     string            `json:"rootId"`
	URL        string            `json:"url"`
	Revision   string            `json:"revision"`
	Resolvable bool              `json:"resolvable"`
	Outdated   bool              `json:"outdated"`
	Native     json.RawMessage   `json:"native,omitempty"`
	MessageIDs map[string]string `json:"messageIds,omitempty"`
}
type RemoteMessage struct {
	ID        string    `json:"id"`
	ParentID  string    `json:"parentId,omitempty"`
	Body      string    `json:"bodyMarkdown"`
	Author    Actor     `json:"author"`
	CreatedAt time.Time `json:"createdAt"`
}
type RemoteComment struct {
	Binding          Binding         `json:"binding"`
	Body             string          `json:"bodyMarkdown"`
	Author           Actor           `json:"author"`
	CreatedAt        time.Time       `json:"createdAt"`
	Resolved         bool            `json:"resolved"`
	ResolutionReason string          `json:"resolutionReason,omitempty"`
	Messages         []RemoteMessage `json:"messages"`
}
type Publication struct {
	Provider string   `json:"provider"`
	Comment  Comment  `json:"comment"`
	Binding  *Binding `json:"binding,omitempty"`
	Message  *Message `json:"message,omitempty"`
}
type Provider interface {
	Name() string
	Fetch(context.Context, Principal, Target) ([]RemoteComment, error)
	Apply(context.Context, Principal, string, Publication) (RemoteComment, error)
}
type ProviderStore interface {
	Binding(context.Context, string, string, string) (Binding, error)
	Bindings(context.Context, string, string) ([]Binding, error)
	ApplyRemote(context.Context, Comment, RemoteComment, Actor) (Comment, error)
}
type PublicationRequest struct {
	CommentID       string `json:"commentId"`
	ExpectedVersion int64  `json:"expectedVersion"`
	RequestID       string `json:"requestId"`
	Provider        string `json:"provider"`
	Action          string `json:"action"`
	MessageID       string `json:"messageId,omitempty"`
}

func (s *Service) Publish(ctx context.Context, req PublicationRequest) (Operation, error) {
	if req.RequestID == "" || req.ExpectedVersion < 1 {
		return Operation{}, ErrInvalid
	}
	if _, ok := s.providers[req.Provider]; !ok {
		return Operation{}, fmt.Errorf("%w: provider %s is not registered", ErrInvalid, req.Provider)
	}
	store, ok := s.store.(ProviderStore)
	if !ok {
		return Operation{}, fmt.Errorf("provider store is required")
	}
	p, err := s.principal(ctx)
	if err != nil {
		return Operation{}, err
	}
	c, err := s.Get(ctx, req.CommentID)
	if err != nil {
		return Operation{}, err
	}
	operations, ok := s.store.(OperationStore)
	if !ok {
		return Operation{}, ErrInvalid
	}
	previous, lookupErr := operations.OperationByRequest(ctx, p.TenantID, req.RequestID)
	if lookupErr == nil {
		var admitted Publication
		if err = json.Unmarshal(previous.Payload, &admitted); err != nil {
			return Operation{}, err
		}
		messageID := ""
		if admitted.Message != nil {
			messageID = admitted.Message.ID
		}
		if previous.CommentID != c.ID || previous.Kind != req.Action || previous.Principal.Actor.Kind != p.Actor.Kind || previous.Principal.Actor.ID != p.Actor.ID || admitted.Provider != req.Provider || admitted.Comment.Version != req.ExpectedVersion || messageID != req.MessageID {
			return Operation{}, ErrConflict
		}
		if err = s.access.Authorize(ctx, p, req.Action, c.Target); err != nil {
			return Operation{}, err
		}
		return previous, nil
	}
	if lookupErr != ErrNotFound {
		return Operation{}, lookupErr
	}
	publication := Publication{Provider: req.Provider, Comment: c}
	binding, err := store.Binding(ctx, p.TenantID, c.ID, req.Provider)
	if err == nil {
		publication.Binding = &binding
	} else if err != ErrNotFound {
		return Operation{}, err
	}
	switch req.Action {
	case "publish":
		if publication.Binding != nil {
			return Operation{}, fmt.Errorf("%w: already published", ErrConflict)
		}
		if c.Phase == Final {
			return Operation{}, ErrConflict
		}
	case "publish-reply":
		if publication.Binding == nil || binding.ThreadID == "" {
			return Operation{}, ErrInvalid
		}
		for _, m := range c.Messages {
			if m.ID == req.MessageID {
				copy := m
				publication.Message = &copy
			}
		}
		if publication.Message == nil {
			return Operation{}, fmt.Errorf("%w: reply does not belong to comment", ErrInvalid)
		}
		for _, localID := range binding.MessageIDs {
			if localID == req.MessageID {
				return Operation{}, ErrConflict
			}
		}
	case "resolve-native", "reopen-native":
		if publication.Binding == nil || !binding.Resolvable {
			return Operation{}, fmt.Errorf("%w: native thread does not support resolution", ErrInvalid)
		}
	default:
		return Operation{}, ErrInvalid
	}
	payload, err := json.Marshal(publication)
	if err != nil {
		return Operation{}, err
	}
	return s.Queue(ctx, QueueRequest{CommentID: c.ID, ExpectedVersion: req.ExpectedVersion, Kind: req.Action, RequestID: req.RequestID, Payload: payload})
}
func (s *Service) Sync(ctx context.Context, provider string, target Target) ([]Comment, error) {
	adapter, ok := s.providers[provider]
	if !ok {
		return nil, fmt.Errorf("%w: unknown provider", ErrInvalid)
	}
	store, ok := s.store.(ProviderStore)
	if !ok {
		return nil, ErrInvalid
	}
	p, err := s.principal(ctx)
	if err != nil {
		return nil, err
	}
	target, err = s.targets.Resolve(ctx, p, target)
	if err != nil {
		return nil, err
	}
	if target.URI == "" || target.ScopeURI == "" || target.Kind == "" {
		return nil, ErrInvalid
	}
	if err = s.access.Authorize(ctx, p, "sync", target); err != nil {
		return nil, err
	}
	snapshots, err := adapter.Fetch(ctx, p, target)
	if err != nil {
		return nil, err
	}
	out := make([]Comment, 0, len(snapshots))
	for _, snapshot := range snapshots {
		if err = validateRemote(provider, snapshot); err != nil {
			return out, err
		}
		c := Comment{TenantID: p.TenantID, Target: target, WorkflowID: NativePR().ID}
		c, err = store.ApplyRemote(ctx, c, snapshot, p.Actor)
		if err != nil {
			return out, err
		}
		out = append(out, c)
	}
	return out, nil
}
func validateRemote(provider string, r RemoteComment) error {
	if r.Binding.Provider != provider || r.Binding.RemoteKey == "" || r.Binding.RootID == "" || r.Binding.Revision == "" || r.Author.ID == "" || strings.TrimSpace(r.Body) == "" || r.CreatedAt.IsZero() {
		return fmt.Errorf("%w: incomplete native comment", ErrInvalid)
	}
	seen := map[string]bool{r.Binding.RootID: true}
	for _, m := range r.Messages {
		if m.ID == "" || seen[m.ID] || m.Author.ID == "" || m.CreatedAt.IsZero() || strings.TrimSpace(m.Body) == "" {
			return fmt.Errorf("%w: invalid native reply", ErrInvalid)
		}
		seen[m.ID] = true
	}
	return nil
}

type providerExecutor struct{ service *Service }

func (s *Service) ProcessProvider(ctx context.Context, lease time.Duration) (Operation, error) {
	return s.Process(ctx, providerExecutor{service: s}, lease)
}
func (e providerExecutor) Execute(ctx context.Context, op Operation) (json.RawMessage, error) {
	var publication Publication
	if err := json.Unmarshal(op.Payload, &publication); err != nil {
		return nil, err
	}
	s := e.service
	adapter, ok := s.providers[publication.Provider]
	if !ok {
		return nil, ErrInvalid
	}
	if publication.Comment.ID != op.CommentID || publication.Comment.TenantID != op.TenantID {
		return nil, ErrForbidden
	}
	snapshot, err := adapter.Apply(ctx, op.Principal, op.Kind, publication)
	receipt, marshalErr := json.Marshal(snapshot)
	if marshalErr != nil {
		return nil, marshalErr
	}
	if err != nil {
		return receipt, err
	}
	if err = validateRemote(publication.Provider, snapshot); err != nil {
		return receipt, err
	}
	store, ok := s.store.(ProviderStore)
	if !ok {
		return receipt, ErrInvalid
	}
	if publication.Message != nil {
		found := false
		for _, m := range snapshot.Messages {
			found = found || snapshot.Binding.MessageIDs[m.ID] == publication.Message.ID
		}
		if !found {
			return receipt, fmt.Errorf("native receipt must identify the published reply")
		}
	}
	_, err = store.ApplyRemote(ctx, publication.Comment, snapshot, op.Principal.Actor)
	return receipt, err
}

func (s *Service) Bindings(ctx context.Context, id string) ([]Binding, error) {
	c, err := s.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	store, ok := s.store.(ProviderStore)
	if !ok {
		return nil, fmt.Errorf("provider store is required")
	}
	return store.Bindings(ctx, c.TenantID, c.ID)
}
