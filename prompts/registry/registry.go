package registry

import (
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/flanksource/captain/pkg/ai/prompt"
	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/clicky"
	clickyapi "github.com/flanksource/clicky/api"
	gavelai "github.com/flanksource/gavel/ai"
	"github.com/flanksource/gavel/commit"
	gavelgit "github.com/flanksource/gavel/git"
	"github.com/flanksource/gavel/prompts"
	"github.com/flanksource/gavel/status"
	"github.com/flanksource/gavel/testrunner/outline"
	todoprompt "github.com/flanksource/gavel/todos/prompt"
	"github.com/flanksource/gavel/verify"
)

const (
	defaultGroupModel      = "claude-sonnet-4-5"
	defaultTodoVerifyModel = "claude-code-sonnet"
)

var promptSourcePattern = regexp.MustCompile(
	`^(?:(?:#[^\n]*|[ \t]*)\n)*---\s*(?:\r\n|\r|\n)([\s\S]*?)(?:\r\n|\r|\n)---\s*(?:\r\n|\r|\n)([\s\S]*)$`)

// All returns every overridable prompt in stable command-family order.
func All() []prompts.Prompt {
	var all []prompts.Prompt
	all = append(all, verify.Prompts()...)
	all = append(all, gavelgit.Prompts()...)
	all = append(all, commit.Prompts()...)
	all = append(all, todoprompt.Prompts()...)
	all = append(all, status.Prompts()...)
	all = append(all, outline.Prompts()...)
	return all
}

// ResolvedPrompt is the config-time view of one registered prompt.
type ResolvedPrompt struct {
	ID             string         `json:"id" yaml:"id"`
	Title          string         `json:"title" yaml:"title"`
	Description    string         `json:"description,omitempty" yaml:"description,omitempty"`
	ConfigPath     string         `json:"configPath" yaml:"configPath"`
	Source         string         `json:"source" yaml:"source"` // builtin | inline | file
	Path           string         `json:"path,omitempty" yaml:"path,omitempty"`
	Raw            string         `json:"raw" yaml:"raw"`
	Body           string         `json:"body" yaml:"body"`
	Frontmatter    map[string]any `json:"frontmatter,omitempty" yaml:"frontmatter,omitempty"`
	Declared       api.Spec       `json:"declared" yaml:"declared"`
	EffectiveModel api.Model      `json:"effectiveModel" yaml:"effectiveModel"`
	ModelSource    string         `json:"modelSource" yaml:"modelSource"`
	ModelNote      string         `json:"modelNote,omitempty" yaml:"modelNote,omitempty"`
}

// Resolve expands every registered prompt against a merged config trace.
func Resolve(trace verify.GavelConfigTrace) ([]ResolvedPrompt, error) {
	all := All()
	resolved := make([]ResolvedPrompt, 0, len(all))
	for _, desc := range all {
		item, err := resolveOne(trace, desc)
		if err != nil {
			return nil, fmt.Errorf("resolve %s (%s): %w", desc.ID, desc.ConfigPath, err)
		}
		resolved = append(resolved, item)
	}
	return resolved, nil
}

func resolveOne(trace verify.GavelConfigTrace, desc prompts.Prompt) (ResolvedPrompt, error) {
	override, err := promptOverrideAt(trace.Merged, desc.ConfigPath)
	if err != nil {
		return ResolvedPrompt{}, err
	}

	source := "builtin"
	path := ""
	if override.HasInline() {
		source = "inline"
	} else if override.File != "" {
		source = "file"
		path = override.ResolvedFilePath(trace.TargetDir)
	}

	raw, err := override.Resolve(trace.TargetDir, desc.Default)
	if err != nil {
		return ResolvedPrompt{}, err
	}
	declared, body, frontmatter, err := parsePromptForResolution(raw)
	if err != nil {
		return ResolvedPrompt{}, err
	}
	if declared.Prompt.User == "" {
		declared.Prompt.User = body
	}
	effective, modelSource, note := effectiveModel(trace.Merged, desc.ModelPolicy, declared.Model)

	return ResolvedPrompt{
		ID: desc.ID, Title: desc.Title, Description: desc.Description,
		ConfigPath: desc.ConfigPath, Source: source, Path: path,
		Raw: raw, Body: body, Frontmatter: frontmatter,
		Declared: declared, EffectiveModel: effective,
		ModelSource: modelSource, ModelNote: note,
	}, nil
}

