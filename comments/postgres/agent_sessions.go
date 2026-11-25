package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/flanksource/gavel/comments"
	"gorm.io/gorm"
)

// managedSessionRow is a comment_agent_sessions row: one managed session of a
// scope. recorded_at is when the session started; a row without a prompt was
// recorded before sessions kept what they run, and only names its group's session.
type managedSessionRow struct {
	SessionID       string
	ScopeURI        string
	GroupKey        string
	Prompt          *string
	RunOptions      json.RawMessage
	Label           *string
	CWD             *string `gorm:"column:cwd"`
	Source          *string
	Model           *string
	RecordedAt      time.Time
	LastDeliveredAt *time.Time
}

// GroupSession is the agent session last started for a delivery group in scope.
func (s *Store) GroupSession(ctx context.Context, tenant, scope, group string) (string, error) {
	var row struct{ SessionID string }
	err := s.db.WithContext(ctx).Table("comment_agent_sessions").Select("session_id").Where("tenant_id = ? AND scope_uri = ? AND group_key = ?", tenant, scope, group).Order("recorded_at DESC").Take(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return "", comments.ErrNotFound
	}
	return row.SessionID, err
}

// RecordSession records a managed session of its scope as delivered to at
// session.LastDeliveredAt: a new session with what it runs, a known one (a
// resume) only as delivered to again. A session with a GroupKey becomes its
// group's when it is new.
func (s *Store) RecordSession(ctx context.Context, tenant string, session comments.ManagedSession) error {
	return s.db.WithContext(ctx).Exec(`INSERT INTO comment_agent_sessions
		(tenant_id,scope_uri,session_id,group_key,prompt,run_options,label,cwd,source,model,recorded_at,last_delivered_at)
		VALUES (?,?,?,?,?,?::jsonb,?,?,?,?,?,?)
		ON CONFLICT (tenant_id,scope_uri,session_id) DO UPDATE SET last_delivered_at=EXCLUDED.last_delivered_at`,
		tenant, session.ScopeURI, session.ID, session.GroupKey, session.Prompt, nullJSON(session.RunOptions), session.Label, session.CWD,
		session.Source, session.Model, session.StartedAt, session.LastDeliveredAt).Error
}

// TouchSession records a delivery at to session id of scope when it is a managed session.
func (s *Store) TouchSession(ctx context.Context, tenant, scope, id string, at time.Time) error {
	return s.db.WithContext(ctx).Exec("UPDATE comment_agent_sessions SET last_delivered_at=? WHERE tenant_id=? AND scope_uri=? AND session_id=?", at, tenant, scope, id).Error
}

// ManagedSessions are the scope's managed sessions that recorded what they run, newest first.
func (s *Store) ManagedSessions(ctx context.Context, tenant, scope string) ([]comments.ManagedSession, error) {
	var rows []managedSessionRow
	err := s.db.WithContext(ctx).Table("comment_agent_sessions").Where("tenant_id = ? AND scope_uri = ? AND prompt IS NOT NULL", tenant, scope).Order("recorded_at DESC").Find(&rows).Error
	if err != nil {
		return nil, err
	}
	sessions := make([]comments.ManagedSession, 0, len(rows))
	for _, row := range rows {
		sessions = append(sessions, row.session())
	}
	return sessions, nil
}

func (r managedSessionRow) session() comments.ManagedSession {
	text := func(value *string) string {
		if value == nil {
			return ""
		}
		return *value
	}
	session := comments.ManagedSession{
		ID: r.SessionID, ScopeURI: r.ScopeURI, GroupKey: r.GroupKey, Prompt: text(r.Prompt), RunOptions: r.RunOptions,
		Label: text(r.Label), CWD: text(r.CWD), Source: text(r.Source), Model: text(r.Model), StartedAt: r.RecordedAt.UTC(),
	}
	if r.LastDeliveredAt != nil {
		session.LastDeliveredAt = r.LastDeliveredAt.UTC()
	}
	return session
}

func nullJSON(raw json.RawMessage) any {
	if len(raw) == 0 {
		return nil
	}
	return string(raw)
}
