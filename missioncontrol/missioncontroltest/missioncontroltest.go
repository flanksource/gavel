// Package missioncontroltest runs a fake Mission Control for tests: the resource search the
// platform dashboard reads, answering one access token and refusing others like Mission Control.
package missioncontroltest

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
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
	searches  []clientapi.SearchResourcesRequest
}

// New starts the server and closes it when the test ends.
func New(t TB) *Server {
	t.Helper()
	s := &Server{}
	s.Server = httptest.NewServer(http.HandlerFunc(s.serve))
	t.Cleanup(s.Close)
	return s
}

// SetResources makes every resource search answer resources.
func (s *Server) SetResources(resources clientapi.SearchResourcesResponse) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.resources = resources
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
		_ = json.NewEncoder(w).Encode(s.resources)
	default:
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "not found"})
	}
}
