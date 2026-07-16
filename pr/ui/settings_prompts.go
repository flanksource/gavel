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
// config layer: its source (default/inline/file) and the raw .prompt text
// (echoed back by the editor as the merge base on PUT). When the frontmatter
// parses, Spec and Body carry the parsed view; when it does not, ParseError
// carries the parser message and Spec/Body are absent — the raw text is the
// only repair source, so it is always retained.
type promptDetailResponse struct {
	ID         string          `json:"id"`
	Scope      string          `json:"scope"`
	Source     string          `json:"source"` // default | inline | file
	Path       string          `json:"path,omitempty"`
	Spec       *map[string]any `json:"spec,omitempty"`
	Body       *string         `json:"body,omitempty"`
	Raw        string          `json:"raw"`
	ParseError string          `json:"parseError,omitempty"`
}

// promptDetailRequest is the editor's save payload, a strict union keyed by
// which fields are present:
//   - Source "default": reset the override; no content fields.
//   - structured edit: Spec and Body present, Raw absent — merged over BaseRaw.
//   - raw repair: Raw present, Spec and Body absent — validated and persisted
//     verbatim (BaseRaw, which may be malformed, is never parsed here).
//
// The pointers distinguish an absent field from a valid empty map/string.
type promptDetailRequest struct {
	Source  string          `json:"source"` // default | inline | file
	Path    string          `json:"path,omitempty"`
	Spec    *map[string]any `json:"spec,omitempty"`
	Body    *string         `json:"body,omitempty"`
	BaseRaw *string         `json:"baseRaw,omitempty"`
	Raw     *string         `json:"raw,omitempty"`
}

// promptDetailInput is the resolved raw prompt for one layer, before parsing.
type promptDetailInput struct {
	ID, Scope, Source, Path, Raw string
}

// buildPromptDetailResponse parses the resolved raw text into a detail. Raw is
// always retained; a clean parse fills Spec and Body, a syntax error fills
// ParseError (leaving Spec/Body absent rather than inventing empty values) so
// the editor can render a recoverable repair state instead of a dead row.
func buildPromptDetailResponse(input promptDetailInput) promptDetailResponse {
	resp := promptDetailResponse{
		ID: input.ID, Scope: input.Scope, Source: input.Source, Path: input.Path, Raw: input.Raw,
	}
	doc, err := dotprompt.Parse(input.Raw)
	if err != nil {
		resp.ParseError = err.Error()
		return resp
	}
	spec := specToMap(doc.Spec)
	resp.Spec = &spec
	resp.Body = &doc.Body
	return resp
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

// getPromptDetail resolves the effective .prompt for the layer and returns it.
// A frontmatter syntax error is not fatal: it comes back as HTTP 200 with the
// raw text and parseError so the dialog can repair it. Resolution, config, and
// file errors remain non-2xx.
func (s *Server) getPromptDetail(w http.ResponseWriter, desc prompts.Prompt, scope, dir string) {
	ov, err := loadPromptSpec(dir, desc.ConfigPath)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	text, err := promptSpecRaw(ov, dir, desc.Default)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, buildPromptDetailResponse(promptDetailInput{
		ID: desc.ID, Scope: scope, Source: overrideSource(ov), Path: ov.File, Raw: text,
	}))
}

// putPromptDetail persists an edit as one of three strict variants: a reset to
// default, a structured edit (spec+body merged over the base frontmatter), or a
// raw repair (exact validated source). It validates before writing anything, so
// an invalid edit leaves both .gavel.yaml and any prompt file unchanged.
func (s *Server) putPromptDetail(w http.ResponseWriter, r *http.Request, desc prompts.Prompt, scope, dir string) {
	var req promptDetailRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request: "+err.Error())
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

	switch req.Source {
	case "default":
		if req.Spec != nil || req.Body != nil || req.Raw != nil {
			respondError(w, http.StatusBadRequest, `source "default" does not accept spec, body, or raw`)
			return
		}
		*ov = verify.PromptSpec{}
		if err := verify.SaveGavelConfig(dir, cfg); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		s.getPromptDetail(w, desc, scope, dir)
		return
	case "inline", "file":
		// handled below
	default:
		respondError(w, http.StatusBadRequest, `source must be "default", "inline" or "file"`)
		return
	}

	if req.Source == "file" && strings.TrimSpace(req.Path) == "" {
		respondError(w, http.StatusBadRequest, "file source requires a path")
		return
	}

	newText, status, err := buildPromptText(dir, desc, ov, req)
	if err != nil {
		respondError(w, status, err.Error())
		return
	}
	if err := persistPromptOverride(dir, &cfg, ov, req.Source, req.Path, newText); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, buildPromptDetailResponse(promptDetailInput{
		ID: desc.ID, Scope: scope, Source: req.Source, Path: ov.File, Raw: newText,
	}))
}

// buildPromptText produces the validated .prompt text to persist for an
// inline/file save. Exactly one content form must be present: a structured edit
// (spec+body) merges over the valid base frontmatter; a raw repair (raw) is the
// exact source, validated but never merged with the possibly-malformed base. It
// returns the HTTP status to use on error.
func buildPromptText(dir string, desc prompts.Prompt, ov *verify.PromptSpec, req promptDetailRequest) (string, int, error) {
	switch {
	case req.Raw != nil && req.Spec == nil && req.Body == nil:
		newText := *req.Raw
		if _, err := dotprompt.Parse(newText); err != nil {
			return "", http.StatusBadRequest, err
		}
		return newText, http.StatusOK, nil
	case req.Raw == nil && req.Spec != nil && req.Body != nil:
		return mergeStructuredPrompt(dir, desc, ov, req)
	default:
		return "", http.StatusBadRequest, errors.New("provide either spec and body (structured edit) or raw (repair), not both or neither")
	}
}

