package native

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

const MinShortReferenceLength = 8

var (
	ErrNotFound              = errors.New("native todo record not found")
	ErrVersionConflict       = errors.New("native todo version conflict")
	ErrWorkspaceConflict     = errors.New("native todo workspace already exists")
	ErrInvalidInput          = errors.New("invalid native todo input")
	ErrNoChanges             = errors.New("native todo mutation has no changes")
	ErrAliasConflict         = errors.New("native todo alias already exists")
	ErrAmbiguousReference    = errors.New("native todo reference is ambiguous")
	ErrSelfRelationship      = errors.New("native todo self relationship")
	ErrCrossWorkspace        = errors.New("native todo relationship crosses workspaces")
	ErrRelationshipExists    = errors.New("native todo relationship already exists")
	ErrRelationshipCycle     = errors.New("native todo dependency cycle")
	ErrRelationshipNotFound  = errors.New("native todo relationship not found")
	ErrIssueHasRelationships = errors.New("native todo issue has relationships")
	ErrLinkConflict          = errors.New("native todo Captain link conflict")
	ErrEventConflict         = errors.New("native todo event source already exists")
)

// Repository persists native TODO state. It deliberately accepts an injected
// GORM pool and does not migrate or auto-create tables.
type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) (*Repository, error) {
	if db == nil {
		return nil, fmt.Errorf("%w: database is nil", ErrInvalidInput)
	}
	return &Repository{db: db}, nil
}

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

