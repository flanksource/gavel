package ui

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/gavel/prompts"
	promptregistry "github.com/flanksource/gavel/prompts/registry"
	"github.com/flanksource/gavel/verify"
)

// promptCatalogEntry is one prompt as the Prompts table and page see it: the
// document gavel would actually run for the requested scope, the runtime it
// resolves to, and which config layer supplies each part of it.
type promptCatalogEntry struct {
	ID              string                         `json:"id"`
	Title           string                         `json:"title"`
	Description     string                         `json:"description,omitempty"`
	ConfigPath      string                         `json:"configPath"`
	Owner           string                         `json:"owner"`
	UsedBy          []string                       `json:"usedBy,omitempty"`
	Source          string                         `json:"source"` // builtin | inline | file
	Path            string                         `json:"path,omitempty"`
	Raw             string                         `json:"raw,omitempty"`
	Version         string                         `json:"version,omitempty"`
	Body            string                         `json:"body,omitempty"`
	Variables       []string                       `json:"variables,omitempty"`
	ParseError      string                         `json:"parseError,omitempty"`
	Effective       promptCatalogRuntime           `json:"effective"`
	Provenance      map[string]string              `json:"provenance,omitempty"`
	Layers          []promptCatalogLayer           `json:"layers"`
	Spec            api.Spec                       `json:"spec"`
	FieldProvenance map[string]api.FieldProvenance `json:"fieldProvenance,omitempty"`
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
	Mode        string   `json:"mode,omitempty"`
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
	// Every runnable todo prompt is a registered one. `todos.prompts` used to let
	// a project declare extra names here; a project prompt is a lifecycle step
	// now, and the catalog no longer has a second, undeclared source.
	return entries, nil
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
	item, err := promptregistry.ResolveOne(promptregistry.ResolveOptions{Trace: scope.Trace, Preview: true}, desc)
	if err != nil {
		entry.ParseError = err.Error()
		return entry
	}
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
	entry.Spec, entry.FieldProvenance = item.Effective, item.Provenance
	entry.Provenance = catalogProvenance(entry.Layers, item)
	entry.Effective = catalogRuntime(item.Effective.Model, catalogModelSource(item.Provenance["/model"].Source))
	return entry
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
// .gavel.yaml (model, mode, effort, file, prompt.user, prompt.system, budget, …).
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

// catalogRuntime projects the shared composition without resolving it again.
func catalogRuntime(model api.Model, modelSource string) promptCatalogRuntime {
	runtime := promptCatalogRuntime{
		Model: model.Name, Mode: string(model.Mode), Effort: string(model.Effort), ModelSource: modelSource,
	}
	for _, fallback := range model.Fallbacks {
		runtime.Fallbacks = append(runtime.Fallbacks, fallback.Name)
	}
	return runtime
}

func catalogProvenance(layers []promptCatalogLayer, item promptregistry.ResolvedPrompt) map[string]string {
	provenance := map[string]string{}
	for key, path := range map[string]string{"body": "/prompt/user", "model": "/model", "mode": "/mode", "effort": "/effort"} {
		source := item.Provenance[path].Source
		origin := catalogModelSource(source)
		for _, layer := range layers {
			if source.Name == layer.Path || source.LayerID == "file" && source.Name == layer.FilePath {
				origin = layer.Origin
			}
		}
		provenance[key] = origin
	}
	return provenance
}

func catalogModelSource(source api.FieldSource) string {
	switch {
	case source.Kind == "saved":
		return "saved defaults"
	case source.Kind == "catalog":
		return "model catalog"
	case strings.HasPrefix(source.Name, "built-in "):
		return "prompt default"
	case strings.HasPrefix(source.Key, "ai") || source.LayerID == "ai":
		return "ai base"
	case source.Kind == "layer":
		return "operation"
	default:
		return "runtime"
	}
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
