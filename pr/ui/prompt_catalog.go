package ui

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/captain/pkg/api/registry"
	"github.com/flanksource/gavel/prompts"
	promptregistry "github.com/flanksource/gavel/prompts/registry"
	todoprompt "github.com/flanksource/gavel/todos/prompt"
	"github.com/flanksource/gavel/verify"
)

// promptCatalogEntry is one prompt as the Prompts table and page see it: the
// document gavel would actually run for the requested scope, the runtime it
// resolves to, and which config layer supplies each part of it.
type promptCatalogEntry struct {
	ID          string               `json:"id"`
	Title       string               `json:"title"`
	Description string               `json:"description,omitempty"`
	ConfigPath  string               `json:"configPath"`
	Owner       string               `json:"owner"`
	UsedBy      []string             `json:"usedBy,omitempty"`
	Source      string               `json:"source"` // builtin | inline | file
	Path        string               `json:"path,omitempty"`
	Raw         string               `json:"raw,omitempty"`
	Version     string               `json:"version,omitempty"`
	Body        string               `json:"body,omitempty"`
	Variables   []string             `json:"variables,omitempty"`
	ParseError  string               `json:"parseError,omitempty"`
	Effective   promptCatalogRuntime `json:"effective"`
	Provenance  map[string]string    `json:"provenance,omitempty"`
	Layers      []promptCatalogLayer `json:"layers"`
}

// promptCatalogLayer is one .gavel.yaml in the scope's chain and what it says
// about this prompt. Scope is the settings query that edits the layer; it is
// empty for a layer the dashboard cannot write (a target directory that is not
// a registered project), which is then shown read-only.
type promptCatalogLayer struct {
	Origin   string   `json:"origin"`
	Path     string   `json:"path"`
	Scope    string   `json:"scope,omitempty"`
	Editable bool     `json:"editable"`
	Source   string   `json:"source"` // none | inline | file
	FilePath string   `json:"filePath,omitempty"`
	Fields   []string `json:"fields,omitempty"`
}

// promptCatalogRuntime is the model the prompt resolves to once compact
// selectors are expanded — what the runtime will actually see.
type promptCatalogRuntime struct {
	Model       string   `json:"model,omitempty"`
	Backend     string   `json:"backend,omitempty"`
	Effort      string   `json:"effort,omitempty"`
	Fallbacks   []string `json:"fallbacks,omitempty"`
	ModelSource string   `json:"modelSource"`
	Error       string   `json:"error,omitempty"`
}

// promptCatalogScope is one request's view: the layered config to resolve
// against and, keyed by canonical layer directory, the settings query that edits
// each layer the dashboard is allowed to write.
type promptCatalogScope struct {
	Trace    verify.GavelConfigTrace
	Editable map[string]string
}

const promptOwnerGavel = "gavel"

func buildPromptCatalog(scope promptCatalogScope) ([]promptCatalogEntry, error) {
	descriptors := registeredPrompts()
	entries := make([]promptCatalogEntry, 0, len(descriptors))
	for _, desc := range descriptors {
		entries = append(entries, registeredPromptEntry(scope, desc))
	}
	named, err := namedTodoPromptEntries(scope)
	if err != nil {
		return nil, err
	}
	return append(entries, named...), nil
}

func registeredPromptEntry(scope promptCatalogScope, desc prompts.Prompt) promptCatalogEntry {
	entry := promptCatalogEntry{
		ID: desc.ID, Title: desc.Title, Description: desc.Description, ConfigPath: desc.ConfigPath,
		Owner: promptOwnerGavel, UsedBy: desc.UsedBy,
		Layers: promptLayers(scope, func(cfg *verify.GavelConfig) (verify.PromptSpec, bool) {
			ov, err := promptOverridePtr(cfg, desc.ConfigPath)
			if err != nil {
				return verify.PromptSpec{}, false
			}
			return *ov, true
		}),
	}
	item, err := promptregistry.ResolveOne(scope.Trace, desc)
	if err != nil {
		entry.ParseError = err.Error()
		return entry
	}
	defaultSpec, _, _, _ := promptregistry.ParsePromptSource(desc.Default)
	entry.Source, entry.Path, entry.Raw, entry.Body = item.Source, item.Path, item.Raw, item.Body
	// The detail endpoint shows an inline override as the default document with
	// the override's keys laid over; the catalog must hash the same text so the
	// versions the page compares against agree.
	merged := scope.Trace.Merged
	if ov, err := promptOverridePtr(&merged, desc.ConfigPath); err == nil {
		if raw, err := promptSpecRaw(ov, scope.Trace.TargetDir, desc.Default); err == nil {
			entry.Raw = raw
		}
	}
	entry.Version = promptSourceVersion(entry.Raw)
	entry.Variables = templateVariables(item.Body)
	entry.Effective = catalogRuntime(item.EffectiveModel, item.ModelSource)
	entry.Provenance = promptProvenance(entry.Layers, item.Source, item.Declared, defaultSpec, scope.Trace.Merged.AI, item.ModelSource)
	return entry
}

