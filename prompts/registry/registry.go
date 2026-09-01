package registry

import (
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/flanksource/captain/pkg/ai/prompt"
	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/captain/pkg/api/registry"
	"github.com/flanksource/clicky"
	clickyapi "github.com/flanksource/clicky/api"
	"github.com/flanksource/gavel/ai/aifix"
	"github.com/flanksource/gavel/ai/prfix"
	"github.com/flanksource/gavel/commit"
	gavelgit "github.com/flanksource/gavel/git"
	"github.com/flanksource/gavel/prompts"
	"github.com/flanksource/gavel/status"
	"github.com/flanksource/gavel/testrunner/outline"
	todoprompt "github.com/flanksource/gavel/todos/prompt"
	"github.com/flanksource/gavel/verify"
)

var promptSourcePattern = regexp.MustCompile(
	`^(?:(?:#[^\n]*|[ \t]*)\n)*---\s*(?:\r\n|\r|\n)([\s\S]*?)(?:\r\n|\r|\n)---\s*(?:\r\n|\r|\n)([\s\S]*)$`)

// All returns every overridable prompt in stable command-family order.
func All() []prompts.Prompt {
	var all []prompts.Prompt
	all = append(all, aifix.Prompts()...)
	all = append(all, prfix.Prompts()...)
	all = append(all, gavelgit.Prompts()...)
	all = append(all, commit.Prompts()...)
	all = append(all, todoprompt.Prompts()...)
	all = append(all, status.Prompts()...)
	all = append(all, outline.Prompts()...)
	return all
}

// ResolvedPrompt is the config-time view of one registered prompt.
type ResolvedPrompt struct {
	ID          string         `json:"id" yaml:"id"`
	Title       string         `json:"title" yaml:"title"`
	Description string         `json:"description,omitempty" yaml:"description,omitempty"`
	ConfigPath  string         `json:"configPath" yaml:"configPath"`
	Source      string         `json:"source" yaml:"source"` // builtin | inline | file
	Path        string         `json:"path,omitempty" yaml:"path,omitempty"`
	Raw         string         `json:"raw" yaml:"raw"`
	Body        string         `json:"body" yaml:"body"`
	Frontmatter map[string]any `json:"frontmatter,omitempty" yaml:"frontmatter,omitempty"`
	// Declared is the operation's own spec at config time (the built-in default's
	// spec when unset, otherwise the inline/file override spec).
	Declared api.Spec `json:"declared" yaml:"declared"`
	// EffectiveModel is the model chosen by layering the base ai: spec, the
	// built-in default prompt, and the operation override (in that precedence).
	EffectiveModel api.Model `json:"effectiveModel" yaml:"effectiveModel"`
	// ModelSource labels which layer supplied EffectiveModel's name: "operation",
	// "prompt default", "ai base", or "runtime" (none set — chosen at run time).
	ModelSource string `json:"modelSource" yaml:"modelSource"`
}

// Resolve expands every registered prompt against a merged config trace.
func Resolve(trace verify.GavelConfigTrace) ([]ResolvedPrompt, error) {
	all := All()
	resolved := make([]ResolvedPrompt, 0, len(all))
	for _, desc := range all {
		item, err := ResolveOne(trace, desc)
		if err != nil {
			return nil, fmt.Errorf("resolve %s (%s): %w", desc.ID, desc.ConfigPath, err)
		}
		resolved = append(resolved, item)
	}
	return resolved, nil
}

func ResolveOne(trace verify.GavelConfigTrace, desc prompts.Prompt) (ResolvedPrompt, error) {
	override, err := promptSpecAt(trace.Merged, desc.ConfigPath)
	if err != nil {
		return ResolvedPrompt{}, err
	}

	var source string
	path := ""
	switch {
	case override.IsEmpty():
		source = "builtin"
	case override.File != "":
		source = "file"
		path = override.ResolvedFilePath(trace.TargetDir)
	default:
		source = "inline"
	}

	raw, err := override.TemplateSource(trace.TargetDir, desc.Default)
	if err != nil {
		return ResolvedPrompt{}, err
	}
	opSpec, body, frontmatter, err := ParsePromptSource(raw)
	if err != nil {
		return ResolvedPrompt{}, err
	}
	// Fold the override's structured spec fields (model/effort/system/budget an
	// inline object supplies but a body-only TemplateSource omits) over the parse.
	opSpec = opSpec.Merge(override.Spec)
	if opSpec.Prompt.User == "" {
		opSpec.Prompt.User = body
	}

	defaultSpec, _, _, err := ParsePromptSource(desc.Default)
	if err != nil {
		return ResolvedPrompt{}, err
	}

	// Only a real override contributes an operation layer; a built-in inherits the
	// base ai: spec and the default prompt alone.
	opOverride := api.Spec{}
	declared := defaultSpec
	if source != "builtin" {
		opOverride = opSpec
		declared = opSpec
	}

	effective, modelSource := effectiveModelFor(trace.Merged.AI, defaultSpec, opOverride)

	return ResolvedPrompt{
		ID: desc.ID, Title: desc.Title, Description: desc.Description,
		ConfigPath: desc.ConfigPath, Source: source, Path: path,
		Raw: raw, Body: body, Frontmatter: frontmatter,
		Declared: declared, EffectiveModel: effective, ModelSource: modelSource,
	}, nil
}

