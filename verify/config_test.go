package verify

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/gavel/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	yamlv3 "gopkg.in/yaml.v3"
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
	dir := t.TempDir()
	require.NoError(t, os.Mkdir(filepath.Join(dir, ".git"), 0o755))

	cfgData := []byte(`checks:
  enabled: true
  maxIterations: 5
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
	assert.Equal(t, 5, cfg.Checks.MaxIterations)
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
	homeCfg := []byte("checks:\n  enabled: false\n  maxIterations: 2\n")
	require.NoError(t, os.WriteFile(filepath.Join(home, ".gavel.yaml"), homeCfg, 0o644))
	repoCfg := []byte("checks:\n  enabled: true\n")
	require.NoError(t, os.WriteFile(filepath.Join(repo, ".gavel.yaml"), repoCfg, 0o644))

	cfg, err := LoadGavelConfig(repo)
	require.NoError(t, err)
	// repo turns checks on; home's maxIterations survives where repo is silent.
	require.NotNil(t, cfg.Checks.Enabled)
	assert.True(t, *cfg.Checks.Enabled)
	assert.Equal(t, 2, cfg.Checks.MaxIterations)
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
	assert.Equal(t, DefaultVerifyModel, defaults.Todos.Verify.Model.Name)
	assert.NotEqual(t, DefaultAIModel, defaults.Todos.Verify.Model.Name)

	repo := GavelConfig{Todos: TodosConfig{Verify: api.Spec{Model: api.Model{Name: "claude-code-opus"}}}}
	merged := MergeGavelConfig(defaults, repo)
	assert.Equal(t, "claude-code-opus", merged.Todos.Verify.Model.Name)
	assert.Equal(t, DefaultAIModel, merged.AI.Model.Name, "the grader layer is not the ai: base")
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

	cfgData := []byte(`pre:
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

func TestMerge_SSHConfig(t *testing.T) {
	t.Run("override replaces cmd", func(t *testing.T) {
		merged := Merge(SSHConfig{Cmd: "make old"}, SSHConfig{Cmd: "make new"})
		assert.Equal(t, "make new", merged.Cmd)
	})
	t.Run("empty override keeps base", func(t *testing.T) {
		merged := Merge(SSHConfig{Cmd: "make old"}, SSHConfig{})
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
commit:
  maxCommits: 7
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
	assert.Contains(t, body, "maxCommits: 7")
}

func TestMerge_SecretsConfig(t *testing.T) {
	t.Run("zero + zero", func(t *testing.T) {
		out := Merge(SecretsConfig{}, SecretsConfig{})
		assert.False(t, out.Disabled)
		assert.Empty(t, out.Configs)
	})

	t.Run("disabled propagates", func(t *testing.T) {
		out := Merge(SecretsConfig{}, SecretsConfig{Disabled: true})
		assert.True(t, out.Disabled)
	})

	t.Run("configs append and dedupe", func(t *testing.T) {
		base := SecretsConfig{Configs: []string{"/a.toml", "/b.toml"}}
		override := SecretsConfig{Configs: []string{"/b.toml", "/c.toml"}}
		out := Merge(base, override)
		assert.Equal(t, []string{"/a.toml", "/b.toml", "/c.toml"}, out.Configs)
	})
}

func TestMerge_CommitGitIgnoreAndAllow(t *testing.T) {
	t.Run("gitignore concatenates across layers with dedup", func(t *testing.T) {
		base := CommitConfig{GitIgnore: []string{"*.log", ".env"}}
		override := CommitConfig{GitIgnore: []string{".env", "**/secrets/**"}}
		out := Merge(base, override)
		assert.Equal(t, []string{"*.log", ".env", "**/secrets/**"}, out.GitIgnore)
	})

	t.Run("allow concatenates with dedup", func(t *testing.T) {
		base := CommitConfig{Allow: []string{"a.log"}}
		override := CommitConfig{Allow: []string{"b.log", "a.log"}}
		out := Merge(base, override)
		assert.Equal(t, []string{"a.log", "b.log"}, out.Allow)
	})

	t.Run("empty override leaves base untouched", func(t *testing.T) {
		base := CommitConfig{GitIgnore: []string{"*.log"}, Allow: []string{"ok.log"}}
		out := Merge(base, CommitConfig{})
		assert.Equal(t, []string{"*.log"}, out.GitIgnore)
		assert.Equal(t, []string{"ok.log"}, out.Allow)
	})

	t.Run("precommit mode override wins when non-empty", func(t *testing.T) {
		base := CommitConfig{Precommit: PrecommitConfig{Mode: "prompt"}}
		out := Merge(base, CommitConfig{Precommit: PrecommitConfig{Mode: "fail"}})
		assert.Equal(t, CheckMode("fail"), out.Precommit.Mode)
	})

	t.Run("precommit empty override preserves base mode", func(t *testing.T) {
		base := CommitConfig{Precommit: PrecommitConfig{Mode: "skip"}}
		out := Merge(base, CommitConfig{})
		assert.Equal(t, CheckMode("skip"), out.Precommit.Mode)
	})

	t.Run("maxCommits override wins when non-zero", func(t *testing.T) {
		base := CommitConfig{MaxCommits: 3}
		out := Merge(base, CommitConfig{MaxCommits: 9})
		assert.Equal(t, 9, out.MaxCommits)
	})

	t.Run("maxCommits zero override preserves base", func(t *testing.T) {
		base := CommitConfig{MaxCommits: 5}
		out := Merge(base, CommitConfig{})
		assert.Equal(t, 5, out.MaxCommits)
	})

	t.Run("message spec override wins when set", func(t *testing.T) {
		base := CommitConfig{Message: PromptSpec{Spec: api.Spec{Model: api.Model{Name: "base-m"}}}}
		override := CommitConfig{Message: PromptSpec{Spec: api.Spec{Model: api.Model{Name: "over-m"}}}}
		out := Merge(base, override)
		assert.Equal(t, "over-m", out.Message.Spec.Model.Name)
	})

	t.Run("message empty override preserves base spec", func(t *testing.T) {
		base := CommitConfig{Message: PromptSpec{Spec: api.Spec{Model: api.Model{Name: "base-m"}}}}
		out := Merge(base, CommitConfig{})
		assert.Equal(t, "base-m", out.Message.Spec.Model.Name)
	})
}

func TestRestartPolicyUnmarshal(t *testing.T) {
	jsonCases := map[string]RestartPolicy{
		`"on-failure"`: RestartOnFailure,
		`"always"`:     RestartAlways,
		`"no"`:         RestartNo,
		`true`:         RestartOnFailure,
		`false`:        RestartNo,
		`null`:         "",
	}
	for in, want := range jsonCases {
		var p RestartPolicy
		require.NoError(t, json.Unmarshal([]byte(in), &p), in)
		assert.Equal(t, want, p, in)
	}
	var bad RestartPolicy
	assert.Error(t, json.Unmarshal([]byte(`"maybe"`), &bad), "invalid enum is a loud error")

	yamlCases := map[string]RestartPolicy{
		"on-failure": RestartOnFailure,
		"always":     RestartAlways,
		"no":         RestartNo, // YAML 1.2 string "no"
		"true":       RestartOnFailure,
		"false":      RestartNo,
	}
	for in, want := range yamlCases {
		var p RestartPolicy
		require.NoError(t, yamlv3.Unmarshal([]byte(in), &p), in)
		assert.Equal(t, want, p, in)
	}
}

func TestMerge_ProcfileConfig(t *testing.T) {
	t.Run("scalar overrides win when set, omitted keep base", func(t *testing.T) {
		base := ProcfileConfig{AutoRestart: RestartNo, MaxRestarts: 3, Profile: "dev"}
		out := Merge(base, ProcfileConfig{AutoRestart: RestartAlways, MaxRestarts: 9})
		assert.Equal(t, RestartAlways, out.AutoRestart)
		assert.Equal(t, 9, out.MaxRestarts)
		assert.Equal(t, "dev", out.Profile, "omitted profile override keeps base")
	})

	t.Run("empty override preserves base scalars", func(t *testing.T) {
		base := ProcfileConfig{AutoRestart: RestartOnFailure, MaxRestarts: 2, Profile: "prod"}
		out := Merge(base, ProcfileConfig{})
		assert.Equal(t, RestartOnFailure, out.AutoRestart)
		assert.Equal(t, 2, out.MaxRestarts)
		assert.Equal(t, "prod", out.Profile)
	})

	t.Run("env merges key-by-key with override winning", func(t *testing.T) {
		base := ProcfileConfig{Env: map[string]string{"A": "1", "B": "2"}}
		override := ProcfileConfig{Env: map[string]string{"B": "20", "C": "3"}}
		out := Merge(base, override)
		assert.Equal(t, map[string]string{"A": "1", "B": "20", "C": "3"}, out.Env)
	})

	t.Run("resource limits + profile override when set and persist when omitted", func(t *testing.T) {
		base := ProcfileConfig{Mem: "256Mi", CPU: 50, Profile: "dev"}
		out := Merge(base, ProcfileConfig{Mem: "1Gi", Profile: "prod"})
		assert.Equal(t, "1Gi", out.Mem, "non-empty override wins")
		assert.Equal(t, 50.0, out.CPU, "omitted override keeps base")
		assert.Equal(t, "prod", out.Profile)

		kept := Merge(base, ProcfileConfig{})
		assert.Equal(t, "256Mi", kept.Mem)
		assert.Equal(t, 50.0, kept.CPU)
		assert.Equal(t, "dev", kept.Profile)
	})
}

func TestLoadGavelConfig_WithProcfile(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.Mkdir(filepath.Join(dir, ".git"), 0o755))

	cfgData := []byte(`procfile:
  profile: dev
  autoRestart: on-failure
  maxRestarts: 5
  mem: 512Mi
  env:
    RAILS_ENV: development
`)
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".gavel.yaml"), cfgData, 0o644))

	cfg, err := LoadGavelConfig(dir)
	require.NoError(t, err)
	assert.Equal(t, RestartOnFailure, cfg.Procfile.AutoRestart)
	assert.Equal(t, 5, cfg.Procfile.MaxRestarts)
	assert.Equal(t, "dev", cfg.Procfile.Profile)
	assert.Equal(t, "512Mi", cfg.Procfile.Mem)
	assert.Equal(t, "development", cfg.Procfile.Env["RAILS_ENV"])
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
  precommit:
    mode: false