// namedTodoPromptEntries lists the prompts declared under todos.prompts that
// are not built-ins: they are runnable (`gavel todos run --prompt <name>`) but
// live outside the static registry.
func namedTodoPromptEntries(scope promptCatalogScope) ([]promptCatalogEntry, error) {
	catalog, err := todoprompt.NewCatalog(scope.Trace.Merged.Todos)
	if err != nil {
		return nil, fmt.Errorf("todos.prompts: %w", err)
	}
	var out []promptCatalogEntry
	for _, def := range catalog.List() {
		if def.Builtin != "" {
			continue
		}
		name := def.Name
		entry := promptCatalogEntry{
			ID: "todos.prompts." + name, Title: firstNonEmpty(def.Title, name), Description: def.Description,
			ConfigPath: "todos.prompts." + name, Owner: promptOwnerGavel,
			UsedBy: []string{"gavel todos run --prompt " + name},
			Source: strings.Replace(overrideSource(&def.Override), "default", "none", 1),
			Path:   def.Override.ResolvedFilePath(scope.Trace.TargetDir),
			Layers: promptLayers(scope, func(cfg *verify.GavelConfig) (verify.PromptSpec, bool) {
				named, ok := cfg.Todos.Prompts[name]
				return named.PromptSpec, ok
			}),
		}
		// The detail endpoint addresses registered prompts only; a named prompt is
		// edited through the settings form, so its layers are shown read-only.
		for i := range entry.Layers {
			entry.Layers[i].Editable, entry.Layers[i].Scope = false, ""
		}
		raw, err := def.Template(scope.Trace.TargetDir)
		if err != nil {
			entry.ParseError = err.Error()
			out = append(out, entry)
			continue
		}
		declared, body, _, err := promptregistry.ParsePromptSource(raw)
		if err != nil {
			entry.ParseError = err.Error()
			out = append(out, entry)
			continue
		}
		entry.Raw, entry.Body, entry.Version = raw, body, promptSourceVersion(raw)
		entry.Variables = templateVariables(body)
		effective := scope.Trace.Merged.AI.Merge(declared).Merge(def.Override.Spec).Model
		modelSource := modelSourceFor(def.Override.Spec.Name, declared.Name, scope.Trace.Merged.AI.Name)
		entry.Effective = catalogRuntime(effective, modelSource)
		entry.Provenance = promptProvenance(entry.Layers, entry.Source, declared, api.Spec{}, scope.Trace.Merged.AI, modelSource)
		out = append(out, entry)
	}
	return out, nil
}

// promptLayers describes each .gavel.yaml in the trace for one prompt: what it
// sets (as the flat keys its override serializes to) and whether the dashboard
// can edit it.
func promptLayers(scope promptCatalogScope, at func(*verify.GavelConfig) (verify.PromptSpec, bool)) []promptCatalogLayer {
	layers := make([]promptCatalogLayer, 0, len(scope.Trace.Sources))
	for _, source := range scope.Trace.Sources {
		cfg := source.Config
		layer := promptCatalogLayer{Origin: source.Origin, Path: source.Path, Source: "none"}
		if query, ok := scope.Editable[canonicalDir(filepath.Dir(source.Path))]; ok {
			layer.Scope, layer.Editable = query, true
		}
		if ov, ok := at(&cfg); ok && !ov.IsEmpty() {
			layer.Source = overrideSource(&ov)
			layer.FilePath = ov.ResolvedFilePath(filepath.Dir(source.Path))
			layer.Fields = promptSpecFields(ov)
		}
		layers = append(layers, layer)
	}
	return layers
}

// promptSpecFields lists the keys an override sets, in the flat form it takes in
// .gavel.yaml (model, backend, effort, file, prompt.user, prompt.system, budget, …).
func promptSpecFields(ov verify.PromptSpec) []string {
	data, err := json.Marshal(ov)
	if err != nil {
		return nil
	}
	flat := map[string]json.RawMessage{}
	if err := json.Unmarshal(data, &flat); err != nil {
		return nil
	}
	var fields []string
	for key, value := range flat {
		if key != "prompt" {
			fields = append(fields, key)
			continue
		}
		sub := map[string]json.RawMessage{}
		if err := json.Unmarshal(value, &sub); err != nil {
			fields = append(fields, key)
			continue
		}
		for subKey := range sub {
			fields = append(fields, "prompt."+subKey)
		}
	}
	sort.Strings(fields)
	return fields
}

