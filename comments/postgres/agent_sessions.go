package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/flanksource/gavel/comments"
	"gorm.io/gorm"
)

// GroupSession is the agent session last recorded for a delivery group in scope.
func (s *Store) GroupSession(ctx context.Context, tenant, scope, group string) (string, error) {
	var row struct{ SessionID string }
	err := s.db.WithContext(ctx).Table("comment_agent_sessions").Select("session_id").Where("tenant_id = ? AND scope_uri = ? AND group_key = ?", tenant, scope, group).Take(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return "", comments.ErrNotFound
	}
	return row.SessionID, err
}

// RecordGroupSession makes session the one a delivery group in scope reuses.
func (s *Store) RecordGroupSession(ctx context.Context, tenant, scope, group, session string) error {
	return s.db.WithContext(ctx).Exec("INSERT INTO comment_agent_sessions (tenant_id,scope_uri,group_key,session_id,recorded_at) VALUES (?,?,?,?,?) ON CONFLICT (tenant_id,scope_uri,group_key) DO UPDATE SET session_id=EXCLUDED.session_id, recorded_at=EXCLUDED.recorded_at", tenant, scope, group, session, time.Now().UTC()).Error
}
