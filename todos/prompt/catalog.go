package prompt

import (
	"fmt"
	"sort"
	"strings"

	"github.com/flanksource/gavel/todos/types"
	"github.com/flanksource/gavel/verify"
)

// EnvelopeKind names the structured result a prompt returns. It is a separate
// axis from the behaviour class because two prompts can behave identically —
// read-only, no commits — while reporting entirely different things: a plan run
// reports a plan file, a triage run reports edits to apply.
type EnvelopeKind string

const (
	// EnvelopeResult is types.ResultEnvelope: summary, endStatus, questions.
	EnvelopeResult EnvelopeKind = "result"
	// EnvelopePlan is types.PlanEnvelope.
	EnvelopePlan EnvelopeKind = "plan"
	// EnvelopeTriage is types.TriageEnvelope.
	EnvelopeTriage EnvelopeKind = "triage"
)

// DefaultName is the prompt a run uses when none is named.
const DefaultName = "run"

// Definition is one runnable named prompt: which template it renders, which
// behaviour class it executes as, and which envelope it returns.
//
// Separating Name from Class is the point of this type. Before it, template
// selection was keyed off types.RunMode, so a third prompt could not exist
// without a fourth behaviour class — and a fourth class meant widening a
// database CHECK constraint. Now RunMode means only "how does this run treat
// commits and verification", and any number of prompts can share one.
type Definition struct {
	Name        string
	Class       types.RunMode
	Envelope    EnvelopeKind
	Title       string
	Description string
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
		return "", fmt.Errorf("resolve todos.%s template: %w", d.Name, err)
	}
	if strings.TrimSpace(source) == "" {
		return "", fmt.Errorf("prompt %q has no template: declare a `file:` under todos.%s", d.Name, d.Name)
	}
	return source, nil
}

// Catalog resolves a prompt name to its definition.
type Catalog struct {
	byName map[string]Definition
}

// builtins returns the compiled-in prompts. Each pairs its template with the
// class it runs as and the envelope it returns.
func builtins() []Definition {
	return []Definition{
		{
			Name:        "run",
			Class:       types.ModeRun,
			Envelope:    EnvelopeResult,
			Title:       "Todo run prompt",
			Description: "The agent prompt for `gavel todos run`: framing, the TODO items, and instructions.",
			Builtin:     runTemplate,
		},
		{
			Name:        "plan",
			Class:       types.ModePlan,
			Envelope:    EnvelopePlan,
			Title:       "Todo plan prompt",
			Description: "The agent prompt for plan-mode runs: read-only investigation that produces a reviewable implementation plan.",
			Builtin:     planTemplate,
		},
		{
			Name:     "triage",
			Class:    types.ModePlan,
			Envelope: EnvelopeTriage,
			Title:    "Todo triage prompt",
			Description: "The agent prompt for triage runs: a read-only pass that compacts the description, reviews the " +
				"verification fixture, and reports the edits for gavel to apply.",
			Builtin: triageTemplate,
		},
	}
}

// NewCatalog assembles the runnable prompts from the built-ins and the project's
// configuration. Layering, lowest precedence first:
//
//	builtin  <  .gavel.yaml todos.<name>
//
// The typed fields (todos.run, todos.plan, todos.triage) are how a built-in is
// overridden. The set of prompts is a code contract — each pairs a behaviour
// class with the envelope the run loop parses its result against — so it is
// exactly these three and configuration re-points them rather than extending them.
func NewCatalog(cfg verify.TodosConfig) (*Catalog, error) {
	catalog := &Catalog{byName: map[string]Definition{}}
	typed := map[string]verify.PromptSpec{
		"run":    cfg.Run,
		"plan":   cfg.Plan,
		"triage": cfg.Triage,
	}

	for _, def := range builtins() {
		def.Override = typed[def.Name]
		def.Origin = originFor(def.Name, def.Override)
		catalog.byName[def.Name] = def
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

// List returns every prompt, ordered with the built-ins first and the rest
// alphabetically, so CLI and dashboard listings agree.
func (c *Catalog) List() []Definition {
	defs := make([]Definition, 0, len(c.byName))
	for _, name := range c.Names() {
		defs = append(defs, c.byName[name])
	}
	return defs
}

// Names returns every prompt name in List order.
func (c *Catalog) Names() []string {
	rank := map[string]int{"run": 0, "plan": 1, "triage": 2}
	names := make([]string, 0, len(c.byName))
	for name := range c.byName {
		names = append(names, name)
	}
	sort.Slice(names, func(i, j int) bool {
		ri, iOK := rank[names[i]]
		rj, jOK := rank[names[j]]
		if iOK != jOK {
			return iOK
		}
		if iOK && jOK {
			return ri < rj
		}
		return names[i] < names[j]
	})
	return names
}

// envelopeFor is the envelope a behaviour class returns, for the callers that
// know only the class.
func envelopeFor(class types.RunMode) EnvelopeKind {
	if class == types.ModePlan {
		return EnvelopePlan
	}
	return EnvelopeResult
}

func originFor(name string, override verify.PromptSpec) string {
	switch {
	case override.File != "":
		return override.ResolvedFilePath("")
	case strings.TrimSpace(override.Spec.Prompt.User) != "":
		return ".gavel.yaml todos." + name
	default:
		return "builtin"
	}
}
