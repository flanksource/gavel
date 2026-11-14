package agents

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"

	"github.com/flanksource/gavel/projects"
)

// PathTarget is where the threads of a file or directory live: under the
// gavel://project/<name> scope of the most specific project holding it, at
// <scope>/files/<rel>. URI is empty for the project root itself.
type PathTarget struct {
	Project string
	// Dir is the project's canonical root, where an agent working a thread runs.
	Dir   string
	Scope string
	URI   string
	IsDir bool
}

// ProjectScope is the comment scope of the project named name.
func ProjectScope(name string) string { return "gavel://project/" + url.PathEscape(name) }

// ProjectPathTarget is the inverse of ProjectFile: the project scope and file
// URI path lies at. Outside every project it fails with projects.ErrNotFound.
func ProjectPathTarget(ctx context.Context, catalog projects.Service, path string) (PathTarget, error) {
	resolved, err := projects.CanonicalPath(path)
	if err != nil {
		return PathTarget{}, fmt.Errorf("resolve %s: %w", path, err)
	}
	// A path that does not exist yet (a file about to be created) is a file.
	info, err := os.Stat(resolved)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return PathTarget{}, err
	}
	isDir := err == nil && info.IsDir()
	dir := resolved
	if !isDir {
		dir = filepath.Dir(resolved)
	}
	project, err := catalog.ForDir(ctx, dir)
	if err != nil {
		return PathTarget{}, err
	}
	root, err := projects.CanonicalPath(project.ResolvedDir())
	if err != nil {
		return PathTarget{}, fmt.Errorf("resolve project %q directory: %w", project.Name, err)
	}
	rel, err := filepath.Rel(root, resolved)
	if err != nil {
		return PathTarget{}, fmt.Errorf("place %s under project %q: %w", path, project.Name, err)
	}
	target := PathTarget{Project: project.Name, Dir: root, Scope: ProjectScope(project.Name), IsDir: isDir}
	if rel != "." {
		target.URI = target.Scope + "/files/" + filepath.ToSlash(rel)
	}
	return target, nil
}
