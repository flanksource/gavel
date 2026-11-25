package ui

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strings"

	"github.com/flanksource/gavel/github"
	"github.com/flanksource/gavel/projects"

	rpchttp "github.com/flanksource/clicky/rpc/http"
	"github.com/flanksource/commons/logger"
	gavelctx "github.com/flanksource/gavel/context"
	"github.com/flanksource/gavel/filters"
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
	case errors.Is(err, projects.ErrNotFound):
		return http.StatusNotFound
	case errors.Is(err, projects.ErrExists):
		return http.StatusConflict
	case errors.Is(err, projects.ErrInvalid):
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}

// newProjectInfo resolves a stored Project into the wire shape returned by the
// projects entity: the directory is ~-expanded, hasProcfile reflects the
// directory's current contents, and counts is the project's own entry from
// loadProjectTodoCounts. A project the caller hides is marked and carries no counts.
func newProjectInfo(ctx context.Context, p projects.Project, counts todoCountsResult, hidden bool) (projectInfo, error) {
	dir := p.ResolvedDir()
	stopFile := rpchttp.Track(ctx, "file")
	hasProcfile := dir != "" && procfile.Find(dir, "") != ""
	stopFile()
	info := projectInfo{
		Name:          p.Name,
		Short:         query.ShortProjectName(p.Name),
		Dir:           dir,
		Repos:         p.Repos,
		ReposInferred: p.ReposInferred,
		HasProcfile:   hasProcfile,
		TodoBackend:   "db",
		Hidden:        hidden,
	}
	if hidden {
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
func loadProjectTodoCounts(ctx context.Context, ps []projects.Project, hidden workspaceSet) []todoCountsResult {
	visible := slices.DeleteFunc(slices.Clone(ps), func(p projects.Project) bool { return hidden.has(p.ResolvedDir()) })
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
		if !hidden.has(p.ResolvedDir()) {
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
func listProjectInfos(ctx context.Context, ps []projects.Project, hidden workspaceSet) []projectInfo {
	counts := loadProjectTodoCounts(ctx, ps, hidden)
	out := make([]projectInfo, len(ps))
	for i, p := range ps {
		info, err := newProjectInfo(ctx, p, counts[i], hidden.has(p.ResolvedDir()))
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
//	GET  /api/projects → list (clicky table when negotiated, else []projectInfo);
//	                     ?exclude=hidden-projects drops the projects that filter holds
//	POST /api/projects → create (409 if the name is already taken)
//
// Per-entity reads and mutations live in handleProjectByName.
func (s *Server) handleProjects(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		excluded, err := s.excludedValues(r, filters.KindWorkspace)
		if err != nil {
			respondError(w, filterErrorStatus(err), err.Error())
			return
		}
		out, err := s.projectList(r.Context())
		if err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		out = slices.DeleteFunc(out, func(p projectInfo) bool { return workspaceSet(excluded).has(p.Dir) })
		s.mu.RLock()
		withRepoIcons(out, s.prs)
		s.mu.RUnlock()
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
		created, err := createProject(r.Context(), p)
		if err != nil {
			respondError(w, statusForProjectErr(err), err.Error())
			return
		}
		respondJSON(w, http.StatusCreated, created)
	default:
		respondError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

type projectRepoIcon struct {
	Repo        string `json:"repo"`
	HomepageURL string `json:"homepageUrl"`
}

// withRepoIcons sets each project's RepoIcon from the homepages its tracked
// PRs report. Only those homepages pass the favicon endpoint's knownHomepage
// guard, so a project whose repos have no tracked PR keeps its letter avatar.
func withRepoIcons(projects []projectInfo, prs github.PRSearchResults) {
	homepages := make(map[string]string, len(prs))
	for _, pr := range prs {
		if pr.RepoHomepageURL != "" {
			homepages[strings.ToLower(pr.Repo)] = pr.RepoHomepageURL
		}
	}
	for i := range projects {
		projects[i].RepoIcon = nil
		for _, repo := range projects[i].Repos {
			if homepage, ok := homepages[strings.ToLower(repo)]; ok {
				projects[i].RepoIcon = &projectRepoIcon{Repo: repo, HomepageURL: homepage}
				break
			}
		}
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
		p, err := getProject(r.Context(), name)
		stopFile()
		if err != nil {
			respondError(w, statusForProjectErr(err), err.Error())
			return
		}
		hidden, err := hiddenProjects(r.Context(), s)
		if err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		counts := loadProjectTodoCounts(r.Context(), []projects.Project{p}, hidden)[0]
		info, err := newProjectInfo(r.Context(), p, counts, hidden.has(p.ResolvedDir()))
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
		updated, err := updateProject(r.Context(), name, p)
		if err != nil {
			respondError(w, statusForProjectErr(err), err.Error())
			return
		}
		respondJSON(w, http.StatusOK, updated)
	case http.MethodDelete:
		if err := deleteProject(r.Context(), name); err != nil {
			respondError(w, statusForProjectErr(err), err.Error())
			return
		}
		w.WriteHeader(http.StatusNoContent)
	default:
		respondError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

// untrackProjectGit stops the git tracker scanning dir, a project the user just
// hid; a process without a tracker has nothing to stop.
func (s *Server) untrackProjectGit(dir string) error {
	tracker, err := s.context().GitTracker()
	if errors.Is(err, gavelctx.ErrNoGitTracker) {
		return nil
	}
	if err != nil {
		return err
	}
	if err := tracker.Untrack(dir); err != nil {
		return fmt.Errorf("stop tracking git state of hidden project %s: %w", dir, err)
	}
	return nil
}

func decodeProject(r *http.Request) (projects.Project, error) {
	var p projects.Project
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		return projects.Project{}, fmt.Errorf("invalid json: %w", err)
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
