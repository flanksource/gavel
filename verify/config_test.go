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

func TestLoadGavelConfig_WithFixtures(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.Mkdir(filepath.Join(dir, ".git"), 0o755))

	cfgData := []byte(`fixtures:
  enabled: true
  files:
    - "specs/*.fixture.md"
    - "tests/**/*.fixture.md"
`)
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".gavel.yaml"), cfgData, 0o644))

	cfg, err := LoadGavelConfig(dir)
	require.NoError(t, err)
	assert.True(t, cfg.Fixtures.Enabled)
	assert.Equal(t, []string{"specs/*.fixture.md", "tests/**/*.fixture.md"}, cfg.Fixtures.Files)
}

func TestFixturesConfig_ResolvedFiles_Default(t *testing.T) {
	empty := FixturesConfig{}
	assert.Equal(t, []string{DefaultFixturesGlob}, empty.ResolvedFiles())

	custom := FixturesConfig{Files: []string{"a.md", "b.md"}}
	assert.Equal(t, []string{"a.md", "b.md"}, custom.ResolvedFiles())
}

func TestMergeFixturesConfig(t *testing.T) {
	t.Run("override enables", func(t *testing.T) {
		merged := MergeFixturesConfig(FixturesConfig{}, FixturesConfig{Enabled: true})
		assert.True(t, merged.Enabled)
	})
	t.Run("override files replace base", func(t *testing.T) {
		base := FixturesConfig{Files: []string{"old.md"}}
		override := FixturesConfig{Files: []string{"new.md"}}
		merged := MergeFixturesConfig(base, override)
		assert.Equal(t, []string{"new.md"}, merged.Files)
	})
	t.Run("override empty keeps base files", func(t *testing.T) {
		base := FixturesConfig{Enabled: true, Files: []string{"base.md"}}
		merged := MergeFixturesConfig(base, FixturesConfig{})
		assert.Equal(t, []string{"base.md"}, merged.Files)
		assert.True(t, merged.Enabled)
	})
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

func TestLoadGavelConfig_WithPushHooksAndSSH(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.Mkdir(filepath.Join(dir, ".git"), 0o755))

	cfgData := []byte(`verify:
  model: claude
pre:
  - name: deps
    run: make deps
  - run: echo warming
post:
  - name: notify
    run: slack post "$RESULT"
ssh:
  cmd: make ci
`)
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".gavel.yaml"), cfgData, 0o644))

	cfg, err := LoadGavelConfig(dir)
	require.NoError(t, err)

	require.Len(t, cfg.Pre, 2)
	assert.Equal(t, "deps", cfg.Pre[0].Name)
	assert.Equal(t, "make deps", cfg.Pre[0].Run)
	assert.Equal(t, "", cfg.Pre[1].Name)
	assert.Equal(t, "echo warming", cfg.Pre[1].Run)

	require.Len(t, cfg.Post, 1)
	assert.Equal(t, "notify", cfg.Post[0].Name)
	assert.Equal(t, `slack post "$RESULT"`, cfg.Post[0].Run)

	assert.Equal(t, "make ci", cfg.SSH.Cmd)
}

func TestMergeSSHConfig(t *testing.T) {
	t.Run("override replaces cmd", func(t *testing.T) {
		merged := MergeSSHConfig(SSHConfig{Cmd: "make old"}, SSHConfig{Cmd: "make new"})
		assert.Equal(t, "make new", merged.Cmd)
	})
	t.Run("empty override keeps base", func(t *testing.T) {
		merged := MergeSSHConfig(SSHConfig{Cmd: "make old"}, SSHConfig{})
		assert.Equal(t, "make old", merged.Cmd)
	})
}

// TestLoadGavelConfig_RepoRoot asserts that the .gavel.yaml committed at the
// repo root parses into the current schema. It doubles as a smoke test that
// every checked-in config key has a Go field and as dogfooding so a typo in
// .gavel.yaml fails CI instead of silently breaking the SSH push flow.
func TestLoadGavelConfig_RepoRoot(t *testing.T) {
	// Locate the repo root from this test file (verify/config_test.go).
	wd, err := os.Getwd()
	require.NoError(t, err)
	repoRoot := filepath.Dir(wd) // parent of verify/
	path := filepath.Join(repoRoot, ".gavel.yaml")
	if _, err := os.Stat(path); err != nil {
		t.Skipf("no .gavel.yaml at %s: %v", path, err)
	}

	cfg, err := LoadGavelConfig(repoRoot)
	require.NoError(t, err)

	require.NotEmpty(t, cfg.Pre, "expected at least one top-level pre hook")
	assert.Equal(t, "deps", cfg.Pre[0].Name)
	assert.NotEmpty(t, cfg.Pre[0].Run)
	assert.NotEmpty(t, cfg.SSH.Cmd, "expected ssh.cmd to be set")
}

func TestMergePrePostHooks_Append(t *testing.T) {
	// Pre/Post hooks from multiple config sources accumulate in declaration
	// order (home → repo → cwd), so a user's personal hooks don't get
	// silently wiped by a repo config and vice versa.
	home := t.TempDir()
	repo := t.TempDir()
	require.NoError(t, os.Mkdir(filepath.Join(repo, ".git"), 0o755))

	t.Setenv("HOME", home)

	homeCfg := []byte(`pre:
  - name: home-pre
    run: echo home
post:
  - name: home-post
    run: echo done-home
`)
	require.NoError(t, os.WriteFile(filepath.Join(home, ".gavel.yaml"), homeCfg, 0o644))

	repoCfg := []byte(`pre:
  - name: repo-pre
    run: make deps
`)
	require.NoError(t, os.WriteFile(filepath.Join(repo, ".gavel.yaml"), repoCfg, 0o644))

	cfg, err := LoadGavelConfig(repo)
	require.NoError(t, err)

	require.Len(t, cfg.Pre, 2)
	assert.Equal(t, "home-pre", cfg.Pre[0].Name)
	assert.Equal(t, "repo-pre", cfg.Pre[1].Name)

	require.Len(t, cfg.Post, 1)
	assert.Equal(t, "home-post", cfg.Post[0].Name)
}

func TestMergeSecretsConfig(t *testing.T) {
	t.Run("zero + zero", func(t *testing.T) {
		out := MergeSecretsConfig(SecretsConfig{}, SecretsConfig{})
		assert.False(t, out.Disabled)
		assert.Empty(t, out.Configs)
	})

	t.Run("disabled propagates", func(t *testing.T) {
		out := MergeSecretsConfig(SecretsConfig{}, SecretsConfig{Disabled: true})
		assert.True(t, out.Disabled)
	})

	t.Run("configs append and dedupe", func(t *testing.T) {
		base := SecretsConfig{Configs: []string{"/a.toml", "/b.toml"}}
		override := SecretsConfig{Configs: []string{"/b.toml", "/c.toml"}}
		out := MergeSecretsConfig(base, override)
		assert.Equal(t, []string{"/a.toml", "/b.toml", "/c.toml"}, out.Configs)
	})
}
