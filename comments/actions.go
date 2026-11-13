package comments

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

type ActionRequest struct {
	ID              string          `json:"id"`
	ExpectedVersion int64           `json:"expectedVersion"`
	Action          string          `json:"action"`
	Body            string          `json:"bodyMarkdown,omitempty"`
	ParentID        string          `json:"parentId,omitempty"`
	Summary         string          `json:"summary,omitempty"`
	Evidence        []string        `json:"evidence,omitempty"`
	Content         json.RawMessage `json:"content,omitempty"`
	Assignee        *Actor          `json:"assignee,omitempty"`
	Target          *Target         `json:"target,omitempty"`
	MessageID       string          `json:"messageId,omitempty"`
}

func (s *Service) Act(ctx context.Context, req ActionRequest) (Comment, error) {
	p, err := s.principal(ctx)
	if err != nil {
		return Comment{}, err
	}
	if req.ExpectedVersion < 1 {
		return Comment{}, fmt.Errorf("%w: expectedVersion is required", ErrInvalid)
	}
	return s.store.Mutate(ctx, p.TenantID, req.ID, req.ExpectedVersion, func(c Comment) (Change, error) {
		if err := s.access.Authorize(ctx, p, req.Action, c.Target); err != nil {
			return Change{}, err
		}
		if c.DeletedAt != nil && req.Action != "restore" {
			return Change{}, ErrNotFound
		}
		if req.Action != "reply" && req.Action != "assign" && req.Action != "acknowledge" && req.Action != "metadata" {
			if store, ok := s.store.(ProviderStore); ok {
				bindings, err := store.Bindings(ctx, p.TenantID, c.ID)
				if err != nil {
					return Change{}, err
				}
				for _, b := range bindings {
					if req.Action == "update" || b.ThreadID != "" {
						return Change{}, fmt.Errorf("%w: use explicit native actions for published comments", ErrConflict)
					}
				}
			}
		}
		before := struct {
			Status     string
			Outcome    *Outcome
			Resolution *Outcome
			Target     Target
			Body       string
			Content    json.RawMessage
			Assignee   *Actor
			DeletedAt  *time.Time
			Message    *Message
		}{Status: c.Status, Outcome: c.Outcome, Resolution: c.Resolution, Target: c.Target, Body: c.Body, Content: c.Content, Assignee: c.Assignee, DeletedAt: c.DeletedAt}
		for _, m := range c.Messages {
			if m.ID == req.MessageID {
				before.Message = &m
				break
			}
		}
		now := time.Now().UTC()
		var msg *Message
		edit, handled, err := s.edit(ctx, p, &c, req, now)
		if err != nil {
			return Change{}, err
		}
		if !handled {
			switch req.Action {
			case "assign":
				if p.Actor.Kind != "user" || c.Phase != Active {
					return Change{}, fmt.Errorf("%w: assigning requires a human and active comment", ErrInvalid)
				}
				if req.Assignee != nil && (req.Assignee.ID == "" || (req.Assignee.Kind != "user" && req.Assignee.Kind != "agent")) {
					return Change{}, ErrInvalid
				}
				c.Assignee = req.Assignee
				if c.WorkflowID == "human-review" {
					c.Status = "open"
					if req.Assignee != nil {
						c.Status = "in_progress"
					}
				}
			case "update":
				if c.Phase == Final || strings.TrimSpace(req.Body) == "" {
					return Change{}, ErrInvalid
				}
				c.Body = req.Body
			case "reply":
				if c.Phase == Final || strings.TrimSpace(req.Body) == "" {
					return Change{}, fmt.Errorf("%w: reply requires an active comment and body", ErrInvalid)
				}
				if req.ParentID != "" {
					found := false
					for _, m := range c.Messages {
						found = found || m.ID == req.ParentID
					}
					if !found {
						return Change{}, fmt.Errorf("%w: reply parent belongs to another comment", ErrInvalid)
					}
				}
				if len(req.Content) > 0 && !json.Valid(req.Content) {
					return Change{}, fmt.Errorf("%w: malformed content", ErrInvalid)
				}
				msg = &Message{ID: uuid.NewString(), ParentID: req.ParentID, Body: req.Body, Author: p.Actor, Content: req.Content, CreatedAt: now}
			default:
				w, ok := s.workflows[c.WorkflowID]
				if !ok {
					return Change{}, fmt.Errorf("%w: unknown workflow %s", ErrInvalid, c.WorkflowID)
				}
				status, phase, err := w.Apply(c.Status, req.Action, p.Actor)
				if err != nil {
					return Change{}, err
				}
				if req.Action == "acknowledge" {
					if c.Assignee != nil && (c.Assignee.ID != p.Actor.ID || c.Assignee.Kind != p.Actor.Kind) && p.Actor.Kind != "user" {
						return Change{}, ErrConflict
					}
					actor := p.Actor
					c.Assignee = &actor
				}
				c.Status, c.Phase = status, phase
				result := &Outcome{Kind: req.Action, Summary: req.Summary, Evidence: req.Evidence, Actor: p.Actor, At: now}
				if phase == AwaitingOutcome {
					c.Resolution = result
				}
				if phase == Final {
					c.Outcome = result
				}
				if req.Action == "reopen" {
					c.Outcome = nil
					c.Resolution = nil
					c.Assignee = nil
				}
			}
		}
		c.UpdatedAt = now
		c.Version++
		data, err := json.Marshal(struct {
			Before  any           `json:"before"`
			Request ActionRequest `json:"request"`
		}{before, req})
		if err != nil {
			return Change{}, err
		}
		return Change{Comment: c, Message: msg, EditMessage: edit.EditMessage, DeleteMessageID: edit.DeleteMessageID, Event: Event{CommentID: c.ID, Kind: req.Action, Actor: p.Actor, At: now, Data: data}}, nil
	})
}
func (s *Service) Events(ctx context.Context, scope string, after int64, limit int) ([]Event, error) {
	p, err := s.principal(ctx)
	if err != nil {
		return nil, err
	}
	if scope == "" || after < 0 || limit < 1 || limit > 200 {
		return nil, ErrInvalid
	}
	if err = s.access.Authorize(ctx, p, "events", Target{ScopeURI: scope}); err != nil {
		return nil, err
	}
	// A tenant feed is filtered through the same resource authorization as reads.
	events, err := s.store.Events(ctx, p.TenantID, scope, after, limit)
	if err != nil {
		return nil, err
	}
	out := make([]Event, 0, len(events))
	for _, e := range events {
		c, err := s.store.Get(ctx, p.TenantID, e.CommentID)
		if err != nil {
			return nil, err
		}
		if err = s.access.Authorize(ctx, p, "read", c.Target); err != nil {
			return nil, err
		}
		if c.Target.ScopeURI == scope {
			out = append(out, e)
		}
	}
	return out, nil
}
