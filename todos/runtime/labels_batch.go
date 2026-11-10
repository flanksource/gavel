package runtime

import (
	"context"
	"errors"

	"github.com/flanksource/gavel/internal/database"
	"github.com/flanksource/gavel/tenant"
	"github.com/flanksource/gavel/todos/labels"
	"github.com/flanksource/gavel/todos/native"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type LabelIndex struct {
	Definitions labels.Definitions `json:"definitions"`
	Counts      map[string]int     `json:"counts"`
	Err         error              `json:"-"`
}

// ReadLabelIndexes reads taxonomy and visible usage without initializing workspaces.
func ReadLabelIndexes(ctx context.Context, db *gorm.DB, projects []WorkspaceOptions) ([]LabelIndex, error) {
	if db == nil {
		var err error
		db, err = database.Require(ctx, "native TODO labels")
		if err != nil {
			return nil, err
		}
	}
	repository, err := native.NewRepository(db)
	if err != nil {
		return nil, err
	}
	repository = repository.ForOrg(tenant.FromContext(ctx).OrgID)
	result := make([]LabelIndex, len(projects))
	normalized := make([]WorkspaceOptions, len(projects))
	var paths, keys []string
	for i, project := range projects {
		normalized[i], result[i].Err = normalizeWorkspaceOptions(project)
		if result[i].Err == nil {
			paths = append(paths, normalized[i].RootPath)
			keys = append(keys, normalized[i].Repositories...)
		}
	}
	lookup, err := repository.LookupWorkspaces(ctx, paths, keys)
	if err != nil {
		return nil, err
	}
	ids := make([]uuid.UUID, len(projects))
	var found []uuid.UUID
	for i, project := range normalized {
		if result[i].Err != nil {
			continue
		}
		workspace, _, err := resolveWorkspace(ctx, lookup, project)
		if errors.Is(err, native.ErrNotFound) {
			continue
		}
		if err != nil {
			result[i].Err = err
			continue
		}
		ids[i] = workspace.ID
		found = append(found, workspace.ID)
	}
	definitions, err := repository.LabelDefinitionsForWorkspaces(ctx, found)
	if err != nil {
		return nil, err
	}
	counts, err := repository.LabelCountsForWorkspaces(ctx, found)
	if err != nil {
		return nil, err
	}
	global := labels.Definitions{}
	byWorkspace := map[uuid.UUID]labels.Definitions{}
	for _, row := range definitions {
		if row.WorkspaceID == nil {
			global = append(global, definitionFromNative(row))
		} else {
			byWorkspace[*row.WorkspaceID] = append(byWorkspace[*row.WorkspaceID], definitionFromNative(row))
		}
	}
	for i, id := range ids {
		if result[i].Err != nil {
			continue
		}
		result[i].Definitions = labels.NewResolver(byWorkspace[id], global).All()
		result[i].Counts = counts[id]
		if result[i].Counts == nil {
			result[i].Counts = map[string]int{}
		}
	}
	return result, nil
}
