// Package ai adapts captain's pkg/ai to the surface gavel consumers previously
// used from clicky/ai, backed entirely by captain. It also performs
// gavel-specific env normalization so that common alternate API-key env vars
// are accepted (e.g. CLAUDE_API_KEY as an alias for ANTHROPIC_API_KEY).
package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"

	captainai "github.com/flanksource/captain/pkg/ai"
	"github.com/flanksource/captain/pkg/ai/middleware"
	_ "github.com/flanksource/captain/pkg/ai/provider"
	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/captain/pkg/collections"
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

// NewProvider builds the captain provider for cfg (env-normalized) with captain's
// default middleware applied: request/response logging always, plus response
// caching when cfg configures it (CacheTTL/CacheDBPath with NoCache unset). This
// is the batteries-included middleware.NewProvider, which also makes the
// --ai-cache-* flags actually take effect. Callers that need the full request
// surface — a working directory, agentic tool/permission knobs — drive
// provider.Execute with a captainai.Request, which the named-prompt PromptRequest
// wrapper cannot express. NewAgent is the higher-level surface built on top.
func NewProvider(cfg AgentConfig) (captainai.Provider, error) {
	NormalizeEnv()
	// Resolve once, here, so the provider is built from a model that already
	// carries its concrete backend rather than one re-derived from a bare name.
	resolved, err := captainai.ResolveModelSelectors(cfg.Model)
	if err != nil {
		return nil, err
	}
	runCfg := cfg
	runCfg.Model = resolved
	provider, err := middleware.NewProvider(runCfg)
	if err != nil {
		// The hint quotes the ORIGINAL cfg: an error should echo the model the
		// user actually typed ("api:terra"), not the id it resolved to.
		return nil, withCredentialHint(cfg, err)
	}
	return provider, nil
}

// CloseProvider releases a provider built by NewProvider. Process-backed
// backends (claude-agent, cmux, codex-appserver) run their agent as a
// clicky-supervised child process that only Close stops; leaving it running
// keeps its task incomplete and blocks process exit indefinitely, which is
// exactly how `gavel pr status --ai-fix` ended up spinning on
// "Still waiting after 8m0s for: .../claude-agent/node_modules/.bin/tsx".
//
// The lookup goes through captain's ProviderAs so it still finds Close through
// the middleware wrapper NewProvider returns. Backends with nothing to release
// are a no-op.
func CloseProvider(provider captainai.Provider) error {
	if closer, ok := api.ProviderAs[api.CloseableProvider](provider); ok {
		return closer.Close()
	}
	return nil
}

// NewAgent builds a captain-backed agent from cfg after normalizing env keys.
// The backend is inferred from the model by captain. The provider carries
// captain's default middleware (via NewProvider) so the prompt source, rendered
// input, schema-in and schema-out print under -v (Debug/Trace); captain's own
// ai.NewAgent omits middleware, so gavel routes through middleware.NewProvider.
func NewAgent(cfg AgentConfig) (Agent, error) {
	provider, err := NewProvider(cfg)
	if err != nil {
		return nil, err
	}
	return captainai.NewAgentWithProvider(provider, cfg), nil
}

// GetDefaultAgent returns an agent built from DefaultConfig.
func GetDefaultAgent() (Agent, error) {
	return NewAgent(DefaultConfig())
}

