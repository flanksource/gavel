package verify

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/flanksource/captain/pkg/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadGavelConfig_WithPushHooksAndSSH(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
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
	t.Setenv("HOME", t.TempDir())
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

	// Unlike gitignore/allow, types must REPLACE rather than append: a repo that
	// narrows generation to feat/fix cannot narrow anything if a broader layer's
	// entries survive the merge.
	t.Run("types override replaces the base list", func(t *testing.T) {
		base := CommitConfig{Types: []string{"feat", "fix", "chore", "docs"}}
		out := Merge(base, CommitConfig{Types: []string{"feat", "fix"}})
		assert.Equal(t, []string{"feat", "fix"}, out.Types)
	})

	t.Run("types empty override preserves base", func(t *testing.T) {
		base := CommitConfig{Types: []string{"feat", "fix"}}
		out := Merge(base, CommitConfig{})
		assert.Equal(t, []string{"feat", "fix"}, out.Types)
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
