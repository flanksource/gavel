package editor

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	captaincli "github.com/flanksource/captain/pkg/cli"
	"github.com/flanksource/captain/pkg/session/deliver"
	"github.com/flanksource/gavel/comments"
	commentapi "github.com/flanksource/gavel/comments/clicky"
	"github.com/flanksource/gavel/internal/database"
)

// AgentSession is a live Claude or Codex session a comment thread can be sent to.
type AgentSession struct {
	ID     string `json:"id"`
	Source string `json:"source"`
	Title  string `json:"title"`
	Model  string `json:"model,omitempty"`
	CWD    string `json:"cwd,omitempty"`
	Status string `json:"status,omitempty"`
	InCmux bool   `json:"inCmux"`
}

// AgentRun is a thread's hand-off to an agent session, kept under "agentRun" in the
// comment's content in clicky-ui's CommentAgentRun shape.
type AgentRun struct {
	SessionID   string         `json:"sessionId"`
	RequestedAt time.Time      `json:"requestedAt"`
	State       string         `json:"state"`
	Error       string         `json:"error,omitempty"`
	Target      AgentRunTarget `json:"target"`
}

// AgentRunTarget is where the thread went, and the channel captain delivered it through.
type AgentRunTarget struct {
	Label   string `json:"label"`
	Channel string `json:"channel,omitempty"`
}

// AgentBridge reaches captain's live sessions; a nil field uses captain in-process.
type AgentBridge struct {
	Live    func(context.Context) ([]captaincli.SessionRecord, error)
	Deliver func(ctx context.Context, sessionID, text string) (deliver.Result, error)
	// Launch starts a managed session from a library prompt, as captaincli.LaunchPrompt.
	Launch func(context.Context, captaincli.PromptInvocation) (captaincli.PromptRunResult, error)
}

// openCaptain opens gavel's process database first, which hands its pool to
// captain: a captain call that ran first would open a second pool on its own,
// and captain then refuses gavel's, failing every later comments request.
func openCaptain(ctx context.Context) error {
	_, err := database.Require(ctx, "Agent sessions")
	return err
}

// captainSessions serves captain's session routes once gavel's pool is captain's.
func captainSessions() http.Handler {
	sessions := http.StripPrefix("/api/captain/sessions", captaincli.SessionHandler())
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := openCaptain(r.Context()); err != nil {
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
			return
		}
		sessions.ServeHTTP(w, r)
	})
}

func (b AgentBridge) live(ctx context.Context) ([]captaincli.SessionRecord, error) {
	if b.Live != nil {
		return b.Live(ctx)
	}
	if err := openCaptain(ctx); err != nil {
		return nil, err
	}
	result, err := captaincli.RunSessionLive(ctx, captaincli.SessionLiveOptions{Source: "all", All: true, Limit: 100})
	return result.Sessions, err
}

func (b AgentBridge) deliver(ctx context.Context, sessionID, text string) (deliver.Result, error) {
	if b.Deliver != nil {
		return b.Deliver(ctx, sessionID, text)
	}
	if err := openCaptain(ctx); err != nil {
		return deliver.Result{}, err
	}
	return captaincli.RunSessionDeliver(ctx, captaincli.SessionDeliverOptions{ID: sessionID, Text: text})
}

func agentSession(record captaincli.SessionRecord) AgentSession {
	session := AgentSession{ID: record.ID, Source: record.Source, Model: record.Model, CWD: record.CWD}
	session.Title = strings.TrimSpace(strings.SplitN(record.InitialPrompt, "\n", 2)[0])
	for _, title := range []string{record.Slug, record.Title} {
		if title != "" {
			session.Title = title
		}
	}
	if session.Title == "" {
		session.Title = record.ID
	}
	if record.Live != nil {
		session.Status = record.Live.Status
		session.InCmux = record.Live.Surface != nil && record.Live.Surface.SurfaceID != ""
		if record.Live.CWD != "" {
			session.CWD = record.Live.CWD
		}
	}
	return session
}

// agentSessions serves GET /api/agent-sessions: the live sessions working on an
// edited file (whose cwd holds one), or every live session with ?all=1.
func (s *Server) agentSessions(w http.ResponseWriter, r *http.Request) {
	records, err := s.Agents.live(r.Context())
	if err != nil {
		http.Error(w, "listing live agent sessions: "+err.Error(), http.StatusBadGateway)
		return
	}
	all := r.URL.Query().Get("all") == "1"
	sessions := []AgentSession{}
	for _, record := range records {
		session := agentSession(record)
		if record.Live != nil && (all || s.coversEditedFile(session.CWD)) {
			sessions = append(sessions, session)
		}
	}
	respond(w, sessions, nil)
}

func (s *Server) coversEditedFile(cwd string) bool {
	if cwd == "" {
		return false
	}
	for _, file := range s.Store.Files() {
		if rel, err := filepath.Rel(cwd, file.Path); err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
			return true
		}
	}
	return false
}

type sendError struct {
	status int
	err    error
}

func (e sendError) Error() string { return e.err.Error() }

// sendComment serves POST /api/file/{id}/comments/{cid}/send {sessionId}: it hands
// the thread to a live session as its agent (the user assigns, the agent
// acknowledges) and delivers the thread into it, recording the hand-off as agentRun.
func (s *Server) sendComment(w http.ResponseWriter, r *http.Request) {
	var input struct {
		SessionID string `json:"sessionId"`
	}
	if !decode(w, r, &input) {
		return
	}
	comment, err := s.sendThread(r.Context(), sendRequest{FileID: r.PathValue("id"), CommentID: r.PathValue("cid"), SessionID: input.SessionID})
	var failed sendError
	if errors.As(err, &failed) {
		http.Error(w, failed.Error(), failed.status)
		return
	}
	respond(w, comment, err)
}

