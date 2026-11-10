package runtime

import (
	"context"

	"github.com/flanksource/gavel/todos/githubpush"
	"github.com/flanksource/gavel/todos/labels"
	"github.com/flanksource/gavel/todos/native"
	"github.com/flanksource/gavel/todos/types"
	"github.com/google/uuid"
)

func prepareWorkspaceLists(ctx context.Context, providers []*Provider, ids []uuid.UUID, issues []native.Issue) (map[uuid.UUID]error, error) {
	repository := providers[0].repository
	aliases, err := repository.ListAliasesByKind(ctx, ids, githubpush.AliasKind)
	if err != nil {
		return nil, err
	}
	definitions, err := repository.LabelDefinitionsForWorkspaces(ctx, ids)
	if err != nil {
		return nil, err
	}
	rows, err := repository.ListIssuePhaseRuns(ctx, ids...)
	if err != nil {
		return nil, err
	}
	workspaceByIssue := map[uuid.UUID]uuid.UUID{}
	for _, issue := range issues {
		workspaceByIssue[issue.ID] = issue.WorkspaceID
	}
	phases := map[uuid.UUID]types.PhaseRuns{}
	failures := map[uuid.UUID]error{}
	for _, row := range rows {
		run, err := phaseRunFromNative(row)
		if err != nil {
			failures[workspaceByIssue[row.IssueID]] = err
			continue
		}
		if phases[row.IssueID] == nil {
			phases[row.IssueID] = types.PhaseRuns{}
		}
		phases[row.IssueID][types.Phase(row.Phase)] = run
	}
	for _, provider := range providers {
		var local, global labels.Definitions
		for _, row := range definitions {
			if row.WorkspaceID == nil {
				global = append(global, definitionFromNative(row))
			} else if *row.WorkspaceID == provider.workspace.ID {
				local = append(local, definitionFromNative(row))
			}
		}
		provider.listAliases = aliases
		provider.labelsCache = map[uuid.UUID]*labels.Resolver{provider.workspace.ID: labels.NewResolver(local, global)}
		provider.phasesCache = map[uuid.UUID]map[uuid.UUID]types.PhaseRuns{provider.workspace.ID: phases}
	}
	return failures, nil
}
