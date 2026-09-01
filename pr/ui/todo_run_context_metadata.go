package ui

import (
	"encoding/json"
	"strings"

	"github.com/flanksource/captain/pkg/api/registry"
	captaincli "github.com/flanksource/captain/pkg/cli"
	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/todos/drivers"
	todoprompt "github.com/flanksource/gavel/todos/prompt"
	"github.com/flanksource/gavel/todos/types"
)

type todoRunToolOption struct {
	Name        string `json:"name"`
	Label       string `json:"label"`
	Group       string `json:"group,omitempty"`
	DefaultMode string `json:"defaultMode,omitempty"`
}

// todoRunToolCatalog maps the default agent toolset (drivers.DefaultTools) onto
// the tool-preferences catalog, grouped for the picker.
func todoRunToolCatalog() []todoRunToolOption {
	group := map[string]string{
		"Read": "Files", "Edit": "Files", "Write": "Files",
		"Bash": "Shell",
		"Glob": "Search", "Grep": "Search",
	}
	tools := drivers.DefaultTools()
	out := make([]todoRunToolOption, 0, len(tools))
	for _, name := range tools {
		// Bash defaults to ask (brokered), matching the run's default posture where
		// command execution is surfaced for approval; the rest auto-run.
		mode := "enabled"
		if name == "Bash" {
			mode = "ask"
		}
		out = append(out, todoRunToolOption{Name: name, Label: name, Group: group[name], DefaultMode: mode})
	}
	return out
}

// todoRunInputSchemas collects each mode's prompt input schema; a mode with no
// inputs (plan, verify) is simply absent.
func todoRunInputSchemas() map[string]json.RawMessage {
	out := map[string]json.RawMessage{}
	for _, mode := range []types.RunMode{types.ModeRun, types.ModePlan} {
		raw, err := todoprompt.InputSchema(mode)
		if err != nil {
			logger.Warnf("todo run context: input schema for %s: %v", mode, err)
			continue
		}
		if raw != nil {
			out[string(mode)] = json.RawMessage(raw)
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// defaultTodoRunMode picks the run dialog's preselected runtime from the
// user's saved default. The provider and the mode are read as the two fields
// they are; the composite adapter id this used to parse carried both in one
// token and had to be split apart again here.
func defaultTodoRunMode(who captaincli.WhoamiResult, modes []todoRunModeOption) string {
	defaults, ok := who.ProviderDefaults[who.DefaultProvider]
	if !ok {
		return ""
	}
	provider, known := registry.ProviderByName(strings.TrimSpace(who.DefaultProvider))
	if !known {
		return ""
	}
	mode := registry.RuntimeMode(strings.TrimSpace(defaults.Mode))
	// A cli default prefers the agent runtime when that one has models: the
	// dashboard drives agents headlessly, where the SDK is the richer surface.
	if mode == registry.ModeCLI &&
		todoRunModeHasModels(modes, provider.Name, string(registry.ModeAgent)) {
		return string(registry.ModeAgent)
	}
	if todoRunModeHasModels(modes, provider.Name, string(mode)) {
		return string(mode)
	}
	return ""
}

func todoRunModeHasModels(modes []todoRunModeOption, provider, id string) bool {
	for _, option := range modes {
		if option.Provider == provider && option.ID == id {
			return len(option.Models) > 0
		}
	}
	return false
}
