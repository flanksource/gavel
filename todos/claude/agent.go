// Package claude maps coding-agent model names to their agent; execution
// itself is fully captain-provided (see todos/headless).
package claude

import (
	"strings"

	"github.com/flanksource/captain/pkg/api/registry"
)

// ResolveAgent maps a model name to the coding agent that serves it ("claude" or
// "codex") and the residual model flag. An empty model defaults to claude; a
// model in the openai/codex family selects codex, everything else claude.
//
// The family test is captain's — a model in the OpenAI provider family is codex —
// rather than the local strings.HasPrefix(lower, "gpt-") switch that duplicated
// captain's classification. The two coding-agent binaries are claude and codex,
// so a gemini/deepseek model still maps to claude (they are not driver agents),
// preserving the prior contract. The residual modelFlag is empty only for the
// bare "claude"/"codex" sentinels.
func ResolveAgent(model string) (agent string, modelFlag string) {
	if strings.TrimSpace(model) == "" {
		return "claude", ""
	}
	if p, _, ok := registry.ProviderForToken(model); ok && p == registry.OpenAI {
		if strings.EqualFold(model, "codex") {
			return "codex", ""
		}
		return "codex", model
	}
	if strings.EqualFold(model, "claude") {
		return "claude", ""
	}
	return "claude", model
}
