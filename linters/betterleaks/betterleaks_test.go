package betterleaks

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	commonsContext "github.com/flanksource/commons/context"
	"github.com/flanksource/gavel/models"
	"github.com/hairyhenderson/toml"
	"github.com/stretchr/testify/require"
)

func TestParseFindings(t *testing.T) {
	data, err := os.ReadFile("testdata/betterleaks-output.json")
	require.NoError(t, err)

	b := &Betterleaks{}
	b.WorkDir = "/workspace"

	violations, err := b.parseFindings(data)
	require.NoError(t, err)
	require.Len(t, violations, 2)

	first := violations[0]
	require.Equal(t, filepath.Join("/workspace", "config/prod.env"), first.File)
	require.Equal(t, 12, first.Line)
	require.Equal(t, 20, first.Column)
	require.Equal(t, models.SeverityError, first.Severity)
	require.Equal(t, "betterleaks", first.Source)
	require.NotNil(t, first.Rule)
	require.Equal(t, "aws-access-token", first.Rule.Method)

	require.NotNil(t, first.Message)
	msg := *first.Message
	require.Contains(t, msg, "AWS Access Key")
	require.NotContains(t, msg, "AKIAIOSFODNN7EXAMPLE", "secret must not appear in violation message")
	require.Contains(t, msg, "redacted")

	second := violations[1]
	require.NotNil(t, second.Message)
	msg2 := *second.Message
	require.NotContains(t, msg2, "abcd1234efgh5678ijkl9012mnop3456", "secret must not appear in violation message")
}

func TestParseFindingsEmpty(t *testing.T) {
	b := &Betterleaks{}
	b.WorkDir = "/workspace"

	for _, in := range []string{"", "   ", "null", "[]"} {
		violations, err := b.parseFindings([]byte(in))
		require.NoError(t, err)
		require.Empty(t, violations, "input %q", in)
	}
}

func TestBuildArgs(t *testing.T) {
	b := &Betterleaks{}
	b.WorkDir = "/workspace"

	t.Run("no files -> scan .", func(t *testing.T) {
		args := b.buildArgs("", "/workspace/.tmp/report.json")
		require.Equal(t, "dir", args[0])
		require.Equal(t, ".", args[1])
		require.Contains(t, args, "--report-format")
		require.Contains(t, args, "json")
		require.Contains(t, args, "--report-path")
		require.Contains(t, args, "/workspace/.tmp/report.json")
		require.Contains(t, args, "--exit-code")
		require.NotContains(t, args, "-c", "no -c flag when tomlPath empty")
	})

	t.Run("with files and config", func(t *testing.T) {
		b2 := &Betterleaks{}
		b2.WorkDir = "/workspace"
		b2.Files = []string{"a.go", "b.go"}
		args := b2.buildArgs("/workspace/.betterleaks.toml", "/workspace/.tmp/report.json")
		require.Equal(t, "dir", args[0])
		require.Equal(t, "a.go", args[1])
		require.Equal(t, "b.go", args[2])
		require.Contains(t, args, "-c")
		require.Contains(t, args, "/workspace/.betterleaks.toml")
	})
}

func TestResolveConfigEmpty(t *testing.T) {
	path, err := ResolveConfig(t.TempDir(), nil)
	require.NoError(t, err)
	require.Empty(t, path, "empty input means skip the linter")
}

func TestResolveConfigSinglePassthrough(t *testing.T) {
	tmp := t.TempDir()
	cfg := filepath.Join(tmp, "only.toml")
	require.NoError(t, os.WriteFile(cfg, []byte(`title = "solo"`+"\n"), 0o644))

	path, err := ResolveConfig(tmp, []string{cfg})
	require.NoError(t, err)
	require.Equal(t, cfg, path, "single config passed through unchanged")
	_, statErr := os.Stat(filepath.Join(tmp, ".tmp", "betterleaks.toml"))
	require.True(t, os.IsNotExist(statErr), ".tmp/betterleaks.toml must not be created for single config")
}

func TestRunCreatesReportDirForSingleConfig(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	workDir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(workDir, ".betterleaks.toml"), []byte(`title = "solo"`+"\n"), 0o644))

	binDir := t.TempDir()
	fakeBetterleaks := filepath.Join(binDir, "betterleaks")
	script := `#!/bin/sh
report=""
while [ "$#" -gt 0 ]; do
  if [ "$1" = "--report-path" ]; then
    shift
    report="$1"
    break
  fi
  shift
done

if [ -z "$report" ]; then
  echo "missing --report-path" >&2
  exit 1
fi

report_dir=${report%/*}
if [ "$report_dir" = "$report" ]; then
  report_dir=.
fi
if [ ! -d "$report_dir" ]; then
  echo "report dir missing: $report_dir" >&2
  exit 1
fi

printf '[]' >"$report"
`
	require.NoError(t, os.WriteFile(fakeBetterleaks, []byte(script), 0o755))
	t.Setenv("PATH", binDir)

	b := NewBetterleaks(workDir)
	violations, err := b.Run(commonsContext.NewContext(context.Background()), nil)
	require.NoError(t, err)
	require.Empty(t, violations)

	reportPath := filepath.Join(workDir, ".tmp", "betterleaks-report.json")
	data, err := os.ReadFile(reportPath)
	require.NoError(t, err)
	require.JSONEq(t, "[]", string(data))
}