// mergeStructuredPrompt merges the editor's spec over the base frontmatter (so
// unmodeled keys survive), reserializes, and re-validates. An invalid base
// frontmatter is a 400: a structured edit cannot merge onto malformed YAML —
// that override must be repaired through the raw path first.
func mergeStructuredPrompt(dir string, desc prompts.Prompt, ov *verify.PromptSpec, req promptDetailRequest) (string, int, error) {
	baseText := ""
	if req.BaseRaw != nil {
		baseText = *req.BaseRaw
	}
	if strings.TrimSpace(baseText) == "" {
		resolved, err := promptSpecRaw(ov, dir, desc.Default)
		if err != nil {
			return "", http.StatusInternalServerError, err
		}
		baseText = resolved
	}
	baseDoc, err := dotprompt.Parse(baseText)
	if err != nil {
		return "", http.StatusBadRequest, fmt.Errorf("base prompt: %w", err)
	}
	newText, err := (&dotprompt.Document{
		Frontmatter: mergeFrontmatter(baseDoc.Frontmatter, *req.Spec),
		Body:        *req.Body,
	}).String()
	if err != nil {
		return "", http.StatusInternalServerError, err
	}
	if _, err := dotprompt.Parse(newText); err != nil {
		return "", http.StatusBadRequest, err
	}
	return newText, http.StatusOK, nil
}

// persistPromptOverride writes the validated text by source: inline stores the
// parsed api.Spec structurally in the layer's .gavel.yaml; file writes the
// referenced .prompt file with the config pointing at it. Called only after
// validation, so it never persists bad text. cfg is taken by pointer because ov
// points into it: mutating *ov and then saving *cfg must reflect the same
// struct, not a copy.
func persistPromptOverride(dir string, cfg *verify.GavelConfig, ov *verify.PromptSpec, source, path, newText string) error {
	switch source {
	case "inline":
		spec, err := promptTextToSpec(newText)
		if err != nil {
			return err
		}
		*ov = verify.PromptSpec{Spec: spec}
	case "file":
		target := path
		if !filepath.IsAbs(target) {
			target = filepath.Join(dir, target)
		}
		if err := os.WriteFile(target, []byte(newText), 0o644); err != nil {
			return err
		}
		*ov = verify.PromptSpec{File: path}
	}
	return verify.SaveGavelConfig(dir, *cfg)
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

// loadPromptSpec reads the layer's .gavel.yaml and returns the PromptSpec at the
// descriptor's dotted config path.
func loadPromptSpec(dir, configPath string) (*verify.PromptSpec, error) {
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

// promptSpecRaw returns the full .prompt document for one layer's override: the
// built-in default when unset, the file's contents for a file override, or the
// inline spec reserialized to a .prompt document (frontmatter carries
// model/budget/effort/…, the body carries prompt.user) for an inline override.
func promptSpecRaw(ov *verify.PromptSpec, dir, def string) (string, error) {
	switch {
	case ov.IsZero():
		return def, nil
	case ov.File != "":
		return ov.TemplateSource(dir, def)
	default:
		return specToPromptText(ov.Spec)
	}
}

// specToPromptText serializes an inline api.Spec back into a .prompt document so
// the editor round-trips it: the body carries prompt.user and the remaining spec
// fields (model, budget, effort, prompt.system, …) become the frontmatter.
func specToPromptText(spec api.Spec) (string, error) {
	body := spec.Prompt.User
	fm := specToMap(spec)
	if p, ok := fm["prompt"].(map[string]any); ok {
		delete(p, "user")
		if len(p) == 0 {
			delete(fm, "prompt")
		} else {
			fm["prompt"] = p
		}
	}
	if m, ok := fm["model"].(string); ok && m == "" {
		delete(fm, "model")
	}
	if len(fm) == 0 {
		return body, nil
	}
	return (&dotprompt.Document{Frontmatter: fm, Body: body}).String()
}

// promptTextToSpec parses a validated .prompt document into the api.Spec stored
// as an inline override: frontmatter → spec, body → prompt.user when the
// frontmatter does not set it.
func promptTextToSpec(text string) (api.Spec, error) {
	doc, err := dotprompt.Parse(text)
	if err != nil {
		return api.Spec{}, err
	}
	spec := doc.Spec
	if spec.Prompt.User == "" {
		spec.Prompt.User = doc.Body
	}
	return spec, nil
}

// overrideSource classifies an override for the row's source badge.
func overrideSource(ov *verify.PromptSpec) string {
	switch {
	case ov.IsZero():
		return "default"
	case ov.File != "":
		return "file"
	default:
		return "inline"
	}
}

// promptOverridePtr walks cfg by the descriptor's dotted json path (e.g.
// "commit.message") to the settable *PromptSpec, so any registered prompt is
// addressable with no per-id code. It fails loud on a bad path or a
// non-PromptSpec target.
func promptOverridePtr(cfg *verify.GavelConfig, dotted string) (*verify.PromptSpec, error) {
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
	ov, ok := v.Addr().Interface().(*verify.PromptSpec)
	if !ok {
		return nil, fmt.Errorf("prompt config path %q resolves to %s, not a PromptSpec", dotted, v.Type())
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
