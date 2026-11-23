package agents

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"path/filepath"
	"slices"
	"strings"
	"time"

	captaincli "github.com/flanksource/captain/pkg/cli"
	"github.com/flanksource/captain/pkg/session/amp"
	"github.com/flanksource/gavel/comments"
)

// SessionStore records the managed sessions delivery launches per scope, and
// the one each delivery group reuses (comment_agent_sessions).
type SessionStore interface {
	// GroupSession is the session last started for group; comments.ErrNotFound when none is.
	GroupSession(ctx context.Context, tenant, scope, group string) (string, error)
	// RecordSession records a launched session, or a resumed one as delivered to again.
	RecordSession(ctx context.Context, tenant string, session comments.ManagedSession) error
	// TouchSession records a delivery into session id when it is a managed session of scope.
	TouchSession(ctx context.Context, tenant, scope, id string, at time.Time) error
	// ManagedSessions are the scope's resumable managed sessions, newest first.
	ManagedSessions(ctx context.Context, tenant, scope string) ([]comments.ManagedSession, error)
}

// Options opt a comments host into agent delivery.
type Options struct {
	Bridge       Bridge
	Groups       SessionStore
	Amp          AmpThreads
	Repositories func(ctx context.Context, scope string) ([]string, error)
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
	if options.Amp != nil && options.Repositories == nil {
		return nil, errors.New("Amp thread discovery requires a project repository resolver")
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
// A batch operation delivers all its threads in one message and records the
// same run, or the same failure, on each.
func (d delivery) Execute(ctx context.Context, op comments.Operation) (json.RawMessage, error) {
	var payload comments.DeliverOperation
	if err := json.Unmarshal(op.Payload, &payload); err != nil {
		return nil, fmt.Errorf("decode delivery target: %w", err)
	}
	threads, err := d.threads(ctx, op.CommentID, payload.Threads)
	if err != nil {
		return nil, err
	}
	run, err := AgentRun{}, requireEdits(payload.DeliverTarget)
	if err == nil {
		run, err = d.deliver(ctx, threads, payload.DeliverTarget)
	}
	if err != nil {
		if run.RequestedAt.IsZero() {
			run.RequestedAt = time.Now().UTC()
		}
		run.State, run.Error, run.Edits = "error", err.Error(), payload.Edits
		for _, thread := range threads {
			_, recordErr := setAgentRun(context.WithoutCancel(ctx), d.service, thread.ID, run)
			err = errors.Join(err, recordErr)
		}
	}
	receipt, marshalErr := json.Marshal(run)
	return receipt, errors.Join(err, marshalErr)
}

// threads are the operation's threads: its own, or a batch led by it.
func (d delivery) threads(ctx context.Context, lead string, batch []string) ([]comments.Comment, error) {
	if len(batch) == 0 {
		batch = []string{lead}
	}
	if batch[0] != lead {
		return nil, fmt.Errorf("delivery batch is led by thread %s, not its operation's thread %s", batch[0], lead)
	}
	threads := make([]comments.Comment, 0, len(batch))
	for _, id := range batch {
		thread, err := d.service.Get(ctx, id)
		if err != nil {
			return nil, fmt.Errorf("read thread %s of the delivery: %w", id, err)
		}
		threads = append(threads, thread)
	}
	return threads, nil
}

// RecordFailure records cause as the thread's errored agentRun. The target's
// edits policy is recorded as resolved; it is empty when resolving it is what
// failed, and the failure is recorded all the same.
func (d delivery) RecordFailure(ctx context.Context, commentID string, target comments.DeliverTarget, cause error) error {
	run := AgentRun{SessionID: target.SessionID, RequestedAt: time.Now().UTC(), State: "error", Error: cause.Error(), Target: AgentRunTarget{Label: targetLabel(target)}, Edits: target.Edits}
	_, err := setAgentRun(ctx, d.service, commentID, run)
	return err
}

// deliver hands threads (one scope's, the first leading) to target: a live
// session, an ended managed session resumed, or a new run. A run target with a
// group reuses the group's session, live or resumed, unless Fresh.
func (d delivery) deliver(ctx context.Context, threads []comments.Comment, target comments.DeliverTarget) (AgentRun, error) {
	edits := target.Edits
	if target.Kind == "session" {
		return d.toSession(ctx, threads, target.SessionID, edits)
	}
	target.GroupKey = target.EffectiveGroupKey()
	if target.Fresh {
		return d.launch(ctx, threads, target)
	}
	lead := threads[0]
	session, err := d.Groups.GroupSession(ctx, lead.TenantID, lead.Target.ScopeURI, target.GroupKey)
	if errors.Is(err, comments.ErrNotFound) {
		return d.launch(ctx, threads, target)
	}
	if err != nil {
		return AgentRun{Edits: edits}, fmt.Errorf("read the session of group %s: %w", target.GroupKey, err)
	}
	live, err := d.live(ctx, session)
	if err != nil {
		return AgentRun{SessionID: session, Edits: edits}, err
	}
	if live != nil {
		return d.toLive(ctx, threads, *live, edits)
	}
	managed, err := d.managed(ctx, lead, session)
	if err != nil {
		return AgentRun{SessionID: session, Edits: edits}, err
	}
	if managed == nil {
		return d.launch(ctx, threads, target)
	}
	return d.resume(ctx, threads, *managed, edits)
}

func (d delivery) NormalizeTarget(target comments.DeliverTarget) (comments.DeliverTarget, error) {
	if target.Kind == "run" {
		_, err := decodeRunOptions(target.RunOptions)
		return target, err
	}
	ref := strings.TrimSpace(target.SessionID)
	if strings.HasPrefix(ref, "T-") || strings.Contains(ref, "://") {
		id, err := amp.ParseThreadRef(ref)
		if err != nil {
			return target, fmt.Errorf("%w: %v", comments.ErrInvalid, err)
		}
		target.SessionID = id
	}
	return target, nil
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

// managed is the managed session id of the thread's scope, or nil when gavel did not launch it there.
func (d delivery) managed(ctx context.Context, thread comments.Comment, id string) (*comments.ManagedSession, error) {
	sessions, err := d.Groups.ManagedSessions(ctx, thread.TenantID, thread.Target.ScopeURI)
	if err != nil {
		return nil, fmt.Errorf("listing the managed sessions of %s: %w", thread.Target.ScopeURI, err)
	}
	for _, session := range sessions {
		if session.ID == id {
			return &session, nil
		}
	}
	return nil, nil
}

// Sessions lists live sessions, narrowed to those working in the scope's
// directory unless All, and with a scope its managed sessions: a live one
// marked managed, an ended one as a resumable "ended" session.
func (d delivery) Sessions(ctx context.Context, q comments.SessionQuery) ([]comments.AgentSession, error) {
	if strings.HasPrefix(q.SessionID, "T-") {
		session, err := d.resolveAmp(ctx, q.SessionID)
		if err != nil {
			return nil, fmt.Errorf("%w: %v", comments.ErrInvalid, err)
		}
		return []comments.AgentSession{session}, nil
	}
	dir := ""
	if !q.All && q.ScopeURI != "" {
		var err error
		if dir, err = d.Workdir(ctx, comments.Target{ScopeURI: q.ScopeURI}); errors.Is(err, comments.ErrNoWorkdir) && d.Amp != nil {
			return d.ampSessions(ctx, q), nil
		} else if err != nil {
			return nil, err
		}
	}
	records, err := d.Bridge.Live(ctx)
	if err != nil {
		return nil, fmt.Errorf("listing live agent sessions: %w", err)
	}
	return d.sessions(ctx, q, dir, records)
}

// sessions are q's sessions among records, narrowed to those working in dir
// unless it is empty, as Sessions lists them.
func (d delivery) sessions(ctx context.Context, q comments.SessionQuery, dir string, records []captaincli.SessionRecord) ([]comments.AgentSession, error) {
	selected := ""
	if q.GroupKey != "" {
		var err error
		selected, err = d.Groups.GroupSession(ctx, q.TenantID, q.ScopeURI, q.GroupKey)
		if errors.Is(err, comments.ErrNotFound) {
			return []comments.AgentSession{}, nil
		}
		if err != nil {
			return nil, fmt.Errorf("looking up comment group %s: %w", q.GroupKey, err)
		}
	}
	var err error
	var managed []comments.ManagedSession
	if q.ScopeURI != "" {
		if managed, err = d.Groups.ManagedSessions(ctx, q.TenantID, q.ScopeURI); err != nil {
			return nil, fmt.Errorf("listing the managed sessions of %s: %w", q.ScopeURI, err)
		}
	}
	managedByID := map[string]comments.ManagedSession{}
	for _, session := range managed {
		managedByID[session.ID] = session
	}
	sessions, live := []comments.AgentSession{}, map[string]bool{}
	for _, record := range records {
		session := agentSession(record)
		m, isManaged := managedByID[record.ID]
		if record.Live != nil && (dir == "" || isManaged || overlaps(dir, session.CWD)) {
			session.Managed, live[record.ID] = isManaged, true
			if m.Label != "" {
				session.Title = m.Label
			}
			sessions = append(sessions, session)
		}
	}
	for _, session := range managed {
		if !live[session.ID] {
			sessions = append(sessions, endedSession(session))
		}
	}
	if selected != "" {
		sessions = slices.DeleteFunc(sessions, func(session comments.AgentSession) bool { return session.ID != selected })
	}
	return append(sessions, d.ampSessions(ctx, q)...), nil
}

// endedSession lists a managed session that is no longer live, for a delivery to resume.
func endedSession(m comments.ManagedSession) comments.AgentSession {
	title := m.Label
	if title == "" {
		title = m.ID
	}
	return comments.AgentSession{ID: m.ID, Source: m.Source, Title: title, Model: m.Model, CWD: m.CWD, Status: "ended", Managed: true, Resumable: true}
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
	if c := record.Context; c != nil && c.WindowTokens > 0 {
		session.Context = &comments.AgentSessionContext{UsedTokens: c.UsedTokens, WindowTokens: c.WindowTokens}
	}
	session.Title = firstLine(record.InitialPrompt)
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
