package verify

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/flanksource/gavel/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLintIgnoreRule_MatchesViolation(t *testing.T) {
	tests := []struct {
		name  string
		rule  LintIgnoreRule
		v     models.Violation
		match bool
	}{
		{
			name:  "rule only matches",
			rule:  LintIgnoreRule{Rule: "errcheck"},
			v:     models.Violation{Source: "golangci-lint", Rule: &models.Rule{Method: "errcheck"}},
			match: true,
		},
		{
			name:  "rule only no match",
			rule:  LintIgnoreRule{Rule: "errcheck"},
			v:     models.Violation{Source: "golangci-lint", Rule: &models.Rule{Method: "unused"}},
			match: false,
		},
		{
			name:  "source only matches",
			rule:  LintIgnoreRule{Source: "eslint"},
			v:     models.Violation{Source: "eslint", Rule: &models.Rule{Method: "no-unused-vars"}},
			match: true,
		},
		{
			name:  "source only no match",
			rule:  LintIgnoreRule{Source: "eslint"},
			v:     models.Violation{Source: "ruff"},
			match: false,
		},
		{
			name:  "source + rule matches",
			rule:  LintIgnoreRule{Source: "golangci-lint", Rule: "errcheck"},
			v:     models.Violation{Source: "golangci-lint", Rule: &models.Rule{Method: "errcheck"}},
			match: true,
		},
		{
			name:  "source matches rule does not",
			rule:  LintIgnoreRule{Source: "golangci-lint", Rule: "errcheck"},
			v:     models.Violation{Source: "golangci-lint", Rule: &models.Rule{Method: "unused"}},
			match: false,
		},
		{
			name:  "rule + file matches",
			rule:  LintIgnoreRule{Rule: "errcheck", File: "pkg/foo.go"},
			v:     models.Violation{Source: "golangci-lint", Rule: &models.Rule{Method: "errcheck"}, File: "pkg/foo.go"},
			match: true,
		},
		{
			name:  "rule matches file does not",
			rule:  LintIgnoreRule{Rule: "errcheck", File: "pkg/foo.go"},
			v:     models.Violation{Source: "golangci-lint", Rule: &models.Rule{Method: "errcheck"}, File: "pkg/bar.go"},
			match: false,
		},
		{
			name:  "file glob matches",
			rule:  LintIgnoreRule{Rule: "errcheck", File: "pkg/**/*.go"},
			v:     models.Violation{Source: "golangci-lint", Rule: &models.Rule{Method: "errcheck"}, File: "pkg/sub/foo.go"},
			match: true,
		},
		{
			name:  "nil rule on violation",
			rule:  LintIgnoreRule{Rule: "errcheck"},
			v:     models.Violation{Source: "golangci-lint"},
			match: false,
		},
		{
			name:  "empty rule and source invalid",
			rule:  LintIgnoreRule{File: "pkg/foo.go"},
			v:     models.Violation{File: "pkg/foo.go"},
			match: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.match, tt.rule.MatchesViolation(tt.v))
		})
	}
}

func TestLoadGavelConfig_WithLintIgnore(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.Mkdir(filepath.Join(dir, ".git"), 0o755))

	cfgData := []byte(`verify:
  model: gemini
lint:
  ignore:
    - rule: errcheck
      source: golangci-lint
    - rule: unused-import
      file: "pkg/foo.go"
`)
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".gavel.yaml"), cfgData, 0o644))

	cfg, err := LoadGavelConfig(dir)
	require.NoError(t, err)

	assert.Equal(t, "gemini", cfg.Verify.Model)
	assert.Len(t, cfg.Lint.Ignore, 2)
	assert.Equal(t, "errcheck", cfg.Lint.Ignore[0].Rule)
	assert.Equal(t, "golangci-lint", cfg.Lint.Ignore[0].Source)
	assert.Equal(t, "unused-import", cfg.Lint.Ignore[1].Rule)
	assert.Equal(t, "pkg/foo.go", cfg.Lint.Ignore[1].File)
}

func TestSaveGavelConfig_RoundTrip(t *testing.T) {
	dir := t.TempDir()

	cfg := GavelConfig{
		Verify: VerifyConfig{Model: "claude"},
		Lint: LintConfig{
			Ignore: []LintIgnoreRule{
				{Rule: "errcheck", Source: "golangci-lint"},
				{Rule: "no-unused-vars", File: "src/legacy.ts"},
			},
		},
	}

	require.NoError(t, SaveGavelConfig(dir, cfg))

	require.NoError(t, os.Mkdir(filepath.Join(dir, ".git"), 0o755))
	loaded, err := LoadGavelConfig(dir)
	require.NoError(t, err)

	assert.Equal(t, cfg.Verify.Model, loaded.Verify.Model)
	assert.Equal(t, cfg.Lint.Ignore, loaded.Lint.Ignore)
}

func TestMergeLintConfig(t *testing.T) {
	base := LintConfig{
		Ignore: []LintIgnoreRule{{Rule: "errcheck"}},
	}
	override := LintConfig{
		Ignore: []LintIgnoreRule{{Rule: "unused", Source: "ruff"}},
	}
	merged := MergeLintConfig(base, override)
	assert.Len(t, merged.Ignore, 2)
	assert.Equal(t, "errcheck", merged.Ignore[0].Rule)
	assert.Equal(t, "unused", merged.Ignore[1].Rule)
}
