package ui

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/flanksource/commons/logger"
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
	// TodoProvider pins which todo backend this workspace uses ("grite" or
	// "todos"). Empty means auto-detect (.todos files if present, else Grite).
	TodoProvider string `json:"todoProvider,omitempty"`
}

var projectsPath = filepath.Join(os.Getenv("HOME"), ".config", "gavel", "projects.json")

// LoadProjects reads ~/.config/gavel/projects.json. A missing file is the
// normal "no projects configured yet" state and returns nil with no error,
// mirroring LoadSettings. A present-but-unparseable file is logged and treated
// as empty so a corrupt file never wedges the UI.
func LoadProjects() []Project {
	data, err := os.ReadFile(projectsPath)
	if err != nil {
		return nil
	}
	var ps []Project
	if err := json.Unmarshal(data, &ps); err != nil {
		logger.Warnf("failed to parse %s: %v", projectsPath, err)
		return nil
	}
	return ps
}

// SaveProjects writes the project list back to ~/.config/gavel/projects.json.
func SaveProjects(ps []Project) {
	if err := os.MkdirAll(filepath.Dir(projectsPath), 0o755); err != nil {
		logger.Warnf("failed to create config dir: %v", err)
		return
	}
	data, err := json.MarshalIndent(ps, "", "  ")
	if err != nil {
		logger.Warnf("failed to marshal projects: %v", err)
		return
	}
	if err := os.WriteFile(projectsPath, data, 0o644); err != nil {
		logger.Warnf("failed to write %s: %v", projectsPath, err)
	}
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
func GetProject(name string) (Project, bool) {
	return projectByName(LoadProjects(), name)
}

// CreateProject persists a new project, failing if required fields are missing
// (ErrProjectInvalid) or the name is already taken (ErrProjectExists).
func CreateProject(p Project) error {
	if p.Name == "" || p.Dir == "" {
		return ErrProjectInvalid
	}
	ps := LoadProjects()
	if _, ok := projectByName(ps, p.Name); ok {
		return fmt.Errorf("%w: %q", ErrProjectExists, p.Name)
	}
	SaveProjects(append(ps, p))
	return nil
}

// UpdateProject replaces the named project. The name is the entity id, so the
// path/argument is authoritative and the body cannot rename it. Missing project
// → ErrProjectNotFound; empty dir → ErrProjectInvalid.
func UpdateProject(name string, p Project) error {
	ps := LoadProjects()
	if _, ok := projectByName(ps, name); !ok {
		return fmt.Errorf("%w: %q", ErrProjectNotFound, name)
	}
	p.Name = name
	if p.Dir == "" {
		return ErrProjectInvalid
	}
	SaveProjects(upsertProject(ps, p))
	return nil
}

// DeleteProject removes the named project, returning ErrProjectNotFound if it
// does not exist.
func DeleteProject(name string) error {
	ps := LoadProjects()
	next, ok := deleteProject(ps, name)
	if !ok {
		return fmt.Errorf("%w: %q", ErrProjectNotFound, name)
	}
	SaveProjects(next)
	return nil
}