type issueRecord struct {
	ID                uuid.UUID
	WorkspaceID       uuid.UUID
	Title             string
	Body              string
	Verification      string
	Labels            pq.StringArray `gorm:"type:text[]"`
	Priority          string
	Status            string
	ExecutionState    string
	ActivePromptRunID *uuid.UUID
	SelectedPlanID    *uuid.UUID
	Version           int64
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

func (r issueRecord) issue() *Issue {
	labels := append([]string(nil), r.Labels...)
	if labels == nil {
		labels = []string{}
	}
	return &Issue{
		ID:                r.ID,
		WorkspaceID:       r.WorkspaceID,
		Title:             r.Title,
		Body:              r.Body,
		Verification:      r.Verification,
		Labels:            labels,
		Priority:          Priority(r.Priority),
		Status:            IssueStatus(r.Status),
		ExecutionState:    ExecutionState(r.ExecutionState),
		ActivePromptRunID: r.ActivePromptRunID,
		SelectedPlanID:    r.SelectedPlanID,
		Version:           r.Version,
		CreatedAt:         r.CreatedAt,
		UpdatedAt:         r.UpdatedAt,
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

const issueColumns = `
	issue.id, issue.workspace_id, issue.title, issue.body, issue.verification,
	issue.labels, issue.priority, issue.status,
	COALESCE(runtime.execution_state, 'idle') AS execution_state,
	issue.active_prompt_run_id, issue.selected_plan_id, issue.version,
	issue.created_at, issue.updated_at`

const issueFrom = `
	todo_issues AS issue
	LEFT JOIN todo_issue_runtime AS runtime ON runtime.issue_id = issue.id`

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

func (r *Repository) CreateIssue(ctx context.Context, input CreateIssueInput) (*Issue, error) {
	if input.WorkspaceID == uuid.Nil {
		return nil, fmt.Errorf("%w: workspace ID is required", ErrInvalidInput)
	}
	title := strings.TrimSpace(input.Title)
	if title == "" {
		return nil, fmt.Errorf("%w: issue title is required", ErrInvalidInput)
	}
	priority := input.Priority
	if priority == "" {
		priority = PriorityMedium
	}
	if !priority.valid() {
		return nil, fmt.Errorf("%w: unsupported priority %q", ErrInvalidInput, priority)
	}
	status := input.Status
	if status == "" {
		status = StatusOpen
	}
	if !status.valid() {
		return nil, fmt.Errorf("%w: unsupported status %q", ErrInvalidInput, status)
	}
	aliases, err := normalizeAliases(input.Aliases)
	if err != nil {
		return nil, err
	}
	labels := normalizeStrings(input.Labels)
	id := input.ID
	if id == uuid.Nil {
		id = uuid.New()
	}

	err = r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var workspaceExists bool
		if err := tx.Raw(`SELECT EXISTS(SELECT 1 FROM todo_workspaces WHERE id = ?)`, input.WorkspaceID).Scan(&workspaceExists).Error; err != nil {
			return err
		}
		if !workspaceExists {
			return fmt.Errorf("%w: workspace %s", ErrNotFound, input.WorkspaceID)
		}

		result := tx.Exec(`
			INSERT INTO todo_issues
				(id, workspace_id, title, body, verification, labels, priority, status,
				 version, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, 1, now(), now())`,
			id, input.WorkspaceID, title, input.Body, input.Verification, pq.Array(labels),
			priority, status,
		)
		if result.Error != nil {
			return result.Error
		}
		if err := insertAliases(tx, input.WorkspaceID, id, aliases); err != nil {
			return err
		}
		_, err := insertEvent(tx, id, 1, EventInput{
			Kind:  "created",
			Actor: input.Actor,
			Payload: map[string]any{
				"title":     title,
				"labels":    labels,
				"priority":  priority,
				"status":    status,
				"workspace": input.WorkspaceID,
			},
		})
		return err
	})
	if err != nil {
		return nil, err
	}
	return r.GetIssue(ctx, id)
}

func (r *Repository) GetIssue(ctx context.Context, id uuid.UUID) (*Issue, error) {
	return getIssue(r.db.WithContext(ctx), `id = ?`, id)
}

// CountIssuesByStatus groups a workspace's issues by the four columns the
// derived TODO status is a function of. Callers fold the groups through the
// same derivation List uses, so counting never has to materialize issue bodies.
func (r *Repository) CountIssuesByStatus(ctx context.Context, workspaceID uuid.UUID) ([]IssueStatusCount, error) {
	var counts []IssueStatusCount
	result := r.db.WithContext(ctx).Raw(`
		SELECT issue.status,
		       COALESCE(runtime.execution_state, 'idle') AS execution_state,
		       COALESCE(active_link.step_kind::text, '') AS step_kind,
		       COALESCE(plan.approval_state::text, '')   AS approval_state,
		       COUNT(*)                                  AS count
		FROM `+issueFrom+`
		LEFT JOIN todo_issue_prompt_runs AS active_link
		  ON active_link.issue_id = issue.id
		 AND active_link.prompt_run_id = issue.active_prompt_run_id
		LEFT JOIN captain_plans AS plan ON plan.id = issue.selected_plan_id
		WHERE issue.workspace_id = ?
		GROUP BY 1, 2, 3, 4`,
		workspaceID,
	).Scan(&counts)
	if result.Error != nil {
		return nil, result.Error
	}
	return counts, nil
}

func (r *Repository) ListIssues(ctx context.Context, workspaceID uuid.UUID) ([]Issue, error) {
	var records []issueRecord
	result := r.db.WithContext(ctx).Raw(
		`SELECT `+issueColumns+` FROM `+issueFrom+` WHERE issue.workspace_id = ? ORDER BY issue.updated_at DESC, issue.id`,
		workspaceID,
	).Scan(&records)
	if result.Error != nil {
		return nil, result.Error
	}
	issues := make([]Issue, 0, len(records))
	for _, record := range records {
		issues = append(issues, *record.issue())
	}
	return issues, nil
}

// GetIssueByRef resolves exact workspace aliases before UUIDs so imported
// UUID-shaped aliases remain authoritative. It then accepts native UUID and
// unambiguous UUID/alias prefixes of at least MinShortReferenceLength.
func (r *Repository) GetIssueByRef(ctx context.Context, workspaceID uuid.UUID, ref string) (*Issue, error) {
	ref = normalizeToken(ref)
	if ref == "" {
		return nil, fmt.Errorf("%w: issue reference is required", ErrInvalidInput)
	}

	var exactAlias struct{ IssueID uuid.UUID }
	aliasResult := r.db.WithContext(ctx).Raw(`
		SELECT issue_id FROM todo_issue_aliases
		WHERE workspace_id = ? AND alias = ?`, workspaceID, ref,
	).Scan(&exactAlias)
	if aliasResult.Error != nil {
		return nil, aliasResult.Error
	}
	if aliasResult.RowsAffected > 0 {
		return getIssue(r.db.WithContext(ctx), `workspace_id = ? AND id = ?`, workspaceID, exactAlias.IssueID)
	}

	if id, err := uuid.Parse(ref); err == nil {
		issue, err := getIssue(r.db.WithContext(ctx), `workspace_id = ? AND id = ?`, workspaceID, id)
		if err == nil || !errors.Is(err, ErrNotFound) {
			return issue, err
		}
	}

	compact := strings.ReplaceAll(ref, "-", "")
	if len(compact) < MinShortReferenceLength {
		return nil, fmt.Errorf("%w: short issue reference must contain at least %d characters", ErrInvalidInput, MinShortReferenceLength)
	}
	prefix := escapeLike(ref) + "%"
	compactPrefix := escapeLike(compact) + "%"
	var matches []struct{ IssueID uuid.UUID }
	result := r.db.WithContext(ctx).Raw(`
		SELECT DISTINCT issue_id
		FROM (
			SELECT id AS issue_id
			FROM todo_issues
			WHERE workspace_id = ?
			  AND replace(id::text, '-', '') LIKE ? ESCAPE '\'
			UNION
			SELECT issue_id
			FROM todo_issue_aliases
			WHERE workspace_id = ? AND alias LIKE ? ESCAPE '\'
		) AS matched_refs
		LIMIT 2`, workspaceID, compactPrefix, workspaceID, prefix,
	).Scan(&matches)
	if result.Error != nil {
		return nil, result.Error
	}
	switch len(matches) {
	case 0:
		return nil, fmt.Errorf("%w: issue reference %q", ErrNotFound, ref)
	case 1:
		return getIssue(r.db.WithContext(ctx), `workspace_id = ? AND id = ?`, workspaceID, matches[0].IssueID)
	default:
		return nil, fmt.Errorf("%w: issue reference %q", ErrAmbiguousReference, ref)
	}
}

// GetIssueByGlobalRef resolves issue references without consulting a runtime
// provider or requiring a workspace. Exact aliases remain authoritative over
// UUIDs, including when an alias happens to be a valid native UUID. References
// that identify issues in multiple workspaces are rejected as ambiguous.
func (r *Repository) GetIssueByGlobalRef(ctx context.Context, ref string) (*Issue, error) {
	ref = normalizeToken(ref)
	if ref == "" {
		return nil, fmt.Errorf("%w: issue reference is required", ErrInvalidInput)
	}

	var exactAliases []struct{ IssueID uuid.UUID }
	aliasResult := r.db.WithContext(ctx).Raw(`
		SELECT DISTINCT issue_id
		FROM todo_issue_aliases
		WHERE alias = ?
		LIMIT 2`, ref,
	).Scan(&exactAliases)
	if aliasResult.Error != nil {
		return nil, aliasResult.Error
	}
	switch len(exactAliases) {
	case 1:
		return getIssue(r.db.WithContext(ctx), `id = ?`, exactAliases[0].IssueID)
	case 2:
		return nil, fmt.Errorf("%w: issue reference %q", ErrAmbiguousReference, ref)
	}

	if id, err := uuid.Parse(ref); err == nil {
		issue, err := getIssue(r.db.WithContext(ctx), `id = ?`, id)
		if err == nil || !errors.Is(err, ErrNotFound) {
			return issue, err
		}
	}

	compact := strings.ReplaceAll(ref, "-", "")
	if len(compact) < MinShortReferenceLength {
		return nil, fmt.Errorf("%w: short issue reference must contain at least %d characters", ErrInvalidInput, MinShortReferenceLength)
	}
	prefix := escapeLike(ref) + "%"
	compactPrefix := escapeLike(compact) + "%"
	var matches []struct{ IssueID uuid.UUID }
	result := r.db.WithContext(ctx).Raw(`
		SELECT DISTINCT issue_id
		FROM (
			SELECT id AS issue_id
			FROM todo_issues
			WHERE replace(id::text, '-', '') LIKE ? ESCAPE '\'
			UNION
			SELECT issue_id
			FROM todo_issue_aliases
			WHERE alias LIKE ? ESCAPE '\'
		) AS matched_refs
		LIMIT 2`, compactPrefix, prefix,
	).Scan(&matches)
	if result.Error != nil {
		return nil, result.Error
	}
	switch len(matches) {
	case 0:
		return nil, fmt.Errorf("%w: issue reference %q", ErrNotFound, ref)
	case 1:
		return getIssue(r.db.WithContext(ctx), `id = ?`, matches[0].IssueID)
	default:
		return nil, fmt.Errorf("%w: issue reference %q", ErrAmbiguousReference, ref)
	}
}

// GetIssueBySessionRef resolves either a Captain-native session UUID or a
// provider session UUID to the one native issue whose prompt run is linked to
// that session. A provider session can have separate Gavel orchestration and
// Claude/Codex transcript rows; matching the shared provider identity bridges
// those rows without making Captain own Gavel's issue links.
func (r *Repository) GetIssueBySessionRef(ctx context.Context, ref string) (*Issue, string, error) {
	ref = normalizeToken(ref)
	parsed, err := uuid.Parse(ref)
	if err != nil {
		return nil, "", fmt.Errorf("%w: session reference must be a UUID", ErrInvalidInput)
	}
	ref = parsed.String()

	var matches []struct {
		IssueID         uuid.UUID
		SessionIdentity string
	}
	result := r.db.WithContext(ctx).Raw(`
		WITH target_sessions AS (
			SELECT id, NULLIF(btrim(provider_session_id), '') AS provider_session_id
			FROM captain_sessions
			WHERE id = ? OR provider_session_id = ?
		), candidate_sessions AS (
			SELECT DISTINCT
				session.id,
				COALESCE(NULLIF(btrim(session.provider_session_id), ''), session.id::text) AS session_identity
			FROM captain_sessions AS session
			WHERE session.id IN (SELECT id FROM target_sessions)
			   OR NULLIF(btrim(session.provider_session_id), '') IN (
				SELECT provider_session_id
				FROM target_sessions
				WHERE provider_session_id IS NOT NULL
			   )
		)
		SELECT
			link.issue_id,
			MIN(candidate.session_identity) AS session_identity
		FROM todo_issue_prompt_runs AS link
		JOIN captain_prompt_runs AS run ON run.id = link.prompt_run_id
		JOIN candidate_sessions AS candidate ON candidate.id = run.session_id
		GROUP BY link.issue_id
		LIMIT 2`, parsed, ref,
	).Scan(&matches)
	if result.Error != nil {
		return nil, "", result.Error
	}
	switch len(matches) {
	case 0:
		return nil, "", fmt.Errorf("%w: session reference %q", ErrNotFound, ref)
	case 1:
		issue, err := getIssue(r.db.WithContext(ctx), `id = ?`, matches[0].IssueID)
		return issue, matches[0].SessionIdentity, err
	default:
		return nil, "", fmt.Errorf("%w: session reference %q", ErrAmbiguousReference, ref)
	}
}

// MoveIssueWorkspace transfers one issue between native workspaces without
// changing its identity or deleting its history and Captain links. Issues with
// relationships must be detached first because relationships are
// workspace-scoped.
func (r *Repository) MoveIssueWorkspace(
	ctx context.Context,
	issueID, targetWorkspaceID uuid.UUID,
	expectedVersion int64,
	actor string,
) (*Issue, error) {
	if issueID == uuid.Nil || targetWorkspaceID == uuid.Nil {
		return nil, fmt.Errorf("%w: issue and target workspace IDs are required", ErrInvalidInput)
	}

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		sourceWorkspaceID, err := issueWorkspace(tx, issueID)
		if err != nil {
			return err
		}
		// Relationship mutations take this same workspace-scoped advisory lock.
		// Taking it before the issue row lock preserves their lock ordering.
		if err := lockWorkspaceRelationships(tx, sourceWorkspaceID); err != nil {
			return err
		}
		locked, err := lockIssue(tx, issueID, expectedVersion)
		if err != nil {
			return err
		}
		if locked.WorkspaceID != sourceWorkspaceID {
			return fmt.Errorf("%w: issue %s workspace changed during move", ErrVersionConflict, issueID)
		}
		if locked.WorkspaceID == targetWorkspaceID {
			return ErrNoChanges
		}

		var targetWorkspace struct{ ID uuid.UUID }
		targetResult := tx.Raw(`SELECT id FROM todo_workspaces WHERE id = ? FOR SHARE`, targetWorkspaceID).Scan(&targetWorkspace)
		if targetResult.Error != nil {
			return targetResult.Error
		}
		if targetResult.RowsAffected == 0 {
			return fmt.Errorf("%w: workspace %s", ErrNotFound, targetWorkspaceID)
		}

		var hasRelationships bool
		if err := tx.Raw(`
			SELECT EXISTS(
				SELECT 1 FROM todo_issue_relationships
				WHERE issue_id = ? OR target_issue_id = ?
			)`, issueID, issueID,
		).Scan(&hasRelationships).Error; err != nil {
			return err
		}
		if hasRelationships {
			return fmt.Errorf("%w: issue %s", ErrIssueHasRelationships, issueID)
		}

		var aliases []Alias
		if err := tx.Raw(`
			SELECT workspace_id, alias, issue_id, COALESCE(kind, '') AS kind, created_at
			FROM todo_issue_aliases
			WHERE issue_id = ?
			ORDER BY alias`, issueID,
		).Scan(&aliases).Error; err != nil {
			return err
		}
		var conflictingAlias string
		conflictResult := tx.Raw(`
			SELECT source.alias
			FROM todo_issue_aliases AS source
			JOIN todo_issue_aliases AS target
			  ON target.workspace_id = ? AND target.alias = source.alias
			WHERE source.issue_id = ?
			LIMIT 1`, targetWorkspaceID, issueID,
		).Scan(&conflictingAlias)
		if conflictResult.Error != nil {
			return conflictResult.Error
		}
		if conflictResult.RowsAffected > 0 {
			return fmt.Errorf("%w: workspace %s alias %q", ErrAliasConflict, targetWorkspaceID, conflictingAlias)
		}

		// The alias foreign key includes workspace_id and is immediate. Remove
		// and reinsert aliases inside this transaction so the issue and aliases
		// never expose different workspaces while retaining alias metadata.
		if err := tx.Exec(`DELETE FROM todo_issue_aliases WHERE issue_id = ?`, issueID).Error; err != nil {
			return err
		}
		moveResult := tx.Exec(`
			UPDATE todo_issues
			SET workspace_id = ?
			WHERE id = ? AND version = ?`, targetWorkspaceID, issueID, expectedVersion,
		)
		if moveResult.Error != nil {
			return moveResult.Error
		}
		if moveResult.RowsAffected != 1 {
			return fmt.Errorf("%w: issue %s expected version %d", ErrVersionConflict, issueID, expectedVersion)
		}
		for _, alias := range aliases {
			result := tx.Exec(`
				INSERT INTO todo_issue_aliases
					(workspace_id, alias, issue_id, kind, created_at)
				VALUES (?, ?, ?, NULLIF(?, ''), ?)`,
				targetWorkspaceID, alias.Alias, issueID, alias.Kind, alias.CreatedAt,
			)
			if result.Error != nil {
				return mapUniqueError(result.Error, ErrAliasConflict, "workspace %s alias %q", targetWorkspaceID, alias.Alias)
			}
		}

		_, err = recordMutation(tx, locked, EventInput{
			Kind:  "workspace_moved",
			Actor: actor,
			Payload: map[string]any{
				"fromWorkspaceId": sourceWorkspaceID,
				"toWorkspaceId":   targetWorkspaceID,
			},
		})
		return err
	})
	if err != nil {
		return nil, err
	}
	return r.GetIssue(ctx, issueID)
}

