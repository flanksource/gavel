package runtime

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/native"
	"github.com/flanksource/gavel/todos/types"
	"github.com/google/uuid"
)

var _ todos.ParentProvider = (*Provider)(nil)

// SetParent makes todo a child of the TODO named by parentRef, or top-level
// again when parentRef is empty. The single-level rules are enforced by the
// repository inside one transaction.
func (p *Provider) SetParent(ctx context.Context, todo *types.TODO, parentRef string) error {
	id, version, err := p.mutationIdentity(todo)
	if err != nil {
		return err
	}
	var parentID *uuid.UUID
	if ref := strings.TrimSpace(parentRef); ref != "" {
		parent, err := p.parentIssue(ctx, ref)
		if err != nil {
			return err
		}
		parentID = &parent.ID
	}
	issue, err := p.repository.SetIssueParent(ctx, id, parentID, version, mutationActor)
	if err != nil {
		return err
	}
	return p.replaceTODO(ctx, todo, issue, p.workDir)
}

// applyLineage resolves where a new TODO belongs and records it on the issue
// input: an explicit parent wins; otherwise a TODO created inside a run becomes a
// child of the running TODO's top-level TODO, unless the request opts out or the
// running TODO is not in this workspace.
//
// The origin is recorded either way, as it was given. It is only read when it
// is the thing deciding the parent, so an origin this workspace cannot make
// sense of never fails a create that names its own parent or asks for none.
func (p *Provider) applyLineage(ctx context.Context, request todos.CreateRequest, input *native.CreateIssueInput) error {
	parentRef := strings.TrimSpace(request.Parent)
	if parentRef != "" && request.NoParent {
		return fmt.Errorf("a parent (%q) and no-parent cannot both be set", parentRef)
	}
	if request.Origin != nil {
		input.Origin = native.IssueOrigin{
			IssueID:   strings.TrimSpace(request.Origin.IssueID),
			SessionID: strings.TrimSpace(request.Origin.SessionID),
		}
	}
	if parentRef != "" {
		parent, err := p.parentIssue(ctx, parentRef)
		if err != nil {
			return err
		}
		input.ParentID = &parent.ID
		return nil
	}
	if request.NoParent || input.Origin.IssueID == "" {
		return nil
	}
	running, found, err := p.originIssue(ctx, input.Origin.IssueID)
	if err != nil || !found {
		return err
	}
	input.Origin.IssueID = running.ID.String()
	top := running
	if running.ParentID != nil {
		if top, err = p.repository.GetIssue(ctx, *running.ParentID); err != nil {
			return fmt.Errorf("resolve the parent of origin TODO %s: %w", running.ID, err)
		}
	}
	// A closed TODO takes no children, and a run that outlived it must still be
	// able to file what it found.
	if !top.Status.IsClosed() {
		input.ParentID = &top.ID
	}
	return nil
}

// originIssue resolves the TODO a run was working on, as an id, short id or
// alias in this workspace. The reference is free text a run exported, so one
// this workspace does not know — another workspace's TODO, another database's,
// or no TODO at all — is reported as not found rather than as a failure.
func (p *Provider) originIssue(ctx context.Context, ref string) (*native.Issue, bool, error) {
	issue, err := p.repository.GetIssueByRef(ctx, p.workspace.ID, ref)
	switch {
	case err == nil:
		return issue, true, nil
	case errors.Is(err, native.ErrNotFound), errors.Is(err, native.ErrInvalidInput),
		errors.Is(err, native.ErrAmbiguousReference):
		return nil, false, nil
	default:
		return nil, false, fmt.Errorf("resolve origin TODO %q: %w", ref, err)
	}
}

// parentIssue resolves a parent ref in this workspace and refuses a child,
// naming the TODO the caller most likely meant, or a closed TODO. The
// repository enforces both again inside the write.
func (p *Provider) parentIssue(ctx context.Context, ref string) (*native.Issue, error) {
	parent, err := p.repository.GetIssueByRef(ctx, p.workspace.ID, ref)
	if err != nil {
		return nil, fmt.Errorf("resolve parent TODO %q: %w", ref, err)
	}
	short := types.ShortID(parent.ID.String())
	if parent.ParentID != nil {
		grandparent := types.ShortID(parent.ParentID.String())
		return nil, fmt.Errorf("%w: %s (%q) is a child of %s, and a child cannot have children; use %s as the parent",
			native.ErrInvalidParent, short, parent.Title, grandparent, grandparent)
	}
	if parent.Status.IsClosed() {
		return nil, fmt.Errorf("%w: %s (%q) is closed, and a closed TODO cannot take children; reopen it first",
			native.ErrInvalidParent, short, parent.Title)
	}
	return parent, nil
}
