package verify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	dotprompt "github.com/flanksource/captain/pkg/ai/prompt"
	"github.com/flanksource/captain/pkg/api"
)

// PromptSpec is one AI operation's configuration: a captain api.Spec (model,
// prompt, budget, effort, …) plus an optional File pointing at a .prompt on
// disk. It replaces the old string-only PromptOverride — the model/budget/effort
// that used to live in scalar sibling fields (commit.model, verify.model, …) now
// hang off the same object as the prompt, and a base spec supplies defaults each
// operation overrides field-wise (see Resolve).
//
// In .gavel.yaml an operation accepts a bare string (shorthand for the prompt
// body), a serialized spec, or the full object form:
//
//	commit:
//	  message: "Write a terse commit message for {{diff}}"      # → prompt.user
//	  grouping:
//	    model: claude-sonnet-4-5
//	    prompt: { user: "Group these changes: {{diff}}" }
//	    file: .gavel/prompts/grouping.prompt                    # or a .prompt file
type PromptSpec struct {
	// Spec is a named field, not embedded. api.Spec declares a value-receiver
	// MarshalJSON/MarshalYAML to omit its empty sections; embedding promoted both
	// onto PromptSpec, where they shadowed the default struct encoding and emitted
	// only the spec — so File was silently dropped from .gavel.yaml on every
	// SaveGavelConfig, and the next load fell back to the built-in prompt.
	//
	// The on-disk shape stays flat (model/prompt/budget beside file, not nested
	// under a `spec:` key); MarshalJSON below is what keeps that contract.
	Spec api.Spec `json:",inline" yaml:",inline"`
	// File is a path to a .prompt file. Relative paths resolve against the
	// directory of the .gavel.yaml that declared this override.
	File string `json:"file,omitempty" yaml:"file,omitempty"`

	// baseDir is populated by the config loader and survives layered merges. It is
	// intentionally not serialized; programmatic values fall back to Resolve's dir.
	baseDir string
}

// IsEmpty reports whether this operation configures nothing at all — no prompt
// file and no spec field of any kind. The spec half is answered reflectively by
// captain, so a field added to api.Spec counts immediately.
func (s PromptSpec) IsEmpty() bool {
	return s.File == "" && api.IsEmpty(s.Spec)
}

// Merge layers override onto this spec field-wise, so a repo .gavel.yaml naming
// only a model keeps the home config's budget and permissions.
//
// File is the exception structure cannot express: it travels with baseDir, the
// directory its relative path resolves against. The config loader stamps baseDir
// on every PromptSpec it decodes — set or not — so the two must move together or
// an inherited file would be resolved against the overriding layer's directory.
func (s PromptSpec) Merge(override PromptSpec) PromptSpec {
	merged := PromptSpec{Spec: s.Spec.Merge(override.Spec), File: s.File, baseDir: s.baseDir}
	if override.File != "" {
		merged.File, merged.baseDir = override.File, override.baseDir
	}
	return merged
}

// UnmarshalJSON accepts three forms and sniffs a string's content conservatively:
//  1. an object/mapping — decoded straight into the spec (+ file);
//  2. a string whose content is itself inline JSON (leading '{') or a .prompt
//     document (leading '---') — parsed and adopted as the spec;
//  3. any other string — the plain prompt body, mapped to prompt.user.
//
// Only '{' or '---' triggers struct parsing, so prose containing ':' is never
// mis-read as a YAML mapping. ghodss/yaml routes .gavel.yaml through JSON, so this
// covers both `message: "..."` and `message: { model: ..., prompt: ... }`.
func (s *PromptSpec) UnmarshalJSON(data []byte) error {
	trimmed := bytes.TrimSpace(data)
	if len(trimmed) == 0 || bytes.Equal(trimmed, []byte("null")) {
		*s = PromptSpec{}
		return nil
	}
	if trimmed[0] == '"' {
		var str string
		if err := json.Unmarshal(trimmed, &str); err != nil {
			return err
		}
		return s.adoptString(str)
	}
	// The object form is flat: the spec's own keys sit beside `file`, so each
	// half decodes from the same object. api.Spec declares no UnmarshalJSON, so
	// this is ordinary struct decoding and unknown keys (`file`) are ignored.
	var spec api.Spec
	if err := json.Unmarshal(trimmed, &spec); err != nil {
		return err
	}
	var file struct {
		File string `json:"file"`
	}
	if err := json.Unmarshal(trimmed, &file); err != nil {
		return err
	}
	*s = PromptSpec{Spec: spec, File: file.File}
	return nil
}

