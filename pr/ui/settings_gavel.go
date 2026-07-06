package ui

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/flanksource/gavel/commit"
	gavelgit "github.com/flanksource/gavel/git"
	"github.com/flanksource/gavel/prompts"
	"github.com/flanksource/gavel/status"
	"github.com/flanksource/gavel/testrunner/outline"
	todoprompt "github.com/flanksource/gavel/todos/prompt"
	"github.com/flanksource/gavel/verify"
)

// gavelSettingsResponse is the read payload for one .gavel.yaml scope: the merged
// view is not used here — the editor loads and saves a single file so each layer
// stays independent (editing the project file must not bake in home/global values).
type gavelSettingsResponse struct {
	Scope  string             `json:"scope"`
	Path   string             `json:"path"`
	Exists bool               `json:"exists"`
	Config verify.GavelConfig `json:"config"`
}

// handleSettingsSchema serves the JSON Schema for .gavel.yaml so the settings
// form can render and validate against the same source of truth as gavel.schema.json.
func (s *Server) handleSettingsSchema(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		respondError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	schema, err := verify.ConfigJSONSchema()
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, _ = io.WriteString(w, schema)
}

// registeredPrompts composes every package's overridable prompt descriptors so
// the settings UI can render one editor per prompt and show each built-in
// default. Composed explicitly (not via init) so a dropped import surfaces as a
// missing prompt at compile time rather than a silent gap in the UI.
func registeredPrompts() []prompts.Prompt {
	var all []prompts.Prompt
	all = append(all, verify.Prompts()...)
	all = append(all, gavelgit.Prompts()...)
	all = append(all, commit.Prompts()...)
	all = append(all, todoprompt.Prompts()...)
	all = append(all, status.Prompts()...)
	all = append(all, outline.Prompts()...)
	return all
}

// handleSettingsPrompts serves the prompt registry (ID, metadata, and embedded
// default per overridable prompt) so the settings form can show defaults and
// which prompts are overridden. Keyed by the schema's x-prompt-id.
func (s *Server) handleSettingsPrompts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		respondError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	respondJSON(w, http.StatusOK, registeredPrompts())
}

// resolveSettingsDir maps the request's scope to the directory whose .gavel.yaml
// is edited. scope=global edits ~/.gavel.yaml; otherwise project=<name> edits the
// registered workspace's root .gavel.yaml. An unknown scope/project is rejected so
// the endpoint can never write an arbitrary path.
func resolveSettingsDir(r *http.Request) (scope, dir string, err error) {
	if r.URL.Query().Get("scope") == "global" {
		home, herr := os.UserHomeDir()
		if herr != nil {
			return "", "", herr
		}
		return "global", home, nil
	}
	name := r.URL.Query().Get("project")
	if name == "" {
		return "", "", errors.New("scope=global or project=<name> is required")
	}
	p, ok := GetProject(name)
	if !ok {
		return "", "", errors.New("unknown project " + name)
	}
	return name, p.ResolvedDir(), nil
}

// handleSettingsGavel reads (GET) or writes (PUT) a single .gavel.yaml file for
// the requested scope.
func (s *Server) handleSettingsGavel(w http.ResponseWriter, r *http.Request) {
	scope, dir, err := resolveSettingsDir(r)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	path := filepath.Join(dir, ".gavel.yaml")

	switch r.Method {
	case http.MethodGet:
		cfg, err := verify.LoadSingleGavelConfig(path)
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		respondJSON(w, http.StatusOK, gavelSettingsResponse{
			Scope:  scope,
			Path:   path,
			Exists: err == nil,
			Config: cfg,
		})
	case http.MethodPut:
		var cfg verify.GavelConfig
		if err := json.NewDecoder(r.Body).Decode(&cfg); err != nil {
			respondError(w, http.StatusBadRequest, "invalid config: "+err.Error())
			return
		}
		if err := verify.SaveGavelConfig(dir, cfg); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		respondJSON(w, http.StatusOK, gavelSettingsResponse{Scope: scope, Path: path, Exists: true, Config: cfg})
	default:
		respondError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}
