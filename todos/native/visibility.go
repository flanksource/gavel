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

// requireChangeable fails with ErrForbidden unless the transaction's user created
// the issue or is an org admin. Local mode and system jobs may change any issue.
func requireChangeable(tx *gorm.DB, id uuid.UUID, createdBy *uuid.UUID) error {
	scope := viewer(tx.Statement.Context)
	if scope == nil || scope.Role == orgAdminRole || (createdBy != nil && *createdBy == *scope.UserID) {
		return nil
	}
	return fmt.Errorf("%w: issue %s belongs to another user; only its creator or an org admin can change it", ErrForbidden, id)
}
