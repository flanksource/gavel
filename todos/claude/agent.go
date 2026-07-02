// Package claude maps coding-agent model names to their agent; execution
// itself is fully captain-provided (see todos/headless).
package claude

import "strings"

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
