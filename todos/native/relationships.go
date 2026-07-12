package native

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AddRelationship inserts one canonical relationship and appends the matching
// issue event in the same transaction. Dependency cycle checks are serialized
// per workspace so two concurrent inverse insertions cannot both commit.
func (r *Repository) AddRelationship(
	ctx context.Context,
	issueID, targetIssueID uuid.UUID,
	relation RelationshipKind,
	expectedVersion int64,
	actor string,
) (*Relationship, error) {
	if issueID == uuid.Nil || targetIssueID == uuid.Nil {
		return nil, fmt.Errorf("%w: relationship issue IDs are required", ErrInvalidInput)
	}
	if issueID == targetIssueID {
		return nil, fmt.Errorf("%w: issue %s", ErrSelfRelationship, issueID)
	}
	if !relation.valid() {
		return nil, fmt.Errorf("%w: unsupported relationship %q", ErrInvalidInput, relation)
	}

	var created Relationship
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		workspaceID, err := issueWorkspace(tx, issueID)
		if err != nil {
			return err
		}
		if err := lockWorkspaceRelationships(tx, workspaceID); err != nil {
			return err
		}
		locked, err := lockIssue(tx, issueID, expectedVersion)
		if err != nil {
			return err
		}
		targetWorkspaceID, err := issueWorkspace(tx, targetIssueID)
		if err != nil {
			return err
		}
		if locked.WorkspaceID != targetWorkspaceID {
			return fmt.Errorf("%w: issue %s belongs to %s, target %s belongs to %s",
				ErrCrossWorkspace, issueID, locked.WorkspaceID, targetIssueID, targetWorkspaceID)
		}

		storedIssueID, storedTargetID := canonicalRelationship(issueID, targetIssueID, relation)
		var exists bool
		if err := tx.Raw(`
			SELECT EXISTS(
				SELECT 1 FROM todo_issue_relationships
				WHERE workspace_id = ? AND issue_id = ? AND target_issue_id = ? AND relation = ?
			)`, workspaceID, storedIssueID, storedTargetID, relation,
		).Scan(&exists).Error; err != nil {
			return err
		}
		if exists {
			return fmt.Errorf("%w: %s %s %s", ErrRelationshipExists, issueID, relation, targetIssueID)
		}
		if relation == RelationshipDependsOn {
			cycle, err := dependencyPathExists(tx, workspaceID, targetIssueID, issueID)
			if err != nil {
				return err
			}
			if cycle {
				return fmt.Errorf("%w: adding %s -> %s", ErrRelationshipCycle, issueID, targetIssueID)
			}
		}

		result := tx.Raw(`
			INSERT INTO todo_issue_relationships
				(workspace_id, issue_id, target_issue_id, relation, created_at)
			VALUES (?, ?, ?, ?, now())
			RETURNING workspace_id, issue_id, target_issue_id, relation, created_at`,
			workspaceID, storedIssueID, storedTargetID, relation,
		).Scan(&created)
		if result.Error != nil {
			return mapUniqueError(result.Error, ErrRelationshipExists, "%s %s %s", issueID, relation, targetIssueID)
		}
		_, err = recordMutation(tx, locked, EventInput{
			Kind:  "relationship_added",
			Actor: actor,
			Payload: map[string]any{
				"relation":      relation,
				"targetIssueId": targetIssueID,
			},
		})
		return err
	})
	if err != nil {
		return nil, err
	}
	return &created, nil
}

func (r *Repository) DeleteRelationship(
	ctx context.Context,
	issueID, targetIssueID uuid.UUID,
	relation RelationshipKind,
	expectedVersion int64,
	actor string,
) error {
	if issueID == uuid.Nil || targetIssueID == uuid.Nil {
		return fmt.Errorf("%w: relationship issue IDs are required", ErrInvalidInput)
	}
	if issueID == targetIssueID {
		return fmt.Errorf("%w: issue %s", ErrSelfRelationship, issueID)
	}
	if !relation.valid() {
		return fmt.Errorf("%w: unsupported relationship %q", ErrInvalidInput, relation)
	}

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		workspaceID, err := issueWorkspace(tx, issueID)
		if err != nil {
			return err
		}
		if err := lockWorkspaceRelationships(tx, workspaceID); err != nil {
			return err
		}
		locked, err := lockIssue(tx, issueID, expectedVersion)
		if err != nil {
			return err
		}
		targetWorkspaceID, err := issueWorkspace(tx, targetIssueID)
		if err != nil {
			return err
		}
		if locked.WorkspaceID != targetWorkspaceID {
			return fmt.Errorf("%w: issue %s and target %s", ErrCrossWorkspace, issueID, targetIssueID)
		}
		storedIssueID, storedTargetID := canonicalRelationship(issueID, targetIssueID, relation)
		result := tx.Exec(`
			DELETE FROM todo_issue_relationships
			WHERE workspace_id = ? AND issue_id = ? AND target_issue_id = ? AND relation = ?`,
			workspaceID, storedIssueID, storedTargetID, relation,
		)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return fmt.Errorf("%w: %s %s %s", ErrRelationshipNotFound, issueID, relation, targetIssueID)
		}
		_, err = recordMutation(tx, locked, EventInput{
			Kind:  "relationship_removed",
			Actor: actor,
			Payload: map[string]any{
				"relation":      relation,
				"targetIssueId": targetIssueID,
			},
		})
		return err
	})
}