func (r *Repository) UpdateIssue(ctx context.Context, id uuid.UUID, expectedVersion int64, patch IssuePatch) (*Issue, error) {
	updates := map[string]any{}
	payload := map[string]any{}
	if patch.Title != nil {
		title := strings.TrimSpace(*patch.Title)
		if title == "" {
			return nil, fmt.Errorf("%w: issue title is required", ErrInvalidInput)
		}
		updates["title"] = title
		payload["title"] = title
	}
	if patch.Body != nil {
		updates["body"] = *patch.Body
		payload["body"] = *patch.Body
	}
	if patch.Verification != nil {
		updates["verification"] = *patch.Verification
		payload["verification"] = *patch.Verification
	}
	if patch.Labels != nil {
		labels := normalizeStrings(*patch.Labels)
		updates["labels"] = pq.Array(labels)
		payload["labels"] = labels
	}
	if patch.Priority != nil {
		if !patch.Priority.valid() {
			return nil, fmt.Errorf("%w: unsupported priority %q", ErrInvalidInput, *patch.Priority)
		}
		updates["priority"] = *patch.Priority
		payload["priority"] = *patch.Priority
	}
	if patch.Status != nil {
		if !patch.Status.valid() {
			return nil, fmt.Errorf("%w: unsupported status %q", ErrInvalidInput, *patch.Status)
		}
		updates["status"] = *patch.Status
		payload["status"] = *patch.Status
	}
	if len(updates) == 0 {
		return nil, ErrNoChanges
	}

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		locked, err := lockIssue(tx, id, expectedVersion)
		if err != nil {
			return err
		}
		current, err := getIssue(tx, `id = ?`, id)
		if err != nil {
			return err
		}
		pruneUnchangedIssueUpdates(current, updates, payload)
		if len(updates) == 0 {
			return nil
		}
		if err := tx.Table("todo_issues").Where("id = ?", id).Updates(updates).Error; err != nil {
			return err
		}
		_, err = recordMutation(tx, locked, EventInput{
			Kind:    "updated",
			Actor:   patch.Actor,
			Payload: payload,
		})
		return err
	})
	if err != nil {
		return nil, err
	}
	return r.GetIssue(ctx, id)
}

