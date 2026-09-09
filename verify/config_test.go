package verify

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/flanksource/captain/pkg/api"
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
		{
			name:  "rule glob matches prefix",
			rule:  LintIgnoreRule{Rule: "acme-*", Source: "betterleaks"},
			v:     models.Violation{Source: "betterleaks", Rule: &models.Rule{Method: "acme-brand-mention"}},
			match: true,
		},
		{
			name:  "rule glob does not match other prefix",
			rule:  LintIgnoreRule{Rule: "acme-*"},
			v:     models.Violation{Rule: &models.Rule{Method: "generic-api-key"}},
			match: false,
		},
		{
			name:  "rule glob with nil violation rule",
			rule:  LintIgnoreRule{Rule: "acme-*"},
			v:     models.Violation{Source: "betterleaks"},
			match: false,
		},
		{
			name:  "source glob matches",
			rule:  LintIgnoreRule{Source: "go*"},
			v:     models.Violation{Source: "golangci-lint", Rule: &models.Rule{Method: "errcheck"}},
			match: true,
		},
		{
			name:  "rule wildcard matches anything",
			rule:  LintIgnoreRule{Rule: "*", Source: "betterleaks"},
			v:     models.Violation{Source: "betterleaks", Rule: &models.Rule{Method: "anything"}},
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
	// LoadGavelConfig layers ~/.gavel.yaml under the repo's, so a test that does
	// not redirect HOME asserts against whatever the developer happens to have.
	t.Setenv("HOME", t.TempDir())
	dir := t.TempDir()
	require.NoError(t, os.Mkdir(filepath.Join(dir, ".git"), 0o755))

	cfgData := []byte(`ai:
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

	assert.Equal(t, "gemini", cfg.AI.Model.Name)
	assert.Len(t, cfg.Lint.Ignore, 2)
	assert.Equal(t, "errcheck", cfg.Lint.Ignore[0].Rule)
	assert.Equal(t, "golangci-lint", cfg.Lint.Ignore[0].Source)
	assert.Equal(t, "unused-import", cfg.Lint.Ignore[1].Rule)
	assert.Equal(t, "pkg/foo.go", cfg.Lint.Ignore[1].File)
}

// TestLoadGavelConfig_SurfacesUnreadableLayer pins the load seam as fail-loud. A
// broken .gavel.yaml used to be discarded layer-by-layer, so a single typo ran
// the whole project on built-in defaults — the settings were declared, ignored,
// and nothing said so. A missing file stays the ordinary case.
func TestLoadGavelConfig_SurfacesUnreadableLayer(t *testing.T) {
	for _, tc := range []struct {
		name string
		yaml string
		want string
	}{
		{"malformed yaml", "ai:\n model: [unclosed\n", "parse"},
		{"invalid tool policy", "todos:\n  run:\n    permissions:\n      tools:\n        Bash: sometimes\n", `invalid tool policy "sometimes"`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("HOME", t.TempDir())
			dir := t.TempDir()
			require.NoError(t, os.WriteFile(filepath.Join(dir, ".gavel.yaml"), []byte(tc.yaml), 0o644))

			_, err := LoadGavelConfig(dir)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tc.want)
		})
	}

	t.Setenv("HOME", t.TempDir())
	_, err := LoadGavelConfig(t.TempDir())
	require.NoError(t, err, "a directory with no .gavel.yaml is not an error")
}

func TestSaveGavelConfig_RoundTrip(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	dir := t.TempDir()

	cfg := GavelConfig{
		AI: api.Spec{Model: api.Model{Name: "claude"}},
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

	assert.Equal(t, cfg.AI.Model.Name, loaded.AI.Model.Name)
	assert.Equal(t, cfg.Lint.Ignore, loaded.Lint.Ignore)
}

func TestLoadGavelConfig_WithFixtures(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
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

func TestLoadGavelConfig_WithChecks(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	dir := t.TempDir()
	require.NoError(t, os.Mkdir(filepath.Join(dir, ".git"), 0o755))

	cfgData := []byte(`checks:
  enabled: true
  test:
    changed: true
    timeout: 3m
  lint:
    changed: true
`)
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".gavel.yaml"), cfgData, 0o644))

	cfg, err := LoadGavelConfig(dir)
	require.NoError(t, err)
	require.NotNil(t, cfg.Checks.Enabled)
	assert.True(t, *cfg.Checks.Enabled)
	require.NotNil(t, cfg.Checks.Test)
	assert.True(t, cfg.Checks.Test.Changed)
	assert.Equal(t, "3m", cfg.Checks.Test.Timeout)
	require.NotNil(t, cfg.Checks.Lint)
	assert.True(t, cfg.Checks.Lint.Changed)
}

func TestLoadGavelConfig_RepoChecksOverrideHome(t *testing.T) {
	home := t.TempDir()
	repo := t.TempDir()
	require.NoError(t, os.Mkdir(filepath.Join(repo, ".git"), 0o755))

	t.Setenv("HOME", home)
	const homeRetry = "verify.summary.failed > 0"
	homeCfg := []byte("checks:\n  enabled: false\n  retry: " + homeRetry + "\n")
	require.NoError(t, os.WriteFile(filepath.Join(home, ".gavel.yaml"), homeCfg, 0o644))
	repoCfg := []byte("checks:\n  enabled: true\n")
	require.NoError(t, os.WriteFile(filepath.Join(repo, ".gavel.yaml"), repoCfg, 0o644))

	cfg, err := LoadGavelConfig(repo)
	require.NoError(t, err)
	// repo turns checks on; home's retry predicate survives where repo is silent.
	require.NotNil(t, cfg.Checks.Enabled)
	assert.True(t, *cfg.Checks.Enabled)
	assert.Equal(t, homeRetry, cfg.Checks.Retry)
}

func TestFixturesConfig_ResolvedFiles_Default(t *testing.T) {
	empty := FixturesConfig{}
	assert.Equal(t, []string{DefaultFixturesGlob}, empty.ResolvedFiles())

	custom := FixturesConfig{Files: []string{"a.md", "b.md"}}
	assert.Equal(t, []string{"a.md", "b.md"}, custom.ResolvedFiles())
}

func TestMerge_FixturesConfig(t *testing.T) {
	t.Run("override enables", func(t *testing.T) {
		merged := Merge(FixturesConfig{}, FixturesConfig{Enabled: true})
		assert.True(t, merged.Enabled)
	})
	t.Run("override files replace base", func(t *testing.T) {
		base := FixturesConfig{Files: []string{"old.md"}}
		override := FixturesConfig{Files: []string{"new.md"}}
		merged := Merge(base, override)
		assert.Equal(t, []string{"new.md"}, merged.Files)
	})
	t.Run("override empty keeps base files", func(t *testing.T) {
		base := FixturesConfig{Enabled: true, Files: []string{"base.md"}}
		merged := Merge(base, FixturesConfig{})
		assert.Equal(t, []string{"base.md"}, merged.Files)
		assert.True(t, merged.Enabled)
	})
}

func TestMerge_LintConfig(t *testing.T) {
	enabledFalse := false
	enabledTrue := true
	base := LintConfig{
		Fix:    PromptSpec{Spec: api.Spec{Model: api.Model{Name: "agent:sonnet"}}},
		Ignore: []LintIgnoreRule{{Rule: "errcheck"}},
		Linters: map[string]LintLinterConfig{
			"jscpd": {Enabled: &enabledFalse},
		},
	}
	override := LintConfig{
		Fix:    PromptSpec{Spec: api.Spec{Model: api.Model{Name: "agent:opus"}}},
		Ignore: []LintIgnoreRule{{Rule: "unused", Source: "ruff"}},
		Linters: map[string]LintLinterConfig{
			"jscpd": {Enabled: &enabledTrue},
		},
	}
	merged := Merge(base, override)
	assert.Equal(t, "agent:opus", merged.Fix.Spec.Model.Name)
	assert.Len(t, merged.Ignore, 2)
	assert.Equal(t, "errcheck", merged.Ignore[0].Rule)
	assert.Equal(t, "unused", merged.Ignore[1].Rule)
	if assert.Contains(t, merged.Linters, "jscpd") {
		assert.NotNil(t, merged.Linters["jscpd"].Enabled)
		assert.True(t, *merged.Linters["jscpd"].Enabled)
	}
}

// A prompt spec is layered configuration like everything else in .gavel.yaml: a
// repo layer that speaks about one field must leave the home layer's siblings
// alone. Whole-struct replacement made `todos: {run: {model: opus}}` silently
// discard the home budget.
func TestMergeGavelConfig_PromptSpecKeepsUnmentionedFields(t *testing.T) {
	home := GavelConfig{Todos: TodosConfig{Run: PromptSpec{Spec: api.Spec{
		Model:  api.Model{Name: "claude-sonnet-4-5"},
		Budget: api.Budget{MaxTokens: 200, Timeout: "30m"},
	}}}}
	repo := GavelConfig{Todos: TodosConfig{Run: PromptSpec{Spec: api.Spec{
		Model: api.Model{Name: "claude-opus-4-1"},
	}}}}

	merged := MergeGavelConfig(home, repo)

	assert.Equal(t, "claude-opus-4-1", merged.Todos.Run.Spec.Model.Name)
	assert.Equal(t, 200, merged.Todos.Run.Spec.Budget.MaxTokens)
	assert.Equal(t, "30m", merged.Todos.Run.Spec.Budget.Timeout)
}

// Whether an override is "configured" is a question about every field, not the
// six a hand-written IsZero happened to list. permissions.mode was invisible to
// it, so a repo layer setting only that was dropped outright rather than merged.
func TestMergeGavelConfig_PromptSpecAppliesEveryField(t *testing.T) {
	home := GavelConfig{Todos: TodosConfig{Run: PromptSpec{Spec: api.Spec{
		Model: api.Model{Name: "claude-sonnet-4-5"},
	}}}}
	repo := GavelConfig{Todos: TodosConfig{Run: PromptSpec{Spec: api.Spec{
		Permissions: api.Permissions{Mode: api.PermissionPlan},
	}}}}

	merged := MergeGavelConfig(home, repo)

	assert.Equal(t, api.PermissionPlan, merged.Todos.Run.Spec.Permissions.Mode)
	assert.Equal(t, "claude-sonnet-4-5", merged.Todos.Run.Spec.Model.Name)
}

// The grader that marks a definition of done resolves through
// request > todos.verify > ai: > captain. Its floor is a config value in the
// layer the chain names — not an `if model == ""` at the point of use — so a repo
// can override it and the settings trace can say where it came from. It must stay
// an agentic backend: the grader is told to inspect the repository with its own
// tools, and the ai: floor is an API model with none.
func TestDefaultGavelConfig_SeedsTheVerifyGrader(t *testing.T) {
	defaults := DefaultGavelConfig()
	// The grader pins a MECHANISM, not a model: it must run agentically or it
	// grades without reading the diff. The model itself is configuration, so the
	// built-in layer names none.
	assert.Equal(t, DefaultVerifyMode, defaults.Todos.Verify.Spec.Model.Mode)
	assert.Empty(t, defaults.Todos.Verify.Spec.Model.Name)
	assert.Empty(t, defaults.AI.Model.Name, "the ai: base must not pick a model either")

	repo := GavelConfig{Todos: TodosConfig{Verify: PromptSpec{Spec: api.Spec{Model: api.Model{Name: "claude-code-opus"}}}}}
	merged := MergeGavelConfig(defaults, repo)
	assert.Equal(t, "claude-code-opus", merged.Todos.Verify.Spec.Model.Name)
	assert.Empty(t, merged.AI.Model.Name, "the grader layer is not the ai: base")
}

// File and baseDir are one fact — which prompt, and the directory its relative
// path resolves against. A layer that names its own file brings its own
// directory; a layer silent about the file leaves both alone, because the config
// loader stamps baseDir on every spec it decodes, set or not.
func TestMergeGavelConfig_PromptSpecFileTravelsWithItsDirectory(t *testing.T) {
	home := GavelConfig{Todos: TodosConfig{Run: PromptSpec{File: "run.prompt", baseDir: "/home/user"}}}
	repo := GavelConfig{Todos: TodosConfig{Run: PromptSpec{File: "prompts/run.prompt", baseDir: "/repo"}}}

	overridden := MergeGavelConfig(home, repo)
	assert.Equal(t, "/repo/prompts/run.prompt", overridden.Todos.Run.ResolvedFilePath(""))

	silent := MergeGavelConfig(home, GavelConfig{Todos: TodosConfig{Run: PromptSpec{
		Spec:    api.Spec{Model: api.Model{Name: "claude-opus-4-1"}},
		baseDir: "/repo",
	}}})
	assert.Equal(t, "/home/user/run.prompt", silent.Todos.Run.ResolvedFilePath(""))
}

// Accumulating lists are the exception structural merging cannot infer, so each
// says so on its field with a merge:"append" tag: hooks, ignore rules and path
// allowlists compose across layers instead of the repo erasing the home's. Paths
// dedupe; ordered command lists do not.
func TestMergeGavelConfig_ListsAccumulateAcrossLayers(t *testing.T) {
	home := GavelConfig{
		Pre:     []HookStep{{Run: "make deps"}},
		Commit:  CommitConfig{Hooks: []CommitHook{{Name: "fmt", Run: "gofmt -l ."}}, GitIgnore: []string{"*.env"}},
		Lint:    LintConfig{Ignore: []LintIgnoreRule{{Rule: "errcheck"}}},
		Secrets: SecretsConfig{Configs: []string{"~/.betterleaks.toml"}},
	}
	repo := GavelConfig{
		Pre:     []HookStep{{Run: "make generate"}},
		Commit:  CommitConfig{Hooks: []CommitHook{{Name: "vet", Run: "go vet ./..."}}, GitIgnore: []string{"*.env", "dist/"}},
		Lint:    LintConfig{Ignore: []LintIgnoreRule{{Rule: "unused"}}},
		Secrets: SecretsConfig{Configs: []string{".betterleaks.toml"}},
	}

	merged := MergeGavelConfig(home, repo)

	assert.Equal(t, []HookStep{{Run: "make deps"}, {Run: "make generate"}}, merged.Pre)
	assert.Equal(t, []string{"fmt", "vet"}, []string{merged.Commit.Hooks[0].Name, merged.Commit.Hooks[1].Name})
	assert.Equal(t, []string{"*.env", "dist/"}, merged.Commit.GitIgnore)
	assert.Equal(t, []string{"errcheck", "unused"}, []string{merged.Lint.Ignore[0].Rule, merged.Lint.Ignore[1].Rule})
	assert.Equal(t, []string{"~/.betterleaks.toml", ".betterleaks.toml"}, merged.Secrets.Configs)
}

func TestLoadGavelConfig_WithLintLinterEnablement(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
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
