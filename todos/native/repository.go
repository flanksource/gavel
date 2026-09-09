package native

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"

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