// promptProvenance names where each part of the effective prompt comes from:
// the highest layer that sets it, else the built-in default's frontmatter, else
// the base ai: spec, else the runtime. A file override attributes the fields its
// frontmatter declares to the layer that points at the file.
func promptProvenance(
	layers []promptCatalogLayer, source string, declared, defaultSpec, base api.Spec, modelSource string,
) map[string]string {
	fileLayer := lastLayerSetting(layers, "file")
	fromLayers := func(fields ...string) (string, bool) {
		if origin := lastLayerSetting(layers, fields...); origin != "" {
			return origin, true
		}
		return "", false
	}
	prov := map[string]string{}
	if origin, ok := fromLayers("file", "prompt.user"); ok {
		prov["body"] = origin
	} else {
		prov["body"] = "prompt default"
	}
	attribute := func(key, field string, declaredSet, defaultSet, baseSet bool) {
		switch origin, ok := fromLayers(field); {
		case ok:
			prov[key] = origin
		case source == "file" && fileLayer != "" && declaredSet:
			prov[key] = fileLayer
		case defaultSet:
			prov[key] = "prompt default"
		case baseSet:
			prov[key] = "ai base"
		default:
			prov[key] = "runtime"
		}
	}
	attribute("model", "model", declared.Name != "", defaultSpec.Name != "", base.Name != "")
	attribute("backend", "backend", declared.Mode != "", defaultSpec.Mode != "", base.Mode != "")
	attribute("effort", "effort", declared.Effort != "", defaultSpec.Effort != "", base.Effort != "")
	if modelSource == "runtime" && prov["model"] == "runtime" {
		prov["model"] = modelSource
	}
	return prov
}

func lastLayerSetting(layers []promptCatalogLayer, fields ...string) string {
	for i := len(layers) - 1; i >= 0; i-- {
		for _, field := range fields {
			for _, set := range layers[i].Fields {
				if set == field {
					return layers[i].Origin
				}
			}
		}
	}
	return ""
}

func modelSourceFor(operation, promptDefault, base string) string {
	switch {
	case operation != "":
		return "operation"
	case promptDefault != "":
		return "prompt default"
	case base != "":
		return "ai base"
	default:
		return "runtime"
	}
}

// catalogRuntime expands a compact model selector (`agent:opus:medium`,
// `claude`, a fallback list) into the plain name, backend, effort and fallback
// chain the drivers see, so the table shows what will run — not the shorthand.
func catalogRuntime(model api.Model, modelSource string) promptCatalogRuntime {
	runtime := promptCatalogRuntime{
		Model: model.Name, Backend: string(model.Mode), Effort: string(model.Effort), ModelSource: modelSource,
	}
	if model.Name == "" {
		return runtime
	}
	expanded, err := registry.ResolveModel(model)
	if err != nil {
		runtime.Error = err.Error()
		return runtime
	}
	runtime.Model, runtime.Backend, runtime.Effort = expanded.Name, string(expanded.Mode), string(expanded.Effort)
	for _, fallback := range expanded.Fallbacks {
		runtime.Fallbacks = append(runtime.Fallbacks, fallback.Name)
	}
	return runtime
}

var (
	handlebarsToken   = regexp.MustCompile(`\{\{\{?\s*([#^/]?)\s*([A-Za-z_][\w.]*)(?:\s+([A-Za-z_][\w.]*))?`)
	handlebarsHelpers = map[string]bool{
		"if": true, "each": true, "unless": true, "with": true, "else": true, "this": true,
		"role": true, "history": true, "media": true, "section": true, "json": true, "lookup": true, "log": true,
	}
)

// templateVariables lists the top-level Handlebars variables a body references
// (`{{diff}}`, `{{{body}}}`, `{{#if linters}}`, `{{#each commits}}`), so the
// table can say what data a prompt consumes without rendering it.
func templateVariables(body string) []string {
	seen := map[string]bool{}
	for _, match := range handlebarsToken.FindAllStringSubmatch(body, -1) {
		if match[1] == "/" {
			continue
		}
		name := match[2]
		if handlebarsHelpers[name] {
			name = match[3]
		}
		name, _, _ = strings.Cut(name, ".")
		if name == "" || handlebarsHelpers[name] {
			continue
		}
		seen[name] = true
	}
	names := make([]string, 0, len(seen))
	for name := range seen {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func promptSourceVersion(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])[:16]
}

func canonicalDir(dir string) string {
	if resolved, err := filepath.EvalSymlinks(dir); err == nil {
		return resolved
	}
	return filepath.Clean(dir)
}
