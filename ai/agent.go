// Package ai adapts captain's pkg/ai to the surface gavel consumers previously
// used from clicky/ai, backed entirely by captain. It also performs
// gavel-specific env normalization so that common alternate API-key env vars
// are accepted (e.g. CLAUDE_API_KEY as an alias for ANTHROPIC_API_KEY).
package ai

import (
	"context"
	"os"
	"sync"

	captainai "github.com/flanksource/captain/pkg/ai"
	"github.com/flanksource/captain/pkg/ai/middleware"
)

// Type aliases to captain so existing call sites compile unchanged.
type (
	PromptRequest  = captainai.PromptRequest
	PromptResponse = captainai.PromptResponse
	Costs          = captainai.Costs
	Cost           = captainai.Cost
)

// Agent is the named-prompt / batch / cost surface gavel codes against.
// captain's *ai.Agent satisfies it, and so do gavel test mocks that carry the
// extra GetType/GetConfig/ListModels methods.
type Agent interface {
	ExecutePrompt(ctx context.Context, req PromptRequest) (*PromptResponse, error)
	ExecuteBatch(ctx context.Context, reqs []PromptRequest) (map[string]*PromptResponse, error)
	GetCosts() Costs
	Close() error
}

var envAliases = map[string][]string{
	"ANTHROPIC_API_KEY": {"CLAUDE_API_KEY", "ANTHROPIC_KEY"},
	"OPENAI_API_KEY":    {"OPENAI_KEY"},
	"GEMINI_API_KEY":    {"GOOGLE_GENERATIVE_AI_API_KEY", "GOOGLE_API_KEY"},
}

var normalizeOnce sync.Once

// NewProvider builds the captain provider for cfg (env-normalized, wrapped with
// the logging middleware) and returns it directly. Callers that need the full
// request surface — a working directory, agentic tool/permission knobs — drive
// provider.Execute with a captainai.Request, which the named-prompt PromptRequest
// wrapper cannot express. NewAgent is the higher-level surface built on top.
func NewProvider(cfg AgentConfig) (captainai.Provider, error) {
	NormalizeEnv()
	provider, err := captainai.NewProvider(cfg.toCaptain())
	if err != nil {
		return nil, err
	}
	return middleware.Wrap(provider, middleware.WithLogging())
}

// NewAgent builds a captain-backed agent from cfg after normalizing env keys.
// The backend is inferred from the model by captain. The provider is wrapped with
// captain's logging middleware so the prompt source, rendered input, schema-in and
// schema-out print under -v (Debug/Trace); captain's NewAgent omits middleware, so
// gavel applies it here (the ai package cannot import middleware without a cycle).
func NewAgent(cfg AgentConfig) (Agent, error) {
	provider, err := NewProvider(cfg)
	if err != nil {
		return nil, err
	}
	return captainai.NewAgentWithProvider(provider, cfg.toCaptain()), nil
}

// GetDefaultAgent returns an agent built from DefaultConfig.
func GetDefaultAgent() (Agent, error) {
	return NewAgent(DefaultConfig())
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
