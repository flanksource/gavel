package comments

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
)

var errNoProposals = fmt.Errorf("%w: this host cannot hold proposed patches", ErrInvalid)

// proposedReply is a reply's proposal the host stored before the reply itself.
type proposedReply struct {
	ID       string
	Proposal Proposal
}

// threadAgentRun is what the service reads of a root's content.agentRun.
type threadAgentRun struct {
	SessionID string `json:"sessionId"`
	Edits     string `json:"edits"`
}

// proposeReply stores req.Patch or req.Edits as a proposal under a new reply ID, then the
// reply holding it; a reply that is never stored discards its proposal.
func (s *Service) proposeReply(ctx context.Context, p Principal, req ActionRequest) (Comment, error) {
	if s.proposals == nil {
		return Comment{}, errNoProposals
	}
	if strings.TrimSpace(req.Body) == "" {
		return Comment{}, fmt.Errorf("%w: a proposal reply needs a body", ErrInvalid)
	}
	if _, err := proposalContent(req.Content, Proposal{}); err != nil {
		return Comment{}, err
	}
	c, err := s.Get(ctx, req.ID)
	if err != nil {
		return Comment{}, err
	}
	replyID := uuid.NewString()
	proposal, err := s.proposals.Propose(ctx, c, replyID, ProposeInput{Patch: []byte(req.Patch), Edits: req.Edits})
	if err != nil {
		return Comment{}, fmt.Errorf("proposing a patch on comment %s: %w", c.ID, err)
	}
	if proposal.State != ProposalProposed {
		err = fmt.Errorf("proposal host answered a %q proposal, not %q", proposal.State, ProposalProposed)
	} else {
		var out Comment
		if out, err = s.mutate(ctx, p, req, &proposedReply{ID: replyID, Proposal: proposal}); err == nil {
			return out, nil
		}
	}
	return Comment{}, errors.Join(err, s.proposals.Discard(context.WithoutCancel(ctx), c, proposal))
}

// requestRevision records the feedback and hands the thread back to the
// agent session that proposed; a thread without one fails before any write.
func (s *Service) requestRevision(ctx context.Context, p Principal, req ActionRequest) (Comment, error) {
	c, err := s.Get(ctx, req.ID)
	if err != nil {
		return Comment{}, err
	}
	target, err := s.revisionTarget(c)
	if err != nil {
		return Comment{}, err
	}
	revised, err := s.mutate(ctx, p, req, nil)
	if err != nil {
		return Comment{}, err
	}
	return s.Deliver(ctx, DeliverRequest{CommentID: revised.ID, ExpectedVersion: revised.Version, RequestID: fmt.Sprintf("revision:%s:%d", revised.ID, revised.Version), Target: target})
}

// revisionTarget is the agent session c's revision requests go to.
func (s *Service) revisionTarget(c Comment) (DeliverTarget, error) {
	var run threadAgentRun
	ok, err := contentValue(c.Content, "agentRun", &run)
	if err != nil {
		return DeliverTarget{}, fmt.Errorf("comment %s agentRun: %w", c.ID, err)
	}
	if !ok || strings.TrimSpace(run.SessionID) == "" {
		return DeliverTarget{}, fmt.Errorf("%w: comment %s has no agent session to ask for a revision", ErrInvalid, c.ID)
	}
	target := DeliverTarget{Kind: "session", SessionID: run.SessionID, Edits: EditsPropose}
	return target, s.checkTarget(target)
}

