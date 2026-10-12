package todos

import (
	"context"
	"fmt"
	"strings"

	"github.com/flanksource/gavel/todos/types"
)

// ChildrenDisposition is what closing a TODO does with its open children. A
// closed child is never touched: it stays attached to the closed parent.
type ChildrenDisposition string

const (
	ChildrenArchive ChildrenDisposition = "archive"
	ChildrenDetach  ChildrenDisposition = "detach"
)

// ChildrenChoice completes a provider's refusal to close a TODO that still has
// open children. It names the parameter as the HTTP API and the CLI spell it,
// because the refusal is shown verbatim on both.
const ChildrenChoice = "archive them too with children=archive, or make them top-level TODOs with children=detach (--children archive|detach)"

// ParseChildrenDisposition reads the `children` parameter. Empty means nobody
// chose, which only matters for a TODO that turns out to have open children.
func ParseChildrenDisposition(raw string) (ChildrenDisposition, error) {
	switch disposition := ChildrenDisposition(strings.ToLower(strings.TrimSpace(raw))); disposition {
	case "", ChildrenArchive, ChildrenDetach:
		return disposition, nil
	default:
		return "", fmt.Errorf("unknown children value %q: use archive or detach", raw)
	}
}

// ArchiveOptions decides what closing a TODO does with its open children.
type ArchiveOptions struct {
	// Children is the choice a person made. Left empty, a TODO with open
	// children is refused by the provider and nothing is written.
	Children ChildrenDisposition
	// Survivor is the TODO the work was handed to, for a retirement that cannot
	// ask anyone. The open children move under it, or under its parent when it is
	// itself a child; Children applies only when neither is open to take them.
	Survivor *types.TODO
}

// ChildrenOutcome is what happened to a closed TODO's open children.
type ChildrenOutcome struct {
	Children    types.TODOS
	Disposition ChildrenDisposition
	// Adopter is the TODO the children moved under, nil when they did not move.
	Adopter *types.TODO
}

// String reads as the end of a sentence about the closed TODO, and is empty
// when it had no open children.
func (o ChildrenOutcome) String() string {
	if len(o.Children) == 0 {
		return ""
	}
	refs := make([]string, 0, len(o.Children))
	for _, child := range o.Children {
		refs = append(refs, triageRef(child))
	}
	what := "made top-level TODOs"
	switch {
	case o.Adopter != nil:
		what = "moved to " + triageRef(o.Adopter)
	case o.Disposition == ChildrenArchive:
		what = "archived too"
	}
	return fmt.Sprintf("open children %s: %s", what, strings.Join(refs, ", "))
}

// Sentence is String as a sentence for the comment a retirement leaves on the
// TODO it closed.
func (o ChildrenOutcome) Sentence() string {
	if len(o.Children) == 0 {
		return ""
	}
	return "Its " + o.String() + "."
}

// Archive closes a TODO, settling its open children first so that a failure
// part-way leaves the parent open and nothing hidden under a closed one.
func Archive(ctx context.Context, provider Provider, todo *types.TODO, opts ArchiveOptions) (ChildrenOutcome, error) {
	outcome, err := SettleChildren(ctx, provider, todo, opts)
	if err != nil {
		return outcome, err
	}
	return outcome, provider.Delete(ctx, todo)
}

// SettleChildren applies opts to the open children of a TODO that is about to
// be closed, and reports what it did. Provider.Delete refuses a TODO that still
// has open children, so this is the only way one gets closed.
func SettleChildren(ctx context.Context, provider Provider, todo *types.TODO, opts ArchiveOptions) (ChildrenOutcome, error) {
	if _, err := ParseChildrenDisposition(string(opts.Children)); err != nil {
		return ChildrenOutcome{}, err
	}
	if opts.Children == "" && opts.Survivor == nil {
		return ChildrenOutcome{}, nil
	}
	// An empty ParentID filters nothing, which would settle the whole backlog.
	if strings.TrimSpace(todo.ID) == "" {
		return ChildrenOutcome{}, fmt.Errorf("%s has no id to look its children up by", triageRef(todo))
	}
	children, err := provider.List(ctx, DiscoveryFilters{
		ParentID: todo.ID, ExcludeStatuses: []types.Status{types.StatusCompleted},
	})
	if err != nil {
		return ChildrenOutcome{}, fmt.Errorf("list the open children of %s: %w", triageRef(todo), err)
	}
	if len(children) == 0 {
		return ChildrenOutcome{}, nil
	}
	adopter, err := childrenAdopter(ctx, provider, todo, opts.Survivor)
	if err != nil {
		return ChildrenOutcome{}, err
	}
	outcome := ChildrenOutcome{Disposition: opts.Children, Adopter: adopter}
	for _, child := range children {
		if adopter != nil && child.ID == adopter.ID {
			continue
		}
		if err := settleChild(ctx, provider, child, outcome); err != nil {
			return outcome, fmt.Errorf("settle child %s of %s: %w", triageRef(child), triageRef(todo), err)
		}
		outcome.Children = append(outcome.Children, child)
	}
	return outcome, nil
}

// childrenAdopter is the TODO that takes a retired TODO's open children: the
// survivor, or its parent when the survivor is a child, because the hierarchy
// is one level deep. It is nil when neither is open, and the children are then
// settled by the disposition instead.
//
// A survivor that is a child of the retired TODO is promoted here, through the
// caller's own pointer so its version stays current, and adopts its siblings.
func childrenAdopter(ctx context.Context, provider Provider, retired, survivor *types.TODO) (*types.TODO, error) {
	if survivor == nil {
		return nil, nil
	}
	adopter := survivor
	if survivor.ParentID != "" && survivor.ParentID != retired.ID {
		parent, err := provider.Get(ctx, survivor.ParentID)
		if err != nil {
			return nil, fmt.Errorf("resolve the parent of %s: %w", triageRef(survivor), err)
		}
		adopter = parent
	}
	if adopter.Status == types.StatusCompleted {
		return nil, nil
	}
	if survivor.ParentID != retired.ID {
		return adopter, nil
	}
	if err := setParent(ctx, provider, survivor, ""); err != nil {
		return nil, fmt.Errorf("make %s top-level so it can take the children of %s: %w",
			triageRef(survivor), triageRef(retired), err)
	}
	return survivor, nil
}

func settleChild(ctx context.Context, provider Provider, child *types.TODO, outcome ChildrenOutcome) error {
	switch {
	case outcome.Adopter != nil:
		return setParent(ctx, provider, child, outcome.Adopter.ID)
	case outcome.Disposition == ChildrenArchive:
		return provider.Delete(ctx, child)
	case outcome.Disposition == ChildrenDetach:
		return setParent(ctx, provider, child, "")
	default:
		return fmt.Errorf("nobody chose what happens to it; %s", ChildrenChoice)
	}
}

func setParent(ctx context.Context, provider Provider, todo *types.TODO, parentRef string) error {
	parents, ok := provider.(ParentProvider)
	if !ok {
		return fmt.Errorf("this TODO provider has no hierarchy to move %s in", triageRef(todo))
	}
	return parents.SetParent(ctx, todo, parentRef)
}
