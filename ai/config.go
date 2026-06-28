package ai

import (
	"time"

	captainai "github.com/flanksource/captain/pkg/ai"
	"github.com/spf13/pflag"
)

// AgentType mirrors clicky/ai.AgentType.
type AgentType string

// Supported agent types.
const (
	AgentTypeClaude AgentType = "claude"
	AgentTypeAider  AgentType = "aider"
)

// Model mirrors clicky/ai.Model for the fields gavel and its tests reference.
type Model struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Provider    string            `json:"provider"`
	Metadata    map[string]string `json:"metadata,omitempty"`
	InputPrice  float64           `json:"input_price_per_token,omitempty"`
	OutputPrice float64           `json:"output_price_per_token,omitempty"`
	MaxTokens   int               `json:"max_tokens,omitempty"`
}

// AgentConfig mirrors clicky/ai.AgentConfig for the fields gavel uses. It is
// converted to a captain ai.Config when building an agent.
type AgentConfig struct {
	Type            AgentType     `json:"type"`
	Model           string        `json:"model"`
	CacheDBPath     string        `json:"cache_db_path,omitempty"`
	ProjectName     string        `json:"project_name,omitempty"`
	CacheTTL        time.Duration `json:"cache_ttl,omitempty"`
	Temperature     float64       `json:"temperature,omitempty"`
	MaxTokens       int           `json:"max_tokens"`
	MaxConcurrent   int           `json:"max_concurrent"`
	Debug           bool          `json:"debug"`
	Verbose         bool          `json:"verbose"`
	StrictMCPConfig bool          `json:"strict_mcp_config"`
	NoCache         bool          `json:"no_cache,omitempty"`
}

// toCaptain maps the gavel config onto captain's ai.Config. Backend is left
// empty so captain infers it from the model.
func (c AgentConfig) toCaptain() captainai.Config {
	return captainai.Config{
		Model:         c.Model,
		MaxTokens:     c.MaxTokens,
		Temperature:   c.Temperature,
		MaxConcurrent: c.MaxConcurrent,
		NoCache:       c.NoCache,
		CacheTTL:      c.CacheTTL,
		CacheDBPath:   c.CacheDBPath,
		ProjectName:   c.ProjectName,
	}
}

var defaultConfig = AgentConfig{
	Type:            AgentTypeClaude,
	Model:           "claude-haiku-4-5",
	MaxTokens:       10000,
	MaxConcurrent:   4,
	StrictMCPConfig: true,
	CacheTTL:        24 * time.Hour,
}

// DefaultConfig returns the default agent configuration.
func DefaultConfig() AgentConfig {
	return defaultConfig
}

// BindFlags adds AI-related flags to the flag set, mutating the package-level
// default config the same way clicky/ai did.
func BindFlags(flags *pflag.FlagSet) {
	agentType := string(defaultConfig.Type)
	flags.StringVar(&agentType, "agent", agentType, "AI agent type (claude, aider)")
	flags.BoolVar(&defaultConfig.Debug, "ai-debug", defaultConfig.Debug, "Enable AI debug output")
	flags.BoolVar(&defaultConfig.Verbose, "ai-verbose", defaultConfig.Verbose, "Enable AI verbose logging")
	flags.StringVar(&defaultConfig.Model, "ai-model", defaultConfig.Model, "AI model to use")
	flags.IntVar(&defaultConfig.MaxTokens, "ai-max-tokens", defaultConfig.MaxTokens, "Maximum tokens per request")
	flags.IntVar(&defaultConfig.MaxConcurrent, "ai-max-concurrent", defaultConfig.MaxConcurrent, "Maximum concurrent AI requests")
	flags.Float64Var(&defaultConfig.Temperature, "ai-temperature", defaultConfig.Temperature, "AI temperature (0.0-2.0)")
	flags.BoolVar(&defaultConfig.StrictMCPConfig, "ai-strict-mcp", defaultConfig.StrictMCPConfig, "Use strict MCP configuration (Claude only)")
	flags.DurationVar(&defaultConfig.CacheTTL, "ai-cache-ttl", defaultConfig.CacheTTL, "AI cache TTL (e.g., 24h, 7d)")
	flags.BoolVar(&defaultConfig.NoCache, "ai-no-cache", defaultConfig.NoCache, "Disable AI response caching")
	flags.StringVar(&defaultConfig.CacheDBPath, "ai-cache-db", defaultConfig.CacheDBPath, "Path to AI cache database (default: ~/.cache/clicky-ai.db)")
	flags.StringVar(&defaultConfig.ProjectName, "ai-project", defaultConfig.ProjectName, "Project name for cache grouping")
	flags.Bool("aider", false, "Use Aider agent (shorthand for --agent=aider)")
	flags.Bool("claude", false, "Use Claude agent (shorthand for --agent=claude)")
	defaultConfig.Type = AgentType(agentType)
}
