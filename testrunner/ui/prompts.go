package testui

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/flanksource/clicky/entity"
	"github.com/flanksource/clicky/prompt"

	"github.com/flanksource/gavel/fixtures"
	"github.com/flanksource/gavel/fixtures/kinds/dbconn"
	"github.com/flanksource/gavel/fixtures/kinds/form"
)

// PromptsMount is where the server's prompt API lives: GET prompts,
// prompts/stream and POST prompts/{id}/answer under it.
const PromptsMount = "/api"

// GavelUserHeader names the answerer of a prompt when the server has no
// authenticated user.
const GavelUserHeader = "X-Gavel-User"

func (s *Server) prompter() *prompt.Manager {
	s.promptsOnce.Do(func() {
		s.prompts = prompt.NewManager(nil)
		s.prompts.SetIdentity(s.promptAnswerer)
	})
	return s.prompts
}

// Prompter answers the form steps of every fixture run this server starts;
// the run view shows its pending prompts for the run's PromptOwner.
func (s *Server) Prompter() fixtures.Prompter { return s.prompter() }

// PromptOwner is the current run's prompt owner (the fixture RunID), set by
// BeginRun.
func (s *Server) PromptOwner() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.metadata == nil {
		return ""
	}
	return s.metadata.PromptOwner
}

// SetPromptIdentity names the authenticated user of a request, for an
// embedder that authenticates; an empty name falls back to the X-Gavel-User
// header, then the request's remote address.
func (s *Server) SetPromptIdentity(identify func(*http.Request) string) {
	s.identify = identify
}

func (s *Server) promptAnswerer(r *http.Request) string {
	if s.identify != nil {
		if user := s.identify(r); user != "" {
			return user
		}
	}
	if user := strings.TrimSpace(r.Header.Get(GavelUserHeader)); user != "" {
		return user
	}
	return r.RemoteAddr
}

// SetConnectionCatalog points GET /api/v1/connection, the form connection
// pickers' lookup, at the gavel connection catalog as workDir sees it.
func (s *Server) SetConnectionCatalog(workDir string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.connectionDir = workDir
}

// handleConnectionLookup answers a `format: connection` picker in clicky's
// lookup wire shape: the catalog's connections of the requested `types`
// whose name contains `__lookup_q`, keyed by the value the picker commits.
func (s *Server) handleConnectionLookup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "connections are read-only here", http.StatusMethodNotAllowed)
		return
	}
	s.mu.RLock()
	dir := s.connectionDir
	s.mu.RUnlock()
	if dir == "" {
		http.Error(w, "this server has no connection catalog", http.StatusServiceUnavailable)
		return
	}
	service, err := dbconn.Service(r.Context(), dir, dbconn.Options{})
	if err != nil {
		http.Error(w, fmt.Sprintf("open the connection catalog: %v", err), http.StatusInternalServerError)
		return
	}
	query := r.URL.Query()
	matched, err := dbconn.Lookup(service, dbconn.LookupQuery{Types: strings.Split(query.Get("types"), ","), Query: query.Get("__lookup_q")})
	if err != nil {
		http.Error(w, fmt.Sprintf("list connections: %v", err), http.StatusInternalServerError)
		return
	}
	options := make(map[string]entity.LookupNode, len(matched))
	for _, connection := range matched {
		label := dbconn.PickerLabel(connection)
		options[dbconn.PickerValue(connection)] = entity.LookupNode{Kind: "text", Text: label, Plain: label}
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	_ = json.NewEncoder(w).Encode(entity.LookupResponse{Filters: map[string]entity.LookupFilter{
		form.ConnectionLookupFilter: {Label: "Connection", Options: options, Total: len(matched)},
	}})
}
