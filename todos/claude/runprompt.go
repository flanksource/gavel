package claude

import (
	"strings"

	"github.com/flanksource/gavel/todos/types"
)

// BuildRunPrompt assembles the prompt body a coding-agent run is given: the
// effort directive (when any) followed by the group prompt. The cmux driver
// moved to captain, but prompt assembly stays here — the run executor builds this
// body and hands it to the provider as the verbatim prompt.
func BuildRunPrompt(todoList []*types.TODO, workDir, effort string) string {
	prompt := BuildGroupPrompt(todoList, workDir)
	if directive := EffortDirective(effort); directive != "" {
		return directive + "\n\n" + prompt
	}
	return prompt
}

// EffortDirective renders the leading reasoning-effort instruction for a run, or
// "" for an unknown tier.
func EffortDirective(effort string) string {
	switch effort {
	case "low":
		return "Be concise."
	case "medium", "":
		return "Think carefully before implementing."
	case "high":
		return "Think hard and reason thoroughly; consider edge cases before implementing."
	case "xhigh":
		return "Think very hard, reason exhaustively, and validate edge cases before implementing."
	default:
		return ""
	}
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
