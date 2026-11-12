package testui

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/flanksource/clicky/entity"
	"github.com/flanksource/clicky/prompt"
	connections "github.com/flanksource/commons-db/cmd/query/connections"
	"github.com/flanksource/commons-db/cmd/query/secretcatalog"

	"github.com/flanksource/gavel/fixtures"
	"github.com/flanksource/gavel/fixtures/kinds/dbconn"
	"github.com/flanksource/gavel/fixtures/kinds/form"
	"github.com/flanksource/gavel/fixtures/kinds/lookup"
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

// ConnectionsMount is the connection API's base, the form's
// `connectionApiBase`: {base}/connection lists (a picker's lookup), creates
// and describes (its JSON Schema) connections, and commons-db serves the
// connection form's test, storage, file, secret and workload routes under it.
const ConnectionsMount = "/api/v1"

// SetConnectionCatalog points the connection API at the gavel connection
// catalog as workDir sees it; a connection created there lands in workDir's
// `.connections` unless the form picks another writable layer.
func (s *Server) SetConnectionCatalog(workDir string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.connectionDir = workDir
}

// handleConnections serves the connection API over the catalog: commons-db's
// connection Service.Handler, in front of the picker lookup and create.
func (s *Server) handleConnections(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	dir := s.connectionDir
	s.mu.RUnlock()
	if dir == "" {
		http.Error(w, "this server has no connection catalog", http.StatusServiceUnavailable)
		return
	}
	service, err := dbconn.Service(r.Context(), dir, dbconn.Options{Secrets: &secretcatalog.Options{}})
	if err != nil {
		http.Error(w, fmt.Sprintf("open the connection catalog: %v", err), http.StatusInternalServerError)
		return
	}
	service.Handler(ConnectionsMount, connectionRoutes{service: service, workDir: dir}).ServeHTTP(w, r)
}

type connectionRoutes struct {
	service *connections.Service
	workDir string
}

func (c connectionRoutes) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != ConnectionsMount+"/connection" {
		http.NotFound(w, r)
		return
	}
	switch r.Method {
	case http.MethodGet:
		c.lookup(w, r)
	case http.MethodPost:
		c.create(w, r)
	default:
		w.Header().Set("Allow", "GET, POST")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// lookup answers a `format: connection` picker in clicky's lookup wire shape:
// the catalog's connections of the requested `types` whose name contains
// `__lookup_q`, keyed by the value the picker commits.
func (c connectionRoutes) lookup(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	var types []string
	if raw := query.Get("types"); raw != "" {
		types = strings.Split(raw, ",")
	}
	result, err := lookup.ConnectionOptions(r.Context(), lookup.ConnectionQuery{WorkDir: c.workDir, Types: types, Query: query.Get("__lookup_q")})
	if err != nil {
		http.Error(w, fmt.Sprintf("list connections: %v", err), http.StatusInternalServerError)
		return
	}
	writeLookup(w, lookup.Response(form.ConnectionLookupFilter, lookup.Result{LookupResult: result, Label: "Connection"}))
}

// createdConnection is what create answers: enough for the picker to select
// the new connection by its connection:// ref, and nothing of its URL, which
// can carry a password in its userinfo.
type createdConnection struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Namespace string `json:"namespace,omitempty"`
	Type      string `json:"type"`
	Ref       string `json:"ref"`
}

// create saves the connection a picker's inline form posts.
func (c connectionRoutes) create(w http.ResponseWriter, r *http.Request) {
	var body map[string]any
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, fmt.Sprintf("decode the connection: %v", err), http.StatusBadRequest)
		return
	}
	created, err := c.service.Create(r.Context(), body)
	if err != nil {
		entity.WriteError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(createdConnection{ID: created.ID.String(), Name: created.Name,
		Namespace: created.Namespace, Type: created.Type, Ref: dbconn.PickerValue(created)})
}

// handleLookup answers a field bound by its form step (lookup.Register):
// GET {lookup.URLPrefix}<token>/<field pointer>?__lookup_filter=<filter>,
// with the search text in `__lookup_q` and the linked connection, once
// chosen, in `connection`.
func (s *Server) handleLookup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", "GET")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	token, pointer, _, err := lookup.ParseURL(r.URL.String())
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	query := r.URL.Query()
	filter := query.Get("__lookup_filter")
	if filter == "" {
		http.Error(w, "a lookup request names its field's filter in __lookup_filter", http.StatusBadRequest)
		return
	}
	req := lookup.Request{Token: token, Pointer: pointer, Query: query.Get("__lookup_q")}
	if query.Has("connection") {
		req.Scope = map[string]any{"connection": query.Get("connection")}
	}
	result, err := lookup.Resolve(r.Context(), req)
	switch {
	case errors.Is(err, lookup.ErrNotFound):
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	case err != nil:
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeLookup(w, lookup.Response(filter, result))
}

func writeLookup(w http.ResponseWriter, response entity.LookupResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	_ = json.NewEncoder(w).Encode(response)
}