// effectiveModelFor layers the base ai: spec, the built-in default prompt's
// spec, and the operation override (lowest to highest precedence) and returns
// the resulting model plus a label for which layer supplied its name. It mirrors
// the runtime PromptSpec.Resolve precedence without rendering or validating.
func effectiveModelFor(base, defaultSpec, opOverride api.Spec) (api.Model, string) {
	model := base.Merge(defaultSpec).Merge(opOverride).Model

	source := "runtime"
	switch {
	case opOverride.Name != "":
		source = "operation"
	case defaultSpec.Name != "":
		source = "prompt default"
	case base.Name != "":
		source = "ai base"
	}

	// Fill only the mode, from the family the name claims. This is a reporting
	// path: EffectiveModel is the selector after layering, not the driver-ready
	// id, so it deliberately stops short of a full Resolve — which would collapse
	// a fallback chain to its primary and rewrite an alias the operator wrote.
	// A name that claims no family (a compact multi-model selector) simply keeps
	// an empty mode; execution resolves it properly and fails loudly there.
	if model.Name != "" && model.Mode == "" {
		if provider, err := registry.ProviderFor(model.Name); err == nil {
			model.Mode = provider.DefaultMode
		}
	}
	return model, source
}

// ParsePromptSource returns a prompt's config-time spec, original unrendered
// body, and frontmatter. Prompts with templated frontmatter are rendered once
// with empty config-time data before parsing, matching Resolve while retaining
// the source body the settings editor must round-trip.
func ParsePromptSource(raw string) (api.Spec, string, map[string]any, error) {
	doc, err := prompt.Parse(raw)
	if err == nil {
		return doc.Spec, doc.Body, doc.Frontmatter, nil
	}

	match := promptSourcePattern.FindStringSubmatch(raw)
	if match == nil || !strings.Contains(match[1], "{{") {
		return api.Spec{}, "", nil, err
	}

	// Some built-ins template their YAML frontmatter (for example a conditional
	// maxItems constraint). Parse cannot decode that source before templating, so
	// render once with empty config-time data to fold the declared spec while
	// retaining the original, unrendered body for inspection.
	req, cfg, renderErr := prompt.Load(raw).Render(map[string]any{}, nil)
	if renderErr != nil {
		return api.Spec{}, "", nil, err
	}
	declared := api.Spec(req)
	if declared.Name == "" {
		declared.Model = cfg.Model
	}
	body := match[2]
	declared.Prompt.User = body
	return declared, body, nil, nil
}

// promptSpecAt walks cfg by the descriptor's dotted json path to the PromptSpec
// override field (e.g. "commit.message" → GavelConfig.Commit.Message), failing
// loud on a bad path or a non-PromptSpec target.
func promptSpecAt(cfg verify.GavelConfig, dotted string) (verify.PromptSpec, error) {
	v := reflect.ValueOf(cfg)
	for _, segment := range strings.Split(dotted, ".") {
		if v.Kind() != reflect.Struct {
			return verify.PromptSpec{}, fmt.Errorf("config path %q: %q is not a struct", dotted, segment)
		}
		var found bool
		t := v.Type()
		for i := 0; i < t.NumField(); i++ {
			name, _, _ := strings.Cut(t.Field(i).Tag.Get("json"), ",")
			if name == "" {
				name = t.Field(i).Name
			}
			if name == segment {
				v = v.Field(i)
				found = true
				break
			}
		}
		if !found {
			return verify.PromptSpec{}, fmt.Errorf("config path %q has no field %q", dotted, segment)
		}
	}
	spec, ok := v.Interface().(verify.PromptSpec)
	if !ok {
		return verify.PromptSpec{}, fmt.Errorf("config path %q resolves to %s, not PromptSpec", dotted, v.Type())
	}
	return spec, nil
}

// Pretty renders prompt provenance, declared/effective model details, and the
// complete source document without selecting an output format inside the model.
func (p ResolvedPrompt) Pretty() clickyapi.Text {
	t := clicky.Text(p.Title, "font-bold text-purple-600").
		Append("  ").Append(p.ID, "font-mono text-muted").NewLine().
		Append("  config: ", "text-muted").Append(p.ConfigPath, "font-mono").NewLine().
		Append("  source: ", "text-muted").Append(p.Source, "font-medium")
	if p.Path != "" {
		t = t.Append("  ").Append(p.Path, "font-mono text-muted")
	}
	t = t.NewLine().Append("  effective model: ", "text-muted")
	if p.EffectiveModel.Name == "" {
		t = t.Append("inherited", "font-medium")
	} else {
		t = t.Append(p.EffectiveModel.Name, "font-medium")
		if p.EffectiveModel.Mode != "" {
			t = t.Append("  mode=").Append(string(p.EffectiveModel.Mode), "font-mono")
		}
	}
	t = t.Append("  from=").Append(p.ModelSource, "font-mono")
	if p.Declared.Name != "" || p.Declared.Mode != "" || p.Declared.Effort != "" || p.Declared.Temperature != nil {
		data, _ := json.Marshal(p.Declared.Model)
		t = t.NewLine().Append("  declared model: ", "text-muted").Append(string(data), "font-mono")
	}
	return t.NewLine().Add(clicky.CodeBlock("markdown", p.Raw)).NewLine()
}
