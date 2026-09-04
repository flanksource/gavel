package ui

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	todoruntime "github.com/flanksource/gavel/todos/runtime"
)

// Project CRUD errors. Callers map these to a transport status (HTTP code / CLI
// exit) with errors.Is, so the create/update/delete semantics live in one place
// for both the dashboard API and the `gavel projects` command.
var (
	ErrProjectNotFound = errors.New("unknown project")
	ErrProjectExists   = errors.New("project already exists")
	ErrProjectInvalid  = errors.New("name and dir are required")
)

// Project associates one or more GitHub repos with a local workspace directory
// on disk. The directory is where Gavel discovers a Procfile so the PR UI can
// supervise the project's processes alongside its pull requests.
type Project struct {
	Name  string   `json:"name"`
	Dir   string   `json:"dir"`
	Repos []string `json:"repos"`
}

var projectsPath = filepath.Join(os.Getenv("HOME"), ".config", "gavel", "projects.json")

// withRepos returns the project with a non-nil Repos list. Go marshals a nil
// slice as JSON null, and every consumer of the catalog — the dashboard's
// Project type, the CLI table, the TODO runtime — treats repos as a list, so a
// project registered without repos must stay an empty list on the wire and on
// disk rather than becoming null.
func (p Project) withRepos() Project {
	if p.Repos == nil {
		p.Repos = []string{}
	}
	return p
}

// normalizeRepos normalizes a whole catalog. Both ends of the file are
// normalized so a catalog written by an older gavel is healed on read.
func normalizeRepos(ps []Project) []Project {
	for i := range ps {
		ps[i] = ps[i].withRepos()
	}
	return ps
}

// LoadProjects reads ~/.config/gavel/projects.json. A missing file is the
// normal "no projects configured yet" state. Read and parse failures name the
// exact catalog file instead of being mistaken for an empty configuration.
func LoadProjects() ([]Project, error) {
	data, err := os.ReadFile(projectsPath)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read gavel projects file %s: %w", projectsPath, err)
	}
	var ps []Project
	if err := json.Unmarshal(data, &ps); err != nil {
		return nil, fmt.Errorf("parse gavel projects file %s: %w", projectsPath, err)
	}
	return normalizeRepos(ps), nil
}

// SaveProjects writes the project list back to ~/.config/gavel/projects.json.
func SaveProjects(ps []Project) error {
	if err := os.MkdirAll(filepath.Dir(projectsPath), 0o755); err != nil {
		return fmt.Errorf("create gavel projects directory for %s: %w", projectsPath, err)
	}
	data, err := json.MarshalIndent(normalizeRepos(ps), "", "  ")
	if err != nil {
		return fmt.Errorf("marshal gavel projects file %s: %w", projectsPath, err)
	}
	if err := os.WriteFile(projectsPath, data, 0o644); err != nil {
		return fmt.Errorf("write gavel projects file %s: %w", projectsPath, err)
	}
	return nil
}

// ResolvedDir expands a leading "~" in Dir to the user's home directory so the
// stored path can be portable across machines. Other paths are returned as-is.
func (p Project) ResolvedDir() string {
	dir := strings.TrimSpace(p.Dir)
	if dir == "~" || strings.HasPrefix(dir, "~/") {
		if home, err := os.UserHomeDir(); err == nil {
			return filepath.Join(home, strings.TrimPrefix(dir[1:], "/"))
		}
	}
	return dir
}

// WorkspaceOptions projects the catalog record into the native TODO runtime
// without giving the runtime a second configuration source.
func (p Project) WorkspaceOptions() todoruntime.WorkspaceOptions {
	return todoruntime.WorkspaceOptions{
		Name:         p.Name,
		RootPath:     p.ResolvedDir(),
		Repositories: append([]string(nil), p.Repos...),
	}
}

// ProjectForRepo returns the first project whose Repos list contains repo.
func ProjectForRepo(ps []Project, repo string) (Project, bool) {
	for _, p := range ps {
		for _, r := range p.Repos {
			if r == repo {
				return p, true
			}
		}
	}
	return Project{}, false
}

