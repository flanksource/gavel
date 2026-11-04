package postgres

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"

	"github.com/flanksource/gavel/comments"
	"gorm.io/gorm"
)

func (s *Store) MarkRead(ctx context.Context, p comments.Principal, m comments.ReadMarker) (comments.ReadMarker, error) {
	err := s.db.WithContext(ctx).Exec("INSERT INTO comment_reads (tenant_id,actor_kind,actor_id,comment_id,version,read_at) VALUES (?, ?, ?, ?, ?, ?) ON CONFLICT (tenant_id,actor_kind,actor_id,comment_id) DO UPDATE SET version=EXCLUDED.version, read_at=EXCLUDED.read_at WHERE comment_reads.version < EXCLUDED.version", p.TenantID, p.Actor.Kind, p.Actor.ID, m.CommentID, m.Version, m.At).Error
	if err != nil {
		return m, err
	}
	rows, err := s.ReadMarkers(ctx, p, []string{m.CommentID})
	if err != nil {
		return m, err
	}
	if len(rows) != 1 {
		return m, comments.ErrNotFound
	}
	return rows[0], nil
}
func (s *Store) ReadMarkers(ctx context.Context, p comments.Principal, ids []string) ([]comments.ReadMarker, error) {
	out := make([]comments.ReadMarker, 0)
	if len(ids) == 0 {
		return out, nil
	}
	err := s.db.WithContext(ctx).Table("comment_reads").Select("comment_id, version, read_at AS at").Where("tenant_id = ? AND actor_kind = ? AND actor_id = ? AND comment_id IN ?", p.TenantID, p.Actor.Kind, p.Actor.ID, ids).Order("comment_id").Scan(&out).Error
	return out, err
}
func (s *Store) Page(ctx context.Context, p comments.Principal, f comments.Filter) (comments.Page, error) {
	page := comments.Page{Items: []comments.Comment{}, Read: []comments.ReadMarker{}, Limit: f.Limit, Offset: f.Offset}
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		q := tx.Table("comments c").Joins("LEFT JOIN comment_reads r ON r.tenant_id=c.tenant_id AND r.comment_id=c.id AND r.actor_kind=? AND r.actor_id=?", p.Actor.Kind, p.Actor.ID).Where("c.tenant_id=? AND c.scope_uri=?", p.TenantID, f.ScopeURI)
		if f.TargetURI != "" {
			q = q.Where("c.target_uri=?", f.TargetURI)
		}
		if f.Phase != "" {
			q = q.Where("c.phase=?", f.Phase)
		}
		if f.Unread {
			q = q.Where("COALESCE(r.version,0) < c.version")
		}
		if err := q.Count(&page.Total).Error; err != nil {
			return err
		}
		var rows []struct{ Payload json.RawMessage }
		if err := q.Select("c.payload").Order("c.id").Limit(f.Limit).Offset(f.Offset).Find(&rows).Error; err != nil {
			return err
		}
		ids := make([]string, 0, len(rows))
		for _, row := range rows {
			var c comments.Comment
			if err := json.Unmarshal(row.Payload, &c); err != nil {
				return err
			}
			page.Items = append(page.Items, c)
			ids = append(ids, c.ID)
		}
		var err error
		page.Read, err = (&Store{db: tx}).ReadMarkers(ctx, p, ids)
		return err
	})
	return page, err
}
func (s *Store) Put(ctx context.Context, tenant, id string, data []byte) error {
	return s.db.WithContext(ctx).Exec("INSERT INTO comment_blobs (tenant_id,id,data) VALUES (?, ?, ?)", tenant, id, data).Error
}
func (s *Store) Open(ctx context.Context, tenant, id string) (io.ReadCloser, error) {
	var row struct{ Data []byte }
	err := s.db.WithContext(ctx).Table("comment_blobs").Where("tenant_id=? AND id=?", tenant, id).Take(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, comments.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return io.NopCloser(bytes.NewReader(row.Data)), nil
}
func (s *Store) Delete(ctx context.Context, tenant, id string) error {
	return s.db.WithContext(ctx).Exec("DELETE FROM comment_blobs WHERE tenant_id=? AND id=?", tenant, id).Error
}
