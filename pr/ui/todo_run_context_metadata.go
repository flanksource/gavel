package ui

import (
	"encoding/json"
	"strings"

	captainai "github.com/flanksource/captain/pkg/ai"
	"github.com/flanksource/captain/pkg/api"
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

func defaultTodoRunBackend(who captaincli.WhoamiResult, backends []todoRunBackendOption) string {
	defaults, ok := who.ProviderDefaults[who.DefaultProvider]
	if !ok {
		return ""
	}
	preferred := captainai.Backend(strings.TrimSpace(defaults.Agent))
	_, mode, known := registry.ProviderFor(preferred)
	provider := api.CatalogPrefixFor(preferred)
	if known && mode == registry.ModeCLI {
		if todoRunBackendHasModels(backends, provider, string(registry.ModeAgent)) {
			return string(registry.ModeAgent)
		}
	}
	if known && todoRunBackendHasModels(backends, provider, string(mode)) {
		return string(mode)
	}
	return ""
}

func todoRunBackendHasModels(backends []todoRunBackendOption, provider, id string) bool {
	for _, backend := range backends {
		if backend.Provider == provider && backend.ID == id {
			return len(backend.Models) > 0
		}
	}
	return false
}
