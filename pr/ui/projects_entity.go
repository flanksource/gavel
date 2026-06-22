package ui

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/flanksource/gavel/procfile"
)

// statusForProjectErr maps the shared CRUD sentinel errors onto HTTP codes.
func statusForProjectErr(err error) int {
	switch {
	case errors.Is(err, ErrProjectNotFound):
		return http.StatusNotFound
	case errors.Is(err, ErrProjectExists):
		return http.StatusConflict
	case errors.Is(err, ErrProjectInvalid):
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}

// newProjectInfo resolves a stored Project into the wire shape returned by the
// projects entity: the directory is ~-expanded, hasProcfile reflects the
// directory's current contents, and todo counts are scoped to the workspace.
func newProjectInfo(ctx context.Context, p Project) projectInfo {
	dir := p.ResolvedDir()
	backend, auto := resolveTodoBackend(dir, p.TodoProvider)
	return projectInfo{
		Name:            p.Name,
		Dir:             dir,
		Repos:           p.Repos,
		HasProcfile:     dir != "" && procfile.Find(dir, "") != "",
		TodoProvider:    p.TodoProvider,
		TodoBackend:     backend,
		TodoBackendAuto: auto,
		TodoCounts:      countProjectTodos(ctx, dir, p.TodoProvider),
	}
}

// handleProjects is the collection endpoint of the projects entity:
//
//	GET  /api/projects → list (clicky table when negotiated, else []projectInfo)
//	POST /api/projects → create (409 if the name is already taken)
//
// Per-entity reads and mutations live in handleProjectByName.
func (s *Server) handleProjects(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		ps := LoadProjects()
		out := make([]projectInfo, 0, len(ps))
		for _, p := range ps {
			out = append(out, newProjectInfo(r.Context(), p))
		}
		if wantsClicky(r) {
			writeProjectsClicky(w, out)
			return
		}
		respondJSON(w, http.StatusOK, out)
	case http.MethodPost:
		p, err := decodeProject(r)
		if err != nil {
			respondError(w, http.StatusBadRequest, err.Error())
			return
		}
		if err := CreateProject(p); err != nil {
			respondError(w, statusForProjectErr(err), err.Error())
			return
		}
		respondJSON(w, http.StatusCreated, p)
	default:
		respondError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

// handleProjectByName is the per-entity endpoint of the projects entity, keyed
// on the project name from the {name} path segment:
//
//	GET    /api/projects/{name} → one project (clicky detail when negotiated)
//	PUT    /api/projects/{name} → update (path name is authoritative)
//	DELETE /api/projects/{name} → remove
func (s *Server) handleProjectByName(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	switch r.Method {
	case http.MethodGet:
		p, ok := GetProject(name)
		if !ok {
			respondError(w, http.StatusNotFound, fmt.Sprintf("unknown project %q", name))
			return
		}
		info := newProjectInfo(r.Context(), p)
		if wantsClicky(r) {
			writeProjectsClicky(w, []projectInfo{info})
			return
		}
		respondJSON(w, http.StatusOK, info)
	case http.MethodPut:
		p, err := decodeProject(r)
		if err != nil {
			respondError(w, http.StatusBadRequest, err.Error())
			return
		}
		if err := UpdateProject(name, p); err != nil {
			respondError(w, statusForProjectErr(err), err.Error())
			return
		}
		p.Name = name // echo the entity id the path identified
		respondJSON(w, http.StatusOK, p)
	case http.MethodDelete:
		if err := DeleteProject(name); err != nil {
			respondError(w, statusForProjectErr(err), err.Error())
			return
		}
		w.WriteHeader(http.StatusNoContent)
	default:
		respondError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func decodeProject(r *http.Request) (Project, error) {
	var p Project
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		return Project{}, fmt.Errorf("invalid json: %w", err)
	}
	return p, nil
}

func respondJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v) //nolint:errcheck
}

func respondError(w http.ResponseWriter, status int, msg string) {
	respondJSON(w, status, map[string]string{"error": msg})
}
