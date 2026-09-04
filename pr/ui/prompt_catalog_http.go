package ui

import (
	"encoding/json"
	"net/http"
	"net/url"
	"os"
	"strings"

	dotprompt "github.com/flanksource/captain/pkg/ai/prompt"
	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/gavel/verify"
)

// handleSettingsPromptCatalog serves every prompt gavel would run for the
// requested scope (scope=global | project=<name>), resolved through the layered
// .gavel.yaml chain: effective source and runtime, plus per-layer provenance.
func (s *Server) handleSettingsPromptCatalog(w http.ResponseWriter, r *http.Request) {
	scope, dir, err := resolveSettingsDir(r)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	catalogScope, err := promptCatalogScopeFor(scope, dir)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	entries, err := buildPromptCatalog(catalogScope)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, entries)
}

// promptCatalogScopeFor loads the config chain for a settings scope and marks
// which of its layers that scope (and the global scope) can edit.
func promptCatalogScopeFor(scope, dir string) (promptCatalogScope, error) {
	trace, err := verify.LoadGavelConfigTrace(dir)
	if err != nil {
		return promptCatalogScope{}, err
	}
	editable := map[string]string{}
	if home, err := os.UserHomeDir(); err == nil {
		editable[canonicalDir(home)] = "scope=global"
	}
	if scope != "global" {
		editable[canonicalDir(dir)] = "project=" + url.QueryEscape(scope)
	}
	return promptCatalogScope{Trace: trace, Editable: editable}, nil
}

// promptRenderRequest renders the effective template — or an unsaved draft —
// with caller-supplied variables and no model call, so an edit can be checked
// against real inputs before it is trusted.
type promptRenderRequest struct {
	Variables map[string]any `json:"variables"`
	Raw       string         `json:"raw,omitempty"`
}

type promptRenderResponse struct {
	User   string `json:"user"`
	System string `json:"system,omitempty"`
	Model  string `json:"model,omitempty"`
	Mode   string `json:"mode,omitempty"`
}

func (s *Server) handleSettingsPromptRender(w http.ResponseWriter, r *http.Request) {
	_, dir, err := resolveSettingsDir(r)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	var req promptRenderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request: "+err.Error())
		return
	}
	raw := req.Raw
	if strings.TrimSpace(raw) == "" {
		if raw, err = effectivePromptSource(r.PathValue("id"), dir); err != nil {
			respondError(w, http.StatusNotFound, err.Error())
			return
		}
	}
	if req.Variables == nil {
		req.Variables = map[string]any{}
	}
	rendered, cfg, err := dotprompt.Load(raw).Render(req.Variables, nil)
	if err != nil {
		respondError(w, http.StatusBadRequest, "render prompt: "+err.Error())
		return
	}
	spec := api.Spec(rendered)
	respondJSON(w, http.StatusOK, promptRenderResponse{
		User: spec.Prompt.User, System: spec.Prompt.System,
		Model: cfg.Model.Name, Mode: string(cfg.Model.Mode),
	})
}

// effectivePromptSource returns the .prompt text gavel would run for a prompt
// id in dir's config chain. Every prompt is a registered one: `todos.prompts`
// declared a second, unregistered source, and a project prompt is a lifecycle
// step now.
func effectivePromptSource(id, dir string) (string, error) {
	trace, err := verify.LoadGavelConfigTrace(dir)
	if err != nil {
		return "", err
	}
	desc, ok := findRegisteredPrompt(id)
	if !ok {
		return "", &unknownPromptError{id: id}
	}
	ov, err := promptOverridePtr(&trace.Merged, desc.ConfigPath)
	if err != nil {
		return "", err
	}
	return promptSpecRaw(ov, trace.TargetDir, desc.Default)
}

type unknownPromptError struct{ id string }

func (e *unknownPromptError) Error() string { return "unknown prompt " + e.id }
