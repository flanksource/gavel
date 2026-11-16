package comments

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
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
	// Patch is a unified diff a reply proposes instead of editing files.
	Patch string `json:"patch,omitempty"`
	// Edits are Edit-tool-style file changes a reply proposes; exclusive with Patch.
	Edits []FileEdit `json:"edits,omitempty"`
	// Mode is ProposalModeApply or ProposalModeStage for ActionAcceptProposal.
	Mode string `json:"mode,omitempty"`
}

// Act applies req to its comment. A person's reply that mentions one agent
// (an @name of Mentionables) then delivers the thread there in the background;
// two different agents fail the reply before it is written.
func (s *Service) Act(ctx context.Context, req ActionRequest) (Comment, error) {
	out, mention, err := s.actMentioning(ctx, req)
	if err != nil {
		return out, err
	}
	s.deliverMention(ctx, mention)
	return out, nil
}

// mentionDelivery is a reply's agent mention, delivered once the reply is
// stored: by Act right away, by Batch only after its transaction commits.
type mentionDelivery struct {
	thread Comment
	route  Route
}

// actMentioning applies req and answers the delivery its reply's mention asks for; nil for none.
func (s *Service) actMentioning(ctx context.Context, req ActionRequest) (Comment, *mentionDelivery, error) {
	p, err := s.principal(ctx)
	if err != nil {
		return Comment{}, nil, err
	}
	route, err := s.replyMention(ctx, p, req)
	if err != nil {
		return Comment{}, nil, err
	}
	out, err := s.act(ctx, p, req)
	if err != nil || route == nil {
		return out, nil, err
	}
	return out, &mentionDelivery{thread: out, route: *route}, nil
}

// deliverMention delivers m's thread through its mention's route in the background.
func (s *Service) deliverMention(ctx context.Context, m *mentionDelivery) {
	if m == nil {
		return
	}
	requestID := fmt.Sprintf("reply:%s:%d", m.thread.ID, m.thread.Version)
	s.deliverInBackground(ctx, m.thread, func(ctx context.Context) (DeliverTarget, error) {
		return m.route.Target, s.deliverThrough(ctx, m.thread.ID, requestID, m.route)
	})
}

// replyMention is the route req, a reply, delivers its thread through by its agent mention; nil for none.
func (s *Service) replyMention(ctx context.Context, p Principal, req ActionRequest) (*Route, error) {
	if req.Action != "reply" || s.delivery == nil || p.Actor.Kind != "user" || len(mentionTokens(req.Body)) == 0 {
		return nil, nil
	}
	c, err := s.Get(ctx, req.ID)
	if err != nil {
		return nil, err
	}
	return s.mentionRoute(ctx, p, c.Target, req.Body)
}

func (s *Service) act(ctx context.Context, p Principal, req ActionRequest) (Comment, error) {
	if req.ExpectedVersion < 1 {
		return Comment{}, fmt.Errorf("%w: expectedVersion is required", ErrInvalid)
	}
	switch {
	case req.Edits != nil && len(req.Edits) == 0:
		return Comment{}, fmt.Errorf("%w: edits must list at least one edit", ErrInvalid)
	case (req.Patch != "" || len(req.Edits) > 0) && req.Action != "reply":
		return Comment{}, fmt.Errorf("%w: only a reply carries a patch or edits", ErrInvalid)
	case req.Patch != "" && len(req.Edits) > 0:
		return Comment{}, fmt.Errorf("%w: a reply carries a patch or edits, not both", ErrInvalid)
	case req.Mode != "" && req.Action != ActionAcceptProposal:
		return Comment{}, fmt.Errorf("%w: only %s takes a mode", ErrInvalid, ActionAcceptProposal)
	case req.Patch != "" || len(req.Edits) > 0:
		return s.proposeReply(ctx, p, req)
	case req.Action == ActionRequestRevision:
		return s.requestRevision(ctx, p, req)
	}
	return s.mutate(ctx, p, req, nil)
}

// mutate applies req to its comment; proposed, when set, is the proposal a
// reply carries, already stored by the host under the reply's ID.
func (s *Service) mutate(ctx context.Context, p Principal, req ActionRequest, proposed *proposedReply) (Comment, error) {
	return s.store.Mutate(ctx, p.TenantID, req.ID, req.ExpectedVersion, func(c Comment) (Change, error) {
		if err := s.access.Authorize(ctx, p, req.Action, c.Target); err != nil {
			return Change{}, err
		}
		if c.DeletedAt != nil && req.Action != "restore" {
			return Change{}, ErrNotFound
		}
		if !slices.Contains([]string{"reply", "assign", "acknowledge", "metadata", ActionRejectProposal, ActionRequestRevision}, req.Action) {
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
		change, handled, err := s.edit(ctx, p, &c, req, now)
		if err != nil {
			return Change{}, err
		}
		if !handled {
			if change, handled, err = s.decideProposal(ctx, p.Actor, &c, req, now); err != nil {
				return Change{}, err
			}
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
				if change.Message, change.EditMessages, err = reply(&c, p.Actor, req, proposed, now); err != nil {
					return Change{}, err
				}
			default:
				w, ok := s.workflows[c.WorkflowID]
				if !ok {
					return Change{}, fmt.Errorf("%w: unknown workflow %s", ErrInvalid, c.WorkflowID)
				}
				if req.Action == "resolve" && p.Actor.Kind == "agent" {
					if err := requireAcceptedProposal(c); err != nil {
						return Change{}, err
					}
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
				settle(&c, status, phase, &Outcome{Kind: req.Action, Summary: req.Summary, Evidence: req.Evidence, Actor: p.Actor, At: now})
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
		change.Comment = c
		change.Event = Event{CommentID: c.ID, Kind: req.Action, Actor: p.Actor, At: now, Data: data}
		return change, nil
	})
}

// settle moves c to status in phase, recording result as its resolution or outcome.
func settle(c *Comment, status string, phase Phase, result *Outcome) {
	c.Status, c.Phase = status, phase
	if phase == AwaitingOutcome {
		c.Resolution = result
	}
	if phase == Final {
		c.Outcome = result
	}
}

// reply is req's new message on c, carrying proposed when set; a proposal
// supersedes every earlier one still proposed, answered as edited messages.
func reply(c *Comment, actor Actor, req ActionRequest, proposed *proposedReply, now time.Time) (*Message, []Message, error) {
	if c.Phase == Final || strings.TrimSpace(req.Body) == "" {
		return nil, nil, fmt.Errorf("%w: reply requires an active comment and body", ErrInvalid)
	}
	if req.ParentID != "" && !slices.ContainsFunc(c.Messages, func(m Message) bool { return m.ID == req.ParentID }) {
		return nil, nil, fmt.Errorf("%w: reply parent belongs to another comment", ErrInvalid)
	}
	if len(req.Content) > 0 && !json.Valid(req.Content) {
		return nil, nil, fmt.Errorf("%w: malformed content", ErrInvalid)
	}
	msg := &Message{ID: uuid.NewString(), ParentID: req.ParentID, Body: req.Body, Author: actor, Content: req.Content, CreatedAt: now}
	if proposed == nil {
		return msg, nil, nil
	}
	content, err := proposalContent(req.Content, proposed.Proposal)
	if err != nil {
		return nil, nil, err
	}
	msg.ID, msg.Content = proposed.ID, content
	superseded, err := supersede(c)
	return msg, superseded, err
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
