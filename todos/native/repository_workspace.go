package native

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type workspaceRecord struct {
	ID          uuid.UUID
	RepoKey     string
	RootPath    string
	DisplayName string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (r workspaceRecord) workspace() *Workspace {
	return &Workspace{
		ID:          r.ID,
		RepoKey:     r.RepoKey,
		RootPath:    r.RootPath,
		DisplayName: r.DisplayName,
		CreatedAt:   r.CreatedAt,
		UpdatedAt:   r.UpdatedAt,
	}
}

const workspaceColumns = `
	workspace.id, COALESCE(workspace.repo_key, '') AS repo_key,
	COALESCE(primary_path.path, '') AS root_path,
	COALESCE(workspace.display_name, '') AS display_name,
	workspace.created_at, workspace.updated_at`

const workspaceFrom = `
	todo_workspaces AS workspace
	LEFT JOIN todo_workspace_paths AS primary_path
	  ON primary_path.workspace_id = workspace.id AND primary_path.is_primary`

func (r *Repository) CreateWorkspace(ctx context.Context, input CreateWorkspaceInput) (*Workspace, error) {
	repoKey := normalizeToken(input.RepoKey)
	rootPath := normalizeWorkspacePath(input.RootPath)
	if repoKey == "" && rootPath == "" {
		return nil, fmt.Errorf("%w: workspace repo key or path is required", ErrInvalidInput)
	}
	id := input.ID
	if id == uuid.Nil {
		id = uuid.New()
	}

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		result := tx.Exec(`
			INSERT INTO todo_workspaces
				(id, repo_key, display_name, created_at, updated_at)
			VALUES (?, NULLIF(?, ''), NULLIF(?, ''), now(), now())`,
			id, repoKey, strings.TrimSpace(input.DisplayName),
		)
		if result.Error != nil {
			return mapUniqueError(result.Error, ErrWorkspaceConflict, "workspace repo key %q", repoKey)
		}
		if rootPath != "" {
			return setWorkspacePrimaryPath(tx, id, rootPath)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return r.GetWorkspace(ctx, id)
}

func (r *Repository) GetWorkspace(ctx context.Context, id uuid.UUID) (*Workspace, error) {
	var record workspaceRecord
	result := r.db.WithContext(ctx).Raw(
		`SELECT `+workspaceColumns+` FROM `+workspaceFrom+` WHERE workspace.id = ?`, id,
	).Scan(&record)
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, fmt.Errorf("%w: workspace %s", ErrNotFound, id)
	}
	return record.workspace(), nil
}

func (r *Repository) GetWorkspaceByRepoKey(ctx context.Context, repoKey string) (*Workspace, error) {
	repoKey = normalizeToken(repoKey)
	if repoKey == "" {
		return nil, fmt.Errorf("%w: workspace repo key is required", ErrInvalidInput)
	}
	var record workspaceRecord
	result := r.db.WithContext(ctx).Raw(
		`SELECT `+workspaceColumns+` FROM `+workspaceFrom+` WHERE workspace.repo_key = ?`, repoKey,
	).Scan(&record)
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, fmt.Errorf("%w: workspace repo key %q", ErrNotFound, repoKey)
	}
	return record.workspace(), nil
}

// GetWorkspaceByPath resolves both the current path and retained locations of
// a moved workspace.
func (r *Repository) GetWorkspaceByPath(ctx context.Context, path string) (*Workspace, error) {
	path = normalizeWorkspacePath(path)
	if path == "" {
		return nil, fmt.Errorf("%w: workspace path is required", ErrInvalidInput)
	}
	var record workspaceRecord
	result := r.db.WithContext(ctx).Raw(
		`SELECT `+workspaceColumns+` FROM `+workspaceFrom+`
		 JOIN todo_workspace_paths AS matched_path ON matched_path.workspace_id = workspace.id
		 WHERE matched_path.path = ?`, path,
	).Scan(&record)
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, fmt.Errorf("%w: workspace path %q", ErrNotFound, path)
	}
	return record.workspace(), nil
}

