package runtime

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/labels"
	"github.com/flanksource/gavel/todos/native"
)

var _ todos.LabelDefinitionProvider = (*Provider)(nil)

// requireWorkspace returns the provider's resolved workspace, or a loud error.
// Label definitions are workspace-scoped, so an unresolved workspace is a bug in
// the caller rather than something to paper over with an empty taxonomy.
func (p *Provider) requireWorkspace(context.Context) (*native.Workspace, error) {
	if p == nil || p.workspace == nil {
		return nil, fmt.Errorf("native TODO provider has no resolved workspace")
	}
	return p.workspace, nil
}

// labelResolver returns the workspace's resolver, loading the whole definition
// set once per provider.
//
// The API opens a provider per request, so this memo is exactly "one definition
// query per request"; a CLI command loads it once for the whole command. The
// alternative — resolving each label as it is rendered — is the N+1 that made
// /api/projects take 46 seconds. Any definition write through this provider
// drops the entry so the next read sees it.
func (p *Provider) labelResolver(ctx context.Context, workspaceID uuid.UUID) (*labels.Resolver, error) {
	// Only the mapping unit tests construct a Provider without a repository —
	// every other method dereferences it, so such a Provider cannot reach
	// production. Definition-free resolution is still correct: built-ins and the
	// hashed palette, exactly what a source with no definition store renders.
	if p.repository == nil {
		return labels.NewResolver(nil, nil), nil
	}

	p.labelsMu.RLock()
	cached, ok := p.labelsCache[workspaceID]
	p.labelsMu.RUnlock()
	if ok {
		return cached, nil
	}

	rows, err := p.repository.ListLabelDefinitions(ctx, workspaceID)
	if err != nil {
		return nil, err
	}

	var workspace, global labels.Definitions
	for _, row := range rows {
		definition := definitionFromNative(row)
		if row.WorkspaceID == nil {
			global = append(global, definition)
			continue
		}
		workspace = append(workspace, definition)
	}
	resolver := labels.NewResolver(workspace, global)

	p.labelsMu.Lock()
	if p.labelsCache == nil {
		p.labelsCache = map[uuid.UUID]*labels.Resolver{}
	}
	p.labelsCache[workspaceID] = resolver
	p.labelsMu.Unlock()

	return resolver, nil
}

// invalidateLabels drops the memo after a definition write.
func (p *Provider) invalidateLabels(workspaceID uuid.UUID) {
	p.labelsMu.Lock()
	delete(p.labelsCache, workspaceID)
	p.labelsMu.Unlock()
}

// LabelDefinitions returns the effective taxonomy for the active workspace:
// stored workspace rows over stored global rows over built-in defaults, each
// carrying the scope it came from.
func (p *Provider) LabelDefinitions(ctx context.Context) (labels.Definitions, error) {
	workspace, err := p.requireWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	resolver, err := p.labelResolver(ctx, workspace.ID)
	if err != nil {
		return nil, err
	}
	return resolver.All(), nil
}

// SetLabelDefinition stores one definition. A global definition applies to every
// workspace; a workspace-scoped one shadows it.
func (p *Provider) SetLabelDefinition(ctx context.Context, definition labels.Definition, global bool) (labels.Definition, error) {
	workspace, err := p.requireWorkspace(ctx)
	if err != nil {
		return labels.Definition{}, err
	}

	input := native.LabelDefinitionInput{
		Name:        definition.Name,
		Color:       string(definition.Color),
		Icon:        definition.Icon,
		Description: definition.Description,
	}
	if !global {
		input.WorkspaceID = &workspace.ID
	}

	row, err := p.repository.SetLabelDefinition(ctx, input)
	if err != nil {
		return labels.Definition{}, err
	}
	p.invalidateLabels(workspace.ID)
	return definitionFromNative(*row), nil
}

// DeleteLabelDefinition retires a label from this workspace: the definition
// goes and the label is stripped from every TODO here, so removing it from the
// project removes it everywhere the project can see it.
//
// A global removal is presentation-only — it spans every workspace, so it drops
// the shared definition and leaves TODO content alone, re-exposing the built-in
// default or the hashed colour.
func (p *Provider) DeleteLabelDefinition(ctx context.Context, name string, global bool) (labels.Removal, error) {
	workspace, err := p.requireWorkspace(ctx)
	if err != nil {
		return labels.Removal{}, err
	}

	var scope *uuid.UUID
	if !global {
		scope = &workspace.ID
	}
	removal, err := p.repository.DeleteLabelDefinition(ctx, scope, name)
	if err != nil {
		return labels.Removal{}, err
	}
	p.invalidateLabels(workspace.ID)
	return removal, nil
}

// LabelCounts reports how many TODOs carry each label, including labels nothing
// defines — the counts behind the dashboard's tag facet.
func (p *Provider) LabelCounts(ctx context.Context) (map[string]int, error) {
	workspace, err := p.requireWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	return p.repository.CountIssuesByLabel(ctx, workspace.ID)
}

func definitionFromNative(row native.LabelDefinition) labels.Definition {
	scope := labels.ScopeGlobal
	if row.WorkspaceID != nil {
		scope = labels.ScopeWorkspace
	}
	return labels.Definition{
		Name:        row.Name,
		Color:       labels.Color(row.Color),
		Icon:        row.Icon,
		Description: row.Description,
		Scope:       scope,
	}
}
