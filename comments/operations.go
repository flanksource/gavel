package comments

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Operation struct {
	ID         string          `json:"id"`
	TenantID   string          `json:"tenantId"`
	CommentID  string          `json:"commentId"`
	Principal  Principal       `json:"principal"`
	Kind       string          `json:"kind"`
	RequestID  string          `json:"requestId"`
	Payload    json.RawMessage `json:"payload"`
	State      string          `json:"state"`
	Lease      string          `json:"lease,omitempty"`
	LeaseUntil time.Time       `json:"leaseUntil,omitempty"`
	Receipt    json.RawMessage `json:"receipt,omitempty"`
	Error      string          `json:"error,omitempty"`
	CreatedAt  time.Time       `json:"createdAt"`
}
type OperationStore interface {
	Queue(context.Context, Operation, int64) (Operation, error)
	Claim(context.Context, string, time.Duration) (Operation, error)
	Finish(context.Context, Operation) (Operation, error)
	Operation(context.Context, string, string) (Operation, error)
	OperationByRequest(context.Context, string, string) (Operation, error)
}
type Executor interface {
	Execute(context.Context, Operation) (json.RawMessage, error)
}
type QueueRequest struct {
	CommentID       string
	ExpectedVersion int64
	Kind            string
	RequestID       string
	Payload         json.RawMessage
}

func (s *Service) Queue(ctx context.Context, req QueueRequest) (Operation, error) {
	store, ok := s.store.(OperationStore)
	if !ok {
		return Operation{}, fmt.Errorf("comment store does not support operations")
	}
	if req.RequestID == "" || req.ExpectedVersion < 1 || !json.Valid(req.Payload) {
		return Operation{}, ErrInvalid
	}
	switch req.Kind {
	case "publish", "publish-reply", "resolve-native", "reopen-native", "deliver-agent":
	default:
		return Operation{}, fmt.Errorf("%w: unknown operation %s", ErrInvalid, req.Kind)
	}
	p, err := s.principal(ctx)
	if err != nil {
		return Operation{}, err
	}
	c, err := s.Get(ctx, req.CommentID)
	if err != nil {
		return Operation{}, err
	}
	if err = s.access.Authorize(ctx, p, req.Kind, c.Target); err != nil {
		return Operation{}, err
	}
	op := Operation{ID: uuid.NewString(), TenantID: p.TenantID, CommentID: c.ID, Principal: p, Kind: req.Kind, RequestID: req.RequestID, Payload: req.Payload, State: "pending", CreatedAt: time.Now().UTC()}
	return store.Queue(ctx, op, req.ExpectedVersion)
}

// Process claims one request. Any unconfirmed delivery is retained as unknown and never automatically retried.
func (s *Service) Process(ctx context.Context, executor Executor, lease time.Duration) (Operation, error) {
	store, ok := s.store.(OperationStore)
	if !ok || executor == nil || lease <= 0 {
		return Operation{}, ErrInvalid
	}
	p, err := s.principal(ctx)
	if err != nil {
		return Operation{}, err
	}
	op, err := store.Claim(ctx, p.TenantID, lease)
	if err != nil {
		return op, err
	}
	c, err := s.store.Get(ctx, op.TenantID, op.CommentID)
	if err == nil {
		err = s.access.Authorize(ctx, op.Principal, op.Kind, c.Target)
	}
	if err != nil {
		op.State = "failed"
		op.Error = err.Error()
		return store.Finish(ctx, op)
	}
	receipt, deliveryErr := executor.Execute(ctx, op)
	if deliveryErr != nil {
		op.State = "unknown"
		op.Error = deliveryErr.Error()
		if json.Valid(receipt) {
			op.Receipt = receipt
		}
	} else if !json.Valid(receipt) {
		op.State = "unknown"
		op.Error = "executor returned an invalid receipt"
	} else {
		op.State = "done"
		op.Receipt = receipt
	}
	// Persistence must outlive a disconnected caller once the external action has completed.
	return store.Finish(context.WithoutCancel(ctx), op)
}