// ListRelationships returns every edge touching issueID from that issue's
// perspective: outgoing dependencies are depends_on, incoming dependencies
// are the derived read-only blocks relation, and related_to is symmetric. The
// returned IssueID is always issueID and TargetIssueID is always the other end.
func (r *Repository) ListRelationships(ctx context.Context, issueID uuid.UUID) ([]Relationship, error) {
	var relationships []Relationship
	result := r.db.WithContext(ctx).Raw(`
		SELECT workspace_id, issue_id, target_issue_id, relation, created_at
		FROM todo_issue_relationships
		WHERE issue_id = ? OR target_issue_id = ?`, issueID, issueID,
	).Scan(&relationships)
	if result.Error != nil {
		return nil, result.Error
	}
	for i := range relationships {
		relationship := &relationships[i]
		if relationship.TargetIssueID != issueID {
			continue
		}
		relationship.IssueID, relationship.TargetIssueID = issueID, relationship.IssueID
		if relationship.Relation == RelationshipDependsOn {
			relationship.Relation = RelationshipBlocks
		}
	}
	sort.Slice(relationships, func(i, j int) bool {
		if relationships[i].Relation != relationships[j].Relation {
			return relationships[i].Relation < relationships[j].Relation
		}
		return relationships[i].TargetIssueID.String() < relationships[j].TargetIssueID.String()
	})
	return relationships, nil
}

// ListDependents returns issues that directly depend on issueID. In UI terms,
// issueID blocks each returned issue.
func (r *Repository) ListDependents(ctx context.Context, issueID uuid.UUID) ([]Issue, error) {
	return r.listRelationshipIssues(ctx, `
		SELECT dependent.*
		FROM todo_issue_relationships AS relationship
		JOIN todo_issues AS dependent ON dependent.id = relationship.issue_id
		WHERE relationship.target_issue_id = ? AND relationship.relation = 'depends_on'
		ORDER BY dependent.updated_at DESC, dependent.id`, issueID)
}

// ListUnsatisfiedDependencies returns the direct dependencies that still block
// issueID. Verified and closed issues satisfy a dependency.
func (r *Repository) ListUnsatisfiedDependencies(ctx context.Context, issueID uuid.UUID) ([]Issue, error) {
	return r.listRelationshipIssues(ctx, `
		SELECT dependency.*
		FROM todo_issue_relationships AS relationship
		JOIN todo_issues AS dependency ON dependency.id = relationship.target_issue_id
		WHERE relationship.issue_id = ?
		  AND relationship.relation = 'depends_on'
		  AND dependency.status NOT IN ('verified', 'closed')
		ORDER BY dependency.updated_at DESC, dependency.id`, issueID)
}

func (r *Repository) listRelationshipIssues(ctx context.Context, query string, args ...any) ([]Issue, error) {
	var records []issueRecord
	if err := r.db.WithContext(ctx).Raw(query, args...).Scan(&records).Error; err != nil {
		return nil, err
	}
	issues := make([]Issue, 0, len(records))
	for _, record := range records {
		issues = append(issues, *record.issue())
	}
	return issues, nil
}

func issueWorkspace(tx *gorm.DB, issueID uuid.UUID) (uuid.UUID, error) {
	var record struct{ WorkspaceID uuid.UUID }
	result := tx.Raw(`SELECT workspace_id FROM todo_issues WHERE id = ?`, issueID).Scan(&record)
	if result.Error != nil {
		return uuid.Nil, result.Error
	}
	if result.RowsAffected == 0 {
		return uuid.Nil, fmt.Errorf("%w: issue %s", ErrNotFound, issueID)
	}
	return record.WorkspaceID, nil
}

func lockWorkspaceRelationships(tx *gorm.DB, workspaceID uuid.UUID) error {
	return tx.Exec(`
		SELECT pg_advisory_xact_lock(hashtextextended(CAST(? AS text), 0))`,
		workspaceID.String(),
	).Error
}

func canonicalRelationship(issueID, targetIssueID uuid.UUID, relation RelationshipKind) (uuid.UUID, uuid.UUID) {
	if relation == RelationshipRelatedTo && strings.Compare(issueID.String(), targetIssueID.String()) > 0 {
		return targetIssueID, issueID
	}
	return issueID, targetIssueID
}

func dependencyPathExists(tx *gorm.DB, workspaceID, fromIssueID, toIssueID uuid.UUID) (bool, error) {
	var exists bool
	err := tx.Raw(`
		WITH RECURSIVE dependencies(issue_id) AS (
			SELECT target_issue_id
			FROM todo_issue_relationships
			WHERE workspace_id = ? AND issue_id = ? AND relation = 'depends_on'
			UNION
			SELECT relationship.target_issue_id
			FROM todo_issue_relationships relationship
			JOIN dependencies ON relationship.issue_id = dependencies.issue_id
			WHERE relationship.workspace_id = ? AND relationship.relation = 'depends_on'
		)
		SELECT EXISTS(SELECT 1 FROM dependencies WHERE issue_id = ?)`,
		workspaceID, fromIssueID, workspaceID, toIssueID,
	).Scan(&exists).Error
	return exists, err
}