// decideProposal handles a human's decision on a reply's proposal.
func (s *Service) decideProposal(ctx context.Context, actor Actor, c *Comment, req ActionRequest, now time.Time) (Change, bool, error) {
	if !slices.Contains([]string{ActionAcceptProposal, ActionRejectProposal, ActionRequestRevision}, req.Action) {
		return Change{}, false, nil
	}
	if actor.Kind != "user" {
		return Change{}, true, fmt.Errorf("%w: %s requires a human", ErrForbidden, req.Action)
	}
	if c.Phase == Final {
		return Change{}, true, fmt.Errorf("%w: comment %s is %s", ErrConflict, c.ID, c.Status)
	}
	i, proposal, err := findProposal(*c, req.MessageID)
	if err != nil {
		return Change{}, true, err
	}
	decided := now
	proposal.DecidedBy, proposal.DecidedAt = &actor, &decided
	var change Change
	switch req.Action {
	case ActionAcceptProposal:
		err = s.acceptProposal(ctx, actor, c, &proposal, req.Mode, now)
	case ActionRejectProposal:
		proposal.State, proposal.Feedback = ProposalRejected, req.Body
	case ActionRequestRevision:
		if strings.TrimSpace(req.Body) == "" {
			return Change{}, true, fmt.Errorf("%w: asking for a revision needs feedback", ErrInvalid)
		}
		proposal.State, proposal.Feedback = ProposalRevisionRequested, req.Body
		change.Message = &Message{ID: uuid.NewString(), ParentID: req.MessageID, Body: req.Body, Author: actor, CreatedAt: now}
	}
	if err != nil {
		return Change{}, true, err
	}
	m, err := setProposal(c, i, proposal)
	change.EditMessages = []Message{m}
	return change, true, err
}

// acceptProposal applies proposal through the host and resolves c, checking
// the resolve transition first so a thread that cannot resolve is never applied.
func (s *Service) acceptProposal(ctx context.Context, actor Actor, c *Comment, proposal *Proposal, mode string, now time.Time) error {
	verbs := map[string]string{ProposalModeApply: "Applied", ProposalModeStage: "Staged"}
	verb, ok := verbs[mode]
	if !ok {
		return fmt.Errorf("%w: mode must be %s or %s, not %q", ErrInvalid, ProposalModeApply, ProposalModeStage, mode)
	}
	if s.proposals == nil {
		return errNoProposals
	}
	w, ok := s.workflows[c.WorkflowID]
	if !ok {
		return fmt.Errorf("%w: unknown workflow %s", ErrInvalid, c.WorkflowID)
	}
	status, phase, err := w.Apply(c.Status, "resolve", actor)
	if err != nil {
		return err
	}
	commit := proposal.Commit[:min(7, len(proposal.Commit))]
	if err = s.proposals.Apply(ctx, *c, *proposal, mode == ProposalModeStage); err != nil {
		return fmt.Errorf("applying proposal %s: %w", commit, err)
	}
	proposal.State = map[string]ProposalState{ProposalModeApply: ProposalApplied, ProposalModeStage: ProposalStaged}[mode]
	settle(c, status, phase, &Outcome{Kind: "resolve", Summary: verb + " proposal " + commit, Actor: actor, At: now})
	return nil
}

// requireAcceptedProposal refuses an agent's resolve of a thread whose agent
// run must answer with a proposal until one of its proposals is accepted.
func requireAcceptedProposal(c Comment) error {
	var run threadAgentRun
	if _, err := contentValue(c.Content, "agentRun", &run); err != nil {
		return fmt.Errorf("comment %s agentRun: %w", c.ID, err)
	}
	if run.Edits != EditsPropose {
		return nil
	}
	for _, m := range c.Messages {
		p, err := messageProposal(m)
		if err != nil {
			return err
		}
		if p != nil && (p.State == ProposalApplied || p.State == ProposalStaged) {
			return nil
		}
	}
	return fmt.Errorf("%w: this thread requires a proposal; reply with edits or a patch", ErrInvalid)
}

// supersede marks every proposal of c still proposed as superseded.
func supersede(c *Comment) ([]Message, error) {
	var edited []Message
	for i, m := range c.Messages {
		p, err := messageProposal(m)
		if err != nil {
			return nil, err
		}
		if p == nil || p.State != ProposalProposed {
			continue
		}
		p.State = ProposalSuperseded
		m, err = setProposal(c, i, *p)
		if err != nil {
			return nil, err
		}
		edited = append(edited, m)
	}
	return edited, nil
}

