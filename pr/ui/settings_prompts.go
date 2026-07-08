package ui

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	dotprompt "github.com/flanksource/captain/pkg/ai/prompt"
	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/gavel/prompts"
	"github.com/flanksource/gavel/verify"
)

// promptDetailResponse is the resolved view of one registered prompt for a
// config layer: its source (default/inline/file), the parsed spec + body, and
// the raw .prompt text (echoed back by the editor as the merge base on PUT).
type promptDetailResponse struct {
	ID     string         `json:"id"`
	Scope  string         `json:"scope"`
	Source string         `json:"source"` // default | inline | file
	Path   string         `json:"path,omitempty"`
	Spec   map[string]any `json:"spec"`
	Body   string         `json:"body"`
	Raw    string         `json:"raw"`
}

// promptDetailRequest is the editor's save payload. Spec carries only the keys
// the editor models; the server merges them over the base frontmatter so
// unmodeled keys (config/input/output) survive. BaseRaw is the raw the editor
// last read, used as a stable merge base.
type promptDetailRequest struct {
	Source  string         `json:"source"` // default | inline | file
	Path    string         `json:"path,omitempty"`
	Spec    map[string]any `json:"spec"`
	Body    string         `json:"body"`
	BaseRaw string         `json:"baseRaw,omitempty"`
}

// handleSettingsPromptDetail resolves (GET) or persists (PUT) one registered
// prompt as a full .prompt document for the requested config layer. The prompt
// id is the schema x-prompt-id; the layer comes from scope=global|project=<name>.
func (s *Server) handleSettingsPromptDetail(w http.ResponseWriter, r *http.Request) {
	desc, ok := findRegisteredPrompt(r.PathValue("id"))
	if !ok {
		respondError(w, http.StatusNotFound, "unknown prompt "+r.PathValue("id"))
		return
	}
	scope, dir, err := resolveSettingsDir(r)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	switch r.Method {
	case http.MethodGet:
		s.getPromptDetail(w, desc, scope, dir)
	case http.MethodPut:
		s.putPromptDetail(w, r, desc, scope, dir)
	default:
		respondError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

// getPromptDetail resolves the effective .prompt for the layer and returns it
// parsed. A parse error is surfaced (400) so the dialog can render it.
func (s *Server) getPromptDetail(w http.ResponseWriter, desc prompts.Prompt, scope, dir string) {
	ov, err := loadPromptOverride(dir, desc.ConfigPath)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	text, err := ov.Resolve(dir, desc.Default)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	doc, err := dotprompt.Parse(text)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, promptDetailResponse{
		ID: desc.ID, Scope: scope, Source: overrideSource(ov), Path: ov.File,
		Spec: specToMap(doc.Spec), Body: doc.Body, Raw: text,
	})
}

// putPromptDetail merges the editor's spec over the base frontmatter, reserializes
// and re-validates the .prompt with captain, then persists by source: inline into
// the layer's .gavel.yaml, file into the referenced .prompt file (config points
// at it). Nothing is written if validation fails.
func (s *Server) putPromptDetail(w http.ResponseWriter, r *http.Request, desc prompts.Prompt, scope, dir string) {
	var req promptDetailRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request: "+err.Error())
		return
	}
	if req.Source != "default" && req.Source != "inline" && req.Source != "file" {
		respondError(w, http.StatusBadRequest, `source must be "default", "inline" or "file"`)
		return
	}
	if req.Source == "file" && strings.TrimSpace(req.Path) == "" {
		respondError(w, http.StatusBadRequest, "file source requires a path")
		return
	}

	cfg, err := loadSingleConfig(dir)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	ov, err := promptOverridePtr(&cfg, desc.ConfigPath)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if req.Source == "default" {
		*ov = verify.PromptOverride{}
		if err := verify.SaveGavelConfig(dir, cfg); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		s.getPromptDetail(w, desc, scope, dir)
		return
	}

	baseText := req.BaseRaw
	if strings.TrimSpace(baseText) == "" {
		if baseText, err = ov.Resolve(dir, desc.Default); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}
	baseDoc, err := dotprompt.Parse(baseText)
	if err != nil {
		respondError(w, http.StatusBadRequest, "base prompt: "+err.Error())
		return
	}

	newText, err := (&dotprompt.Document{
		Frontmatter: mergeFrontmatter(baseDoc.Frontmatter, req.Spec),
		Body:        req.Body,
	}).String()
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	// Re-parse with captain before persisting; on error save nothing.
	if _, err := dotprompt.Parse(newText); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	switch req.Source {
	case "inline":
		*ov = verify.PromptOverride{Inline: newText}
	case "file":
		target := req.Path
		if !filepath.IsAbs(target) {
			target = filepath.Join(dir, target)
		}
		if err := os.WriteFile(target, []byte(newText), 0o644); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		*ov = verify.PromptOverride{File: req.Path}
	}
	if err := verify.SaveGavelConfig(dir, cfg); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	doc, err := dotprompt.Parse(newText)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, promptDetailResponse{
		ID: desc.ID, Scope: scope, Source: req.Source, Path: ov.File,
		Spec: specToMap(doc.Spec), Body: doc.Body, Raw: newText,
	})
}

