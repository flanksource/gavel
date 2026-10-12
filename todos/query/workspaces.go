package query

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/types"
)

// Workspace is one directory a list reads from. Name is the registered project
// it belongs to, and is empty for a directory named directly rather than
// through the projects list.
type Workspace struct {
	Name string
	Dir  string
}

// Short is the workspace's short project name, the name a list is filtered by.
func (w Workspace) Short() string {
	return ShortProjectName(w.Name)
}

// ShortProjectName slugs a registered project name into the form lists show
// and filters take: lowercased, with each run of anything but letters and
// digits collapsed to one "-". "OM Digital Frontend" becomes
// "om-digital-frontend".
func ShortProjectName(name string) string {
	var slug strings.Builder
	pendingDash := false
	for _, r := range strings.ToLower(name) {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			pendingDash = slug.Len() > 0
			continue
		}
		if pendingDash {
			slug.WriteByte('-')
			pendingDash = false
		}
		slug.WriteRune(r)
	}
	return slug.String()
}

// IndexByShortName keys named workspaces by their short project name. Two
// projects that slug to the same short name, or a name with no letters or
// digits to slug, are an error: a filter by that name could not say which
// project it meant.
func IndexByShortName(workspaces []Workspace) (map[string]Workspace, error) {
	index := make(map[string]Workspace, len(workspaces))
	for _, ws := range workspaces {
		short := ws.Short()
		if short == "" {
			return nil, fmt.Errorf("project %q has no letters or digits to form a short name from", ws.Name)
		}
		if existing, ok := index[short]; ok {
			return nil, fmt.Errorf("projects %q and %q share the short name %q; rename one of them", existing.Name, ws.Name, short)
		}
		index[short] = ws
	}
	return index, nil
}

func (w Workspace) label() string {
	if w.Name == "" {
		return w.Dir
	}
	return fmt.Sprintf("project %q (%s)", w.Name, w.Dir)
}

// OpenProvider opens the TODO provider that owns a workspace directory.
type OpenProvider func(ctx context.Context, dir string) (todos.Provider, error)

// ListWorkspaces lists the TODOs of every workspace, reading a directory that
// two projects share only once and tagging each row with the project it came
// from.
//
// Every open and list failure is returned alongside the rows that did load:
// treating an unreachable workspace as an empty one would be a false success.
func ListWorkspaces(ctx context.Context, workspaces []Workspace, open OpenProvider, filters todos.DiscoveryFilters) (types.TODOS, error) {
	seen := map[string]struct{}{}
	var listed types.TODOS
	var failures []error
	for _, ws := range workspaces {
		dir := strings.TrimSpace(ws.Dir)
		if dir == "" {
			failures = append(failures, fmt.Errorf("list native TODOs for project %q: workspace directory is empty", ws.Name))
			continue
		}
		key := filepath.Clean(dir)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}

		provider, err := open(ctx, dir)
		if err != nil {
			failures = append(failures, fmt.Errorf("open native TODO workspace for %s: %w", ws.label(), err))
			continue
		}
		items, err := provider.List(ctx, filters)
		if err != nil {
			failures = append(failures, fmt.Errorf("list native TODOs for %s: %w", ws.label(), err))
			continue
		}
		for _, todo := range items {
			if todo == nil {
				continue
			}
			if ws.Name != "" {
				todo.Workspace = ws.Name
			}
			if todo.CWD == "" {
				todo.CWD = dir
			}
		}
		listed = append(listed, items...)
	}
	listed.Sort()
	return listed, errors.Join(failures...)
}
