// Package missioncontroltest runs a fake Mission Control for tests: the resource search and the
// views the platform dashboard reads, answering one access token and refusing others, and refusing
// change searches on fields Mission Control can't search changes by, like Mission Control.
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
	changes   map[string][]clientapi.SelectedResource
	refused   map[string]string
	searches  []clientapi.SearchResourcesRequest
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
// Checks.
func (s *Server) SetResources(resources clientapi.SearchResourcesResponse) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.resources = resources
}

// SetChanges makes a change search selecting the config item configID answer changes.
func (s *Server) SetChanges(configID string, changes ...clientapi.SelectedResource) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.changes == nil {
		s.changes = map[string][]clientapi.SelectedResource{}
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
	}{{"configs", search.Configs}, {"checks", search.Checks}, {"changes", search.ConfigChanges}} {
		if message, ok := s.refused[kind.name]; ok && len(kind.selectors) > 0 {
			return clientapi.SearchResourcesResponse{}, fmt.Errorf("%s", message)
		}
	}
	if len(search.Configs) > 0 {
		found.Configs = s.resources.Configs
	}
	if len(search.Checks) > 0 {
		found.Checks = s.resources.Checks
	}
	for _, selector := range search.ConfigChanges {
		configIDs, err := changeConfigIDs(selector.Search)
		if err != nil {
			return clientapi.SearchResourcesResponse{}, err
		}
		for _, id := range configIDs {
			found.ConfigChanges = append(found.ConfigChanges, s.changes[id]...)
		}
	}
	return found, nil
}

// changeFields are the fields Mission Control searches changes by: the columns of duty's
// ConfigChangeQueryModel, besides tags.<key>.
var changeFields = map[string]bool{
	"id": true, "config_id": true, "name": true, "type": true, "created_at": true, "severity": true,
	"change_type": true, "summary": true, "count": true, "first_observed": true, "agent_id": true,
}

// changeConfigIDs is the config items a change search selects with config_id=<id>,<id>; it refuses
// a field Mission Control can't search changes by, with Mission Control's error.
func changeConfigIDs(search string) ([]string, error) {
	var ids []string
	for _, term := range strings.Fields(search) {
		field, value, _ := strings.Cut(term, "=")
		field = strings.TrimPrefix(field, "!")
		if !changeFields[field] && !strings.HasPrefix(field, "tags.") {
			return nil, fmt.Errorf("error applying query model: query for column:%s in table:catalog_changes not supported", field)
		}
		if field == "config_id" {
			ids = append(ids, strings.Split(value, ",")...)
		}
	}
	return ids, nil
}
