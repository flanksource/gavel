package runtime

import (
	"context"
	"fmt"
	"github.com/flanksource/gavel/tenant"
	"github.com/flanksource/gavel/todos/native"
	"github.com/flanksource/gavel/todos/types"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Resolved creates request-local state over a workspace prepared at startup.
func Resolved(ctx context.Context, db *gorm.DB, options WorkspaceOptions, workspace *native.Workspace) (*Provider, error) {
	if workspace == nil {
		return nil, fmt.Errorf("resolved TODO workspace is required")
	}
	global, err := newGlobal(db, tenant.FromContext(ctx).OrgID)
	if err != nil {
		return nil, err
	}
	provider := global.withWorkspace(workspace, options.RootPath)
	provider.readOnly = true
	provider.sandboxedRuns = options.SandboxedRuns
	provider.discoverPlanHooks = !options.SandboxedRuns
	return provider, nil
}

// ReconcileWorkspaces catches up committed external answers outside HTTP reads.
func ReconcileWorkspaces(ctx context.Context, providers []*Provider) error {
	indexed := map[uuid.UUID]*Provider{}
	for _, provider := range providers {
		indexed[provider.workspace.ID] = provider
	}
	if len(indexed) == 0 {
		return nil
	}
	return reconcileWorkspaceAnswers(ctx, indexed)
}

func CountResolvedWorkspaces(ctx context.Context, providers []*Provider) (map[uuid.UUID]map[types.Status]int, error) {
	indexed := map[uuid.UUID]*Provider{}
	for _, provider := range providers {
		indexed[provider.workspace.ID] = provider
	}
	return countWorkspaces(ctx, indexed)
}
