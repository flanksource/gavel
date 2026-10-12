package native

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// SetIssueParent makes issueID a child of parentID, or a top-level issue again
// when parentID is nil, and records a parent_changed event. The hierarchy is one
// level deep: the parent must be a top-level issue in the same workspace, and an
// issue that has children cannot become a child. Naming the parent the issue
// already has changes nothing and records nothing.
//
// Parent changes are serialized per workspace, with the lock relationships and
// workspace moves take, so two issues cannot each become the other's child.
func (r *Repository) SetIssueParent(
	ctx context.Context,
	issueID uuid.UUID,
	parentID *uuid.UUID,
	expectedVersion int64,
	actor string,
) (*Issue, error) {
	if issueID == uuid.Nil || (parentID != nil && *parentID == uuid.Nil) {
		return nil, fmt.Errorf("%w: issue and parent IDs are required", ErrInvalidInput)
	}
	if parentID != nil && *parentID == issueID {
		return nil, fmt.Errorf("%w: issue %s cannot be its own parent", ErrInvalidParent, issueID)
	}

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		locked, err := lockIssueHierarchy(tx, issueID, expectedVersion)
		if err != nil {
			return err
		}
		current, err := issueParent(tx, issueID)
		if err != nil {
			return err
		}
		if sameParent(current, parentID) {
			return nil
		}
		if parentID != nil {
			if err := requireTopLevelParent(tx, locked.WorkspaceID, *parentID); err != nil {
				return err
			}
			if err := requireNoChildren(tx, issueID); err != nil {
				return err
			}
		}
		if err := tx.Exec(`UPDATE todo_issues SET parent_issue_id = ? WHERE id = ?`, parentID, issueID).Error; err != nil {
			return err
		}
		_, err = recordMutation(tx, locked, EventInput{
			Kind:    "parent_changed",
			Actor:   actor,
			Payload: map[string]any{"from": current, "to": parentID},
		})
		return err
	})
	if err != nil {
		return nil, err
	}
	return r.GetIssue(ctx, issueID)
}

// lockIssueHierarchy takes the workspace relationship lock and then the issue
// row lock, the order relationships and workspace moves use.
func lockIssueHierarchy(tx *gorm.DB, issueID uuid.UUID, expectedVersion int64) (*lockedIssue, error) {
	workspaceID, err := issueWorkspace(tx, issueID)
	if err != nil {
		return nil, err
	}
	if err := lockWorkspaceRelationships(tx, workspaceID); err != nil {
		return nil, err
	}
	locked, err := lockIssue(tx, issueID, expectedVersion)
	if err != nil {
		return nil, err
	}
	if locked.WorkspaceID != workspaceID {
		return nil, fmt.Errorf("%w: issue %s workspace changed while setting its parent", ErrVersionConflict, issueID)
	}
	return locked, nil
}

// requireTopLevelParent is the rule every new parent link must pass: the parent
// exists, belongs to workspaceID, is not itself a child, and is still open, so
// nothing is attached where a closed parent would hide it. Callers hold the
// workspace relationship lock, which is what keeps the answer true until they
// commit.
func requireTopLevelParent(tx *gorm.DB, workspaceID, parentID uuid.UUID) error {
	var parent struct {
		WorkspaceID   uuid.UUID
		ParentIssueID *uuid.UUID
		Status        IssueStatus
	}
	result := tx.Raw(`SELECT workspace_id, parent_issue_id, status FROM todo_issues WHERE id = ?`, parentID).Scan(&parent)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("%w: parent issue %s", ErrNotFound, parentID)
	}
	if parent.WorkspaceID != workspaceID {
		return fmt.Errorf("%w: parent issue %s belongs to workspace %s, not %s",
			ErrInvalidParent, parentID, parent.WorkspaceID, workspaceID)
	}
	if parent.ParentIssueID != nil {
		return fmt.Errorf("%w: issue %s is itself a child of %s, and a child cannot have children",
			ErrInvalidParent, parentID, *parent.ParentIssueID)
	}
	if parent.Status.IsClosed() {
		return fmt.Errorf("%w: issue %s is %s, and a closed issue cannot take children",
			ErrInvalidParent, parentID, parent.Status)
	}
	return nil
}

