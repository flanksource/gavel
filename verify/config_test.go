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
			name:  "file only matches",
			rule:  LintIgnoreRule{File: "pkg/foo.go"},
			v:     models.Violation{File: "pkg/foo.go"},
			match: true,
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
	enabledFalse := false
	enabledTrue := true
	base := LintConfig{
		Ignore: []LintIgnoreRule{{Rule: "errcheck"}},
		Linters: map[string]LintLinterConfig{
			"jscpd": {Enabled: &enabledFalse},
		},
	}
	override := LintConfig{
		Ignore: []LintIgnoreRule{{Rule: "unused", Source: "ruff"}},
		Linters: map[string]LintLinterConfig{
			"jscpd": {Enabled: &enabledTrue},
		},
	}
	merged := MergeLintConfig(base, override)
	assert.Len(t, merged.Ignore, 2)
	assert.Equal(t, "errcheck", merged.Ignore[0].Rule)
	assert.Equal(t, "unused", merged.Ignore[1].Rule)
	if assert.Contains(t, merged.Linters, "jscpd") {
		assert.NotNil(t, merged.Linters["jscpd"].Enabled)
		assert.True(t, *merged.Linters["jscpd"].Enabled)
	}
}

func TestLoadGavelConfig_WithLintLinterEnablement(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.Mkdir(filepath.Join(dir, ".git"), 0o755))

	cfgData := []byte(`lint:
  linters:
    jscpd:
      enabled: true
`)
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".gavel.yaml"), cfgData, 0o644))

	cfg, err := LoadGavelConfig(dir)
	require.NoError(t, err)
	assert.True(t, cfg.Lint.IsLinterEnabled("jscpd", false))
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
	info, err := os.Stat(path)
	if err != nil {
		t.Skipf("no .gavel.yaml at %s: %v", path, err)
	}
	if info.Size() == 0 {
		t.Skipf(".gavel.yaml at %s is empty", path)
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

// TestSaveAfterLayeredLoad_DoesNotLeakHomeIntoRepo is a regression test for
// a data-leak bug where callers loaded a merged GavelConfig (home+repo+cwd)
// and then wrote it back to the repo's .gavel.yaml via SaveGavelConfig —
// silently promoting every ~/.gavel.yaml field into the repo on the next
// `gavel lint --triage` or UI ignore click.
//
// The fix is to always load the single repo file for the read-modify-write
// cycle. This test guards callers by using the primitives directly and
// asserting the leak does not happen.
func TestSaveAfterLayeredLoad_DoesNotLeakHomeIntoRepo(t *testing.T) {
	home := t.TempDir()
	repo := t.TempDir()
	require.NoError(t, os.Mkdir(filepath.Join(repo, ".git"), 0o755))
	t.Setenv("HOME", home)

	// Home has a global commit.gitignore list the user never wants in any repo.
	require.NoError(t, os.WriteFile(filepath.Join(home, ".gavel.yaml"), []byte(`commit:
  gitignore:
    - .env
    - .claude
`), 0o644))

	// Repo starts with a narrow lint.ignore only.
	repoPath := filepath.Join(repo, ".gavel.yaml")
	require.NoError(t, os.WriteFile(repoPath, []byte(`lint:
  ignore:
    - file: existing.go
`), 0o644))

	// Simulate the lint --triage / UI ignore flow: read just the repo file,
	// append a rule, save back.
	repoCfg, err := LoadSingleGavelConfig(repoPath)
	require.NoError(t, err)
	repoCfg.Lint.Ignore = append(repoCfg.Lint.Ignore, LintIgnoreRule{File: "new.go"})
	require.NoError(t, SaveGavelConfig(repo, repoCfg))

	written, err := os.ReadFile(repoPath)
	require.NoError(t, err)
	body := string(written)

	// The repo file must carry the new rule.
	assert.Contains(t, body, "new.go")
	assert.Contains(t, body, "existing.go")

	// The repo file must NOT have absorbed anything from ~/.gavel.yaml.
	assert.NotContains(t, body, ".env",
		"home-level commit.gitignore must not leak into the repo file")
	assert.NotContains(t, body, ".claude",
		"home-level commit.gitignore must not leak into the repo file")
}

// TestSaveGavelConfig_RoundTripPreservesPreAndSSH guards the other half of
// the regression: once the repo file is loaded via the single-file loader,
// a save round-trip must preserve every top-level field (pre, ssh.cmd, post,
// verify.*). Without this, a future refactor that drops a YAML tag would
// silently eat fields on the next write.
func TestSaveGavelConfig_RoundTripPreservesPreAndSSH(t *testing.T) {
	dir := t.TempDir()
	original := []byte(`pre:
  - name: deps
    run: make tidy
ssh:
  cmd: make all
verify:
  model: claude
`)
	path := filepath.Join(dir, ".gavel.yaml")
	require.NoError(t, os.WriteFile(path, original, 0o644))

	cfg, err := LoadSingleGavelConfig(path)
	require.NoError(t, err)
	require.NoError(t, SaveGavelConfig(dir, cfg))

	written, err := os.ReadFile(path)
	require.NoError(t, err)
	body := string(written)

	assert.Contains(t, body, "name: deps")
	assert.Contains(t, body, "run: make tidy")
	assert.Contains(t, body, "cmd: make all")
	assert.Contains(t, body, "model: claude")
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

func TestMergeCommitConfig_GitIgnoreAndAllow(t *testing.T) {
	t.Run("gitignore concatenates across layers with dedup", func(t *testing.T) {
		base := CommitConfig{GitIgnore: []string{"*.log", ".env"}}
		override := CommitConfig{GitIgnore: []string{".env", "**/secrets/**"}}
		out := MergeCommitConfig(base, override)
		assert.Equal(t, []string{"*.log", ".env", "**/secrets/**"}, out.GitIgnore)
	})

	t.Run("allow concatenates with dedup", func(t *testing.T) {
		base := CommitConfig{Allow: []string{"a.log"}}
		override := CommitConfig{Allow: []string{"b.log", "a.log"}}
		out := MergeCommitConfig(base, override)
		assert.Equal(t, []string{"a.log", "b.log"}, out.Allow)
	})

	t.Run("empty override leaves base untouched", func(t *testing.T) {
		base := CommitConfig{GitIgnore: []string{"*.log"}, Allow: []string{"ok.log"}}
		out := MergeCommitConfig(base, CommitConfig{})
		assert.Equal(t, []string{"*.log"}, out.GitIgnore)
		assert.Equal(t, []string{"ok.log"}, out.Allow)
	})

	t.Run("linkedDeps mode override wins when non-empty", func(t *testing.T) {
		base := CommitConfig{LinkedDeps: LinkedDepsConfig{Mode: "prompt"}}
		out := MergeCommitConfig(base, CommitConfig{LinkedDeps: LinkedDepsConfig{Mode: "fail"}})
		assert.Equal(t, "fail", out.LinkedDeps.Mode)
	})

	t.Run("linkedDeps empty override preserves base mode", func(t *testing.T) {
		base := CommitConfig{LinkedDeps: LinkedDepsConfig{Mode: "skip"}}
		out := MergeCommitConfig(base, CommitConfig{})
		assert.Equal(t, "skip", out.LinkedDeps.Mode)
	})
}

func TestLoadSingleGavelConfig(t *testing.T) {
	t.Run("reads one file without layering", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, ".gavel.yaml")
		require.NoError(t, os.WriteFile(path, []byte(`
commit:
  gitignore:
    - "*.log"
  allow:
    - "keep.log"
`), 0o644))

		cfg, err := LoadSingleGavelConfig(path)
		require.NoError(t, err)
		assert.Equal(t, []string{"*.log"}, cfg.Commit.GitIgnore)
		assert.Equal(t, []string{"keep.log"}, cfg.Commit.Allow)
	})

	t.Run("missing file returns os.ErrNotExist", func(t *testing.T) {
		_, err := LoadSingleGavelConfig(filepath.Join(t.TempDir(), "missing.yaml"))
		require.Error(t, err)
		assert.True(t, os.IsNotExist(err))
	})

	t.Run("malformed yaml returns parse error", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, ".gavel.yaml")
		require.NoError(t, os.WriteFile(path, []byte("not: [valid yaml"), 0o644))
		_, err := LoadSingleGavelConfig(path)
		require.Error(t, err)
	})
}
