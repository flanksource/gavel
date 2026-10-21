package ui

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"slices"

	rpchttp "github.com/flanksource/clicky/rpc/http"
	"github.com/flanksource/commons/logger"
	gavelctx "github.com/flanksource/gavel/context"
	"github.com/flanksource/gavel/procfile"
	"github.com/flanksource/gavel/todos/query"
)

// projectsTodoCounts is the seam the projects entity reads TODO counts through:
// one call for every project it shows, one result per project in the same
// order. Package tests swap it to drive the handlers without PostgreSQL.
var projectsTodoCounts = countProjectsTodos

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
// directory's current contents, and counts is the project's own entry from
// loadProjectTodoCounts.
func newProjectInfo(ctx context.Context, p Project, counts todoCountsResult) (projectInfo, error) {
	dir := p.ResolvedDir()
	stopFile := rpchttp.Track(ctx, "file")
	hasProcfile := dir != "" && procfile.Find(dir, "") != ""
	stopFile()
	info := projectInfo{
		Name:        p.Name,
		Short:       query.ShortProjectName(p.Name),
		Dir:         dir,
		Repos:       p.Repos,
		HasProcfile: hasProcfile,
		TodoBackend: "db",
		Hidden:      p.Hidden,
	}
	if p.Hidden {
		return info, nil
	}
	if counts.Err != nil {
		return info, counts.Err
	}
	info.TodoCounts = &counts.Counts
	return info, nil
}

// loadProjectTodoCounts reads every project's TODO counts in one batch through
// the projectsTodoCounts seam, and checks the batch answers each project: a
// result that cannot be matched to its project is reported on every project
// rather than risk showing one project another's counts. Hidden projects are
// not shown with counts, so they are not queried and get an empty result.
func loadProjectTodoCounts(ctx context.Context, ps []Project) []todoCountsResult {
	visible := slices.DeleteFunc(slices.Clone(ps), func(p Project) bool { return p.Hidden })
	stopDB := rpchttp.Track(ctx, "db")
	counts := projectsTodoCounts(ctx, visible)
	stopDB()
	out := make([]todoCountsResult, len(ps))
	if len(counts) != len(visible) {
		err := fmt.Errorf("load native TODO counts: got %d results for %d projects", len(counts), len(visible))
		for i := range out {
			out[i].Err = err
		}
		return out
	}
	next := 0
	for i, p := range ps {
		if !p.Hidden {
			out[i], next = counts[next], next+1
		}
	}
	return out
}

// listProjectInfos resolves every project's wire shape in projects.json order,
// reading all of their TODO counts in one batch.
//
// A project whose TODO counts cannot be loaded reports the failure on its own
// entry instead of failing the list: the projects payload also drives Processes
// and PRs, so one unreachable workspace must not blank the whole dashboard. The
// failure stays loud — logged, and visible in the response.
func listProjectInfos(ctx context.Context, ps []Project) []projectInfo {
	counts := loadProjectTodoCounts(ctx, ps)
	out := make([]projectInfo, len(ps))
	for i, p := range ps {
		info, err := newProjectInfo(ctx, p, counts[i])
		if err != nil {
			logger.Errorf("load project %q native TODOs: %v", p.Name, err)
			info.Error = err.Error()
		}
		out[i] = info
	}
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
		out, err := s.projectList(r.Context())
		if err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
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
		stopFile := rpchttp.Track(r.Context(), "file")
		p, err := GetProject(name)
		stopFile()
		if err != nil {
			respondError(w, statusForProjectErr(err), err.Error())
			return
		}
		info, err := newProjectInfo(r.Context(), p, loadProjectTodoCounts(r.Context(), []Project{p})[0])
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

// handleProjectHidden is POST /api/projects/{name}/hidden {"hidden": bool}. A
// narrow endpoint so the full-record PUT never owns the flag. Hiding also
// stops the git tracker scanning the project's directory; a process without a
// tracker has nothing to stop.
func (s *Server) handleProjectHidden(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Hidden *bool `json:"hidden"`
	}
	if err := decodeStrict(r, &body); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	if body.Hidden == nil {
		respondError(w, http.StatusBadRequest, `"hidden" is required`)
		return
	}
	p, err := SetProjectHidden(r.PathValue("name"), *body.Hidden)
	if err != nil {
		respondError(w, statusForProjectErr(err), err.Error())
		return
	}
	if p.Hidden {
		if err := s.untrackProjectGit(p); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}
	respondJSON(w, http.StatusOK, listProjectInfos(r.Context(), []Project{p})[0])
}

func (s *Server) untrackProjectGit(p Project) error {
	tracker, err := s.context().GitTracker()
	if errors.Is(err, gavelctx.ErrNoGitTracker) {
		return nil
	}
	if err != nil {
		return err
	}
	if err := tracker.Untrack(p.ResolvedDir()); err != nil {
		return fmt.Errorf("stop tracking git state of hidden project %q: %w", p.Name, err)
	}
	return nil
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
