// Package ai wraps commons-db/llm.NewLLMAgent with gavel-specific env
// normalization so that common alternate API-key env vars are accepted
// (e.g. CLAUDE_API_KEY as an alias for ANTHROPIC_API_KEY).
package ai

import (
	"os"
	"sync"

	clickyai "github.com/flanksource/clicky/ai"
	"github.com/flanksource/commons-db/llm"
)

var envAliases = map[string][]string{
	"ANTHROPIC_API_KEY": {"CLAUDE_API_KEY", "ANTHROPIC_KEY"},
	"OPENAI_API_KEY":    {"OPENAI_KEY"},
	"GEMINI_API_KEY":    {"GOOGLE_GENERATIVE_AI_API_KEY", "GOOGLE_API_KEY"},
}

var normalizeOnce sync.Once

func NewAgent(cfg clickyai.AgentConfig) (*llm.LLMAgent, error) {
	NormalizeEnv()
	return llm.NewLLMAgent(cfg)
}

func NormalizeEnv() {
	normalizeOnce.Do(normalizeEnv)
}

func normalizeEnv() {
	for canonical, aliases := range envAliases {
		if os.Getenv(canonical) != "" {
			continue
		}
		for _, alias := range aliases {
			if v := os.Getenv(alias); v != "" {
				_ = os.Setenv(canonical, v)
				break
			}
		}
	}
}
