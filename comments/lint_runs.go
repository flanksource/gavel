package comments

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type LintRequest struct {
	Target        Target          `json:"target"`
	RequestID     string          `json:"requestId"`
	PromptID      string          `json:"promptId"`
	PromptVersion string          `json:"promptVersion"`
	Selection     json.RawMessage `json:"selection"`
}
type ExecutionRef struct {
	RunID     string `json:"runId"`
	SessionID string `json:"sessionId,omitempty"`
}
type LintRun struct {
	ID         string                     `json:"id"`
	TenantID   string                     `json:"tenantId"`
	Request    LintRequest                `json:"request"`
	Author     Actor                      `json:"author"`
	Executions []ExecutionRef             `json:"executions"`
	Findings   map[string]json.RawMessage `json:"findings"`
	CreatedAt  time.Time                  `json:"createdAt"`
}

func (r LintRun) GetID() string   { return r.ID }
func (r LintRun) GetName() string { return r.Request.PromptID }

type LintStore interface {
	BeginLint(context.Context, LintRun) (LintRun, bool, error)
	LintRun(context.Context, string, string) (LintRun, error)
	RecordLint(context.Context, string, string, ExecutionRef, string, json.RawMessage) (LintRun, error)
}

func (s *Service) BeginLint(ctx context.Context, req LintRequest) (LintRun, bool, error) {
	p, err := s.principal(ctx)
	if err != nil {
		return LintRun{}, false, err
	}
	if req.RequestID == "" || req.PromptID == "" || req.PromptVersion == "" || !json.Valid(req.Selection) {
		return LintRun{}, false, ErrInvalid
	}
	if err = s.access.Authorize(ctx, p, "lint", req.Target); err != nil {
		return LintRun{}, false, err
	}
	req.Target, err = s.targets.Resolve(ctx, p, req.Target)
	if err != nil {
		return LintRun{}, false, err
	}
	if req.Target.URI == "" || req.Target.ScopeURI == "" {
		return LintRun{}, false, ErrInvalid
	}
	if err = s.access.Authorize(ctx, p, "lint", req.Target); err != nil {
		return LintRun{}, false, err
	}
	store, ok := s.store.(LintStore)
	if !ok {
		return LintRun{}, false, fmt.Errorf("comment store does not support lint metadata")
	}
	return store.BeginLint(ctx, LintRun{ID: uuid.NewString(), TenantID: p.TenantID, Request: req, Author: p.Actor, Executions: []ExecutionRef{}, Findings: map[string]json.RawMessage{}, CreatedAt: time.Now().UTC()})
}
func (s *Service) LintRun(ctx context.Context, id string) (LintRun, error) {
	p, err := s.principal(ctx)
	if err != nil {
		return LintRun{}, err
	}
	store, ok := s.store.(LintStore)
	if !ok {
		return LintRun{}, ErrInvalid
	}
	run, err := store.LintRun(ctx, p.TenantID, id)
	if err != nil {
		return run, err
	}
	return run, s.access.Authorize(ctx, p, "read", run.Request.Target)
}
func (s *Service) RecordLint(ctx context.Context, id string, ref ExecutionRef, linter string, findings json.RawMessage) (LintRun, error) {
	run, err := s.LintRun(ctx, id)
	if err != nil {
		return run, err
	}
	p, err := s.principal(ctx)
	if err != nil {
		return run, err
	}
	if ref.RunID == "" || linter == "" || !json.Valid(findings) {
		return run, ErrInvalid
	}
	if err = s.access.Authorize(ctx, p, "lint", run.Request.Target); err != nil {
		return run, err
	}
	return s.store.(LintStore).RecordLint(ctx, p.TenantID, id, ref, linter, findings)
}
