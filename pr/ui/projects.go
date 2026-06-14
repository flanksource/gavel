package ui

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/flanksource/commons/logger"
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