// DecodeStructured unmarshals a prompt response's structured output into target.
// It prefers the provider's validated StructuredData (raw JSON) and falls back
// to the Result text, failing loudly when neither yields the target shape. Use
// it with prompts whose output schema is declared in the .prompt frontmatter
// (PromptRequest.SchemaJSON) rather than bound to a Go struct.
func DecodeStructured(resp *PromptResponse, target any) error {
	if raw, ok := resp.StructuredData.(json.RawMessage); ok && len(raw) > 0 {
		return json.Unmarshal(raw, target)
	}
	if resp.Result != "" {
		return json.Unmarshal([]byte(resp.Result), target)
	}
	return fmt.Errorf("response carried no structured output to decode into %T", target)
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

func withCredentialHint(cfg AgentConfig, err error) error {
	if !captainai.IsMissingAPIKey(err) {
		return err
	}
	backend := credentialBackend(cfg)
	canonical := captainai.AuthEnvVars(backend)
	if len(canonical) == 0 {
		return err
	}

	accepted := append([]string(nil), canonical...)
	for _, name := range canonical {
		accepted = appendUnique(accepted, envAliases[name]...)
	}
	hint := "set " + strings.Join(canonical, " or ")
	aliases := difference(accepted, canonical)
	if len(aliases) > 0 {
		hint += " (also accepts " + strings.Join(aliases, " or ") + ")"
	}
	if suggestions := similarCredentialEnvVars(accepted); len(suggestions) > 0 {
		parts := make([]string, len(suggestions))
		for i, suggestion := range suggestions {
			parts[i] = fmt.Sprintf("%s (did you mean %s?)", suggestion.found, suggestion.expected)
		}
		hint += "; similar environment variable"
		if len(parts) > 1 {
			hint += "s"
		}
		hint += " found: " + strings.Join(parts, ", ")
	}
	return &credentialHintError{
		cause:   err,
		model:   cfg.Model.Name,
		backend: backend,
		hint:    hint,
	}
}

type credentialHintError struct {
	cause   error
	model   string
	backend captainai.Backend
	hint    string
}

func (e *credentialHintError) Error() string {
	return fmt.Sprintf("API key not found for backend %q (model %q; %s)", e.backend, e.model, e.hint)
}

func (e *credentialHintError) Unwrap() error { return e.cause }

// credentialBackend reports which provider's credentials a config needs.
//
// It reads the backend off the model rather than re-deriving it from the name:
// NewProvider resolves cfg.Model before this can be reached, and re-inferring
// from a bare name is how a model's backend gets replaced by a guess.
func credentialBackend(cfg AgentConfig) captainai.Backend {
	if cfg.Model.Backend.Valid() {
		return cfg.Model.Backend
	}
	resolved, err := captainai.ResolveModelSelectors(cfg.Model)
	if err != nil {
		return ""
	}
	return resolved.Backend
}

func appendUnique(values []string, additions ...string) []string {
	seen := make(map[string]bool, len(values)+len(additions))
	for _, value := range values {
		seen[value] = true
	}
	for _, value := range additions {
		if value != "" && !seen[value] {
			values = append(values, value)
			seen[value] = true
		}
	}
	return values
}

func difference(values, excluded []string) []string {
	skip := make(map[string]bool, len(excluded))
	for _, value := range excluded {
		skip[value] = true
	}
	var out []string
	for _, value := range values {
		if !skip[value] {
			out = append(out, value)
		}
	}
	return out
}

type credentialEnvSuggestion struct {
	found    string
	expected string
	distance int
}

func similarCredentialEnvVars(expected []string) []credentialEnvSuggestion {
	expectedSet := make(map[string]bool, len(expected))
	for _, name := range expected {
		expectedSet[name] = true
	}

	var suggestions []credentialEnvSuggestion
	for _, item := range os.Environ() {
		name, _, _ := strings.Cut(item, "=")
		if expectedSet[name] {
			continue
		}
		closest := credentialEnvSuggestion{found: name, distance: -1}
		for _, candidate := range expected {
			distance := collections.Levenshtein(strings.ToUpper(name), candidate)
			if closest.distance < 0 || distance < closest.distance {
				closest.expected = candidate
				closest.distance = distance
			}
		}
		if closest.distance >= 0 && closest.distance <= 2 {
			suggestions = append(suggestions, closest)
		}
	}
	sort.Slice(suggestions, func(i, j int) bool {
		if suggestions[i].distance != suggestions[j].distance {
			return suggestions[i].distance < suggestions[j].distance
		}
		return suggestions[i].found < suggestions[j].found
	})
	if len(suggestions) > 3 {
		suggestions = suggestions[:3]
	}
	return suggestions
}
