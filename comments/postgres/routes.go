package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/flanksource/gavel/comments"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

var _ comments.RouteStore = (*Store)(nil)

func (s *Store) Routes(ctx context.Context, tenant, scope string) ([]comments.Route, error) {
	var rows []struct{ Payload json.RawMessage }
	if err := s.db.WithContext(ctx).Table("comment_routes").Select("payload").Where("tenant_id = ? AND scope_uri = ?", tenant, scope).Order("target_prefix, id").Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]comments.Route, 0, len(rows))
	for _, row := range rows {
		var r comments.Route
		if err := json.Unmarshal(row.Payload, &r); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, nil
}

func (s *Store) Route(ctx context.Context, tenant, id string) (comments.Route, error) {
	var row struct{ Payload json.RawMessage }
	err := s.db.WithContext(ctx).Table("comment_routes").Select("payload").Where("tenant_id = ? AND id = ?", tenant, id).Take(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return comments.Route{}, comments.ErrNotFound
	}
	if err != nil {
		return comments.Route{}, err
	}
	var r comments.Route
	return r, json.Unmarshal(row.Payload, &r)
}

// SaveRoute upserts r by id; another route on the same prefix conflicts.
func (s *Store) SaveRoute(ctx context.Context, r comments.Route) (comments.Route, error) {
	data, err := json.Marshal(r)
	if err != nil {
		return comments.Route{}, err
	}
	err = s.db.WithContext(ctx).Exec("INSERT INTO comment_routes (tenant_id,id,scope_uri,target_prefix,payload) VALUES (?,?,?,?,?::jsonb) ON CONFLICT (tenant_id,id) DO UPDATE SET scope_uri=EXCLUDED.scope_uri, target_prefix=EXCLUDED.target_prefix, payload=EXCLUDED.payload", r.TenantID, r.ID, r.ScopeURI, r.TargetPrefix, string(data)).Error
	var conflict *pgconn.PgError
	if errors.As(err, &conflict) && conflict.ConstraintName == "comment_routes_prefix" {
		return comments.Route{}, comments.ErrConflict
	}
	if err != nil {
		return comments.Route{}, err
	}
	return r, nil
}

// ReplaceRoutes makes routes the scope's whole route set in one transaction:
// the projection a host whose routes are configured elsewhere keeps here.
func (s *Store) ReplaceRoutes(ctx context.Context, tenant, scope string, routes []comments.Route) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("DELETE FROM comment_routes WHERE tenant_id = ? AND scope_uri = ?", tenant, scope).Error; err != nil {
			return err
		}
		for _, r := range routes {
			if r.TenantID != tenant || r.ScopeURI != scope {
				return fmt.Errorf("route %s belongs to %s/%s, not %s/%s", r.ID, r.TenantID, r.ScopeURI, tenant, scope)
			}
			data, err := json.Marshal(r)
			if err != nil {
				return err
			}
			if err := tx.Exec("INSERT INTO comment_routes (tenant_id,id,scope_uri,target_prefix,payload) VALUES (?,?,?,?,?::jsonb)", r.TenantID, r.ID, r.ScopeURI, r.TargetPrefix, string(data)).Error; err != nil {
				return fmt.Errorf("store route %s on %s: %w", r.ID, r.TargetPrefix, err)
			}
		}
		return nil
	})
}

func (s *Store) DeleteRoute(ctx context.Context, tenant, id string) error {
	q := s.db.WithContext(ctx).Exec("DELETE FROM comment_routes WHERE tenant_id = ? AND id = ?", tenant, id)
	if q.Error == nil && q.RowsAffected == 0 {
		return comments.ErrNotFound
	}
	return q.Error
}

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
