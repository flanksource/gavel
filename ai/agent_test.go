package ai

import (
	"errors"
	"github.com/flanksource/captain/pkg/api"
	"os"
	"strings"
	"testing"

	captainai "github.com/flanksource/captain/pkg/ai"
)

func TestNormalizeEnv_CopiesClaudeKeyToAnthropic(t *testing.T) {
	clearAllKnownKeys(t)
	t.Setenv("CLAUDE_API_KEY", "sk-ant-test")

	normalizeEnv()

	if got := os.Getenv("ANTHROPIC_API_KEY"); got != "sk-ant-test" {
		t.Fatalf("ANTHROPIC_API_KEY = %q, want %q", got, "sk-ant-test")
	}
}

func TestNormalizeEnv_PrefersExistingAnthropicKey(t *testing.T) {
	clearAllKnownKeys(t)
	t.Setenv("ANTHROPIC_API_KEY", "keep")
	t.Setenv("CLAUDE_API_KEY", "overwrite")

	normalizeEnv()

	if got := os.Getenv("ANTHROPIC_API_KEY"); got != "keep" {
		t.Fatalf("ANTHROPIC_API_KEY = %q, want %q", got, "keep")
	}
}

func TestNormalizeEnv_AnthropicKeyAlias(t *testing.T) {
	clearAllKnownKeys(t)
	t.Setenv("ANTHROPIC_KEY", "sk-ant-alt")

	normalizeEnv()

	if got := os.Getenv("ANTHROPIC_API_KEY"); got != "sk-ant-alt" {
		t.Fatalf("ANTHROPIC_API_KEY = %q, want %q", got, "sk-ant-alt")
	}
}

func TestNormalizeEnv_ClaudePreferredOverAnthropicKey(t *testing.T) {
	clearAllKnownKeys(t)
	t.Setenv("CLAUDE_API_KEY", "first")
	t.Setenv("ANTHROPIC_KEY", "second")

	normalizeEnv()

	if got := os.Getenv("ANTHROPIC_API_KEY"); got != "first" {
		t.Fatalf("ANTHROPIC_API_KEY = %q, want %q (CLAUDE_API_KEY should win, listed first)", got, "first")
	}
}

func TestNormalizeEnv_OpenAIAlias(t *testing.T) {
	clearAllKnownKeys(t)
	t.Setenv("OPENAI_KEY", "sk-openai-test")

	normalizeEnv()

	if got := os.Getenv("OPENAI_API_KEY"); got != "sk-openai-test" {
		t.Fatalf("OPENAI_API_KEY = %q, want %q", got, "sk-openai-test")
	}
}

func TestNormalizeEnv_GeminiAliasPreferenceOrder(t *testing.T) {
	clearAllKnownKeys(t)
	t.Setenv("GOOGLE_GENERATIVE_AI_API_KEY", "first")
	t.Setenv("GOOGLE_API_KEY", "second")

	normalizeEnv()

	if got := os.Getenv("GEMINI_API_KEY"); got != "first" {
		t.Fatalf("GEMINI_API_KEY = %q, want %q (ai-sdk name should win)", got, "first")
	}
}

func TestNormalizeEnv_NoKeyNoChange(t *testing.T) {
	clearAllKnownKeys(t)

	normalizeEnv()

	for _, canonical := range []string{"ANTHROPIC_API_KEY", "OPENAI_API_KEY", "GEMINI_API_KEY"} {
		if got := os.Getenv(canonical); got != "" {
			t.Errorf("%s = %q, want empty", canonical, got)
		}
	}
}

func TestNewProvider_ReportsBackendKeyAndSimilarEnvName(t *testing.T) {
	clearAllKnownKeys(t)
	t.Setenv("OPEN_AI_API_KEY", "must-not-leak")

	_, err := NewProvider(AgentConfig{Model: api.Model{Name: "api:terra"}})
	if err == nil {
		t.Fatal("expected missing OpenAI API key error")
	}
	message := err.Error()
	for _, want := range []string{
		`API key not found for backend "openai"`,
		`model "api:terra"`,
		"set OPENAI_API_KEY (also accepts OPENAI_KEY)",
		"similar environment variable found: OPEN_AI_API_KEY (did you mean OPENAI_API_KEY?)",
	} {
		if !strings.Contains(message, want) {
			t.Fatalf("error = %q, want %q", message, want)
		}
	}
	if strings.Contains(message, "must-not-leak") {
		t.Fatalf("error leaked environment value: %q", message)
	}
	if strings.Contains(message, "ANTHROPIC_API_KEY") || strings.Contains(message, "GEMINI_API_KEY") {
		t.Fatalf("error contains unrelated provider keys: %q", message)
	}
	if !errors.Is(err, captainai.ErrNoAPIKey) {
		t.Fatalf("error = %v, want ErrNoAPIKey", err)
	}
}

func TestNewProvider_DoesNotSuggestKeyWhenConfigured(t *testing.T) {
	clearAllKnownKeys(t)
	t.Setenv("OPENAI_API_KEY", "configured")

	provider, err := NewProvider(AgentConfig{Model: api.Model{Name: "api:terra"}})
	if err != nil {
		t.Fatalf("NewProvider(api:terra) with OPENAI_API_KEY: %v", err)
	}
	if provider == nil {
		t.Fatal("NewProvider(api:terra) returned nil provider")
	}
}

func TestWithCredentialHint_IgnoresNonCredentialErrors(t *testing.T) {
	cause := errors.New("model unavailable")
	if got := withCredentialHint(AgentConfig{Model: api.Model{Name: "api:terra"}}, cause); got != cause {
		t.Fatalf("non-credential error was decorated: %v", got)
	}
}

// clearAllKnownKeys unsets every canonical + alias in envAliases for the
// duration of the test. Uses t.Setenv so the originals are restored on
// cleanup.
func clearAllKnownKeys(t *testing.T) {
	t.Helper()
	for canonical, aliases := range envAliases {
		t.Setenv(canonical, "")
		_ = os.Unsetenv(canonical)
		for _, a := range aliases {
			t.Setenv(a, "")
			_ = os.Unsetenv(a)
		}
	}
}
