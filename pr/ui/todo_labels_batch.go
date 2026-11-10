package ui

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"

	"github.com/flanksource/gavel/todos/labels"
	todoruntime "github.com/flanksource/gavel/todos/runtime"
)

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
				options[i].RootPath = filepath.Join(s.workspaces.Sources.DataDir, "workspaces", s.orgID().String(), id.String())
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
	json.NewEncoder(w).Encode(struct {
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
	projects, err := LoadProjects(LoadProjectsOptions{IncludeHidden: true})
	if err != nil {
		return nil, err
	}
	result := make([]projectInfo, len(projects))
	for i, project := range projects {
		result[i] = projectInfo{Name: project.Name, Dir: project.ResolvedDir(), Repos: project.Repos}
	}
	return result, nil
}
