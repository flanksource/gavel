// Hosted todo visibility: who may see an issue (its creator, or the whole org when public)
// and who may change it (its creator or an org admin), from the request's tenant scope.
package native

import (
	"context"
	"fmt"

	"github.com/flanksource/gavel/tenant"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Visibility says who in an organization sees an issue.
type Visibility string

const (
	// VisibilityPrivate issues are seen only by their creator. It is the default.
	VisibilityPrivate Visibility = "private"
	// VisibilityPublic issues are seen by every member of the organization.
	VisibilityPublic Visibility = "public"
)

const orgAdminRole = "org:admin"

func normalizeVisibility(value Visibility) (Visibility, error) {
	switch value {
	case "":
		return VisibilityPrivate, nil
	case VisibilityPrivate, VisibilityPublic:
		return value, nil
	default:
		return "", fmt.Errorf("%w: visibility %q is not private or public", ErrInvalidInput, value)
	}
}

// viewer is the user a request acts for, or nil for local mode and system jobs,
// which see and change every issue of their organization.
func viewer(ctx context.Context) *tenant.Scope {
	if ctx == nil {
		return nil
	}
	scope := tenant.FromContext(ctx)
	if scope.UserID == nil {
		return nil
	}
	return &scope
}

// visibleIssue limits an issue alias to the issues the context's user can see.
// It returns an always-true condition with no arguments when there is no user.
func visibleIssue(ctx context.Context, alias string) (string, []any) {
	scope := viewer(ctx)
	if scope == nil {
		return "TRUE", nil
	}
	return "(" + alias + ".visibility = 'public' OR " + alias + ".created_by_user_id = ?)", []any{*scope.UserID}
}

// requireChangeable is RequireChangeable for the transaction's user.
func requireChangeable(tx *gorm.DB, id uuid.UUID, createdBy *uuid.UUID) error {
	return RequireChangeable(tx.Statement.Context, id, createdBy)
}

// RequireChangeable fails with ErrForbidden unless the context's user created issue id, which
// createdBy names, or is an org admin. Local mode and system jobs may change any issue.
func RequireChangeable(ctx context.Context, id uuid.UUID, createdBy *uuid.UUID) error {
	scope := viewer(ctx)
	if scope == nil || scope.Role == orgAdminRole || (createdBy != nil && *createdBy == *scope.UserID) {
		return nil
	}
	return fmt.Errorf("%w: issue %s belongs to another user; only its creator or an org admin can change it", ErrForbidden, id)
}

// PrivateIssueCreator returns the name of the user who created issue id, one of the org's issues
// that the context's user can't see, so a link to it can say whose it is without revealing it.
// The name falls back to the creator's email, and is empty for an issue with no creator. It
// fails with ErrNotFound when the org has no such issue.
func (r *Repository) PrivateIssueCreator(ctx context.Context, id uuid.UUID) (string, error) {
	var rows []struct{ Creator string }
	if err := r.db.WithContext(ctx).Raw(`
		SELECT COALESCE(NULLIF(u.name, ''), u.email, '') AS creator
		FROM todo_issues AS issue
		LEFT JOIN users AS u ON u.id = issue.created_by_user_id
		WHERE issue.id = ? AND `+inOrgWorkspace("issue.workspace_id"), id, r.org).Scan(&rows).Error; err != nil {
		return "", fmt.Errorf("look up the creator of todo %s: %w", id, err)
	}
	if len(rows) == 0 {
		return "", fmt.Errorf("%w: todo %s", ErrNotFound, id)
	}
	return rows[0].Creator, nil
}

// VisibleIssue is the condition, and its arguments, that limits an issue alias to the issues
// the context's user can see, for queries outside this package that join todo_issues.
func VisibleIssue(ctx context.Context, alias string) (string, []any) {
	return visibleIssue(ctx, alias)
}
