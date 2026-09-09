package runtime

import (
	"context"
	"fmt"
	"time"

	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/native"
	"github.com/flanksource/gavel/todos/types"
	"github.com/google/uuid"
)

var _ todos.RelationshipProvider = (*Provider)(nil)

// Link records a relationship from todo to the TODO named by targetRef. Cycle
// detection, cross-workspace rejection and duplicate detection are enforced by
// the repository inside one transaction.
func (p *Provider) Link(
	ctx context.Context,
	todo *types.TODO,
	targetRef string,
	relation types.RelationKind,
) (*todos.Link, error) {
	id, version, target, err := p.relationshipOperands(ctx, todo, targetRef, relation)
	if err != nil {
		return nil, err
	}
	created, err := p.repository.AddRelationship(ctx, id, target.ID, nativeRelation(relation), version, mutationActor)
	if err != nil {
		return nil, err
	}
	if err := p.reloadTODO(ctx, todo, p.workDir); err != nil {
		return nil, err
	}
	link := linkFromIssue(relation, target, created.CreatedAt)
	return &link, nil
}

// Unlink removes an existing relationship. The derived blocks relation is not
// removable: delete the depends_on edge from the blocked TODO instead.
func (p *Provider) Unlink(
	ctx context.Context,
	todo *types.TODO,
	targetRef string,
	relation types.RelationKind,
) error {
	id, version, target, err := p.relationshipOperands(ctx, todo, targetRef, relation)
	if err != nil {
		return err
	}
	if err := p.repository.DeleteRelationship(ctx, id, target.ID, nativeRelation(relation), version, mutationActor); err != nil {
		return err
	}
	return p.reloadTODO(ctx, todo, p.workDir)
}

// Links returns every edge touching todo from its own perspective, resolving
// each target so callers can render a title and status without a second query.
func (p *Provider) Links(ctx context.Context, todo *types.TODO) ([]todos.Link, error) {
	id, _, err := p.mutationIdentity(todo)
	if err != nil {
		return nil, err
	}
	relationships, err := p.repository.ListRelationships(ctx, id)
	if err != nil {
		return nil, err
	}
	links := make([]todos.Link, 0, len(relationships))
	for _, relationship := range relationships {
		target, err := p.repository.GetIssue(ctx, relationship.TargetIssueID)
		if err != nil {
			return nil, fmt.Errorf("resolve linked TODO %s: %w", relationship.TargetIssueID, err)
		}
		links = append(links, linkFromIssue(
			types.RelationKind(relationship.Relation), target, relationship.CreatedAt))
	}
	return links, nil
}

// relationshipOperands validates the writable relation and resolves both ends,
// returning the source issue's optimistic-lock version.
func (p *Provider) relationshipOperands(
	ctx context.Context,
	todo *types.TODO,
	targetRef string,
	relation types.RelationKind,
) (uuid.UUID, int64, *native.Issue, error) {
	parsed, err := types.ParseRelationKind(string(relation))
	if err != nil {
		return uuid.Nil, 0, nil, err
	}
	if parsed != relation {
		return uuid.Nil, 0, nil, fmt.Errorf("relation %q is not writable; use %s or %s",
			relation, types.RelationDependsOn, types.RelationRelatedTo)
	}
	id, version, err := p.mutationIdentity(todo)
	if err != nil {
		return uuid.Nil, 0, nil, err
	}
	target, err := p.repository.GetIssueByRef(ctx, p.workspace.ID, targetRef)
	if err != nil {
		return uuid.Nil, 0, nil, fmt.Errorf("resolve linked TODO %q: %w", targetRef, err)
	}
	return id, version, target, nil
}

func nativeRelation(relation types.RelationKind) native.RelationshipKind {
	return native.RelationshipKind(relation)
}

func linkFromIssue(relation types.RelationKind, issue *native.Issue, createdAt time.Time) todos.Link {
	id := issue.ID.String()
	return todos.Link{
		Relation:      relation,
		TargetID:      id,
		TargetShortID: id[:8],
		TargetTitle:   issue.Title,
		TargetStatus:  todoStatus(issue.Status, issue.ExecutionState),
		CreatedAt:     createdAt,
	}
}