// lockIssueForPatch locks an issue for UpdateIssue. Cancelling is how an issue
// is deleted, and it must not take open children out of sight with it, so that
// one change is decided under the hierarchy lock.
func lockIssueForPatch(tx *gorm.DB, issueID uuid.UUID, expectedVersion int64, patch IssuePatch) (*lockedIssue, error) {
	if patch.Status == nil || *patch.Status != StatusCancelled {
		return lockIssue(tx, issueID, expectedVersion)
	}
	locked, err := lockIssueHierarchy(tx, issueID, expectedVersion)
	if err != nil {
		return nil, err
	}
	return locked, requireNoOpenChildren(tx, issueID)
}

// requireNoOpenChildren refuses to cancel an issue that would take open
// children out of sight with it. Closed children are left attached. The caller
// holds the workspace relationship lock, so no child can attach in between.
func requireNoOpenChildren(tx *gorm.DB, issueID uuid.UUID) error {
	var open int64
	err := tx.Raw(`
		SELECT COUNT(*) FROM todo_issues
		WHERE parent_issue_id = ? AND status NOT IN (?, ?)`, issueID, StatusClosed, StatusCancelled).Scan(&open).Error
	if err != nil {
		return err
	}
	if open > 0 {
		return fmt.Errorf("%w: %d under issue %s", ErrOpenChildren, open, issueID)
	}
	return nil
}

func requireNoChildren(tx *gorm.DB, issueID uuid.UUID) error {
	children, err := countChildren(tx, issueID)
	if err != nil {
		return err
	}
	if children > 0 {
		return fmt.Errorf("%w: issue %s has %d children, and a parent cannot become a child",
			ErrInvalidParent, issueID, children)
	}
	return nil
}

// requireOutsideHierarchy refuses an issue that has a parent or children. The
// parent foreign key is workspace-scoped, so such an issue cannot change
// workspace; this names the reason instead of surfacing the constraint.
func requireOutsideHierarchy(tx *gorm.DB, issueID uuid.UUID) error {
	parent, err := issueParent(tx, issueID)
	if err != nil {
		return err
	}
	if parent != nil {
		return fmt.Errorf("%w: issue %s is a child of %s; detach it first", ErrIssueInHierarchy, issueID, *parent)
	}
	children, err := countChildren(tx, issueID)
	if err != nil {
		return err
	}
	if children > 0 {
		return fmt.Errorf("%w: issue %s has %d children; detach them first", ErrIssueInHierarchy, issueID, children)
	}
	return nil
}

func issueParent(tx *gorm.DB, issueID uuid.UUID) (*uuid.UUID, error) {
	var record struct{ ParentIssueID *uuid.UUID }
	result := tx.Raw(`SELECT parent_issue_id FROM todo_issues WHERE id = ?`, issueID).Scan(&record)
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, fmt.Errorf("%w: issue %s", ErrNotFound, issueID)
	}
	return record.ParentIssueID, nil
}

func countChildren(tx *gorm.DB, issueID uuid.UUID) (int64, error) {
	var children int64
	err := tx.Raw(`SELECT COUNT(*) FROM todo_issues WHERE parent_issue_id = ?`, issueID).Scan(&children).Error
	return children, err
}

func sameParent(current, next *uuid.UUID) bool {
	if current == nil || next == nil {
		return current == next
	}
	return *current == *next
}

// addLineage records where a new issue came from in its created event: the
// parent it was attached to and the run it was created inside.
func addLineage(payload map[string]any, parentID *uuid.UUID, origin IssueOrigin) {
	if parentID != nil {
		payload["parent"] = *parentID
	}
	if issue := strings.TrimSpace(origin.IssueID); issue != "" {
		payload["originIssue"] = issue
	}
	if session := strings.TrimSpace(origin.SessionID); session != "" {
		payload["originSession"] = session
	}
}
