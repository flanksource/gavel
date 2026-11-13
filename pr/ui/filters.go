// Serves /api/filters: the caller's named filters, such as the projects and GitHub orgs
// they hide, and resolves the ?exclude=<filter> references list endpoints take.
package ui

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"path/filepath"
	"slices"
	"strings"

	"github.com/flanksource/gavel/filters"
	"github.com/flanksource/gavel/internal/database"
	"github.com/flanksource/gavel/tenant"
)

// errFilterKind is an ?exclude= filter whose values aren't the kind the list holds.
var errFilterKind = errors.New("filter holds the wrong kind of value")

// SetFilterStore makes store hold the Server's named filters. SetOrg sets the
// organization's; a local dashboard sets the process database's. Call it before serving.
func (s *Server) SetFilterStore(store *filters.Store) { s.filters = store }

// filterStore is the store of the caller's named filters: the one SetFilterStore or
// SetOrg gave the Server, else the process database's.
func (s *Server) filterStore(ctx context.Context) (*filters.Store, error) {
	if s != nil && s.filters != nil {
		return s.filters, nil
	}
	db, err := database.Require(ctx, "Gavel filters")
	if err != nil {
		return nil, err
	}
	return filters.NewStore(db), nil
}

// workspaceSet is a set of todo workspace roots: cleaned absolute directories, or
// repo:<id> in hosted mode, as the filter store keeps them. A project is keyed by
// its ResolvedDir, so one registered with a relative directory can't be hidden.
type workspaceSet map[string]bool

func newWorkspaceSet(values []string) workspaceSet {
	set := make(workspaceSet, len(values))
	for _, value := range values {
		set[value] = true
	}
	return set
}

// has reports whether dir, a project's directory, is in the set.
func (w workspaceSet) has(dir string) bool { return w[filepath.Clean(dir)] }

// hiddenProjects is the seam the project surfaces read the caller's hidden-projects
// filter through. Package tests that don't exercise hiding swap it to skip PostgreSQL.
var hiddenProjects = func(ctx context.Context, s *Server) (workspaceSet, error) {
	store, err := s.filterStore(ctx)
	if err != nil {
		return nil, err
	}
	filter, err := store.Get(ctx, tenant.FromContext(ctx), filters.HiddenProjects)
	if err != nil {
		return nil, err
	}
	return newWorkspaceSet(filter.Values), nil
}

// visibleProjects is the project catalog without the projects the caller hides.
// s is nil for a surface with no Server; it then reads the process database.
func visibleProjects(ctx context.Context, s *Server) ([]Project, error) {
	ps, err := LoadProjects()
	if err != nil {
		return nil, err
	}
	hidden, err := hiddenProjects(ctx, s)
	if err != nil {
		return nil, err
	}
	return slices.DeleteFunc(ps, func(p Project) bool { return hidden.has(p.ResolvedDir()) }), nil
}

// LoadVisibleProjects is the project catalog without the projects ctx's scope
// (the local org outside a request) hides, read from the process database.
func LoadVisibleProjects(ctx context.Context) ([]Project, error) { return visibleProjects(ctx, nil) }

// visibleProject is GetProject for the surfaces that would start tracking a
// project's git state: a hidden project is reported as not found.
func (s *Server) visibleProject(ctx context.Context, name string) (Project, error) {
	project, err := GetProject(name)
	if err != nil {
		return Project{}, err
	}
	hidden, err := hiddenProjects(ctx, s)
	if err != nil {
		return Project{}, err
	}
	if hidden.has(project.ResolvedDir()) {
		return Project{}, fmt.Errorf("%w: %q is hidden by the %s filter", ErrProjectNotFound, name, filters.HiddenProjects)
	}
	return project, nil
}

// excludedValues resolves the request's repeatable ?exclude=<filter> to the union of
// those filters' values. Every filter must hold values of kind.
func (s *Server) excludedValues(r *http.Request, kind string) (map[string]bool, error) {
	names := r.URL.Query()["exclude"]
	excluded := map[string]bool{}
	if len(names) == 0 {
		return excluded, nil
	}
	store, err := s.filterStore(r.Context())
	if err != nil {
		return nil, err
	}
	for _, name := range names {
		filter, err := store.Get(r.Context(), tenant.FromContext(r.Context()), name)
		if err != nil {
			return nil, err
		}
		if filter.Kind != kind {
			return nil, fmt.Errorf("%w: ?exclude=%s holds %s values, this list holds %s", errFilterKind, name, filter.Kind, kind)
		}
		for _, value := range filter.Values {
			excluded[value] = true
		}
	}
	return excluded, nil
}

// filterErrorStatus maps a filter error onto its HTTP status.
func filterErrorStatus(err error) int {
	if errors.Is(err, filters.ErrUnknownFilter) || errors.Is(err, filters.ErrInvalidValue) || errors.Is(err, errFilterKind) {
		return http.StatusBadRequest
	}
	return http.StatusInternalServerError
}

func (s *Server) handleFilterList(w http.ResponseWriter, r *http.Request) {
	store, err := s.filterStore(r.Context())
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err)
		return
	}
	list, err := store.List(r.Context(), tenant.FromContext(r.Context()))
	if err != nil {
		writeJSONError(w, filterErrorStatus(err), err)
		return
	}
	writeJSON(w, map[string][]filters.Filter{"filters": list})
}

func (s *Server) handleFilterGet(w http.ResponseWriter, r *http.Request) {
	store, err := s.filterStore(r.Context())
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err)
		return
	}
	filter, err := store.Get(r.Context(), tenant.FromContext(r.Context()), r.PathValue("name"))
	if err != nil {
		writeJSONError(w, filterErrorStatus(err), err)
		return
	}
	writeJSON(w, filter)
}

func (s *Server) handleFilterAdd(w http.ResponseWriter, r *http.Request) {
	s.editFilter(w, r, (*filters.Store).Add)
}

func (s *Server) handleFilterRemove(w http.ResponseWriter, r *http.Request) {
	s.editFilter(w, r, (*filters.Store).Remove)
}

type filterEdit func(*filters.Store, context.Context, tenant.Scope, string, string) (filters.Filter, error)

// editFilter applies {"value": "..."} to the {name} filter. Hiding a local project also
// stops the git tracker scanning its directory.
func (s *Server) editFilter(w http.ResponseWriter, r *http.Request, edit filterEdit) {
	var body struct {
		Value string `json:"value"`
	}
	if err := decodeStrict(r, &body); err != nil {
		writeJSONError(w, http.StatusBadRequest, err)
		return
	}
	store, err := s.filterStore(r.Context())
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err)
		return
	}
	name := r.PathValue("name")
	filter, err := edit(store, r.Context(), tenant.FromContext(r.Context()), name, body.Value)
	if err != nil {
		writeJSONError(w, filterErrorStatus(err), err)
		return
	}
	if name == filters.HiddenProjects && r.Method == http.MethodPost && !s.hosted {
		if err := s.untrackProjectGit(strings.TrimSpace(body.Value)); err != nil {
			writeJSONError(w, http.StatusInternalServerError, err)
			return
		}
	}
	writeJSON(w, filter)
}