func TestResolveConfigMergesMultiple(t *testing.T) {
	tmp := t.TempDir()

	homeCfg := filepath.Join(tmp, "home.toml")
	repoCfg := filepath.Join(tmp, "repo.toml")

	homeTOML := `
title = "home"
[extend]
useDefault = true
disabledRules = ["aws-access-token", "generic-api-key"]

[[rules]]
id = "home-rule"
description = "home"
regex = "HOME_.*"
keywords = ["home"]

[[rules]]
id = "shared"
description = "from home"
regex = "OLD_.*"
`
	repoTOML := `
title = "repo"
[extend]
disabledRules = ["generic-api-key", "slack-token"]

[[rules]]
id = "shared"
description = "from repo"
regex = "NEW_.*"

[[rules]]
id = "repo-rule"
description = "repo"
regex = "REPO_.*"

[[allowlists]]
description = "repo-wide skip"
paths = ["testdata/.*"]
`
	require.NoError(t, os.WriteFile(homeCfg, []byte(homeTOML), 0o644))
	require.NoError(t, os.WriteFile(repoCfg, []byte(repoTOML), 0o644))

	out, err := ResolveConfig(tmp, []string{homeCfg, repoCfg})
	require.NoError(t, err)
	require.Equal(t, filepath.Join(tmp, ".tmp", "betterleaks.toml"), out)

	data, err := os.ReadFile(out)
	require.NoError(t, err)

	var merged tomlConfig
	require.NoError(t, toml.Unmarshal(data, &merged))

	require.Equal(t, "repo", merged.Title, "override wins for scalars")
	require.True(t, merged.Extend.UseDefault, "useDefault carries from home")
	require.ElementsMatch(t,
		[]string{"aws-access-token", "generic-api-key", "slack-token"},
		merged.Extend.DisabledRules,
		"disabled rules union + dedupe")

	rulesByID := make(map[string]tomlRule, len(merged.Rules))
	for _, r := range merged.Rules {
		rulesByID[r.ID] = r
	}
	require.Len(t, merged.Rules, 3, "home-rule + shared + repo-rule")
	require.Equal(t, "from repo", rulesByID["shared"].Description, "override replaces same id")
	require.Equal(t, "NEW_.*", rulesByID["shared"].Regex)
	require.Contains(t, rulesByID, "home-rule")
	require.Contains(t, rulesByID, "repo-rule")

	require.Len(t, merged.Allowlists, 1, "allowlists append additively")
}

func TestMergeTOMLConfigsEmpty(t *testing.T) {
	out := mergeTOMLConfigs(tomlConfig{}, tomlConfig{})
	require.Empty(t, out.Rules)
	require.Empty(t, out.Extend.DisabledRules)
}

func TestDiscoverConfigsFindsRepoFile(t *testing.T) {
	// Isolate HOME so it can't pollute discovery with the real user's files.
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	// Synthesize a fake git repo with a betterleaks config.
	repo := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(repo, ".git"), 0o755))
	cfg := filepath.Join(repo, ".betterleaks.toml")
	require.NoError(t, os.WriteFile(cfg, []byte(`title = "repo"`+"\n"), 0o644))

	found := DiscoverConfigs(repo)
	require.NotEmpty(t, found)
	// Repo config must be present; absolute paths are returned so compare abs.
	absCfg, _ := filepath.Abs(cfg)
	require.Contains(t, found, absCfg)
}

func TestDiscoverConfigsEmptyWhenAbsent(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	repo := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(repo, ".git"), 0o755))

	found := DiscoverConfigs(repo)
	require.Empty(t, found, "no config files anywhere should yield empty slice")
}

func TestDiscoverConfigsReadsExtraPathsFromGavelYaml(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	repo := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(repo, ".git"), 0o755))

	// Extra config file referenced via .gavel.yaml, relative to repo root.
	extraPath := filepath.Join(repo, "config", "custom.toml")
	require.NoError(t, os.MkdirAll(filepath.Dir(extraPath), 0o755))
	require.NoError(t, os.WriteFile(extraPath, []byte(`title = "extra"`+"\n"), 0o644))

	gavelYaml := `
secrets:
  configs:
    - config/custom.toml
`
	require.NoError(t, os.WriteFile(filepath.Join(repo, ".gavel.yaml"), []byte(gavelYaml), 0o644))

	found := DiscoverConfigs(repo)
	absExtra, _ := filepath.Abs(extraPath)
	require.Contains(t, found, absExtra)
}

func TestDiscoverConfigsResolvesLayerRelativeExtrasInOrder(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	repo := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(repo, ".git"), 0o755))

	workDir := filepath.Join(repo, "service")
	require.NoError(t, os.MkdirAll(workDir, 0o755))

	homeNative := filepath.Join(home, ".betterleaks.toml")
	homeExtra := filepath.Join(home, "home-extra.toml")
	repoNative := filepath.Join(repo, ".gitleaks.toml")
	repoExtra := filepath.Join(repo, "config", "repo-extra.toml")
	cwdNative := filepath.Join(workDir, ".betterleaks.toml")
	cwdExtra := filepath.Join(workDir, "nested", "cwd-extra.toml")

	for _, path := range []string{
		homeNative,
		homeExtra,
		repoNative,
		repoExtra,
		cwdNative,
		cwdExtra,
	} {
		require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
		require.NoError(t, os.WriteFile(path, []byte(`title = "cfg"`+"\n"), 0o644))
	}

	require.NoError(t, os.WriteFile(filepath.Join(home, ".gavel.yaml"), []byte(`
secrets:
  configs:
    - home-extra.toml
`), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(repo, ".gavel.yaml"), []byte(`
secrets:
  configs:
    - config/repo-extra.toml
`), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(workDir, ".gavel.yaml"), []byte(`
secrets:
  configs:
    - nested/cwd-extra.toml
`), 0o644))

	found := DiscoverConfigs(workDir)
	require.Equal(t, []string{
		homeNative,
		homeExtra,
		repoNative,
		repoExtra,
		cwdNative,
		cwdExtra,
	}, found)
}
