package registry

import (
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/flanksource/captain/pkg/ai/prompt"
	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/captain/pkg/captainconfig"
	"github.com/flanksource/clicky"
	clickyapi "github.com/flanksource/clicky/api"
	"github.com/flanksource/gavel/prompts"
	"github.com/flanksource/gavel/verify"
)

var promptSourcePattern = regexp.MustCompile(
	`^(?:(?:#[^\n]*|[ \t]*)\n)*---\s*(?:\r\n|\r|\n)([\s\S]*?)(?:\r\n|\r|\n)---\s*(?:\r\n|\r|\n)([\s\S]*)$`)

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
	Declared   api.Spec                       `json:"declared" yaml:"declared"`
	Effective  api.Spec                       `json:"effective" yaml:"effective"`
	Provenance map[string]api.FieldProvenance `json:"provenance" yaml:"provenance"`
	Trace      []api.SpecLayer                `json:"trace" yaml:"trace"`
	Warnings   []string                       `json:"warnings,omitempty" yaml:"warnings,omitempty"`
}

type ResolveOptions struct {
	Trace     verify.GavelConfigTrace
	Saved     *captainconfig.AIDefaults
	Preview   bool
	Data      map[string]any
	Draft     string
	Normalize func(api.Spec) (api.SpecNormalization, error)
}

func ResolveOne(opts ResolveOptions, desc prompts.Prompt) (ResolvedPrompt, error) {
	trace := opts.Trace
	override, err := promptSpecAt(trace.Merged, desc.ConfigPath)
	if err != nil {
		return ResolvedPrompt{}, err
	}
	if opts.Draft != "" {
		override = verify.PromptSpec{Spec: api.Spec{Prompt: api.Prompt{User: opts.Draft}}}
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
	declared := defaultSpec
	if source != "builtin" {
		declared = opSpec
	}

	resolution, err := resolvePromptSpec(opts, desc, override)
	if err != nil {
		return ResolvedPrompt{}, err
	}

	return ResolvedPrompt{
		ID: desc.ID, Title: desc.Title, Description: desc.Description,
		ConfigPath: desc.ConfigPath, Source: source, Path: path,
		Raw: raw, Body: body, Frontmatter: frontmatter,
		Declared: declared, Effective: resolution.Spec, Provenance: resolution.Provenance,
		Trace: resolution.Trace, Warnings: resolution.Warnings,
	}, nil
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
	req, cfg, renderErr := prompt.Load(raw).Render(prompt.RenderOptions{Data: map[string]any{}, Declared: true})
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
	if p.Effective.Name == "" {
		t = t.Append("inherited", "font-medium")
	} else {
		t = t.Append(p.Effective.Name, "font-medium")
		if p.Effective.Mode != "" {
			t = t.Append("  mode=").Append(string(p.Effective.Mode), "font-mono")
		}
	}
	data, err := json.MarshalIndent(struct {
		Spec       api.Spec                       `json:"spec"`
		Provenance map[string]api.FieldProvenance `json:"provenance"`
	}{p.Effective, p.Provenance}, "", "  ")
	if err != nil {
		return t.NewLine().Append("render effective spec: "+err.Error(), "text-red-600")
	}
	return t.NewLine().Add(clicky.CodeBlock("json", string(data))).NewLine().Add(clicky.CodeBlock("markdown", p.Raw)).NewLine()
}