func pruneUnchangedIssueUpdates(current *Issue, updates, payload map[string]any) {
	if value, ok := updates["title"].(string); ok && current.Title == value {
		delete(updates, "title")
		delete(payload, "title")
	}
	if value, ok := updates["body"].(string); ok && current.Body == value {
		delete(updates, "body")
		delete(payload, "body")
	}
	if value, ok := updates["verification"].(string); ok && current.Verification == value {
		delete(updates, "verification")
		delete(payload, "verification")
	}
	if value, ok := payload["labels"].([]string); ok && slices.Equal(current.Labels, value) {
		delete(updates, "labels")
		delete(payload, "labels")
	}
	if value, ok := updates["priority"].(Priority); ok && current.Priority == value {
		delete(updates, "priority")
		delete(payload, "priority")
	}
	if value, ok := updates["status"].(IssueStatus); ok && current.Status == value {
		delete(updates, "status")
		delete(payload, "status")
	}
}

// DeleteIssue performs an explicit hard delete. Normal workflow deletion
// should use StatusCancelled so append-only history remains available.
func (r *Repository) DeleteIssue(ctx context.Context, id uuid.UUID, expectedVersion int64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if _, err := lockIssue(tx, id, expectedVersion); err != nil {
			return err
		}
		result := tx.Exec(`DELETE FROM todo_issues WHERE id = ?`, id)
		if result.Error != nil {
			return mapDeleteIssueError(result.Error, id)
		}
		if result.RowsAffected == 0 {
			return fmt.Errorf("%w: issue %s", ErrNotFound, id)
		}
		return nil
	})
}

