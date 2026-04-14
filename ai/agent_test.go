package ai

import (
	"os"
	"testing"
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
