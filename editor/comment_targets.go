package editor

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/flanksource/gavel/comments/agents"
	"github.com/flanksource/gavel/projects"
)

// fileTarget is where a file's threads live: under the scope of the project
// holding it, as pr/ui's comment host and `gavel comments` address them.
type fileTarget struct {
	Project string
	// Dir is the project's root, where an agent working the thread runs.
	Dir   string
	Scope string
	URI   string
}

// NoProjectError is a file outside every configured project: its threads have
// no scope until a project is created for it, as Suggestion proposes.
type NoProjectError struct {
	Path       string           `json:"path"`
	Suggestion projects.Project `json:"suggestion"`
}

func (e NoProjectError) Error() string {
	return fmt.Sprintf("%s is not inside a gavel project; create one for %s to comment on it", e.Path, e.Suggestion.Dir)
}

func projectScope(name string) string { return "gavel://project/" + url.PathEscape(name) }

// targetOf resolves the most specific project holding path, as ProjectForDir does for every gavel surface.
func targetOf(ctx context.Context, path string) (fileTarget, error) {
	catalog, err := projects.Shared(ctx)
	if err != nil {
		return fileTarget{}, err
	}
	project, err := catalog.ForDir(ctx, filepath.Dir(path))
	if errors.Is(err, projects.ErrNotFound) {
		return fileTarget{}, NoProjectError{Path: path, Suggestion: suggestProject(ctx, path)}
	}
	if err != nil {
		return fileTarget{}, err
	}
	root, err := projects.CanonicalPath(project.ResolvedDir())
	if err != nil {
		return fileTarget{}, fmt.Errorf("resolve project %q directory: %w", project.Name, err)
	}
	resolved, err := projects.CanonicalPath(path)
	if err != nil {
		return fileTarget{}, fmt.Errorf("resolve %s: %w", path, err)
	}
	rel, err := filepath.Rel(root, resolved)
	if err != nil {
		return fileTarget{}, fmt.Errorf("place %s under project %q: %w", path, project.Name, err)
	}
	scope := projectScope(project.Name)
	return fileTarget{Project: project.Name, Dir: root, Scope: scope, URI: scope + "/files/" + filepath.ToSlash(rel)}, nil
}

// suggestProject proposes the file's git work tree, named after its directory,
// or the file's own directory outside git.
func suggestProject(ctx context.Context, path string) projects.Project {
	dir := filepath.Dir(path)
	if out, err := exec.CommandContext(ctx, "git", "-C", dir, "rev-parse", "--show-toplevel").Output(); err == nil {
		dir = strings.TrimSpace(string(out))
	}
	return projects.Project{Name: filepath.Base(dir), Dir: dir}
}

// createProject serves POST /api/projects {name, dir}: the project a file
// outside every project needs before it can be commented on.
func (s *Server) createProject(w http.ResponseWriter, r *http.Request) {
	var project projects.Project
	if !decode(w, r, &project) {
		return
	}
	project = projects.Project{Name: strings.TrimSpace(project.Name), Dir: strings.TrimSpace(project.Dir)}
	catalog, err := projects.Shared(r.Context())
	if err != nil {
		respond(w, nil, err)
		return
	}
	created, err := catalog.Create(r.Context(), project)
	switch {
	case errors.Is(err, projects.ErrExists):
		http.Error(w, err.Error(), http.StatusConflict)
	case err != nil:
		respond(w, nil, err)
	default:
		respond(w, created, nil)
	}
}

// editedFile is the edited file a project target names, for checks only the editor can make.
func (s *Store) editedFile(ctx context.Context, scope, uri string) (Document, error) {
	catalog, err := projects.Shared(ctx)
	if err != nil {
		return Document{}, err
	}
	_, _, path, err := agents.ProjectFile(ctx, catalog, scope, uri)
	if err != nil {
		return Document{}, err
	}
	for _, file := range s.files {
		if filepath.Join(file.root, file.name) == path {
			return s.Read(file.ID)
		}
	}
	return Document{}, fmt.Errorf("%w: %s is not one of the edited files", errInvalidTarget, path)
}

var errInvalidTarget = errors.New("comment target is not an edited file")
