package prompt

import (
	"fmt"
	"strings"

	"github.com/flanksource/gavel/todos/types"
	"github.com/flanksource/gavel/verify"
)

// DefaultName is the prompt a run uses when none is named.
const DefaultName = "run"

// Definition is one runnable named prompt: which template it renders and which
// behaviour class it executes as. The structured result is the same envelope
// for every prompt; the lifecycle step that runs it declares the events.
//
// Separating Name from Class is the point of this type. Before it, template
// selection was keyed off types.RunMode, so a third prompt could not exist
// without a fourth behaviour class — and a fourth class meant widening a
// database CHECK constraint. Now RunMode means only "how does this run treat
// commits and verification", and any number of prompts can share one.
type Definition struct {
	Name  string
	Class types.RunMode
	// ConfigKey is the `.gavel.yaml todos.<key>` block that overrides this prompt.
	ConfigKey   string
	Title       string
	Description string
	// Backlog renders the workspace backlog and label taxonomy into the prompt,
	// and the verification fixture verbatim, for a prompt that reviews the TODO
	// against the rest of the backlog.
	Backlog bool
	// Builtin is the embedded default template, empty for a prompt that exists
	// only in configuration.
	Builtin string
	// Override is the .gavel.yaml layer for this prompt: its own spec fields and
	// an optional `file:` template.
	Override verify.PromptSpec
	// Origin says where the effective template comes from, for `gavel todos
	// prompts` and the settings UI.
	Origin string
}

// Template returns the template source to render: the override's file or inline
// body when it supplies one, otherwise the embedded builtin. A configured prompt
// with neither is a hard error — it would otherwise render an empty prompt and
// the agent would be asked to do nothing.
func (d Definition) Template(workDir string) (string, error) {
	source, err := d.Override.TemplateSource(workDir, d.Builtin)
	if err != nil {
		return "", fmt.Errorf("resolve %s template: %w", d.Name, err)
	}
	if strings.TrimSpace(source) == "" {
		return "", fmt.Errorf("prompt %q has no template: declare a `file:` for it", d.Name)
	}
	return source, nil
}

// Catalog resolves a prompt name to its definition.
type Catalog struct {
	byName map[string]Definition
	names  []string
}

// builtins returns the compiled-in prompts, in the order listings show them.
func builtins() []Definition {
	return []Definition{
		{
			Name: "run", ConfigKey: "run", Class: types.ModeRun, Title: "Todo run prompt",
			Description: "The agent prompt for `gavel todos run`: framing, the TODO items, and instructions.",
			Builtin:     runTemplate,
		},
		{
			Name: "plan", ConfigKey: "plan", Class: types.ModePlan, Title: "Todo plan prompt",
			Description: "The agent prompt for plan-mode runs: read-only investigation that produces a reviewable implementation plan.",
			Builtin:     planTemplate,
		},
		{
			Name: "triage", ConfigKey: "triage", Class: types.ModePlan, Backlog: true, Title: "Todo triage prompt",
			Description: "The agent prompt for triage runs: compacts the description, reviews the verification fixture, " +
				"and records its edits through the todo tools.",
			Builtin: triageTemplate,
		},
		{
			Name: "triage.new", ConfigKey: "triageNew", Class: types.ModePlan, Backlog: true, Title: "New todo triage prompt",
			Description: "Name and label newly created work and relate it to the backlog through the todo tools.",
			Builtin:     triageNewTemplate,
		},
	}
}

// BuiltinClass is the behaviour class a built-in prompt runs as. Overrides
// re-point a built-in's template but never its class.
func BuiltinClass(name string) (types.RunMode, bool) {
	for _, def := range builtins() {
		if def.Name == name {
			return def.Class, true
		}
	}
	return "", false
}

// NewCatalog assembles the runnable prompts from the built-ins and the project's
// configuration. Layering, lowest precedence first:
//
//	builtin  <  .gavel.yaml todos.<configKey>
//
// The set of prompts is a code contract — each pairs a template with the
// behaviour class it runs as — so configuration re-points them rather than
// extending them.
func NewCatalog(cfg verify.TodosConfig) (*Catalog, error) {
	catalog := &Catalog{byName: map[string]Definition{}}
	typed := map[string]verify.PromptSpec{
		"run":       cfg.Run,
		"plan":      cfg.Plan,
		"triage":    cfg.Triage,
		"triageNew": cfg.TriageNew,
	}
	for _, def := range builtins() {
		override, ok := typed[def.ConfigKey]
		if !ok {
			return nil, fmt.Errorf("prompt %q has no todos.%s configuration block", def.Name, def.ConfigKey)
		}
		def.Override = override
		def.Origin = originFor(def)
		catalog.byName[def.Name] = def
		catalog.names = append(catalog.names, def.Name)
	}
	return catalog, nil
}

// Lookup resolves a prompt name. An empty name is the default run prompt. An
// unknown name enumerates what is available rather than failing bare, because the
// available set is project-specific and not discoverable from the error otherwise.
func (c *Catalog) Lookup(name string) (Definition, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		name = DefaultName
	}
	def, ok := c.byName[name]
	if !ok {
		return Definition{}, fmt.Errorf("unknown todo prompt %q; available prompts: %s", name, strings.Join(c.Names(), ", "))
	}
	return def, nil
}

// List returns every prompt in built-in order, so CLI and dashboard listings agree.
func (c *Catalog) List() []Definition {
	defs := make([]Definition, 0, len(c.names))
	for _, name := range c.names {
		defs = append(defs, c.byName[name])
	}
	return defs
}

// Names returns every prompt name in List order.
func (c *Catalog) Names() []string { return append([]string(nil), c.names...) }

func originFor(def Definition) string {
	switch {
	case def.Override.File != "":
		return def.Override.ResolvedFilePath("")
	case strings.TrimSpace(def.Override.Spec.Prompt.User) != "":
		return ".gavel.yaml todos." + def.ConfigKey
	default:
		return "builtin"
	}
}