`), 0o644))

		cfg, err := LoadSingleGavelConfig(path)
		require.NoError(t, err)
		assert.Equal(t, []string{"*.log"}, cfg.Commit.GitIgnore)
		assert.Equal(t, []string{"keep.log"}, cfg.Commit.Allow)
		assert.Equal(t, CheckMode("skip"), cfg.Commit.Precommit.Mode)
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

func TestLoadGavelConfigTrace_FilePath(t *testing.T) {
	home := t.TempDir()
	repo := t.TempDir()
	targetDir := filepath.Join(repo, "pkg", "api")
	targetFile := filepath.Join(targetDir, "handler.go")

	require.NoError(t, os.Mkdir(filepath.Join(repo, ".git"), 0o755))
	require.NoError(t, os.MkdirAll(targetDir, 0o755))
	require.NoError(t, os.WriteFile(targetFile, []byte("package api\n"), 0o644))
	t.Setenv("HOME", home)

	require.NoError(t, os.WriteFile(filepath.Join(home, ".gavel.yaml"), []byte(`ai:
  model: gemini
pre:
  - name: home
    run: echo home
`), 0o644))

	require.NoError(t, os.WriteFile(filepath.Join(repo, ".gavel.yaml"), []byte(`ai:
  model: claude
pre:
  - name: repo
    run: echo repo
