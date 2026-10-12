package entity

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"slices"
	"sort"
	"strings"

	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/native"
	"github.com/flanksource/gavel/todos/query"
	"github.com/flanksource/gavel/todos/types"
)

// parentMatch is a TODO a parent reference names, and the workspace it is in.
type parentMatch struct {
	workspace query.Workspace
	todo      *types.TODO
}

func (m parentMatch) String() string {
	return fmt.Sprintf("%s %q in %s", m.todo.DisplayID(), m.todo.Title, workspaceLabel(m.workspace))
}

func workspaceLabel(ws query.Workspace) string {
	if short := ws.Short(); short != "" {
		return short
	}
	return ws.Dir
}

// parentMatches collects the TODOs a reference resolved to, once each.
type parentMatches struct {
	found []parentMatch
	// misses are the reasons workspaces gave for not knowing the reference as an
	// id or alias: the detail a mistyped reference needs.
	misses []string
}

func (m *parentMatches) add(ws query.Workspace, todo *types.TODO) {
	if !slices.ContainsFunc(m.found, func(match parentMatch) bool { return match.todo.ID == todo.ID }) {
		m.found = append(m.found, parentMatch{workspace: ws, todo: todo})
	}
}

// resolveParent turns a list's parent reference into the one TODO it names.
//
// A list filters rows by parent id, and only a provider can say which id a
// reference means: comparing the reference against ids answered an alias, a
// title or a mistyped id with an empty list, which reads as "no children". So
// the reference is resolved first — as an id, short id or alias in each
// workspace the list covers, and as a title only when none of them knows it —
// and a reference naming nothing, or TODOs in more than one workspace, is the
// caller's mistake rather than an empty result.
func (d Deps) resolveParent(ctx context.Context, workspaces []query.Workspace, ref string) (parentMatch, error) {
	opened, err := d.openWorkspaces(ctx, workspaces)
	if err != nil {
		return parentMatch{}, err
	}
	var matches parentMatches
	for _, ws := range opened {
		todo, err := ws.provider.Get(ctx, ref)
		switch {
		case err == nil:
			matches.add(ws.workspace, todo)
		// Too short to be an id prefix is still a possible title.
		case errors.Is(err, native.ErrNotFound), errors.Is(err, native.ErrInvalidInput):
			if miss := err.Error(); !slices.Contains(matches.misses, miss) {
				matches.misses = append(matches.misses, miss)
			}
		case errors.Is(err, native.ErrAmbiguousReference):
			return parentMatch{}, rejected(fmt.Errorf("parent %q in %s: %w", ref, workspaceLabel(ws.workspace), err))
		default:
			return parentMatch{}, fmt.Errorf("resolve parent %q in %s: %w", ref, workspaceLabel(ws.workspace), err)
		}
	}
	if len(matches.found) == 0 {
		for _, ws := range opened {
			listed, err := ws.provider.List(ctx, todos.DiscoveryFilters{})
			if err != nil {
				return parentMatch{}, fmt.Errorf("list TODOs in %s to match parent %q by title: %w", workspaceLabel(ws.workspace), ref, err)
			}
			for _, todo := range query.TitleMatches(listed, ref) {
				matches.add(ws.workspace, todo)
			}
		}
	}
	return matches.only(ref, workspaces)
}

// only is the single TODO the reference resolved to; none and several are both
// rejected as the caller's mistake.
func (m parentMatches) only(ref string, workspaces []query.Workspace) (parentMatch, error) {
	switch len(m.found) {
	case 1:
		return m.found[0], nil
	case 0:
		scope := fmt.Sprintf("any of the %d listed projects", len(workspaces))
		if len(workspaces) == 1 {
			scope = workspaceLabel(workspaces[0])
		}
		reason := fmt.Sprintf("parent %q does not name a TODO in %s", ref, scope)
		if len(m.misses) > 0 {
			reason += ": " + strings.Join(m.misses, "; ")
		}
		return parentMatch{}, rejected(errors.New(reason))
	default:
		named := make([]string, len(m.found))
		for i, match := range m.found {
			named[i] = match.String()
		}
		sort.Strings(named)
		return parentMatch{}, rejected(fmt.Errorf("parent %q names %d TODOs: %s; name the parent by its full id",
			ref, len(m.found), strings.Join(named, ", ")))
	}
}

type openedWorkspace struct {
	workspace query.Workspace
	provider  todos.Provider
}

// openWorkspaces opens each distinct workspace directory once, the way
// query.ListWorkspaces reads them: two projects sharing a directory hold the
// same TODOs, and resolving against both would call one TODO ambiguous.
func (d Deps) openWorkspaces(ctx context.Context, workspaces []query.Workspace) ([]openedWorkspace, error) {
	seen := map[string]struct{}{}
	opened := make([]openedWorkspace, 0, len(workspaces))
	for _, ws := range workspaces {
		dir := strings.TrimSpace(ws.Dir)
		if dir == "" {
			return nil, fmt.Errorf("resolve parent TODO in project %q: workspace directory is empty", ws.Name)
		}
		key := filepath.Clean(dir)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		provider, err := d.OpenProvider(ctx, dir)
		if err != nil {
			return nil, fmt.Errorf("open native TODO workspace %s to resolve a parent: %w", workspaceLabel(ws), err)
		}
		opened = append(opened, openedWorkspace{workspace: ws, provider: provider})
	}
	return opened, nil
}