func (r *Repository) SetAliases(ctx context.Context, issueID uuid.UUID, expectedVersion int64, aliases []AliasInput, actor string) (*Issue, error) {
	normalized, err := normalizeAliases(aliases)
	if err != nil {
		return nil, err
	}
	err = r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		locked, err := lockIssue(tx, issueID, expectedVersion)
		if err != nil {
			return err
		}
		if err := tx.Exec(`DELETE FROM todo_issue_aliases WHERE issue_id = ?`, issueID).Error; err != nil {
			return err
		}
		if err := insertAliases(tx, locked.WorkspaceID, issueID, normalized); err != nil {
			return err
		}
		_, err = recordMutation(tx, locked, EventInput{
			Kind:  "aliases_updated",
			Actor: actor,
			Payload: map[string]any{
				"aliases": normalized,
			},
		})
		return err
	})
	if err != nil {
		return nil, err
	}
	return r.GetIssue(ctx, issueID)
}

func (r *Repository) ListAliases(ctx context.Context, issueID uuid.UUID) ([]Alias, error) {
	var aliases []Alias
	result := r.db.WithContext(ctx).Raw(`
		SELECT workspace_id, alias, issue_id, COALESCE(kind, '') AS kind, created_at
		FROM todo_issue_aliases WHERE issue_id = ? ORDER BY alias`, issueID,
	).Scan(&aliases)
	return aliases, result.Error
}

