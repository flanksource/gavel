package todos

import (
	"context"
	"fmt"

	"github.com/flanksource/gavel/todos/types"
)

// Transfer moves a native issue between PostgreSQL workspaces while preserving
// identity and history. Provider implementations that cannot perform that
// atomic native move are rejected; copy/delete fallback is intentionally gone.
func Transfer(ctx context.Context, source, target Provider, ref string) (*types.TODO, error) {
	if source == nil || target == nil {
		return nil, fmt.Errorf("source and target providers are required")
	}
	todo, err := source.Get(ctx, ref)
	if err != nil {
		return nil, fmt.Errorf("load source todo %q: %w", ref, err)
	}
	mover, ok := source.(TransferProvider)
	if !ok {
		return nil, fmt.Errorf("source provider does not support native PostgreSQL workspace transfer")
	}
	return mover.MoveTo(ctx, todo, target)
}
