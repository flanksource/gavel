package ui

import (
	"encoding/json"
	"fmt"

	"github.com/flanksource/gavel/todos/lifecycle"
	todoprompt "github.com/flanksource/gavel/todos/prompt"
)

type todoRunToolOption struct {
	Name        string `json:"name"`
	Label       string `json:"label"`
	Group       string `json:"group,omitempty"`
	DefaultMode string `json:"defaultMode,omitempty"`
}

// todoRunToolCatalog maps the default agent toolset onto the tool-preferences
// catalog, grouped for the picker.
//
// Every tool defaults to enabled. `ask` used to be Bash's default here, which
// made the picker offer a per-tool policy captain refuses on every runtime — the
// run was rejected at the policy gate rather than prompting. Asking before
// acting is the permission MODE plus the approval broker now, not a tool
// setting.
func todoRunToolCatalog() []todoRunToolOption {
	group := map[string]string{
		"Read": "Files", "Edit": "Files", "Write": "Files",
		"Bash": "Shell",
		"Glob": "Search", "Grep": "Search",
	}
	tools := todoRunAgentTools()
	out := make([]todoRunToolOption, 0, len(tools))
	for _, name := range tools {
		out = append(out, todoRunToolOption{Name: name, Label: name, Group: group[name], DefaultMode: "enabled"})
	}
	return out
}

// todoRunAgentTools is the standard edit-capable tool set the picker offers
// when a run declares no explicit per-tool policy.
func todoRunAgentTools() []string {
	return []string{"Read", "Edit", "Write", "Bash", "Glob", "Grep"}
}

// todoRunInputSchemas collects each lifecycle step's prompt input schema, keyed
// by step name. The schema follows the step's behaviour class — an implementing
// step exposes the test/lint options its prompt renders — and a step whose
// class has no inputs (plan, triage, verify) is simply absent.
func todoRunInputSchemas(steps []lifecycle.Step) (map[string]json.RawMessage, error) {
	out := map[string]json.RawMessage{}
	for _, step := range steps {
		raw, err := todoprompt.InputSchema(lifecycle.Class(step))
		if err != nil {
			return nil, fmt.Errorf("input schema for step %s: %w", step.Name, err)
		}
		if raw != nil {
			out[step.Name] = json.RawMessage(raw)
		}
	}
	if len(out) == 0 {
		return nil, nil
	}
	return out, nil
}