// ListAliasesByKind reads one kind of alias for a whole workspace in a single
// query, keyed by issue. Listing decorates every issue with its external links,
// so the per-issue ListAliases would be an N+1 across the entire backlog.
func (r *Repository) ListAliasesByKind(ctx context.Context, workspaceID uuid.UUID, kind string) (map[uuid.UUID][]Alias, error) {
	var aliases []Alias
	result := r.db.WithContext(ctx).Raw(`
		SELECT workspace_id, alias, issue_id, COALESCE(kind, '') AS kind, created_at
		FROM todo_issue_aliases
		WHERE workspace_id = ? AND lower(COALESCE(kind, '')) = lower(?)
		ORDER BY alias`, workspaceID, kind,
	).Scan(&aliases)
	if result.Error != nil {
		return nil, result.Error
	}
	byIssue := make(map[uuid.UUID][]Alias, len(aliases))
	for _, alias := range aliases {
		byIssue[alias.IssueID] = append(byIssue[alias.IssueID], alias)
	}
	return byIssue, nil
}

func (r *Repository) AppendEvent(ctx context.Context, issueID uuid.UUID, expectedVersion int64, input EventInput) (*Event, error) {
	var event *Event
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		locked, err := lockIssue(tx, issueID, expectedVersion)
		if err != nil {
			return err
		}
		event, err = recordMutation(tx, locked, input)
		return err
	})
	return event, err
}

func (r *Repository) AddComment(ctx context.Context, issueID uuid.UUID, expectedVersion int64, actor, body string) (*Event, error) {
	if strings.TrimSpace(body) == "" {
		return nil, fmt.Errorf("%w: comment body is required", ErrInvalidInput)
	}
	return r.AppendEvent(ctx, issueID, expectedVersion, EventInput{
		Kind:  "comment",
		Actor: actor,
		Body:  body,
	})
}