func parsePromptForResolution(raw string) (api.Spec, string, map[string]any, error) {
	doc, err := prompt.Parse(raw)
	if err == nil {
		return doc.Spec, doc.Body, doc.Frontmatter, nil
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
	if declared.Model.Name == "" {
		declared.Model = cfg.Model
	}
	body := raw
	if match := promptSourcePattern.FindStringSubmatch(raw); match != nil {
		body = match[2]
	}
	declared.Prompt.User = body
	return declared, body, nil, nil
}

func effectiveModel(cfg verify.GavelConfig, policy prompts.ModelPolicy, declared api.Model) (api.Model, string, string) {
	var model api.Model
	var source string
	switch policy {
	case prompts.ModelFromVerifyConfig:
		model.Name, source = cfg.Verify.Model, "verify.model"
	case prompts.ModelFromCommitConfig:
		if cfg.Commit.Model != "" {
			model.Name, source = cfg.Commit.Model, "commit.model"
		} else {
			model.Name, source = gavelai.DefaultConfig().Model, "AI default"
		}
	case prompts.ModelFromGroupConfig:
		switch {
		case cfg.Commit.GroupModel != "":
			model.Name, source = cfg.Commit.GroupModel, "commit.groupModel"
		case cfg.Commit.Model != "":
			model.Name, source = cfg.Commit.Model, "commit.model"
		default:
			model.Name, source = defaultGroupModel, "commit grouping default"
		}
	case prompts.ModelFromPrompt:
		model, source = declared, "prompt spec"
	case prompts.ModelFromTodo:
		model.Name, source = defaultTodoVerifyModel, "todo/frontmatter default"
	case prompts.ModelFromSession:
		return api.Model{}, "session", "inherited from the active run session"
	default:
		model.Name, source = gavelai.DefaultConfig().Model, "AI default"
	}

	if model.Name == "" {
		return model, source, "selected at runtime"
	}
	if model.Backend == "" {
		backend, err := model.ResolveBackend()
		if err != nil {
			return model, source, "backend selected at runtime"
		}
		model.Backend = backend
	}
	return model, source, ""
}

func promptOverrideAt(cfg verify.GavelConfig, dotted string) (verify.PromptOverride, error) {
	v := reflect.ValueOf(cfg)
	for _, segment := range strings.Split(dotted, ".") {
		if v.Kind() != reflect.Struct {
			return verify.PromptOverride{}, fmt.Errorf("%q is not a struct", segment)
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
			return verify.PromptOverride{}, fmt.Errorf("config path %q has no field %q", dotted, segment)
		}
	}
	override, ok := v.Interface().(verify.PromptOverride)
	if !ok {
		return verify.PromptOverride{}, fmt.Errorf("config path %q resolves to %s, not PromptOverride", dotted, v.Type())
	}
	return override, nil
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
		if p.EffectiveModel.Backend != "" {
			t = t.Append("  backend=").Append(p.EffectiveModel.Backend, "font-mono")
		}
	}
	t = t.Append("  from=").Append(p.ModelSource, "font-mono")
	if p.ModelNote != "" {
		t = t.Append("  ").Append(p.ModelNote, "text-muted")
	}
	if p.Declared.Model.Name != "" || p.Declared.Model.Backend != "" || p.Declared.Model.Effort != "" || p.Declared.Model.Temperature != nil {
		data, _ := json.Marshal(p.Declared.Model)
		t = t.NewLine().Append("  declared model: ", "text-muted").Append(string(data), "font-mono")
	}
	return t.NewLine().Add(clicky.CodeBlock("markdown", p.Raw)).NewLine()
}
