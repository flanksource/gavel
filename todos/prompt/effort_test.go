package prompt

import dotprompt "github.com/flanksource/captain/pkg/ai/prompt"

func EffortDirective(tier string) string {
	text := map[string]string{
		"low":    "Be concise.",
		"medium": "Think carefully before implementing.",
		"high":   "Think hard and reason thoroughly; consider edge cases before implementing.",
		"xhigh":  "Think very hard, reason exhaustively, and validate edge cases before implementing.",
	}
	if tier == "" {
		tier = "medium"
	}
	if text[tier] == "" {
		return ""
	}
	return dotprompt.NewHookBlocks("gavel:effort").Wrap("effort", text[tier])
}