// findRegisteredPrompt looks up a prompt descriptor by its stable id.
func findRegisteredPrompt(id string) (prompts.Prompt, bool) {
	for _, p := range registeredPrompts() {
		if p.ID == id {
			return p, true
		}
	}
	return prompts.Prompt{}, false
}

// loadPromptOverride reads the layer's .gavel.yaml and returns the override at
// the descriptor's dotted config path.
func loadPromptOverride(dir, configPath string) (*verify.PromptOverride, error) {
	cfg, err := loadSingleConfig(dir)
	if err != nil {
		return nil, err
	}
	return promptOverridePtr(&cfg, configPath)
}

// loadSingleConfig loads one .gavel.yaml layer, tolerating a missing file (a
// zero config the editor can populate).
func loadSingleConfig(dir string) (verify.GavelConfig, error) {
	cfg, err := verify.LoadSingleGavelConfig(filepath.Join(dir, ".gavel.yaml"))
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return verify.GavelConfig{}, err
	}
	return cfg, nil
}

// overrideSource classifies an override for the row's source badge.
func overrideSource(ov *verify.PromptOverride) string {
	switch {
	case ov.IsZero():
		return "default"
	case strings.TrimSpace(ov.Inline) != "":
		return "inline"
	default:
		return "file"
	}
}

// promptOverridePtr walks cfg by the descriptor's dotted json path (e.g.
// "commit.messagePrompt") to the settable *PromptOverride, so any registered
// prompt is addressable with no per-id code. It fails loud on a bad path or a
// non-PromptOverride target.
func promptOverridePtr(cfg *verify.GavelConfig, dotted string) (*verify.PromptOverride, error) {
	v := reflect.ValueOf(cfg).Elem()
	segments := strings.Split(dotted, ".")
	for i, seg := range segments {
		if v.Kind() != reflect.Struct {
			return nil, fmt.Errorf("prompt config path %q: %q is not a struct", dotted, strings.Join(segments[:i], "."))
		}
		field, ok := fieldByJSONName(v, seg)
		if !ok {
			return nil, fmt.Errorf("prompt config path %q: no field %q", dotted, seg)
		}
		v = field
	}
	ov, ok := v.Addr().Interface().(*verify.PromptOverride)
	if !ok {
		return nil, fmt.Errorf("prompt config path %q resolves to %s, not a PromptOverride", dotted, v.Type())
	}
	return ov, nil
}

// fieldByJSONName returns the struct field of v whose json tag (or, absent a
// tag, field name) matches name.
func fieldByJSONName(v reflect.Value, name string) (reflect.Value, bool) {
	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		tag, _, _ := strings.Cut(t.Field(i).Tag.Get("json"), ",")
		if tag == "" {
			tag = t.Field(i).Name
		}
		if tag == name {
			return v.Field(i), true
		}
	}
	return reflect.Value{}, false
}

// mergeFrontmatter merges the editor's spec over the base frontmatter: every
// modeled key is cleared first (so a field the editor cleared disappears rather
// than lingering) then the editor's keys are applied. Unmodeled keys
// (config/input/output/custom) survive untouched.
func mergeFrontmatter(base, editedSpec map[string]any) map[string]any {
	modeled := modeledSpecKeys()
	merged := map[string]any{}
	for k, val := range base {
		if !modeled[k] {
			merged[k] = val
		}
	}
	for k, val := range editedSpec {
		merged[k] = val
	}
	return merged
}

// modeledSpecKeys is the set of top-level json keys the editor owns — every
// field of api.Spec, flattening the inlined Model — derived by reflection so a
// new spec field is covered automatically.
func modeledSpecKeys() map[string]bool {
	keys := map[string]bool{}
	collectJSONKeys(reflect.TypeOf(api.Spec{}), keys)
	return keys
}

func collectJSONKeys(t reflect.Type, out map[string]bool) {
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		name, rest, _ := strings.Cut(f.Tag.Get("json"), ",")
		if f.Anonymous && (name == "" || strings.Contains(rest, "inline")) {
			collectJSONKeys(f.Type, out)
			continue
		}
		if name == "" || name == "-" {
			name = f.Name
		}
		out[name] = true
	}
}

// specToMap renders a spec as a json object for the editor to bind.
func specToMap(spec api.Spec) map[string]any {
	b, err := json.Marshal(spec)
	if err != nil {
		return map[string]any{}
	}
	m := map[string]any{}
	_ = json.Unmarshal(b, &m)
	return m
}
