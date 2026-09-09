package ui

import (
	"encoding/json"
	"net/http"
	"net/url"
	"os"

	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/captain/pkg/captainconfig"
	captaincli "github.com/flanksource/captain/pkg/cli"
	promptregistry "github.com/flanksource/gavel/prompts/registry"
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
	User       string           `json:"user"`
	System     string           `json:"system,omitempty"`
	Model      string           `json:"model,omitempty"`
	Mode       string           `json:"mode,omitempty"`
	Resolution api.ResolvedSpec `json:"resolution"`
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
	desc, ok := findRegisteredPrompt(r.PathValue("id"))
	if !ok {
		respondError(w, http.StatusNotFound, "unknown prompt "+r.PathValue("id"))
		return
	}
	trace, err := verify.LoadGavelConfigTrace(dir)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "load prompt config: "+err.Error())
		return
	}
	saved, _, err := captainconfig.Load()
	if err != nil {
		respondError(w, http.StatusInternalServerError, "load saved AI defaults: "+err.Error())
		return
	}
	item, err := promptregistry.ResolveOne(promptregistry.ResolveOptions{
		Trace: trace, Saved: &saved.AI, Data: req.Variables, Draft: req.Raw,
		Normalize: func(spec api.Spec) (api.SpecNormalization, error) {
			return (captaincli.AIRuntimeOptions{}).Normalize(captaincli.AIRuntimeNormalizeOptions{Spec: spec, Saved: saved, Cwd: trace.TargetDir})
		},
	}, desc)
	if err != nil {
		respondError(w, http.StatusBadRequest, "render prompt: "+err.Error())
		return
	}
	respondJSON(w, http.StatusOK, promptRenderResponse{
		User: item.Effective.Prompt.User, System: item.Effective.Prompt.System,
		Model: item.Effective.Name, Mode: string(item.Effective.Mode),
		Resolution: api.ResolvedSpec{Spec: item.Effective, Trace: item.Trace, Provenance: item.Provenance, Warnings: item.Warnings},
	})
}
