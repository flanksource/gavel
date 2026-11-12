// Package missioncontroltest runs a fake Mission Control for tests: the resource search, the
// changes search and the views the platform dashboard reads, answering one access token only.
package missioncontroltest

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"

	"github.com/flanksource/incident-commander/clientapi"
)

// Token is the only access token the server accepts.
const Token = "mc-token"

// TB is the part of testing.TB the server uses.
type TB interface {
	Helper()
	Cleanup(func())
}

// Server is the fake Mission Control backend.
type Server struct {
	*httptest.Server

	mu        sync.Mutex
	resources clientapi.SearchResourcesResponse
	changes   map[string][]clientapi.ConfigChangeRow
	refused   map[string]string
	searches  []clientapi.SearchResourcesRequest
	changeReq []clientapi.CatalogChangesSearchRequest
	views     map[string]any
	viewReads []map[string]any
}

// New starts the server and closes it when the test ends.
func New(t TB) *Server {
	t.Helper()
	s := &Server{}
	s.Server = httptest.NewServer(http.HandlerFunc(s.serve))
	t.Cleanup(s.Close)
	return s
}

// SetResources makes searches for config items and health checks answer resources' Configs and
// Checks, leaving out their timestamps unless the search asks for them, as Mission Control does.
func (s *Server) SetResources(resources clientapi.SearchResourcesResponse) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.resources = resources
}

// SetChanges makes a changes search for the config item configID answer changes.
func (s *Server) SetChanges(configID string, changes ...clientapi.ConfigChangeRow) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.changes == nil {
		s.changes = map[string][]clientapi.ConfigChangeRow{}
	}
	s.changes[configID] = changes
}

// Refuse makes every search for kind ("configs", "checks" or "changes") answer 500 with message.
func (s *Server) Refuse(kind, message string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.refused == nil {
		s.refused = map[string]string{}
	}
	s.refused[kind] = message
}

// SetView makes the View namespace/name answer view, as the JSON Mission Control renders it.
func (s *Server) SetView(namespace, name string, view any) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.views == nil {
		s.views = map[string]any{}
	}
	s.views[namespace+"/"+name] = view
}

// ViewRequests lists the bodies of the view reads served, in order.
func (s *Server) ViewRequests() []map[string]any {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]map[string]any(nil), s.viewReads...)
}

// Searches lists the resource searches served, in order.
func (s *Server) Searches() []clientapi.SearchResourcesRequest {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]clientapi.SearchResourcesRequest(nil), s.searches...)
}

// ChangeSearches lists the changes searches served, in order.
func (s *Server) ChangeSearches() []clientapi.CatalogChangesSearchRequest {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]clientapi.CatalogChangesSearchRequest(nil), s.changeReq...)
}

func (s *Server) serve(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()
	w.Header().Set("Content-Type", "application/json")
	if r.Header.Get("Authorization") != "Bearer "+Token {
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "invalid access token"})
		return
	}
	switch {
	case r.Method == http.MethodPost && r.URL.Path == "/resources/search":
		var search clientapi.SearchResourcesRequest
		if err := json.NewDecoder(r.Body).Decode(&search); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}
		s.searches = append(s.searches, search)
		found, err := s.search(search)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}
		_ = json.NewEncoder(w).Encode(found)
	case r.Method == http.MethodPost && r.URL.Path == "/catalog/changes":
		var search clientapi.CatalogChangesSearchRequest
		if err := json.NewDecoder(r.Body).Decode(&search); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}
		s.changeReq = append(s.changeReq, search)
		if message, ok := s.refused["changes"]; ok {
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": message})
			return
		}
		var found clientapi.CatalogChangesSearchResponse
		for _, id := range strings.Split(search.CatalogID, ",") {
			found.Changes = append(found.Changes, s.changes[id]...)
		}
		found.Total = int64(len(found.Changes))
		_ = json.NewEncoder(w).Encode(found)
	case r.Method == http.MethodPost && strings.HasPrefix(r.URL.Path, "/view/"):
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}
		s.viewReads = append(s.viewReads, body)
		namespace, name, _ := strings.Cut(strings.TrimPrefix(r.URL.Path, "/view/"), "/")
		view, ok := s.views[namespace+"/"+name]
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "view(namespace=" + namespace + ", name=" + name + ") not found"})
			return
		}
		_ = json.NewEncoder(w).Encode(view)
	default:
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "not found"})
	}
}

// search answers the kinds search asks for.
func (s *Server) search(search clientapi.SearchResourcesRequest) (clientapi.SearchResourcesResponse, error) {
	var found clientapi.SearchResourcesResponse
	for _, kind := range []struct {
		name      string
		selectors []clientapi.ResourceSelector
	}{{"configs", search.Configs}, {"checks", search.Checks}} {
		if message, ok := s.refused[kind.name]; ok && len(kind.selectors) > 0 {
			return clientapi.SearchResourcesResponse{}, fmt.Errorf("%s", message)
		}
	}
	if len(search.Configs) > 0 {
		found.Configs = withTimestamps(s.resources.Configs, search.Timestamps)
	}
	if len(search.Checks) > 0 {
		found.Checks = withTimestamps(s.resources.Checks, search.Timestamps)
	}
	return found, nil
}

// withTimestamps is resources, without their timestamps unless asked for.
func withTimestamps(resources []clientapi.SelectedResource, asked bool) []clientapi.SelectedResource {
	if asked {
		return resources
	}
	found := make([]clientapi.SelectedResource, len(resources))
	for i, resource := range resources {
		resource.CreatedAt, resource.UpdatedAt, resource.DeletedAt = nil, nil, nil
		found[i] = resource
	}
	return found
}
