package ui

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	aitools "github.com/flanksource/captain/pkg/ai/tools"
	"github.com/flanksource/captain/pkg/api"
	captaindb "github.com/flanksource/captain/pkg/database"
	"github.com/flanksource/captain/pkg/runtimeprofiles"
)

type runtimePresetLibraryResponse struct {
	Presets  []runtimeprofiles.Preset     `json:"presets"`
	Profiles []runtimeprofiles.Profile    `json:"profiles"`
	Sources  []runtimeprofiles.SourceInfo `json:"sources"`
}

type runtimePresetCreateRequest struct {
	Target string `json:"target"`
	runtimeprofiles.PresetInput
}

type runtimePresetResolveResponse struct {
	Resolved          api.ResolvedSpec          `json:"resolved"`
	Tools             []todoRunToolOption       `json:"tools"`
	Permissions       map[string]api.ToolPolicy `json:"permissions"`
	PermissionSupport map[string]api.Support    `json:"permissionSupport"`
	EffectivePolicy   api.PermissionPolicy      `json:"effectivePolicy"`
}

func (s *Server) handleSettingsRuntimePresets(w http.ResponseWriter, r *http.Request) {
	dir, err := s.runtimePresetDir(r)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	catalog, err := runtimePresetCatalog(r.Context(), dir)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	serveRuntimePresetLibrary(w, r, catalog)
}

func (s *Server) runtimePresetDir(r *http.Request) (string, error) {
	if project := strings.TrimSpace(r.URL.Query().Get("project")); project != "" {
		configured, err := GetProject(project)
		if err != nil {
			return "", err
		}
		return configured.ResolvedDir(), nil
	}
	if r.URL.Query().Get("scope") != "global" {
		return "", errors.New("scope=global or project=<name> is required")
	}
	return s.todoWorkDir(), nil
}

func runtimePresetCatalog(ctx context.Context, dir string) (*runtimeprofiles.Catalog, error) {
	provider, err := openTodoProvider(ctx, dir)
	if err != nil {
		return nil, fmt.Errorf("open runtime preset provider: %w", err)
	}
	backed, ok := provider.(interface{ Captain() *captaindb.DB })
	if !ok || backed.Captain() == nil {
		return nil, errors.New("runtime preset provider has no Captain database")
	}
	open := func(context.Context) (*captaindb.DB, error) { return backed.Captain(), nil }
	return runtimeprofiles.NewDefaultCatalog(ctx, runtimeprofiles.DefaultCatalogOptions{
		Cwd: dir, Read: open, Write: open,
	})
}

func serveRuntimePresetLibrary(w http.ResponseWriter, r *http.Request, catalog *runtimeprofiles.Catalog) {
	switch r.Method {
	case http.MethodGet:
		presets, err := catalog.ListPresets(r.Context())
		if err != nil {
			respondRuntimePresetError(w, err)
			return
		}
		profiles, err := catalog.ListProfiles(r.Context())
		if err != nil {
			respondRuntimePresetError(w, err)
			return
		}
		respondJSON(w, http.StatusOK, runtimePresetLibraryResponse{
			Presets: presets, Profiles: profiles, Sources: catalog.Sources(),
		})
	case http.MethodPost:
		var request runtimePresetCreateRequest
		if err := decodeRuntimePresetRequest(r, &request); err != nil {
			respondError(w, http.StatusBadRequest, err.Error())
			return
		}
		created, err := catalog.CreatePreset(r.Context(), request.Target, request.PresetInput)
		if err != nil {
			respondRuntimePresetError(w, err)
			return
		}
		respondJSON(w, http.StatusCreated, created)
	case http.MethodPut:
		var input runtimeprofiles.PresetInput
		if err := decodeRuntimePresetRequest(r, &input); err != nil {
			respondError(w, http.StatusBadRequest, err.Error())
			return
		}
		updated, err := catalog.UpdatePreset(r.Context(), r.PathValue("id"), input)
		if err != nil {
			respondRuntimePresetError(w, err)
			return
		}
		respondJSON(w, http.StatusOK, updated)
	case http.MethodDelete:
		if err := catalog.DeletePreset(r.Context(), r.PathValue("id")); err != nil {
			respondRuntimePresetError(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	default:
		respondError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (s *Server) handleSettingsRuntimePresetResolve(w http.ResponseWriter, r *http.Request) {
	var request api.RuntimePresetResolveRequest
	if err := decodeRuntimePresetRequest(r, &request); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	resolved, err := api.ResolveRuntimePresets(request)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	definitions := runtimePresetToolDefinitions()
	options := aitools.ResolveOptions{Preferences: resolved.Spec.ToolPreferences, Policy: resolved.Spec.ToolPolicy}
	permissions, err := aitools.ResolveToolPermissions(definitions, options)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, runtimePresetResolveResponse{
		Resolved: resolved, Tools: todoRunToolCatalog(), Permissions: permissions,
		PermissionSupport: runtimePresetPermissionSupport(resolved.Spec, permissions),
		EffectivePolicy:   options.EffectivePolicy(),
	})
}

func runtimePresetToolDefinitions() []api.ToolDefinition {
	items := todoRunToolCatalog()
	definitions := make([]api.ToolDefinition, 0, len(items))
	for _, item := range items {
		definitions = append(definitions, api.ToolDefinition{
			Name: item.Name, Group: item.Group, DefaultPermission: api.ToolPolicyAllow,
			Handler: func(context.Context, map[string]any) (any, error) {
				return nil, errors.New("runtime preset preview does not execute tools")
			},
		})
	}
	return definitions
}

func runtimePresetPermissionSupport(spec api.Spec, permissions map[string]api.ToolPolicy) map[string]api.Support {
	provider, mode, err := spec.Runtime()
	if err != nil {
		return map[string]api.Support{}
	}
	capabilities := api.PermissionCapabilitiesFor(api.RuntimeOf(provider, mode))
	support := make(map[string]api.Support, len(permissions))
	for name, policy := range permissions {
		support[name] = capabilities.ToolPolicySupport(api.ProvenanceCaller, policy)
	}
	return support
}

func decodeRuntimePresetRequest(r *http.Request, target any) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return fmt.Errorf("invalid request: %w", err)
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		return errors.New("invalid request: expected one JSON object")
	}
	return nil
}

func respondRuntimePresetError(w http.ResponseWriter, err error) {
	var referenced runtimeprofiles.ReferencedError
	switch {
	case errors.As(err, &referenced), errors.Is(err, runtimeprofiles.ErrAmbiguous),
		errors.Is(err, runtimeprofiles.ErrNameTaken), errors.Is(err, runtimeprofiles.ErrReadOnly):
		respondError(w, http.StatusConflict, err.Error())
	case errors.Is(err, runtimeprofiles.ErrNotFound):
		respondError(w, http.StatusNotFound, err.Error())
	case errors.Is(err, runtimeprofiles.ErrInvalid):
		respondError(w, http.StatusBadRequest, err.Error())
	default:
		respondError(w, http.StatusInternalServerError, err.Error())
	}
}