func (r *Repository) ListEvents(ctx context.Context, issueID uuid.UUID) ([]Event, error) {
	type eventRecord struct {
		ID        uuid.UUID
		IssueID   uuid.UUID
		Sequence  int64
		Kind      string
		Actor     sql.NullString
		Body      sql.NullString
		Payload   []byte
		Source    string
		SourceID  sql.NullString
		CreatedAt time.Time
	}
	var records []eventRecord
	result := r.db.WithContext(ctx).Raw(`
		SELECT id, issue_id, sequence, kind, actor, body, payload, source, source_id, created_at
		FROM todo_issue_events WHERE issue_id = ? ORDER BY sequence`, issueID,
	).Scan(&records)
	if result.Error != nil {
		return nil, result.Error
	}
	events := make([]Event, 0, len(records))
	for _, record := range records {
		payload := append(json.RawMessage(nil), record.Payload...)
		if len(payload) == 0 {
			payload = json.RawMessage(`{}`)
		}
		events = append(events, Event{
			ID:        record.ID,
			IssueID:   record.IssueID,
			Sequence:  record.Sequence,
			Kind:      record.Kind,
			Actor:     record.Actor.String,
			Body:      record.Body.String,
			Payload:   payload,
			Source:    record.Source,
			SourceID:  record.SourceID.String,
			CreatedAt: record.CreatedAt,
		})
	}
	return events, nil
}

type lockedIssue struct {
	ID          uuid.UUID
	WorkspaceID uuid.UUID
	Version     int64
}

func lockIssue(tx *gorm.DB, id uuid.UUID, expectedVersion int64) (*lockedIssue, error) {
	var issue lockedIssue
	result := tx.Raw(`
		SELECT id, workspace_id, version FROM todo_issues
		WHERE id = ? FOR UPDATE`, id,
	).Scan(&issue)
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, fmt.Errorf("%w: issue %s", ErrNotFound, id)
	}
	if issue.Version != expectedVersion {
		return nil, fmt.Errorf("%w: issue %s expected version %d, current version %d", ErrVersionConflict, id, expectedVersion, issue.Version)
	}
	return &issue, nil
}

func getIssue(db *gorm.DB, condition string, args ...any) (*Issue, error) {
	var record issueRecord
	result := db.Raw(`SELECT `+issueColumns+` FROM `+issueFrom+` WHERE `+condition, args...).Scan(&record)
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, ErrNotFound
	}
	return record.issue(), nil
}

func insertAliases(tx *gorm.DB, workspaceID, issueID uuid.UUID, aliases []AliasInput) error {
	for _, alias := range aliases {
		result := tx.Exec(`
			INSERT INTO todo_issue_aliases (workspace_id, alias, issue_id, kind, created_at)
			VALUES (?, ?, ?, NULLIF(?, ''), now())`,
			workspaceID, alias.Alias, issueID, alias.Kind,
		)
		if result.Error != nil {
			return mapUniqueError(result.Error, ErrAliasConflict, "workspace %s alias %q", workspaceID, alias.Alias)
		}
	}
	return nil
}

func recordMutation(tx *gorm.DB, issue *lockedIssue, input EventInput) (*Event, error) {
	kind := normalizeToken(input.Kind)
	if kind == "" {
		return nil, fmt.Errorf("%w: event kind is required", ErrInvalidInput)
	}
	var sequence int64
	if err := tx.Raw(`
		SELECT COALESCE(MAX(sequence), 0) + 1
		FROM todo_issue_events WHERE issue_id = ?`, issue.ID,
	).Scan(&sequence).Error; err != nil {
		return nil, err
	}
	nextVersion := issue.Version + 1
	result := tx.Exec(`
		UPDATE todo_issues SET version = ?, updated_at = now()
		WHERE id = ? AND version = ?`, nextVersion, issue.ID, issue.Version,
	)
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected != 1 {
		return nil, fmt.Errorf("%w: issue %s expected version %d", ErrVersionConflict, issue.ID, issue.Version)
	}
	event, err := insertEvent(tx, issue.ID, sequence, input)
	if err != nil {
		return nil, err
	}
	issue.Version = nextVersion
	return event, nil
}

