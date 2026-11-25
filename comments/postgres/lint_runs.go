package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"

	"github.com/flanksource/gavel/comments"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func (s *Store) BeginLint(ctx context.Context, run comments.LintRun) (comments.LintRun, bool, error) {
	created := false
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if _, err := sequence(ctx, tx, run.TenantID); err != nil {
			return err
		}
		var row struct{ Payload json.RawMessage }
		q := tx.Table("comment_lint_runs").Where("tenant_id=? AND request_id=?", run.TenantID, run.Request.RequestID).Take(&row)
		if q.Error == nil {
			var previous comments.LintRun
			if err := json.Unmarshal(row.Payload, &previous); err != nil {
				return err
			}
			var a, b any
			left, err := json.Marshal(previous.Request)
			if err != nil {
				return err
			}
			right, err := json.Marshal(run.Request)
			if err != nil {
				return err
			}
			if err = json.Unmarshal(left, &a); err != nil {
				return err
			}
			if err = json.Unmarshal(right, &b); err != nil {
				return err
			}
			if !reflect.DeepEqual(a, b) || previous.Author.Kind != run.Author.Kind || previous.Author.ID != run.Author.ID {
				return comments.ErrConflict
			}
			run = previous
			return nil
		}
		if !errors.Is(q.Error, gorm.ErrRecordNotFound) {
			return q.Error
		}
		data, err := json.Marshal(run)
		if err != nil {
			return err
		}
		if err = tx.Exec("INSERT INTO comment_lint_runs (tenant_id,id,request_id,payload) VALUES (?, ?, ?, ?::jsonb)", run.TenantID, run.ID, run.Request.RequestID, string(data)).Error; err != nil {
			return err
		}
		created = true
		return nil
	})
	return run, created, err
}
func (s *Store) LintRun(ctx context.Context, tenant, id string) (comments.LintRun, error) {
	var row struct{ Payload json.RawMessage }
	err := s.db.WithContext(ctx).Table("comment_lint_runs").Where("tenant_id=? AND id=?", tenant, id).Take(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return comments.LintRun{}, comments.ErrNotFound
	}
	if err != nil {
		return comments.LintRun{}, err
	}
	var run comments.LintRun
	err = json.Unmarshal(row.Payload, &run)
	return run, err
}
func (s *Store) RecordLint(ctx context.Context, tenant, id string, ref comments.ExecutionRef, linter string, findings json.RawMessage) (comments.LintRun, error) {
	var run comments.LintRun
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var row struct{ Payload json.RawMessage }
		err := tx.Table("comment_lint_runs").Clauses(clause.Locking{Strength: "UPDATE"}).Where("tenant_id=? AND id=?", tenant, id).Take(&row).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return comments.ErrNotFound
		}
		if err != nil {
			return err
		}
		if err = json.Unmarshal(row.Payload, &run); err != nil {
			return err
		}
		found := false
		for i, existing := range run.Executions {
			if existing.RunID == ref.RunID {
				found = true
				if existing.SessionID != "" && ref.SessionID != "" && existing.SessionID != ref.SessionID {
					return comments.ErrConflict
				}
				if ref.SessionID != "" {
					run.Executions[i] = ref
				}
			}
		}
		if !found {
			run.Executions = append(run.Executions, ref)
		}
		run.Findings[linter] = findings
		data, err := json.Marshal(run)
		if err != nil {
			return err
		}
		return tx.Exec("UPDATE comment_lint_runs SET payload=?::jsonb WHERE tenant_id=? AND id=?", string(data), tenant, id).Error
	})
	return run, err
}
