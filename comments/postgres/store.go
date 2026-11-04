package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/flanksource/gavel/comments"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Store struct{ db *gorm.DB }

func New(db *gorm.DB) (*Store, error) {
	if db == nil {
		return nil, fmt.Errorf("comments store requires a database")
	}
	return &Store{db: db}, nil
}

var _ comments.Store = (*Store)(nil)

type record struct {
	ID       string
	TenantID string
	Version  int64
	Payload  json.RawMessage
}

func get(ctx context.Context, db *gorm.DB, tenant, id string, lock bool) (comments.Comment, error) {
	q := db.WithContext(ctx).Table("comments").Where("tenant_id = ? AND id = ?", tenant, id)
	if lock {
		q = q.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	var row record
	if err := q.Take(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return comments.Comment{}, comments.ErrNotFound
		}
		return comments.Comment{}, err
	}
	var c comments.Comment
	if err := json.Unmarshal(row.Payload, &c); err != nil {
		return c, fmt.Errorf("decode comment %s: %w", id, err)
	}
	var payloads []struct{ Payload json.RawMessage }
	if err := db.WithContext(ctx).Table("comment_messages").Select("payload").Where("tenant_id = ? AND comment_id = ?", tenant, id).Order("sequence").Find(&payloads).Error; err != nil {
		return c, err
	}
	for _, p := range payloads {
		var m comments.Message
		if err := json.Unmarshal(p.Payload, &m); err != nil {
			return c, err
		}
		c.Messages = append(c.Messages, m)
	}
	var attachments []struct{ Payload json.RawMessage }
	if err := db.WithContext(ctx).Table("comment_attachments").Select("payload").Where("tenant_id=? AND comment_id=?", tenant, id).Order("id").Find(&attachments).Error; err != nil {
		return c, err
	}
	for _, row := range attachments {
		var a comments.Attachment
		if err := json.Unmarshal(row.Payload, &a); err != nil {
			return c, err
		}
		c.Attachments = append(c.Attachments, a)
	}
	return c, nil
}
func (s *Store) Get(ctx context.Context, tenant, id string) (comments.Comment, error) {
	return get(ctx, s.db, tenant, id, false)
}
func sequence(ctx context.Context, tx *gorm.DB, tenant string) (int64, error) {
	if err := tx.WithContext(ctx).Exec("INSERT INTO comment_sequences (tenant_id, sequence) VALUES (?, 0) ON CONFLICT DO NOTHING", tenant).Error; err != nil {
		return 0, err
	}
	var row struct{ Sequence int64 }
	err := tx.WithContext(ctx).Raw("SELECT sequence FROM comment_sequences WHERE tenant_id = ? FOR UPDATE", tenant).Scan(&row).Error
	return row.Sequence + 1, err
}
func appendEvent(ctx context.Context, tx *gorm.DB, tenant string, seq int64, e comments.Event) error {
	e.Sequence = seq
	data, err := json.Marshal(e)
	if err != nil {
		return err
	}
	if err = tx.WithContext(ctx).Exec("INSERT INTO comment_events (tenant_id, sequence, comment_id, payload) VALUES (?, ?, ?, ?::jsonb)", tenant, seq, e.CommentID, string(data)).Error; err != nil {
		return err
	}
	return tx.WithContext(ctx).Exec("UPDATE comment_sequences SET sequence = ? WHERE tenant_id = ?", seq, tenant).Error
}
func (s *Store) Create(ctx context.Context, c comments.Comment, key, hash string, event comments.Event) (comments.Comment, error) {
	var out comments.Comment
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		seq, err := sequence(ctx, tx, c.TenantID)
		if err != nil {
			return err
		}
		var previous struct {
			ID          string
			RequestHash string
		}
		q := tx.Table("comments").Select("id, request_hash").Where("tenant_id = ? AND request_id = ?", c.TenantID, key).Take(&previous)
		if q.Error == nil {
			if previous.RequestHash != hash {
				return comments.ErrConflict
			}
			out, err = get(ctx, tx, c.TenantID, previous.ID, false)
			return err
		}
		if !errors.Is(q.Error, gorm.ErrRecordNotFound) {
			return q.Error
		}
		data, err := json.Marshal(c)
		if err != nil {
			return err
		}
		err = tx.Exec("INSERT INTO comments (id, tenant_id, scope_uri, target_uri, phase, version, request_id, request_hash, payload) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?::jsonb)", c.ID, c.TenantID, c.Target.ScopeURI, c.Target.URI, c.Phase, c.Version, key, hash, string(data)).Error
		if err != nil {
			return err
		}
		if err = appendEvent(ctx, tx, c.TenantID, seq, event); err != nil {
			return err
		}
		out = c
		return nil
	})
	return out, err
}
func (s *Store) Mutate(ctx context.Context, tenant, id string, expected int64, fn func(comments.Comment) (comments.Change, error)) (comments.Comment, error) {
	var out comments.Comment
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		seq, err := sequence(ctx, tx, tenant)
		if err != nil {
			return err
		}
		c, err := get(ctx, tx, tenant, id, true)
		if err != nil {
			return err
		}
		if c.Version != expected {
			return comments.ErrConflict
		}
		var pending int64
		if err = tx.Table("comment_operations").Where("tenant_id = ? AND comment_id = ? AND state IN ('pending','inflight')", tenant, id).Count(&pending).Error; err != nil {
			return err
		}
		if pending > 0 {
			return comments.ErrConflict
		}
		change, err := fn(c)
		if err != nil {
			return err
		}
		out = change.Comment
		if out.ID != id || out.TenantID != tenant || out.Version != expected+1 || (out.Phase == comments.Final) != (out.Outcome != nil) {
			return fmt.Errorf("%w: invalid aggregate mutation", comments.ErrInvalid)
		}
		stored := out
		stored.Messages = nil
		stored.Attachments = nil
		data, err := json.Marshal(stored)
		if err != nil {
			return err
		}
		result := tx.Exec("UPDATE comments SET phase = ?, version = ?, payload = ?::jsonb WHERE tenant_id = ? AND id = ? AND version = ?", out.Phase, out.Version, string(data), tenant, id, expected)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected != 1 {
			return comments.ErrConflict
		}
		if change.Message != nil {
			m := *change.Message
			data, err = json.Marshal(m)
			if err != nil {
				return err
			}
			if err = tx.Exec("INSERT INTO comment_messages (tenant_id, comment_id, id, sequence, payload) SELECT ?, ?, ?, COALESCE(MAX(sequence),0)+1, ?::jsonb FROM comment_messages WHERE tenant_id = ? AND comment_id = ?", tenant, id, m.ID, string(data), tenant, id).Error; err != nil {
				return err
			}
			out.Messages = append(out.Messages, m)
		}
		if change.Attachment != nil {
			data, err := json.Marshal(change.Attachment)
			if err != nil {
				return err
			}
			if err = tx.Exec("INSERT INTO comment_attachments (tenant_id,comment_id,id,payload) VALUES (?, ?, ?, ?::jsonb)", tenant, id, change.Attachment.ID, string(data)).Error; err != nil {
				return err
			}
			out.Attachments = append(out.Attachments, *change.Attachment)
		}
		return appendEvent(ctx, tx, tenant, seq, change.Event)
	})
	return out, err
}
func (s *Store) List(ctx context.Context, tenant string, f comments.Filter) ([]comments.Comment, error) {
	q := s.db.WithContext(ctx).Table("comments").Select("payload").Where("tenant_id = ? AND scope_uri = ?", tenant, f.ScopeURI)
	if f.TargetURI != "" {
		q = q.Where("target_uri = ?", f.TargetURI)
	}
	if f.Phase != "" {
		q = q.Where("phase = ?", f.Phase)
	}
	var rows []struct{ Payload json.RawMessage }
	if err := q.Order("id").Limit(f.Limit).Offset(f.Offset).Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]comments.Comment, 0, len(rows))
	for _, r := range rows {
		var c comments.Comment
		if err := json.Unmarshal(r.Payload, &c); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, nil
}
func (s *Store) Events(ctx context.Context, tenant, scope string, after int64, limit int) ([]comments.Event, error) {
	var rows []struct{ Payload json.RawMessage }
	err := s.db.WithContext(ctx).Table("comment_events e").Joins("JOIN comments c ON c.tenant_id = e.tenant_id AND c.id = e.comment_id").Select("e.payload").Where("e.tenant_id = ? AND c.scope_uri = ? AND e.sequence > ?", tenant, scope, after).Order("e.sequence").Limit(limit).Find(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make([]comments.Event, 0, len(rows))
	for _, r := range rows {
		var e comments.Event
		if err = json.Unmarshal(r.Payload, &e); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, nil
}
