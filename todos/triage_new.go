package todos

import (
	"context"
	"github.com/flanksource/gavel/todos/types"
)

type TriageNewProvider interface {
	ValidateTriageNew(context.Context, *types.TODO, *types.TriageNewEnvelope) error
	ApplyTriageNew(context.Context, *types.TODO, *types.TriageNewEnvelope, TriageNewApplyOptions) error
}

type TriageNewApplyOptions struct {
	TargetVersion int64
}

type TriageNewApplier interface {
	Apply(context.Context, *types.TODO, *types.TriageNewEnvelope) error
}
