package verify

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ghodss/yaml"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPromptOverride_UnmarshalShorthandString(t *testing.T) {
	var o PromptOverride
	require.NoError(t, yaml.Unmarshal([]byte(`"be terse {{patch}}"`), &o))
	assert.Equal(t, "be terse {{patch}}", o.Inline)
	assert.Empty(t, o.File)
}

func TestPromptOverride_UnmarshalObject(t *testing.T) {
	var o PromptOverride
	require.NoError(t, yaml.Unmarshal([]byte("file: prompts/review.prompt\n"), &o))
	assert.Equal(t, "prompts/review.prompt", o.File)
	assert.Empty(t, o.Inline)
}

func TestPromptOverride_ResolveInlineWins(t *testing.T) {
	o := PromptOverride{Inline: "inline body", File: "ignored.prompt"}
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
	assert.True(t, PromptOverride{Inline: "  "}.IsZero())
	assert.False(t, PromptOverride{Inline: "x"}.IsZero())
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
	assert.Equal(t, "review strictly {{scopeInstruction}}", cfg.Verify.PromptTemplate.Inline)
	assert.Equal(t, ".gavel/prompts/msg.prompt", cfg.Commit.MessagePrompt.File)
}

func TestMergeConfig_PromptOverrideLastWriteWins(t *testing.T) {
	base := GavelConfig{Verify: VerifyConfig{PromptTemplate: PromptOverride{Inline: "base"}}}
	override := GavelConfig{Verify: VerifyConfig{PromptTemplate: PromptOverride{File: "repo.prompt"}}}

	merged := mergeGavelConfig(base, override)
	assert.Equal(t, "repo.prompt", merged.Verify.PromptTemplate.File)
	assert.Empty(t, merged.Verify.PromptTemplate.Inline, "override replaces, not merges, the prompt")

	// An empty override leaves the base prompt untouched.
	kept := mergeGavelConfig(base, GavelConfig{})
	assert.Equal(t, "base", kept.Verify.PromptTemplate.Inline)
}
