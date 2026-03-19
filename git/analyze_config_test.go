package git

import (
	"testing"

	"github.com/flanksource/repomap"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExcludeConfigCompile(t *testing.T) {
	cfg := &repomap.ExcludeConfig{
		Rules: []repomap.ExcludeRule{
			{When: "commit.is_merge"},
			{When: "commit.line_changes > 10000"},
		},
	}

	compiled, err := cfg.Compile()
	require.NoError(t, err)
	assert.Len(t, compiled.CompiledRules(), 2)
}

func TestExcludeConfigCompileInvalid(t *testing.T) {
	cfg := &repomap.ExcludeConfig{
		Rules: []repomap.ExcludeRule{
			{When: "invalid syntax !!!"},
		},
	}

	_, err := cfg.Compile()
	assert.Error(t, err)
}

func TestExcludeConfigMerge(t *testing.T) {
	base := repomap.ExcludeConfig{
		Files:   []string{"*.lock"},
		Authors: []string{"ci-bot"},
	}
	other := repomap.ExcludeConfig{
		Files: []string{"*.svg"},
	}

	merged := base.Merge(other)
	assert.Contains(t, merged.Files, "*.lock")
	assert.Contains(t, merged.Files, "*.svg")
	assert.Contains(t, merged.Authors, "ci-bot")
}

func TestExcludeConfigResolvePresets(t *testing.T) {
	presets := map[string]repomap.Preset{
		"bots":  {Exclude: repomap.ExcludeConfig{Authors: []string{"dependabot*"}}},
		"noise": {Exclude: repomap.ExcludeConfig{Files: []string{"*.lock"}}},
	}

	cfg := repomap.ExcludeConfig{Files: []string{"*.svg"}}
	cfg.ResolvePresets([]string{"preset:bots", "preset:noise"}, presets)

	assert.Contains(t, cfg.Authors, "dependabot*")
	assert.Contains(t, cfg.Files, "*.lock")
	assert.Contains(t, cfg.Files, "*.svg")
}