// projectByName returns the project with the given name.
func projectByName(ps []Project, name string) (Project, bool) {
	for _, p := range ps {
		if p.Name == name {
			return p, true
		}
	}
	return Project{}, false
}

// upsertProject replaces the project with a matching name (or appends a new
// one) and returns the updated list.
func upsertProject(ps []Project, p Project) []Project {
	for i := range ps {
		if ps[i].Name == p.Name {
			ps[i] = p
			return ps
		}
	}
	return append(ps, p)
}

// deleteProject removes the project with the given name and reports whether one
// was found. The three-index slice keeps the result from aliasing the input's
// backing array.
func deleteProject(ps []Project, name string) ([]Project, bool) {
	for i := range ps {
		if ps[i].Name == name {
			return append(ps[:i:i], ps[i+1:]...), true
		}
	}
	return ps, false
}

// GetProject returns the stored project with the given name.
func GetProject(name string) (Project, error) {
	projects, err := LoadProjects()
	if err != nil {
		return Project{}, err
	}
	if project, ok := projectByName(projects, name); ok {
		return project, nil
	}
	return Project{}, fmt.Errorf("%w: %q was not found in %s", ErrProjectNotFound, name, projectsPath)
}

// ProjectForDir resolves dir to the most-specific configured project root.
func ProjectForDir(dir string) (Project, error) {
	absolute, err := filepath.Abs(strings.TrimSpace(dir))
	if err != nil {
		return Project{}, fmt.Errorf("resolve project directory %q using %s: %w", dir, projectsPath, err)
	}
	projects, err := LoadProjects()
	if err != nil {
		return Project{}, err
	}
	var match Project
	matchLength := -1
	for _, project := range projects {
		root, err := filepath.Abs(project.ResolvedDir())
		if err != nil {
			return Project{}, fmt.Errorf("resolve project %q directory %q from %s: %w", project.Name, project.Dir, projectsPath, err)
		}
		relative, err := filepath.Rel(filepath.Clean(root), filepath.Clean(absolute))
		if err != nil {
			return Project{}, fmt.Errorf("compare project %q directory %q using %s: %w", project.Name, root, projectsPath, err)
		}
		if relative != "." && (relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator))) {
			continue
		}
		if len(root) > matchLength {
			match = project
			matchLength = len(root)
		}
	}
	if matchLength >= 0 {
		return match, nil
	}
	return Project{}, fmt.Errorf("%w: project for directory %q was not found in %s", ErrProjectNotFound, filepath.Clean(absolute), projectsPath)
}

// CreateProject persists a new project, failing if required fields are missing
// (ErrProjectInvalid) or the name is already taken (ErrProjectExists).
func CreateProject(p Project) error {
	if p.Name == "" || p.Dir == "" {
		return ErrProjectInvalid
	}
	ps, err := LoadProjects()
	if err != nil {
		return err
	}
	if _, ok := projectByName(ps, p.Name); ok {
		return fmt.Errorf("%w: %q", ErrProjectExists, p.Name)
	}
	return SaveProjects(append(ps, p))
}

// UpdateProject replaces the named project. The name is the entity id, so the
// path/argument is authoritative and the body cannot rename it. Missing project
// → ErrProjectNotFound; empty dir → ErrProjectInvalid.
func UpdateProject(name string, p Project) error {
	ps, err := LoadProjects()
	if err != nil {
		return err
	}
	if _, ok := projectByName(ps, name); !ok {
		return fmt.Errorf("%w: %q was not found in %s", ErrProjectNotFound, name, projectsPath)
	}
	p.Name = name
	if p.Dir == "" {
		return ErrProjectInvalid
	}
	return SaveProjects(upsertProject(ps, p))
}

// DeleteProject removes the named project, returning ErrProjectNotFound if it
// does not exist.
func DeleteProject(name string) error {
	ps, err := LoadProjects()
	if err != nil {
		return err
	}
	next, ok := deleteProject(ps, name)
	if !ok {
		return fmt.Errorf("%w: %q was not found in %s", ErrProjectNotFound, name, projectsPath)
	}
	return SaveProjects(next)
}
