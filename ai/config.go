package ai

import (
	"time"

	captainai "github.com/flanksource/captain/pkg/ai"
	"github.com/flanksource/captain/pkg/api"
	"github.com/spf13/pflag"
)

// AgentConfig is captain's run config. It used to be a flat mirror of it —
// Model/Backend/Temperature as loose strings — which could not express a model's
// effort or fallbacks at all, so those were silently unrepresentable no matter
// what the caller passed. Speaking captain's own type end to end is what keeps a
// model's backend, mode and effort attached to it.
type AgentConfig = captainai.Config

// DefaultConfig returns the default agent configuration.
//
// It returns a value, not a shared package var: BindFlags used to bind pflag
// pointers straight into a package-level default, so every FlagSet that bound it
// (status, git analyze, git amend) aliased the same struct and the last one to
// parse won.
//
// It deliberately carries NO model. It used to pin "claude-haiku-4-5", which made
// the bug class possible twice over: an agent could be built without anyone
// choosing a model, and because BindFlags merges --ai-model onto this struct, the
// hardcoded name always outranked whatever .gavel.yaml configured. The model is
// now the caller's to resolve — see verify.GavelConfig.ModelFor.
func DefaultConfig() AgentConfig {
	return AgentConfig{
		Budget:        api.Budget{MaxTokens: 10000},
		MaxConcurrent: 4,
		CacheTTL:      24 * time.Hour,
	}
}

// modelFlag adapts --ai-model onto a structured api.Model. Expanding on Set is
// what gives the flag captain's compact grammar ("agent:opus:high") rather than a
// bare name whose backend has to be re-guessed later.
type modelFlag struct{ cfg *AgentConfig }

func (f modelFlag) String() string {
	if f.cfg == nil {
		return ""
	}
	return f.cfg.Model.Name
}

func (f modelFlag) Set(value string) error {
	m, err := (api.Model{Name: value}).Expand()
	if err != nil {
		return err
	}
	f.cfg.Model = f.cfg.Model.Merge(m)
	return nil
}

func (modelFlag) Type() string { return "string" }

// BindFlags adds the AI flags to a flag set, writing into cfg.
//
// The flags keep their --ai-* names: gavel already spends --model on `git
// analyze`, and flag names are global to a command, so renaming these would
// collide and panic cobra at init.
func BindFlags(flags *pflag.FlagSet, cfg *AgentConfig) {
	flags.Var(modelFlag{cfg: cfg}, "ai-model", "AI model to use, e.g. claude-sonnet-5 or a compact selector like agent:opus:high")
	flags.IntVar(&cfg.Budget.MaxTokens, "ai-max-tokens", cfg.Budget.MaxTokens, "Maximum tokens per request")
	flags.IntVar(&cfg.MaxConcurrent, "ai-max-concurrent", cfg.MaxConcurrent, "Maximum concurrent AI requests")
	flags.DurationVar(&cfg.CacheTTL, "ai-cache-ttl", cfg.CacheTTL, "AI cache TTL (e.g., 24h, 7d)")
	flags.BoolVar(&cfg.NoCache, "ai-no-cache", cfg.NoCache, "Disable AI response caching")
	flags.StringVar(&cfg.CacheDBPath, "ai-cache-db", cfg.CacheDBPath, "Path to AI cache database (default: ~/.cache/clicky-ai.db)")
	flags.StringVar(&cfg.ProjectName, "ai-project", cfg.ProjectName, "Project name for cache grouping")
}
