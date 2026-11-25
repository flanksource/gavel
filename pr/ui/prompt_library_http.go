package ui

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	dotprompt "github.com/flanksource/captain/pkg/ai/prompt"
	"github.com/flanksource/captain/pkg/api"
	captaincli "github.com/flanksource/captain/pkg/cli"
	"github.com/flanksource/gavel/prompts"
	"github.com/flanksource/gavel/verify"
)

func (s *Server) handleLibraryPromptDetail(w http.ResponseWriter, r *http.Request) {
	scope, dir, err := resolveSettingsDir(r)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	lib, err := loadPromptLibrary(r.Context(), dir)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	name := strings.TrimPrefix(r.PathValue("id"), libraryPromptPrefix)
	source, err := lib.source(r.URL.Query().Get("source"))
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	detail, exists, err := lib.detail(name, source.ID)
	if err != nil {
		respondError(w, http.StatusNotFound, err.Error())
		return
	}
	if r.Method == http.MethodPut {
		if err := saveLibraryPrompt(lib, detail, source, r, exists); err != nil {
			status := http.StatusBadRequest
			if strings.Contains(err.Error(), "changed") || strings.Contains(err.Error(), "already exists") {
				status = http.StatusConflict
			}
			respondError(w, status, err.Error())
			return
		}
		lib, err = loadPromptLibrary(r.Context(), dir)
		if err == nil {
			found := false
			for _, summary := range lib.Prompts {
				found = found || libraryPromptName(summary) == name
			}
			if !found {
				respondJSON(w, http.StatusOK, buildPromptDetailResponse(promptDetailInput{ID: r.PathValue("id"), Scope: scope, Source: "default"}))
				return
			}
			detail, exists, err = lib.detail(name, source.ID)
		}
		if err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
	} else if r.Method != http.MethodGet {
		respondError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	kind := "default"
	if exists && source.Writable {
		kind = "file"
	}
	response := buildPromptDetailResponse(promptDetailInput{ID: r.PathValue("id"), Scope: scope, Source: kind, Path: detail.RelPath, Raw: detail.Content})
	respondJSON(w, http.StatusOK, response)
}

func saveLibraryPrompt(lib promptLibrary, detail captaincli.PromptDetail, source captaincli.PromptSourceInfo, r *http.Request, exists bool) error {
	if !source.Writable {
		return fmt.Errorf("prompt source %s is read-only; select a writable source", source.Label)
	}
	var req promptDetailRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		return err
	}
	if req.BaseRaw == nil {
		return fmt.Errorf("baseRaw is required to save a library prompt")
	}
	if *req.BaseRaw != detail.Content {
		return fmt.Errorf("prompt changed since it was loaded; reload before saving")
	}
	if req.Source == "default" {
		if req.Spec != nil || req.Body != nil || req.Raw != nil {
			return fmt.Errorf("reset does not accept prompt content")
		}
		if !exists {
			return nil
		}
		return captaincli.RemovePrompt(lib.Context, captaincli.PromptRemoveOptions{Directory: lib.Dir, Reference: detail.ID, Version: detail.Version})
	}
	if req.Source != "file" || req.Path != detail.RelPath {
		return fmt.Errorf("library prompts must be saved as file %s", detail.RelPath)
	}
	raw, _, err := buildPromptText(lib.Dir, prompts.Prompt{Default: detail.Content}, &verify.PromptSpec{}, req)
	if err != nil {
		return err
	}
	options := captaincli.PromptSaveOptions{Directory: lib.Dir, Write: captaincli.PromptWriteRequest{Content: raw, Target: source.ID, RelPath: detail.RelPath}}
	if exists {
		options.Reference, options.Write.BaseVersion = detail.ID, detail.Version
	}
	_, err = captaincli.SavePrompt(lib.Context, options)
	return err
}

func (s *Server) renderLibraryPrompt(w http.ResponseWriter, r *http.Request, dir string, req promptRenderRequest) {
	lib, err := loadPromptLibrary(r.Context(), dir)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	name := strings.TrimPrefix(r.PathValue("id"), libraryPromptPrefix)
	detail, err := captaincli.InspectPrompt(lib.Context, dir, name)
	if err != nil {
		respondError(w, http.StatusNotFound, err.Error())
		return
	}
	raw := detail.Content
	if req.Raw != "" {
		raw = req.Raw
	}
	spec, _, err := dotprompt.LoadNamed(name, raw).Render(dotprompt.RenderOptions{Data: req.Variables, Declared: true, Dir: dir})
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, promptRenderResponse{User: spec.Prompt.User, System: spec.Prompt.System, Model: spec.Name, Mode: string(spec.Mode), Resolution: api.ResolvedSpec{Spec: spec}})
}