// ListWorkspacePaths returns the primary path first, followed by retained
// locations. A repo-key-only workspace legitimately returns an empty slice.
func (r *Repository) ListWorkspacePaths(ctx context.Context, workspaceID uuid.UUID) ([]WorkspacePath, error) {
	var paths []WorkspacePath
	result := r.db.WithContext(ctx).Raw(`
		SELECT workspace_id, path, is_primary, created_at, updated_at
		FROM todo_workspace_paths
		WHERE workspace_id = ?
		ORDER BY is_primary DESC, path`, workspaceID,
	).Scan(&paths)
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		var exists bool
		if err := r.db.WithContext(ctx).Raw(
			`SELECT EXISTS(SELECT 1 FROM todo_workspaces WHERE id = ?)`, workspaceID,
		).Scan(&exists).Error; err != nil {
			return nil, err
		}
		if !exists {
			return nil, fmt.Errorf("%w: workspace %s", ErrNotFound, workspaceID)
		}
	}
	if paths == nil {
		paths = []WorkspacePath{}
	}
	return paths, nil
}

func (r *Repository) UpdateWorkspace(ctx context.Context, id uuid.UUID, input UpdateWorkspaceInput) (*Workspace, error) {
	if input.RepoKey == nil && input.RootPath == nil && input.DisplayName == nil {
		return nil, ErrNoChanges
	}
	var repoKey *string
	if input.RepoKey != nil {
		normalized := normalizeToken(*input.RepoKey)
		repoKey = &normalized
	}
	var rootPath *string
	if input.RootPath != nil {
		normalized := normalizeWorkspacePath(*input.RootPath)
		if normalized == "" {
			return nil, fmt.Errorf("%w: workspace path is required", ErrInvalidInput)
		}
		rootPath = &normalized
	}
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var current struct {
			RepoKey sql.NullString
		}
		result := tx.Raw(`SELECT repo_key FROM todo_workspaces WHERE id = ? FOR UPDATE`, id).Scan(&current)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return fmt.Errorf("%w: workspace %s", ErrNotFound, id)
		}

		nextRepoKey := current.RepoKey.String
		if repoKey != nil {
			nextRepoKey = *repoKey
		}
		if nextRepoKey == "" && rootPath == nil {
			var hasPath bool
			if err := tx.Raw(`SELECT EXISTS(SELECT 1 FROM todo_workspace_paths WHERE workspace_id = ?)`, id).Scan(&hasPath).Error; err != nil {
				return err
			}
			if !hasPath {
				return fmt.Errorf("%w: workspace repo key or path is required", ErrInvalidInput)
			}
		}

		updates := map[string]any{"updated_at": gorm.Expr("now()")}
		if repoKey != nil {
			updates["repo_key"] = nullableString(*repoKey)
		}
		if input.DisplayName != nil {
			updates["display_name"] = nullableString(strings.TrimSpace(*input.DisplayName))
		}
		if err := tx.Table("todo_workspaces").Where("id = ?", id).Updates(updates).Error; err != nil {
			return mapUniqueError(err, ErrWorkspaceConflict, "workspace %s", id)
		}
		if rootPath != nil {
			if err := setWorkspacePrimaryPath(tx, id, *rootPath); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return r.GetWorkspace(ctx, id)
}

func (r *Repository) DeleteWorkspace(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Exec(`DELETE FROM todo_workspaces WHERE id = ?`, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("%w: workspace %s", ErrNotFound, id)
	}
	return nil
}

func setWorkspacePrimaryPath(tx *gorm.DB, workspaceID uuid.UUID, path string) error {
	if err := tx.Exec(`
		UPDATE todo_workspace_paths
		SET is_primary = false, updated_at = now()
		WHERE workspace_id = ? AND is_primary`, workspaceID,
	).Error; err != nil {
		return err
	}
	result := tx.Exec(`
		INSERT INTO todo_workspace_paths
			(workspace_id, path, is_primary, created_at, updated_at)
		VALUES (?, ?, true, now(), now())
		ON CONFLICT (workspace_id, path) DO UPDATE
		SET is_primary = true, updated_at = now()`, workspaceID, path,
	)
	if result.Error != nil {
		return mapUniqueError(result.Error, ErrWorkspaceConflict, "workspace path %q", path)
	}
	return nil
}