func insertEvent(tx *gorm.DB, issueID uuid.UUID, sequence int64, input EventInput) (*Event, error) {
	kind := normalizeToken(input.Kind)
	if kind == "" {
		return nil, fmt.Errorf("%w: event kind is required", ErrInvalidInput)
	}
	payload, err := marshalPayload(input.Payload)
	if err != nil {
		return nil, err
	}
	source := normalizeToken(input.Source)
	if source == "" {
		source = "gavel"
	}

	type eventRecord struct {
		ID        uuid.UUID
		IssueID   uuid.UUID
		Sequence  int64
		Kind      string
		Actor     sql.NullString
		Body      sql.NullString
		Payload   []byte
		Source    string
		SourceID  sql.NullString
		CreatedAt time.Time
	}
	var record eventRecord
	result := tx.Raw(`
		INSERT INTO todo_issue_events
			(id, issue_id, sequence, kind, actor, body, payload, source, source_id, created_at)
		VALUES (?, ?, ?, ?, NULLIF(?, ''), NULLIF(?, ''), CAST(? AS jsonb), ?, NULLIF(?, ''), now())
		RETURNING id, issue_id, sequence, kind, actor, body, payload, source, source_id, created_at`,
		uuid.New(), issueID, sequence, kind, strings.TrimSpace(input.Actor), input.Body,
		string(payload), source, strings.TrimSpace(input.SourceID),
	).Scan(&record)
	if result.Error != nil {
		return nil, mapUniqueError(result.Error, ErrEventConflict, "event source %q/%q", source, input.SourceID)
	}
	return &Event{
		ID:        record.ID,
		IssueID:   record.IssueID,
		Sequence:  record.Sequence,
		Kind:      record.Kind,
		Actor:     record.Actor.String,
		Body:      record.Body.String,
		Payload:   append(json.RawMessage(nil), record.Payload...),
		Source:    record.Source,
		SourceID:  record.SourceID.String,
		CreatedAt: record.CreatedAt,
	}, nil
}

func marshalPayload(value any) ([]byte, error) {
	if value == nil {
		return []byte(`{}`), nil
	}
	switch payload := value.(type) {
	case json.RawMessage:
		if !json.Valid(payload) {
			return nil, fmt.Errorf("%w: event payload is not valid JSON", ErrInvalidInput)
		}
		return payload, nil
	case []byte:
		if !json.Valid(payload) {
			return nil, fmt.Errorf("%w: event payload is not valid JSON", ErrInvalidInput)
		}
		return payload, nil
	default:
		encoded, err := json.Marshal(value)
		if err != nil {
			return nil, fmt.Errorf("encode native todo event payload: %w", err)
		}
		return encoded, nil
	}
}

func normalizeAliases(inputs []AliasInput) ([]AliasInput, error) {
	byAlias := make(map[string]AliasInput, len(inputs))
	for _, input := range inputs {
		alias := normalizeToken(input.Alias)
		if alias == "" {
			return nil, fmt.Errorf("%w: alias cannot be empty", ErrInvalidInput)
		}
		kind := normalizeToken(input.Kind)
		if existing, ok := byAlias[alias]; ok && existing.Kind != kind {
			return nil, fmt.Errorf("%w: alias %q has conflicting kinds", ErrInvalidInput, alias)
		}
		byAlias[alias] = AliasInput{Alias: alias, Kind: kind}
	}
	aliases := make([]AliasInput, 0, len(byAlias))
	for _, alias := range byAlias {
		aliases = append(aliases, alias)
	}
	sort.Slice(aliases, func(i, j int) bool { return aliases[i].Alias < aliases[j].Alias })
	return aliases, nil
}

func normalizeStrings(values []string) []string {
	unique := make(map[string]struct{}, len(values))
	for _, value := range values {
		if value = normalizeToken(value); value != "" {
			unique[value] = struct{}{}
		}
	}
	result := make([]string, 0, len(unique))
	for value := range unique {
		result = append(result, value)
	}
	sort.Strings(result)
	return result
}

func normalizeToken(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func normalizeWorkspacePath(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	return filepath.Clean(value)
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

func nullableString(value string) any {
	if value == "" {
		return nil
	}
	return value
}

func escapeLike(value string) string {
	value = strings.ReplaceAll(value, `\`, `\\`)
	value = strings.ReplaceAll(value, `%`, `\%`)
	return strings.ReplaceAll(value, `_`, `\_`)
}

func mapUniqueError(err, sentinel error, format string, args ...any) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return fmt.Errorf("%w: %s", sentinel, fmt.Sprintf(format, args...))
	}
	return err
}

func mapDeleteIssueError(err error, issueID uuid.UUID) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23503" &&
		strings.HasPrefix(pgErr.ConstraintName, "todo_issue_relationships_") {
		return fmt.Errorf("%w: issue %s", ErrIssueHasRelationships, issueID)
	}
	return err
}

func (p Priority) valid() bool {
	switch p {
	case PriorityLow, PriorityMedium, PriorityHigh, PriorityCritical:
		return true
	default:
		return false
	}
}

func (s IssueStatus) valid() bool {
	switch s {
	case StatusDraft, StatusOpen, StatusVerified, StatusClosed, StatusCancelled:
		return true
	default:
		return false
	}
}

func (r RelationshipKind) valid() bool {
	return r == RelationshipDependsOn || r == RelationshipRelatedTo
}

func (s StepKind) valid() bool {
	return s == StepPlan || s == StepRun || s == StepVerify
}
