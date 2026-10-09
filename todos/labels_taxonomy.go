package todos

import (
	"context"
	"fmt"
	"strings"

	"github.com/flanksource/commons/logger"

	"github.com/flanksource/gavel/todos/labels"
)

// AllowedLabels is the label vocabulary a triage agent may draw from: every
// definition the workspace resolves, after workspace rows shadow global ones and
// global rows shadow the built-ins.
//
// It is the single source for both halves of the contract — the `## Labels`
// section the prompt shows the agent, and the check ApplyTriage runs before it
// writes — so the list the agent is offered and the list it is held to cannot
// drift apart.
//
// A provider with no definition store, or one whose store cannot be read, falls
// back to the built-in names rather than to an empty set: an empty vocabulary
// would silently forbid every label instead of offering the ten that always
// resolve. The degradation is logged, because a label proposal drawn from the
// builtins alone will miss a workspace's own taxonomy.
func AllowedLabels(ctx context.Context, provider Provider) []string {
	definitions, ok := provider.(LabelDefinitionProvider)
	if !ok {
		return labels.BuiltinNames()
	}
	resolved, err := definitions.LabelDefinitions(ctx)
	if err != nil {
		logger.Warnf("triage label suggestions are degraded: could not read the label taxonomy: %v", err)
		return labels.BuiltinNames()
	}
	if len(resolved) == 0 {
		return labels.BuiltinNames()
	}
	return resolved.Names()
}

// LabelTaxonomySection renders the vocabulary for a prompt: one line per label
// with its description, and its usage count when the provider can supply one, so
// the agent can tell a label the backlog actually uses from one nobody has
// applied yet. An empty taxonomy renders nothing, which omits the section.
func LabelTaxonomySection(ctx context.Context, provider Provider) string {
	definitions, ok := provider.(LabelDefinitionProvider)
	if !ok {
		return renderLabelNames(labels.BuiltinNames())
	}
	resolved, err := definitions.LabelDefinitions(ctx)
	if err != nil || len(resolved) == 0 {
		return renderLabelNames(labels.BuiltinNames())
	}
	counts, err := definitions.LabelCounts(ctx)
	if err != nil {
		logger.Warnf("triage label usage counts are unavailable: %v", err)
		counts = nil
	}

	var lines []string
	for _, def := range resolved {
		line := "- `" + def.Name + "`"
		if description := strings.TrimSpace(def.Description); description != "" {
			line += " — " + description
		}
		if count := counts[def.Name]; count > 0 {
			line += fmt.Sprintf(" (%d in use)", count)
		}
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n")
}

func renderLabelNames(names []string) string {
	var lines []string
	for _, name := range names {
		lines = append(lines, "- `"+name+"`")
	}
	return strings.Join(lines, "\n")
}
