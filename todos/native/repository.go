package native

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/flanksource/gavel/tenant"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
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
	ErrInvalidParent         = errors.New("invalid native todo parent")
	ErrIssueInHierarchy      = errors.New("native todo issue has a parent or children")
	ErrOpenChildren          = errors.New("native todo issue has open children")
	ErrLinkConflict          = errors.New("native todo Captain link conflict")
	ErrEventConflict         = errors.New("native todo event source already exists")
	ErrForbidden             = errors.New("native todo change forbidden")
)

// Repository persists native TODO state. It deliberately accepts an injected
// GORM pool and does not migrate or auto-create tables. Every read and write is
// scoped to one organization: the local org unless built by ForOrg.
type Repository struct {
	db  *gorm.DB
	org uuid.UUID
}

func NewRepository(db *gorm.DB) (*Repository, error) {
	if db == nil {
		return nil, fmt.Errorf("%w: database is nil", ErrInvalidInput)
	}
	return &Repository{db: db, org: tenant.LocalOrgID}, nil
}

// inTx returns a Repository for the same organization over db, typically an
// open transaction.
func (r *Repository) inTx(db *gorm.DB) *Repository {
	scoped := *r
	scoped.db = db
	return &scoped
}

// ForOrg returns a Repository over org's todos, sharing the connection.
func (r *Repository) ForOrg(org uuid.UUID) *Repository {
	scoped := *r
	scoped.org = org
	return &scoped
}

// inOrgWorkspace limits a workspace_id column to the organization's workspaces.
// The returned condition takes the org ID as its one argument.
func inOrgWorkspace(column string) string {
	return column + ` IN (SELECT id FROM todo_workspaces WHERE org_id = ?)`
}

// inOrgIssue limits an issue_id column to issues in the organization's
// workspaces. The returned condition takes the org ID as its one argument.
func inOrgIssue(column string) string {
	return column + ` IN (SELECT id FROM todo_issues WHERE ` + inOrgWorkspace("workspace_id") + `)`
}

// requestUser returns the user a transaction acts for, taken from its context.
// It is nil in local mode and for system jobs, which leaves attribution NULL.
func requestUser(tx *gorm.DB) *uuid.UUID {
	if tx.Statement.Context == nil {
		return nil
	}
	return tenant.FromContext(tx.Statement.Context).UserID
}

// requireWorkspace fails with ErrNotFound unless the workspace belongs to org.
func requireWorkspace(tx *gorm.DB, org, workspaceID uuid.UUID) error {
	var exists bool
	if err := tx.Raw(`SELECT EXISTS(SELECT 1 FROM todo_workspaces WHERE org_id = ? AND id = ?)`, org, workspaceID).Scan(&exists).Error; err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("%w: workspace %s", ErrNotFound, workspaceID)
	}
	return nil
}

type lockedIssue struct {
	ID              uuid.UUID
	WorkspaceID     uuid.UUID
	Version         int64
	CreatedByUserID *uuid.UUID
}

// lockIssue locks an issue of org for update. Another org's issue, and one the
// transaction's user can't see, is ErrNotFound; one they can see but didn't
// create is ErrForbidden unless they are an org admin.
func lockIssue(tx *gorm.DB, org, id uuid.UUID, expectedVersion int64) (*lockedIssue, error) {
	var issue lockedIssue
	visible, visibleArgs := visibleIssue(tx.Statement.Context, "todo_issues")
	result := tx.Raw(`
		SELECT id, workspace_id, version, created_by_user_id FROM todo_issues
		WHERE id = ? AND `+inOrgWorkspace("workspace_id")+` AND `+visible+` FOR UPDATE`, append([]any{id, org}, visibleArgs...)...,
	).Scan(&issue)
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, fmt.Errorf("%w: issue %s", ErrNotFound, id)
	}
	if err := requireChangeable(tx, id, issue.CreatedByUserID); err != nil {
		return nil, err
	}
	if issue.Version != expectedVersion {
		return nil, fmt.Errorf("%w: issue %s expected version %d, current version %d", ErrVersionConflict, id, expectedVersion, issue.Version)
	}
	return &issue, nil
}

// getIssue reads one issue of org matching condition. Another org's issue, and
// one the context's user can't see, is ErrNotFound.
func getIssue(db *gorm.DB, org uuid.UUID, condition string, args ...any) (*Issue, error) {
	var record issueRecord
	visible, visibleArgs := visibleIssue(db.Statement.Context, "issue")
	query := append(append([]any{org}, visibleArgs...), args...)
	result := db.Raw(`SELECT `+issueColumns+` FROM `+issueFrom+` WHERE `+inOrgWorkspace("issue.workspace_id")+` AND `+visible+` AND (`+condition+`)`,
		query...).Scan(&record)
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, ErrNotFound
	}
	return record.issue(), nil
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
	if errors.As(err, &pgErr) && pgErr.Code == "23503" && pgErr.ConstraintName == "todo_issues_parent_fkey" {
		return fmt.Errorf("%w: issue %s has children; detach or delete them first", ErrIssueInHierarchy, issueID)
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

// IsClosed reports a status no work is left under: finished or cancelled. A
// verified issue is not closed; it is waiting for someone to close it.
func (s IssueStatus) IsClosed() bool {
	return s == StatusClosed || s == StatusCancelled
}

func (r RelationshipKind) valid() bool {
	return r == RelationshipDependsOn || r.symmetric()
}

// symmetric relations read the same from both ends, so each pair is stored once
// in canonical order (see todo_issue_relationships_related_order).
func (r RelationshipKind) symmetric() bool {
	return r == RelationshipRelatedTo || r == RelationshipDuplicates
}

// valid mirrors the todo_issue_prompt_runs.step_kind CHECK exactly: a non-empty,
// trimmed, lower-case name.
//
// It is deliberately NOT an enum any more. Which steps exist is the project's
// lifecycle definition's to say (todos/lifecycle/todos.yaml, overridable from
// .gavel.yaml), so a closed set here would refuse in Go a step the database and
// the definition both accept. Whether a name is a step of the loaded lifecycle
// is checked where the lifecycle is loaded — the host — not in storage.
func (s StepKind) valid() bool {
	name := string(s)
	return name != "" && name == strings.ToLower(strings.TrimSpace(name))
}