// MarshalJSON emits the flat object form — the spec's keys beside `file` — so a
// .gavel.yaml round-trips through SaveGavelConfig unchanged. ghodss/yaml routes
// YAML through JSON, so this governs the YAML shape too and no MarshalYAML is
// needed.
//
// It is written by hand because api.Spec marshals through an unexported
// projection that omits its empty sections; re-deriving those omissions here
// would be the drift this package is trying to delete. Marshalling the spec and
// splicing `file` in keeps captain the single authority on the spec's shape.
func (s PromptSpec) MarshalJSON() ([]byte, error) {
	raw, err := json.Marshal(s.Spec)
	if err != nil {
		return nil, fmt.Errorf("marshal prompt spec: %w", err)
	}
	flat := map[string]json.RawMessage{}
	if err := json.Unmarshal(raw, &flat); err != nil {
		return nil, fmt.Errorf("flatten prompt spec: %w", err)
	}
	if s.File != "" {
		file, err := json.Marshal(s.File)
		if err != nil {
			return nil, fmt.Errorf("marshal prompt file: %w", err)
		}
		flat["file"] = file
	}
	return json.Marshal(flat)
}

// adoptString handles the string form of a PromptSpec (case 2/3 above).
func (s *PromptSpec) adoptString(str string) error {
	content := strings.TrimSpace(str)
	switch {
	case strings.HasPrefix(content, "{"):
		var spec api.Spec
		if err := json.Unmarshal([]byte(content), &spec); err != nil {
			return fmt.Errorf("decode inline prompt spec: %w", err)
		}
		*s = PromptSpec{Spec: spec}
	case strings.HasPrefix(content, "---"):
		doc, err := dotprompt.Parse(str)
		if err != nil {
			return fmt.Errorf("parse inline .prompt document: %w", err)
		}
		spec := doc.Spec
		if spec.Prompt.User == "" {
			spec.Prompt.User = doc.Body
		}
		*s = PromptSpec{Spec: spec}
	default:
		*s = PromptSpec{Spec: api.Spec{Prompt: api.Prompt{User: str}}}
	}
	return nil
}

// Resolve produces the executable spec for this operation by layering, lowest to
// highest precedence: the base spec (ai:) < the built-in default .prompt <
// this operation's own spec. The chosen prompt body (operation override, File,
// or default) is rendered with data — so a default that templates its frontmatter
// (e.g. a maxItems constraint) resolves too. The result is validated.
func (s PromptSpec) Resolve(base api.Spec, defaultPrompt string, data map[string]any, dir string) (api.Spec, error) {
	defaultSpec, err := RenderPromptSpec(defaultPrompt, data, PromptSpecOptions{})
	if err != nil {
		return api.Spec{}, fmt.Errorf("render default prompt: %w", err)
	}

	opSpec := s.Spec
	switch {
	case s.File != "":
		raw, err := os.ReadFile(s.resolvedFilePath(dir))
		if err != nil {
			return api.Spec{}, fmt.Errorf("read prompt override file: %w", err)
		}
		fileSpec, err := RenderPromptSpec(string(raw), data, PromptSpecOptions{})
		if err != nil {
			return api.Spec{}, fmt.Errorf("render prompt override file: %w", err)
		}
		opSpec = fileSpec.Merge(s.Spec) // inline spec fields win over the file
	case strings.TrimSpace(s.Spec.Prompt.User) != "":
		rendered, err := RenderPromptSpec(s.Spec.Prompt.User, data, PromptSpecOptions{})
		if err != nil {
			return api.Spec{}, fmt.Errorf("render inline prompt: %w", err)
		}
		opSpec.Prompt.User = rendered.Prompt.User
	}

	resolved := base.Merge(defaultSpec).Merge(opSpec)
	if err := resolved.Validate(); err != nil {
		return api.Spec{}, fmt.Errorf("resolved prompt spec: %w", err)
	}
	return resolved, nil
}

