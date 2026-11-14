package agents

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	captaincli "github.com/flanksource/captain/pkg/cli"
	"github.com/flanksource/gavel/comments"
)

// SessionStore records the managed session a delivery group reuses (comment_agent_sessions).
type SessionStore interface {
	// GroupSession returns comments.ErrNotFound when no session is recorded.
	GroupSession(ctx context.Context, tenant, scope, group string) (string, error)
	RecordGroupSession(ctx context.Context, tenant, scope, group, session string) error
}

// Options opt a comments host into agent delivery.
type Options struct {
	Bridge Bridge
	Groups SessionStore
	// Workdir is the directory a target's managed sessions run in, and that a
	// scope's session listing is narrowed to.
	Workdir func(context.Context, comments.Target) (string, error)
	// BaseURL is the origin serving the comments API (e.g. http://127.0.0.1:8080)
	// that attachment links in a thread's message are made absolute against;
	// empty leaves them as /api/v1/... paths.
	BaseURL string
}

// Delivery is the comments.Options.Delivery constructor for these options.
func Delivery(options Options) (func(*comments.Service) comments.Delivery, error) {
	if options.Bridge == nil || options.Groups == nil || options.Workdir == nil {
		return nil, errors.New("agent delivery requires a bridge, a session store and a workdir resolver")
	}
	if options.BaseURL != "" {
		if u, err := url.Parse(options.BaseURL); err != nil || u.Scheme == "" || u.Host == "" {
			return nil, fmt.Errorf("agent delivery base URL %q must be an absolute URL", options.BaseURL)
		}
	}
	return func(service *comments.Service) comments.Delivery {
		return delivery{Options: options, service: service}
	}, nil
}

type delivery struct {
	Options
	service *comments.Service
}

// Execute runs one deliver-agent operation; any failure is recorded on the
// thread as agentRun.state "error" as well as returned.
func (d delivery) Execute(ctx context.Context, op comments.Operation) (json.RawMessage, error) {
	var target comments.DeliverTarget
	if err := json.Unmarshal(op.Payload, &target); err != nil {
		return nil, fmt.Errorf("decode delivery target: %w", err)
	}
	comment, err := d.service.Get(ctx, op.CommentID)
	if err != nil {
		return nil, err
	}
	run, err := d.deliver(ctx, comment, target)
	if err != nil {
		if run.RequestedAt.IsZero() {
			run.RequestedAt = time.Now().UTC()
		}
		run.State, run.Error = "error", err.Error()
		_, recordErr := setAgentRun(ctx, d.service, comment.ID, run)
		err = errors.Join(err, recordErr)
	}
	receipt, marshalErr := json.Marshal(run)
	return receipt, errors.Join(err, marshalErr)
}

func (d delivery) RecordFailure(ctx context.Context, commentID string, target comments.DeliverTarget, cause error) error {
	run := AgentRun{SessionID: target.SessionID, RequestedAt: time.Now().UTC(), State: "error", Error: cause.Error(), Target: AgentRunTarget{Label: targetLabel(target)}}
	_, err := setAgentRun(ctx, d.service, commentID, run)
	return err
}

func (d delivery) deliver(ctx context.Context, comment comments.Comment, target comments.DeliverTarget) (AgentRun, error) {
	if target.Kind == "session" {
		return d.toSession(ctx, comment, target.SessionID)
	}
	if target.GroupKey == "" || target.Fresh {
		return d.launch(ctx, comment, target)
	}
	session, err := d.Groups.GroupSession(ctx, comment.TenantID, comment.Target.ScopeURI, target.GroupKey)
	if errors.Is(err, comments.ErrNotFound) {
		return d.launch(ctx, comment, target)
	}
	if err != nil {
		return AgentRun{}, fmt.Errorf("read the session of group %s: %w", target.GroupKey, err)
	}
	live, err := d.live(ctx, session)
	if err != nil {
		return AgentRun{SessionID: session}, err
	}
	if live == nil {
		return d.launch(ctx, comment, target)
	}
	return d.toSession(ctx, comment, session)
}

func (d delivery) Check(target comments.DeliverTarget) error {
	if target.Kind != "run" {
		return nil
	}
	_, err := decodeRunOptions(target.RunOptions)
	return err
}

// live is the live record of session id, or nil when it is no longer live.
func (d delivery) live(ctx context.Context, id string) (*comments.AgentSession, error) {
	records, err := d.Bridge.Live(ctx)
	if err != nil {
		return nil, fmt.Errorf("listing live agent sessions: %w", err)
	}
	for _, record := range records {
		if record.ID == id && record.Live != nil {
			session := agentSession(record)
			return &session, nil
		}
	}
	return nil, nil
}

// Sessions lists live sessions, narrowed to those working in the scope's directory unless All.
func (d delivery) Sessions(ctx context.Context, q comments.SessionQuery) ([]comments.AgentSession, error) {
	records, err := d.Bridge.Live(ctx)
	if err != nil {
		return nil, fmt.Errorf("listing live agent sessions: %w", err)
	}
	dir := ""
	if !q.All && q.ScopeURI != "" {
		if dir, err = d.Workdir(ctx, comments.Target{ScopeURI: q.ScopeURI}); err != nil {
			return nil, err
		}
	}
	sessions := []comments.AgentSession{}
	for _, record := range records {
		session := agentSession(record)
		if record.Live != nil && (dir == "" || overlaps(dir, session.CWD)) {
			sessions = append(sessions, session)
		}
	}
	return sessions, nil
}

// overlaps is whether a session in cwd works in dir: one holds the other.
func overlaps(dir, cwd string) bool {
	if cwd == "" {
		return false
	}
	inside := func(parent, child string) bool {
		rel, err := filepath.Rel(parent, child)
		return err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
	}
	return inside(dir, cwd) || inside(cwd, dir)
}

func agentSession(record captaincli.SessionRecord) comments.AgentSession {
	session := comments.AgentSession{ID: record.ID, Source: record.Source, Model: record.Model, CWD: record.CWD}
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