type sendRequest struct{ FileID, CommentID, SessionID string }

func (s *Server) sendThread(ctx context.Context, req sendRequest) (comments.Comment, error) {
	doc, service, comment, err := s.fileComment(ctx, req.FileID, req.CommentID)
	if err != nil {
		return comments.Comment{}, err
	}
	session, err := s.liveSession(ctx, req.SessionID)
	if err != nil {
		return comments.Comment{}, err
	}
	agent := comments.Actor{Kind: "agent", ID: session.Source + ":" + session.ID, Label: session.Model}
	if agent.Label == "" {
		agent.Label = session.Source
	}
	if comment, err = claimForAgent(ctx, service, comment, agent); err != nil {
		return comments.Comment{}, err
	}
	run := AgentRun{SessionID: session.ID, State: "queued", RequestedAt: time.Now().UTC(), Target: AgentRunTarget{Label: session.Title}}
	if comment, err = setAgentRun(ctx, service, comment, run); err != nil {
		return comments.Comment{}, err
	}
	result, deliverErr := s.Agents.deliver(ctx, session.ID, threadMessage(threadMessageOptions{Comment: comment, Path: doc.Path, Agent: agent}))
	if deliverErr != nil {
		run.State, run.Error = "error", deliverErr.Error()
	} else {
		run.State, run.Target.Channel = string(result.Status), string(result.Channel)
	}
	if comment, err = setAgentRun(ctx, service, comment, run); err != nil {
		return comments.Comment{}, err
	}
	if deliverErr != nil {
		return comments.Comment{}, sendError{http.StatusBadGateway, fmt.Errorf("delivering thread %s to session %s: %w", comment.ID, session.ID, deliverErr)}
	}
	return comment, nil
}

func (s *Server) fileComment(ctx context.Context, fileID, commentID string) (Document, *comments.Service, comments.Comment, error) {
	doc, err := s.Store.Read(fileID)
	if err != nil {
		return Document{}, nil, comments.Comment{}, err
	}
	target, err := targetOf(ctx, doc.Path)
	if err != nil {
		return Document{}, nil, comments.Comment{}, err
	}
	service, err := s.OpenComments(ctx)
	if err != nil {
		return Document{}, nil, comments.Comment{}, err
	}
	comment, err := service.Get(ctx, commentID)
	if err != nil || comment.Target.URI != target.URI {
		return Document{}, nil, comments.Comment{}, sendError{http.StatusNotFound, fmt.Errorf("comment %s is not on file %s", commentID, doc.Path)}
	}
	return doc, service, comment, nil
}

func (s *Server) liveSession(ctx context.Context, id string) (AgentSession, error) {
	if id == "" {
		return AgentSession{}, sendError{http.StatusBadRequest, errors.New("sessionId is required")}
	}
	records, err := s.Agents.live(ctx)
	if err != nil {
		return AgentSession{}, sendError{http.StatusBadGateway, fmt.Errorf("listing live agent sessions: %w", err)}
	}
	for _, record := range records {
		if record.ID == id && record.Live != nil {
			return agentSession(record), nil
		}
	}
	return AgentSession{}, sendError{http.StatusConflict, fmt.Errorf("agent session %s is no longer live; choose another", id)}
}

// claimForAgent assigns the thread to agent as the local user, unless it already
// is, then acknowledges it as that agent, so the agent holds it while it works.
func claimForAgent(ctx context.Context, service *comments.Service, comment comments.Comment, agent comments.Actor) (comments.Comment, error) {
	var err error
	if comment.Assignee == nil || comment.Assignee.Kind != agent.Kind || comment.Assignee.ID != agent.ID {
		if comment, err = service.Act(ctx, comments.ActionRequest{ID: comment.ID, ExpectedVersion: comment.Version, Action: "assign", Assignee: &agent}); err != nil {
			return comments.Comment{}, fmt.Errorf("assigning thread %s to %s: %w", comment.ID, agent.ID, err)
		}
	}
	comment, err = service.Act(commentapi.WithActor(ctx, agent), comments.ActionRequest{ID: comment.ID, ExpectedVersion: comment.Version, Action: "acknowledge"})
	if err != nil {
		return comments.Comment{}, fmt.Errorf("acknowledging thread %s as %s: %w", comment.ID, agent.ID, err)
	}
	return comment, nil
}

// setAgentRun records run under "agentRun" in the comment's content, keeping its other keys.
func setAgentRun(ctx context.Context, service *comments.Service, comment comments.Comment, run AgentRun) (comments.Comment, error) {
	content := map[string]json.RawMessage{}
	if len(comment.Content) > 0 && string(comment.Content) != "null" {
		if err := json.Unmarshal(comment.Content, &content); err != nil {
			return comments.Comment{}, fmt.Errorf("comment %s content is not a JSON object: %w", comment.ID, err)
		}
	}
	encoded, err := json.Marshal(run)
	if err != nil {
		return comments.Comment{}, err
	}
	content["agentRun"] = encoded
	merged, err := json.Marshal(content)
	if err != nil {
		return comments.Comment{}, err
	}
	return service.Act(ctx, comments.ActionRequest{ID: comment.ID, ExpectedVersion: comment.Version, Action: "metadata", Content: merged})
}
