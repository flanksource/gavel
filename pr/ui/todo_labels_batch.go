package ui

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/flanksource/gavel/todos/labels"
	todoruntime "github.com/flanksource/gavel/todos/runtime"
)

type todoBatchRequest struct {
	Dirs *[]string `json:"dirs"`
}

type todoBatchError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func todoBatchErrorFor(err error) *todoBatchError {
	code := "load_failed"
	switch {
	case errors.Is(err, context.Canceled):
		code = "canceled"
	case errors.Is(err, context.DeadlineExceeded):
		code = "deadline_exceeded"
	}
	return &todoBatchError{Code: code, Message: err.Error()}
}

// validateTodoBatchDirs checks dirs are distinct, non-blank workspaces and, when paths is set,
// normalized absolute paths. A hosted dashboard's workspaces are repo:<id> names, which each
// workspace's load checks against the organization's repositories.
func validateTodoBatchDirs(dirs []string, paths bool) error {
	seen := make(map[string]int, len(dirs))
	for i, dir := range dirs {
		if dir == "" || strings.TrimSpace(dir) != dir {
			return fmt.Errorf("dirs[%d] must be a non-blank normalized workspace path", i)
		}
		if paths && !filepath.IsAbs(dir) {
			return fmt.Errorf("dirs[%d] must be absolute: %q", i, dir)
		}
		if paths && filepath.Clean(dir) != dir {
			return fmt.Errorf("dirs[%d] must be normalized: %q", i, dir)
		}
		if first, ok := seen[dir]; ok {
			return fmt.Errorf("dirs[%d] duplicates dirs[%d]: %q", i, first, dir)
		}
		seen[dir] = i
	}
	return nil
}

type todoLabelBatchResult struct {
	Dir    string                  `json:"dir"`
	Labels *todoruntime.LabelIndex `json:"labels,omitempty"`
	Error  *todoBatchError         `json:"error,omitempty"`
}

func (s *Server) handleTodoLabelsBatch(w http.ResponseWriter, r *http.Request) {
	var request todoBatchRequest
	if err := decodeTodoRequest(r, &request); err != nil {
		writeTodoError(w, http.StatusBadRequest, err)
		return
	}
	if request.Dirs == nil {
		writeTodoError(w, http.StatusBadRequest, fmt.Errorf("dirs is required"))
		return
	}
	if err := validateTodoBatchDirs(*request.Dirs, !s.hosted); err != nil {
		writeTodoError(w, http.StatusBadRequest, err)
		return
	}
	results := make([]todoLabelBatchResult, len(*request.Dirs))
	options := make([]todoruntime.WorkspaceOptions, len(results))
	projects, err := s.labelProjects(r)
	if err != nil {
		writeTodoError(w, http.StatusInternalServerError, err)
		return
	}
	for i, dir := range *request.Dirs {
		results[i].Dir = dir
		for _, project := range projects {
			if project.Dir != dir {
				continue
			}
			options[i] = todoruntime.WorkspaceOptions{Name: project.Name, RootPath: project.Dir, Repositories: project.Repos}
			if s.workspaces != nil {
				id, ok := s.workspaces.repoID(s.orgID(), dir)
				if !ok {
					break
				}
				options[i].RootPath = s.workspaces.rootPath(s.orgID(), id)
			}
			break
		}
		if options[i].RootPath == "" {
			results[i].Error = todoBatchErrorFor(fmt.Errorf("workspace %q is not configured for this organization", dir))
		}
	}
	indexes, err := todoruntime.ReadLabelIndexes(r.Context(), s.orgDB, options)
	if err != nil {
		writeTodoError(w, http.StatusInternalServerError, err)
		return
	}
	for i := range results {
		if results[i].Error != nil {
			continue
		}
		if indexes[i].Err != nil {
			results[i].Error = todoBatchErrorFor(indexes[i].Err)
		} else {
			results[i].Labels = &indexes[i]
		}
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(struct {
		Results []todoLabelBatchResult `json:"results"`
		Palette []string               `json:"palette"`
	}{results, labels.PaletteStrings()}) //nolint:errcheck
}

func (s *Server) labelProjects(r *http.Request) ([]projectInfo, error) {
	if s.workspaceData != nil {
		return s.catalogProjects()
	}
	if s.workspaces != nil {
		return s.workspaces.Projects(r.Context(), s.orgID())
	}
	if s.hosted {
		return nil, errNoHostedWorkspaces
	}
	projects, err := LoadProjects()
	if err != nil {
		return nil, err
	}
	result := make([]projectInfo, len(projects))
	for i, project := range projects {
		result[i] = projectInfo{Name: project.Name, Dir: project.ResolvedDir(), Repos: project.Repos}
	}
	return result, nil
}