ssh:
  cmd: make ci
`), 0o644))

	require.NoError(t, os.WriteFile(filepath.Join(targetDir, ".gavel.yaml"), []byte(`ai:
  model: codex
pre:
  - name: target
    run: echo target
lint:
  ignore:
    - file: pkg/api/**
`), 0o644))

	trace, err := LoadGavelConfigTrace(targetFile)
	require.NoError(t, err)

	assert.Equal(t, targetFile, trace.TargetPath)
	assert.Equal(t, targetDir, trace.TargetDir)
	assert.Equal(t, repo, trace.GitRoot)

	require.Len(t, trace.Sources, 3)
	assert.Equal(t, "user-home", trace.Sources[0].Origin)
	assert.Equal(t, filepath.Join(home, ".gavel.yaml"), trace.Sources[0].Path)
	assert.Equal(t, "git-root", trace.Sources[1].Origin)
	assert.Equal(t, filepath.Join(repo, ".gavel.yaml"), trace.Sources[1].Path)
	assert.Equal(t, "parent-directory", trace.Sources[2].Origin)
	assert.Equal(t, filepath.Join(targetDir, ".gavel.yaml"), trace.Sources[2].Path)

	assert.Equal(t, "codex", trace.Merged.AI.Model.Name)
	require.Len(t, trace.Merged.Pre, 3)
	assert.Equal(t, "home", trace.Merged.Pre[0].Name)
	assert.Equal(t, "repo", trace.Merged.Pre[1].Name)
	assert.Equal(t, "target", trace.Merged.Pre[2].Name)
	assert.Equal(t, "make ci", trace.Merged.SSH.Cmd)
	require.Len(t, trace.Merged.Lint.Ignore, 1)
	assert.Equal(t, "pkg/api/**", trace.Merged.Lint.Ignore[0].File)
}

func TestLoadGavelConfigTrace_DedupesGitRootTarget(t *testing.T) {
	home := t.TempDir()
	repo := t.TempDir()
	t.Setenv("HOME", home)

	require.NoError(t, os.Mkdir(filepath.Join(repo, ".git"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(home, ".gavel.yaml"), []byte(`ai:
  model: gemini
`), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(repo, ".gavel.yaml"), []byte(`ai:
  model: claude
`), 0o644))

	trace, err := LoadGavelConfigTrace(repo)
	require.NoError(t, err)

	require.Len(t, trace.Sources, 2)
	assert.Equal(t, "user-home", trace.Sources[0].Origin)
	assert.Equal(t, "git-root", trace.Sources[1].Origin)
	assert.Equal(t, "claude", trace.Merged.AI.Model.Name)
}
