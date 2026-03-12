package git

import (
	"os"
	"path/filepath"
	"testing"

	. "github.com/flanksource/gavel/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadDefaultAnalyzeConfig(t *testing.T) {
	conf, err := loadDefaultAnalyzeConfig()
	require.NoError(t, err)
	require.NotNil(t, conf)

	assert.Contains(t, conf.FilterSets, "bots")
	assert.Contains(t, conf.FilterSets, "noise")
	assert.Contains(t, conf.FilterSets, "merges")
	assert.Contains(t, conf.Includes, "bots")
	assert.Contains(t, conf.Includes, "noise")
	assert.Contains(t, conf.Includes, "merges")
}

func TestMergeConfigs(t *testing.T) {
	base := &GitAnalyzeConfig{
		FilterSets:    map[string]FilterSet{"bots": {IgnoreAuthors: []string{"dependabot*"}}},
		Includes:      []string{"bots"},
		IgnoreFiles:   []string{"*.lock"},
		IgnoreAuthors: []string{"ci-bot"},
	}

	user := &GitAnalyzeConfig{
		FilterSets:  map[string]FilterSet{"custom": {IgnoreFiles: []string{"vendor/*"}}},
		Includes:    []string{"custom"},
		IgnoreFiles: []string{"*.svg"},
	}

	merged := base.Merge(user)

	assert.Contains(t, merged.FilterSets, "bots")
	assert.Contains(t, merged.FilterSets, "custom")
	assert.Equal(t, []string{"custom"}, merged.Includes)
	assert.Contains(t, merged.IgnoreFiles, "*.lock")
	assert.Contains(t, merged.IgnoreFiles, "*.svg")
	assert.Contains(t, merged.IgnoreAuthors, "ci-bot")
}

func TestMergeConfigNilUser(t *testing.T) {
	base := &GitAnalyzeConfig{
		Includes:    []string{"bots"},
		IgnoreFiles: []string{"*.lock"},
	}
	merged := base.Merge(nil)
	assert.Equal(t, base, merged)
}

func TestResolveActiveFilters(t *testing.T) {
	conf := &GitAnalyzeConfig{
		FilterSets: map[string]FilterSet{
			"bots":  {IgnoreAuthors: []string{"dependabot*"}},
			"noise": {IgnoreFiles: []string{"*.lock"}},
			"extra": {IgnoreCommitTypes: []string{"chore"}},
		},
		Includes:    []string{"bots", "noise"},
		IgnoreFiles: []string{"*.svg"},
	}

	t.Run("default includes", func(t *testing.T) {
		resolved := conf.ResolveActiveFilters(nil, nil)
		assert.Contains(t, resolved.IgnoreAuthors, "dependabot*")
		assert.Contains(t, resolved.IgnoreFiles, "*.lock")
		assert.Contains(t, resolved.IgnoreFiles, "*.svg")
		assert.NotContains(t, resolved.IgnoreCommitTypes, "chore")
	})

	t.Run("include adds set", func(t *testing.T) {
		resolved := conf.ResolveActiveFilters([]string{"extra"}, nil)
		assert.Contains(t, resolved.IgnoreCommitTypes, "chore")
		assert.Contains(t, resolved.IgnoreAuthors, "dependabot*")
	})

	t.Run("exclude removes set", func(t *testing.T) {
		resolved := conf.ResolveActiveFilters(nil, []string{"bots"})
		assert.NotContains(t, resolved.IgnoreAuthors, "dependabot*")
		assert.Contains(t, resolved.IgnoreFiles, "*.lock")
	})
}

func TestCompileCELRules(t *testing.T) {
	conf := &GitAnalyzeConfig{
		IgnoreCommitRules: []CommitRule{
			{CEL: "commit.is_merge"},
			{CEL: "commit.line_changes > 10000"},
		},
	}
	err := conf.Compile()
	require.NoError(t, err)
	assert.Len(t, conf.compiledCommitRules, 2)
}

func TestCompileCELRulesInvalid(t *testing.T) {
	conf := &GitAnalyzeConfig{
		IgnoreCommitRules: []CommitRule{
			{CEL: "invalid syntax !!!"},
		},
	}
	err := conf.Compile()
	assert.Error(t, err)
}

func TestFindAnalyzeConfig(t *testing.T) {
	dir := t.TempDir()

	// Create a fake .git directory so FindGitRoot works
	require.NoError(t, os.Mkdir(filepath.Join(dir, ".git"), 0o755))

	configContent := `
ignore_files:
  - "*.svg"
ignore_authors:
  - "bot@example.com"
`
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".gitanalyze.yaml"), []byte(configContent), 0o644))

	subDir := filepath.Join(dir, "sub", "dir")
	require.NoError(t, os.MkdirAll(subDir, 0o755))

	conf, err := findAnalyzeConfig(subDir)
	require.NoError(t, err)
	require.NotNil(t, conf)
	assert.Contains(t, conf.IgnoreFiles, "*.svg")
	assert.Contains(t, conf.IgnoreAuthors, "bot@example.com")
}

func TestMatchesAuthor(t *testing.T) {
	commit := Commit{
		Author:    Author{Name: "dependabot[bot]", Email: "dependabot@github.com"},
		Committer: Author{Name: "GitHub", Email: "noreply@github.com"},
	}

	matched, reason := matchesAuthor(commit, []string{"dependabot*"})
	assert.True(t, matched)
	assert.Contains(t, reason, "dependabot")

	matched, _ = matchesAuthor(commit, []string{"renovate*"})
	assert.False(t, matched)
}

func TestMatchesCommitMessage(t *testing.T) {
	matched, _ := matchesCommitMessage("fixup! some commit", []string{"fixup!*"})
	assert.True(t, matched)

	matched, _ = matchesCommitMessage("feat: new feature", []string{"fixup!*"})
	assert.False(t, matched)
}

func TestMatchesCommitType(t *testing.T) {
	matched, _ := matchesCommitType(CommitType("chore"), []string{"chore", "ci"})
	assert.True(t, matched)

	matched, _ = matchesCommitType(CommitType("feat"), []string{"chore", "ci"})
	assert.False(t, matched)
}

func TestMatchesFile(t *testing.T) {
	matched, _ := matchesFile("package-lock.json", []string{"package-lock.json"})
	assert.True(t, matched)

	matched, _ = matchesFile("src/utils/helper.go", []string{"*.lock"})
	assert.False(t, matched)

	matched, _ = matchesFile("go.sum", []string{"go.sum"})
	assert.True(t, matched)

	matched, _ = matchesFile("path/to/something.lock", []string{"*.lock"})
	assert.True(t, matched)
}