// findProposal is the message of c with id and the proposal it holds.
func findProposal(c Comment, id string) (int, Proposal, error) {
	i := slices.IndexFunc(c.Messages, func(m Message) bool { return m.ID == id })
	if id == "" || i < 0 {
		return 0, Proposal{}, fmt.Errorf("%w: reply %q of comment %s", ErrNotFound, id, c.ID)
	}
	p, err := messageProposal(c.Messages[i])
	if err != nil {
		return 0, Proposal{}, err
	}
	if p == nil {
		return 0, Proposal{}, fmt.Errorf("%w: reply %s holds no proposal", ErrInvalid, id)
	}
	if p.State != ProposalProposed {
		return 0, Proposal{}, fmt.Errorf("%w: the proposal of reply %s is already %s", ErrConflict, id, p.State)
	}
	return i, *p, nil
}

// setProposal stores p in message i of c, answering the edited message.
func setProposal(c *Comment, i int, p Proposal) (Message, error) {
	content, err := withContentValue(c.Messages[i].Content, ProposalContentKey, p)
	if err != nil {
		return Message{}, err
	}
	c.Messages[i].Content = content
	return c.Messages[i], nil
}

// messageProposal is the proposal m holds, nil for a reply without one.
func messageProposal(m Message) (*Proposal, error) {
	var p Proposal
	ok, err := contentValue(m.Content, ProposalContentKey, &p)
	if err != nil || !ok {
		return nil, err
	}
	return &p, nil
}

// proposalContent is a reply's content with p merged in; content must be an
// object of the caller's own keys.
func proposalContent(content json.RawMessage, p Proposal) (json.RawMessage, error) {
	if ok, err := contentValue(content, ProposalContentKey, &json.RawMessage{}); ok || err != nil {
		return nil, fmt.Errorf("%w: a reply's content cannot set %q itself", ErrInvalid, ProposalContentKey)
	}
	return withContentValue(content, ProposalContentKey, p)
}

// contentValue decodes key of an object content into out, reporting whether
// content is an object holding key; other content holds no keys.
func contentValue(content json.RawMessage, key string, out any) (bool, error) {
	trimmed := bytes.TrimSpace(content)
	if len(trimmed) == 0 || trimmed[0] != '{' {
		return false, nil
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(trimmed, &fields); err != nil {
		return false, err
	}
	raw, ok := fields[key]
	if !ok || string(raw) == "null" {
		return false, nil
	}
	return true, json.Unmarshal(raw, out)
}

// withContentValue is content with key set to value; content must be empty or an object.
func withContentValue(content json.RawMessage, key string, value any) (json.RawMessage, error) {
	fields := map[string]json.RawMessage{}
	if trimmed := bytes.TrimSpace(content); len(trimmed) > 0 && string(trimmed) != "null" {
		if err := json.Unmarshal(trimmed, &fields); err != nil {
			return nil, fmt.Errorf("%w: content must be a JSON object to hold %q: %v", ErrInvalid, key, err)
		}
	}
	encoded, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	fields[key] = encoded
	return json.Marshal(fields)
}

// ProposalDiff is the patch reply replyID of thread id proposes.
func (s *Service) ProposalDiff(ctx context.Context, id, replyID string) (ProposalDiff, error) {
	if s.proposals == nil {
		return ProposalDiff{}, errNoProposals
	}
	c, err := s.Get(ctx, id)
	if err != nil {
		return ProposalDiff{}, err
	}
	i := slices.IndexFunc(c.Messages, func(m Message) bool { return m.ID == replyID })
	if i < 0 {
		return ProposalDiff{}, fmt.Errorf("%w: reply %q of comment %s", ErrNotFound, replyID, id)
	}
	p, err := messageProposal(c.Messages[i])
	if err != nil {
		return ProposalDiff{}, err
	}
	if p == nil {
		return ProposalDiff{}, fmt.Errorf("%w: reply %s holds no proposal", ErrInvalid, replyID)
	}
	return s.proposals.Diff(ctx, c, *p)
}
