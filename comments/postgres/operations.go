package postgres

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/flanksource/gavel/comments"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var _ comments.OperationStore = (*Store)(nil)

func (s *Store) OperationByRequest(ctx context.Context, tenant, requestID string) (comments.Operation, error) {
	var row struct{ ID string }
	err := s.db.WithContext(ctx).Table("comment_operations").Select("id").Where("tenant_id=? AND request_id=?", tenant, requestID).Take(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return comments.Operation{}, comments.ErrNotFound
	}
	if err != nil {
		return comments.Operation{}, err
	}
	return s.Operation(ctx, tenant, row.ID)
}

func (s *Store) Queue(ctx context.Context, op comments.Operation, expected int64) (comments.Operation, error) {
	var out comments.Operation
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		seq, err := sequence(ctx, tx, op.TenantID)
		if err != nil {
			return err
		}
		var previous struct{ Payload json.RawMessage }
		q := tx.Table("comment_operations").Select("payload").Where("tenant_id = ? AND request_id = ?", op.TenantID, op.RequestID).Take(&previous)
		if q.Error == nil {
			if err = json.Unmarshal(previous.Payload, &out); err != nil {
				return err
			}
			if out.CommentID != op.CommentID || out.Kind != op.Kind || out.Principal.Actor.ID != op.Principal.Actor.ID || out.Principal.Actor.Kind != op.Principal.Actor.Kind {
				return comments.ErrConflict
			}
			var a, b any
			leftDecoder := json.NewDecoder(bytes.NewReader(out.Payload))
			leftDecoder.UseNumber()
			if err = leftDecoder.Decode(&a); err != nil {
				return err
			}
			rightDecoder := json.NewDecoder(bytes.NewReader(op.Payload))
			rightDecoder.UseNumber()
			if err = rightDecoder.Decode(&b); err != nil {
				return err
			}
			left, err := json.Marshal(a)
			if err != nil {
				return err
			}
			right, err := json.Marshal(b)
			if err != nil {
				return err
			}
			if string(left) != string(right) {
				return comments.ErrConflict
			}
			out, err = (&Store{db: tx}).Operation(ctx, op.TenantID, out.ID)
			return err
		}
		if !errors.Is(q.Error, gorm.ErrRecordNotFound) {
			return q.Error
		}
		c, err := get(ctx, tx, op.TenantID, op.CommentID, true)
		if err != nil {
			return err
		}
		if c.Version != expected {
			return comments.ErrConflict
		}
		data, err := json.Marshal(op)
		if err != nil {
			return err
		}
		if err = tx.Exec("INSERT INTO comment_operations (id,tenant_id,comment_id,request_id,state,payload) VALUES (?, ?, ?, ?, ?, ?::jsonb)", op.ID, op.TenantID, op.CommentID, op.RequestID, op.State, string(data)).Error; err != nil {
			var conflict *pgconn.PgError
			if errors.As(err, &conflict) && conflict.ConstraintName == "comment_operations_active" {
				return comments.ErrConflict
			}
			return err
		}
		out = op
		return appendEvent(ctx, tx, op.TenantID, seq, comments.Event{CommentID: c.ID, Kind: "queued:" + op.Kind, Actor: op.Principal.Actor, At: op.CreatedAt, Data: data})
	})
	return out, err
}
func (s *Store) Operation(ctx context.Context, tenant, id string) (comments.Operation, error) {
	var row struct {
		Payload    json.RawMessage
		State      string
		Lease      string
		LeaseUntil *time.Time
	}
	err := s.db.WithContext(ctx).Table("comment_operations").Where("tenant_id = ? AND id = ?", tenant, id).Take(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return comments.Operation{}, comments.ErrNotFound
	}
	if err != nil {
		return comments.Operation{}, err
	}
	var op comments.Operation
	if err = json.Unmarshal(row.Payload, &op); err != nil {
		return op, err
	}
	op.State = row.State
	op.Lease = row.Lease
	if row.LeaseUntil != nil {
		op.LeaseUntil = *row.LeaseUntil
	}
	return op, nil
}
func (s *Store) Claim(ctx context.Context, req comments.ClaimRequest) (comments.Operation, error) {
	var op comments.Operation
	lease := req.Lease
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("UPDATE comment_operations SET state = 'unknown' WHERE tenant_id = ? AND state = 'inflight' AND lease_until < now()", req.TenantID).Error; err != nil {
			return err
		}
		var row struct {
			ID       string
			TenantID string
		}
		pending := tx.Table("comment_operations").Select("id, tenant_id").Where("tenant_id = ? AND state = 'pending'", req.TenantID)
		if req.OperationID != "" {
			pending = pending.Where("id = ?", req.OperationID)
		}
		if len(req.Kinds) > 0 {
			pending = pending.Where("payload->>'kind' IN ?", req.Kinds)
		}
		q := pending.Order("id").Clauses(clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"}).Limit(1).Scan(&row)
		if q.Error != nil {
			return q.Error
		}
		if row.ID == "" {
			return nil
		}
		store := &Store{db: tx}
		var err error
		op, err = store.Operation(ctx, row.TenantID, row.ID)
		if err != nil {
			return err
		}
		op.State = "inflight"
		op.Lease = uuid.NewString()
		op.LeaseUntil = time.Now().UTC().Add(lease)
		return tx.Exec("UPDATE comment_operations SET state = 'inflight', lease = ?, lease_until = ? WHERE tenant_id = ? AND id = ?", op.Lease, op.LeaseUntil, op.TenantID, op.ID).Error
	})
	if err == nil && op.ID == "" {
		return op, comments.ErrNotFound
	}
	return op, err
}
func (s *Store) Finish(ctx context.Context, op comments.Operation) (comments.Operation, error) {
	if op.State != "done" && op.State != "unknown" && op.State != "failed" {
		return op, comments.ErrInvalid
	}
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		seq, err := sequence(ctx, tx, op.TenantID)
		if err != nil {
			return err
		}
		data, err := json.Marshal(op)
		if err != nil {
			return err
		}
		q := tx.Exec("UPDATE comment_operations SET state = ?, payload = ?::jsonb WHERE tenant_id = ? AND id = ? AND state = 'inflight' AND lease = ? AND lease_until >= now()", op.State, string(data), op.TenantID, op.ID, op.Lease)
		if q.Error != nil {
			return q.Error
		}
		if q.RowsAffected != 1 {
			return comments.ErrConflict
		}
		return appendEvent(ctx, tx, op.TenantID, seq, comments.Event{CommentID: op.CommentID, Kind: "delivery:" + op.State, Actor: op.Principal.Actor, At: time.Now().UTC(), Data: data})
	})
	return op, err
}
