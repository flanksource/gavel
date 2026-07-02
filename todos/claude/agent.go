package claude

import (
	"strings"

	captainai "github.com/flanksource/captain/pkg/ai"
	todoprompt "github.com/flanksource/gavel/todos/prompt"
	"github.com/flanksource/gavel/todos/types"
)

// renderRunPrompt renders the run-mode prompt for the SDK bridge: the resolved
// .gavel.yaml template with the group sections, or the verbatim body override,
// always carrying the envelope schema instruction.
func renderRunPrompt(todoList []*types.TODO, workDir, bodyOverride string) (captainai.Request, error) {
	tmpl, err := todoprompt.ResolveTemplate(workDir, types.ModeRun)
	if err != nil {
		return captainai.Request{}, err
	}
	req, _, err := todoprompt.Render(todoList, todoprompt.Options{
		WorkDir:      workDir,
		Mode:         types.ModeRun,
		Template:     tmpl,
		BodyOverride: bodyOverride,
	})
	return req, err
}

// ResolveAgent maps a model name to the coding agent that serves it ("claude" or
// "codex") and the residual model flag. An empty model defaults to claude; codex
// is selected by the "codex"/"gpt-"/"codex-" prefixes.
func ResolveAgent(model string) (agent string, modelFlag string) {
	if model == "" {
		return "claude", ""
	}
	lower := strings.ToLower(model)
	if lower == "codex" || strings.HasPrefix(lower, "gpt-") || strings.HasPrefix(lower, "codex-") {
		if lower == "codex" {
			return "codex", ""
		}
		return "codex", model
	}
	if lower == "claude" {
		return "claude", ""
	}
	return "claude", model
}