// TemplateSource returns the raw .prompt template source this override supplies —
// a File's contents or the inline prompt body — or fallback when neither is set.
// Unlike Resolve it neither renders nor merges: call sites that render the
// template themselves with their own data (todos run/plan, status/outline
// summaries) use this to get the source string, and resolve the model separately.
// A configured-but-unreadable File is a hard error.
func (s PromptSpec) TemplateSource(dir, fallback string) (string, error) {
	if s.File != "" {
		data, err := os.ReadFile(s.resolvedFilePath(dir))
		if err != nil {
			return "", fmt.Errorf("read prompt override file %q: %w", s.resolvedFilePath(dir), err)
		}
		return string(data), nil
	}
	if strings.TrimSpace(s.Spec.Prompt.User) != "" {
		return s.Spec.Prompt.User, nil
	}
	return fallback, nil
}

// ResolvedFilePath returns File as an absolute path relative to the config layer
// that declared it (see resolvedFilePath). Exported for callers that display the
// effective override location (the settings registry). Empty when File is unset.
func (s PromptSpec) ResolvedFilePath(fallbackDir string) string {
	return s.resolvedFilePath(fallbackDir)
}

// resolvedFilePath returns File as an absolute path relative to the config layer
// that declared it (baseDir), falling back to fallbackDir.
func (s PromptSpec) resolvedFilePath(fallbackDir string) string {
	if s.File == "" || filepath.IsAbs(s.File) {
		return s.File
	}
	base := s.baseDir
	if base == "" {
		base = fallbackDir
	}
	return filepath.Join(base, s.File)
}

// setPromptSpecBaseDirs reflection-walks the config and stamps every PromptSpec's
// baseDir so relative File references resolve against the directory of the
// .gavel.yaml that declared them, surviving later layered merges.
func setPromptSpecBaseDirs(cfg *GavelConfig, dir string) {
	promptSpecType := reflect.TypeOf(PromptSpec{})
	var walk func(reflect.Value)
	walk = func(v reflect.Value) {
		for v.Kind() == reflect.Pointer {
			if v.IsNil() {
				return
			}
			v = v.Elem()
		}
		if v.Type() == promptSpecType {
			v.Addr().Interface().(*PromptSpec).baseDir = dir
			return
		}
		// Map values are not addressable, so a PromptSpec reached through a map
		// cannot be stamped in place: copy the element, walk the copy, and write it
		// back. Without this a relative `file:` inside one would resolve against
		// the process working directory instead of its declaring config.
		if v.Kind() == reflect.Map {
			if v.IsNil() {
				return
			}
			for _, key := range v.MapKeys() {
				element := reflect.New(v.Type().Elem()).Elem()
				element.Set(v.MapIndex(key))
				walk(element)
				v.SetMapIndex(key, element)
			}
			return
		}
		if v.Kind() != reflect.Struct {
			return
		}
		for i := 0; i < v.NumField(); i++ {
			if v.Type().Field(i).PkgPath == "" {
				walk(v.Field(i))
			}
		}
	}
	walk(reflect.ValueOf(cfg))
}

// PromptSpecOptions configures RenderPromptSpec.
type PromptSpecOptions struct {
	// Declared keeps the frontmatter model exactly as the prompt wrote it instead
	// of the driver-ready (name, mode) pair rendering resolves it to.
	//
	// A configuration layer must carry what it declares. Resolution fills the mode
	// from the model's provider default, and once folded that mode is
	// indistinguishable from one the prompt asked for: a built-in template naming
	// `model: claude` would silently outrank the mode a higher layer's compact
	// model named. The folded spec is resolved once, at the end.
	Declared bool
}

// RenderPromptSpec renders a .prompt source (frontmatter + Handlebars body)
// with data, returning the resulting spec. An empty source yields the zero spec.
// Exported for call sites that layer frontmatter one source at a time — Resolve
// merges all three internally, which loses the per-layer contribution a
// provenance trace needs.
func RenderPromptSpec(source string, data map[string]any, opts PromptSpecOptions) (api.Spec, error) {
	if strings.TrimSpace(source) == "" {
		return api.Spec{}, nil
	}
	req, cfg, err := dotprompt.Load(source).Render(data, nil)
	if err != nil {
		return api.Spec{}, err
	}
	spec := api.Spec(req)
	if spec.Name == "" {
		spec.Model = cfg.Model
	}
	// Frontmatter that templates itself, or that mixes in the dotprompt dialect,
	// cannot be read before rendering — it keeps the resolved model, and is the
	// one layer that still contributes a mode nobody wrote.
	if opts.Declared {
		if doc, perr := dotprompt.Parse(source); perr == nil {
			spec.Model = doc.Spec.Model
		}
	}
	return spec, nil
}
