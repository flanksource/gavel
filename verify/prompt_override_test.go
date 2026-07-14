package verify

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/flanksource/captain/pkg/api"
	"github.com/ghodss/yaml"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPromptOverride_UnmarshalShorthandString(t *testing.T) {
	var o PromptOverride
	require.NoError(t, yaml.Unmarshal([]byte(`"be terse {{patch}}"`), &o))
	text, ok := o.InlineText()
	assert.True(t, ok)
	assert.Equal(t, "be terse {{patch}}", text)
	assert.Empty(t, o.File)
}

func TestPromptOverride_UnmarshalObject(t *testing.T) {
	var o PromptOverride
	require.NoError(t, yaml.Unmarshal([]byte("file: prompts/review.prompt\n"), &o))
	assert.Equal(t, "prompts/review.prompt", o.File)
	assert.Empty(t, o.Inline)
}

func TestPromptOverride_ResolveInlineWins(t *testing.T) {
	o := InlinePrompt("inline body")
	o.File = "ignored.prompt"
	got, err := o.Resolve(t.TempDir(), "EMBEDDED")
	require.NoError(t, err)
	assert.Equal(t, "inline body", got)
}

func TestPromptOverride_ResolveFileRelativeToDir(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "review.prompt"), []byte("from file"), 0o644))
	o := PromptOverride{File: "review.prompt"}
	got, err := o.Resolve(dir, "EMBEDDED")
	require.NoError(t, err)
	assert.Equal(t, "from file", got)
}

func TestPromptOverride_ResolveUnsetReturnsFallback(t *testing.T) {
	got, err := PromptOverride{}.Resolve(t.TempDir(), "EMBEDDED")
	require.NoError(t, err)
	assert.Equal(t, "EMBEDDED", got)
}

func TestPromptOverride_ResolveMissingFileIsError(t *testing.T) {
	o := PromptOverride{File: "does-not-exist.prompt"}
	_, err := o.Resolve(t.TempDir(), "EMBEDDED")
	require.Error(t, err)
}

func TestPromptOverride_IsZero(t *testing.T) {
	assert.True(t, PromptOverride{}.IsZero())
	assert.True(t, InlinePrompt("  ").IsZero())
	assert.False(t, InlinePrompt("x").IsZero())
	assert.False(t, PromptOverride{File: "x"}.IsZero())
}

func TestGavelConfig_PromptOverridesRoundTrip(t *testing.T) {
	const src = `verify:
  promptTemplate: "review strictly {{scopeInstruction}}"
commit:
  messagePrompt:
    file: .gavel/prompts/msg.prompt
`
	var cfg GavelConfig
	require.NoError(t, yaml.Unmarshal([]byte(src), &cfg))
	text, ok := cfg.Verify.PromptTemplate.InlineText()
	assert.True(t, ok)
	assert.Equal(t, "review strictly {{scopeInstruction}}", text)
	assert.Equal(t, ".gavel/prompts/msg.prompt", cfg.Commit.MessagePrompt.File)
}

func TestMergeConfig_PromptOverrideLastWriteWins(t *testing.T) {
	base := GavelConfig{Verify: VerifyConfig{PromptTemplate: InlinePrompt("base")}}
	override := GavelConfig{Verify: VerifyConfig{PromptTemplate: PromptOverride{File: "repo.prompt"}}}

	merged := mergeGavelConfig(base, override)
	assert.Equal(t, "repo.prompt", merged.Verify.PromptTemplate.File)
	assert.False(t, merged.Verify.PromptTemplate.HasInline(), "override replaces, not merges, the prompt")

	// An empty override leaves the base prompt untouched.
	kept := mergeGavelConfig(base, GavelConfig{})
	text, ok := kept.Verify.PromptTemplate.InlineText()
	assert.True(t, ok)
	assert.Equal(t, "base", text)
}

func TestPromptOverride_StructuredInlineSpec(t *testing.T) {
	temp := 0.2
	spec := api.Spec{
		Model: api.Model{Name: "claude-sonnet-5", Temperature: &temp},
		Prompt: api.Prompt{
			User:   "Review {{patch}}",
			System: "Be strict",
		},
	}
	o, err := StructuredInlinePrompt(spec)
	require.NoError(t, err)

	source, err := o.Resolve(t.TempDir(), "")
	require.NoError(t, err)
	assert.Contains(t, source, "model: claude-sonnet-5")
	assert.Contains(t, source, "system: Be strict")
	assert.Contains(t, source, "Review {{patch}}")
	assert.NotContains(t, source, "user: Review")
}

func TestPromptOverride_StructuredInlineYAML(t *testing.T) {
	const src = `inline:
  model: claude-sonnet-5
  prompt:
    user: Review {{patch}}
`
	var o PromptOverride
	require.NoError(t, yaml.Unmarshal([]byte(src), &o))
	assert.True(t, json.Valid(o.Inline))
	resolved, err := o.Resolve(t.TempDir(), "")
	require.NoError(t, err)
	assert.Contains(t, resolved, "model: claude-sonnet-5")
	assert.Contains(t, resolved, "Review {{patch}}")
}

func TestPromptOverride_StructuredInlineRejectsInvalidShapes(t *testing.T) {
	for _, raw := range []string{`[]`, `true`, `42`, `{"model":"claude-sonnet-5","prompt":{"user":""}}`} {
		o := PromptOverride{Inline: json.RawMessage(raw)}
		_, err := o.Resolve(t.TempDir(), "")
		require.Error(t, err, raw)
	}
}

func TestPromptOverride_RejectsNullInline(t *testing.T) {
	var o PromptOverride
	err := yaml.Unmarshal([]byte("inline: null\n"), &o)
	require.Error(t, err)
}

func TestPromptOverride_FileResolvesFromDeclaringConfigLayer(t *testing.T) {
	configDir := t.TempDir()
	targetDir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(configDir, "prompts"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(configDir, "prompts", "review.prompt"), []byte("from declaring layer"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(configDir, ".gavel.yaml"), []byte("verify:\n  promptTemplate:\n    file: prompts/review.prompt\n"), 0o644))

	cfg, err := LoadSingleGavelConfig(filepath.Join(configDir, ".gavel.yaml"))
	require.NoError(t, err)
	resolved, err := cfg.Verify.PromptTemplate.Resolve(targetDir, "fallback")
	require.NoError(t, err)
	assert.Equal(t, "from declaring layer", resolved)
}
