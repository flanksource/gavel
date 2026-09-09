package ui

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync"

	rpchttp "github.com/flanksource/clicky/rpc/http"
	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/procfile"
)

var projectTodoCounts = countProjectTodos

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
func newProjectInfo(ctx context.Context, p Project) (projectInfo, error) {
	dir := p.ResolvedDir()
	stopFile := rpchttp.Track(ctx, "file")
	hasProcfile := dir != "" && procfile.Find(dir, "") != ""
	stopFile()
	info := projectInfo{
		Name:        p.Name,
		Dir:         dir,
		Repos:       p.Repos,
		HasProcfile: hasProcfile,
		TodoBackend: "db",
	}
	stopDB := rpchttp.Track(ctx, "db")
	counts, err := projectTodoCounts(ctx, p)
	stopDB()
	if err != nil {
		return info, err
	}
	info.TodoCounts = &counts
	return info, nil
}

// projectInfoConcurrency bounds the parallel per-project TODO count lookups.
// Each one is a single grouped query, so a small fan-out hides the round-trip
// latency without opening a connection per configured project.
const projectInfoConcurrency = 4

// listProjectInfos resolves every project's wire shape concurrently while
// preserving projects.json order.
//
// A project whose TODO counts cannot be loaded reports the failure on its own
// entry instead of failing the list: the projects payload also drives Processes
// and PRs, so one unreachable workspace must not blank the whole dashboard. The
// failure stays loud — logged, and visible in the response.
func listProjectInfos(ctx context.Context, ps []Project) []projectInfo {
	out := make([]projectInfo, len(ps))
	slots := make(chan struct{}, projectInfoConcurrency)
	var wg sync.WaitGroup
	for i, p := range ps {
		wg.Add(1)
		go func() {
			defer wg.Done()
			slots <- struct{}{}
			defer func() { <-slots }()
			info, err := newProjectInfo(ctx, p)
			if err != nil {
				logger.Errorf("load project %q native TODOs: %v", p.Name, err)
				info.Error = err.Error()
			}
			out[i] = info
		}()
	}
	wg.Wait()
	return out
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
		stopFile := rpchttp.Track(r.Context(), "file")
		ps, err := LoadProjects()
		stopFile()
		if err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		out := listProjectInfos(r.Context(), ps)
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
		stopFile := rpchttp.Track(r.Context(), "file")
		p, err := GetProject(name)
		stopFile()
		if err != nil {
			respondError(w, statusForProjectErr(err), err.Error())
			return
		}
		info, err := newProjectInfo(r.Context(), p)
		if err != nil {
			writeTodoError(w, http.StatusInternalServerError, fmt.Errorf("load project %q native TODOs: %w", p.Name, err))
			return
		}
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
	// The create and update handlers echo the decoded project back, so it has to
	// carry the same non-null repos list the collection endpoint serves.
	return p.withRepos(), nil
}

func respondJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v) //nolint:errcheck
}

func respondError(w http.ResponseWriter, status int, msg string) {
	respondJSON(w, status, map[string]string{"error": msg})
}
