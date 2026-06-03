package commit

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClassifyPkgDep(t *testing.T) {
	cases := []struct {
		raw      string
		kind     LinkedDepKind
		path     string
		local    bool
		scenario string
	}{
		{"^1.2.3", "", "", false, "semver"},
		{"workspace:*", "", "", false, "pnpm workspace"},
		{"git+ssh://git@example.com/x.git", "", "", false, "git URL"},
		{"file:../foo", LinkedDepKindPkgJSONFile, "../foo", true, "file:"},
		{"link:./packages/x", LinkedDepKindPkgJSONLink, "./packages/x", true, "link:"},
		{"portal:../../pkg", LinkedDepKindPkgJSONPortal, "../../pkg", true, "portal:"},
		{"./local/pkg", LinkedDepKindPkgJSONRelPath, "./local/pkg", true, "relative ./"},
		{"../outside", LinkedDepKindPkgJSONRelPath, "../outside", true, "relative ../"},
		{"/abs/path", LinkedDepKindPkgJSONRelPath, "/abs/path", true, "absolute"},
	}
	for _, c := range cases {
		t.Run(c.scenario, func(t *testing.T) {
			kind, path, ok := classifyPkgDep(c.raw)
			assert.Equal(t, c.local, ok)
			if c.local {
				assert.Equal(t, c.kind, kind)
				assert.Equal(t, c.path, path)
			}
		})
	}
}

func TestIsLocalReplaceTarget(t *testing.T) {
	assert.True(t, isLocalReplaceTarget("./fork"))
	assert.True(t, isLocalReplaceTarget("../fork"))
	assert.True(t, isLocalReplaceTarget("/abs/path"))
	assert.False(t, isLocalReplaceTarget("example.com/fork"))
	assert.False(t, isLocalReplaceTarget(""))
	assert.False(t, isLocalReplaceTarget("example.com/fork v1.2.3"))
}

func TestEvaluateLinkedDeps_GoModReplace(t *testing.T) {
	t.Run("escaping relative replace is a violation", func(t *testing.T) {
		repo := initCommitRepo(t)
		writeFile(t, repo, "go.mod", `module example.com/app

go 1.22

require github.com/flanksource/commons v1.0.0

replace github.com/flanksource/commons => ../../commons
`)
		gitRun(t, repo, "add", "go.mod")

		vs, err := EvaluateLinkedDeps(repo, repo, []string{"go.mod"}, nil)
		require.NoError(t, err)
		require.Len(t, vs, 1)
		assert.Equal(t, "go.mod", vs[0].File)
		assert.Equal(t, LinkedDepKindGoModReplace, vs[0].Kind)
		assert.Equal(t, "github.com/flanksource/commons", vs[0].Name)
		assert.Equal(t, "../../commons", vs[0].Target)
	})

	t.Run("in-repo relative replace is clean", func(t *testing.T) {
		repo := initCommitRepo(t)
		writeFileInDir(t, repo, "internal/fork/go.mod", "module fork\n\ngo 1.22\n")
		writeFile(t, repo, "go.mod", `module example.com/app

go 1.22

replace fork => ./internal/fork
`)
		gitRun(t, repo, "add", "go.mod", "internal/fork/go.mod")

		vs, err := EvaluateLinkedDeps(repo, repo, []string{"go.mod", "internal/fork/go.mod"}, nil)
		require.NoError(t, err)
		assert.Empty(t, vs)
	})

	t.Run("module-version replace (not local) is clean", func(t *testing.T) {
		repo := initCommitRepo(t)
		writeFile(t, repo, "go.mod", `module example.com/app

go 1.22

replace github.com/foo/bar => github.com/foo/bar v1.2.3
`)
		gitRun(t, repo, "add", "go.mod")

		vs, err := EvaluateLinkedDeps(repo, repo, []string{"go.mod"}, nil)
		require.NoError(t, err)
		assert.Empty(t, vs)
	})

	t.Run("absolute path replace is a violation", func(t *testing.T) {
		repo := initCommitRepo(t)
		writeFile(t, repo, "go.mod", `module example.com/app

go 1.22

replace foo => /tmp/foo
`)
		gitRun(t, repo, "add", "go.mod")

		vs, err := EvaluateLinkedDeps(repo, repo, []string{"go.mod"}, nil)
		require.NoError(t, err)
		require.Len(t, vs, 1)
		assert.Equal(t, "/tmp/foo", vs[0].Target)
	})

	t.Run("nested manifest resolves relative to its dir", func(t *testing.T) {
		repo := initCommitRepo(t)
		writeFileInDir(t, repo, "services/api/go.mod", `module example.com/app/services/api

go 1.22

replace internal/shared => ../shared
`)
		writeFileInDir(t, repo, "services/shared/go.mod", "module example.com/app/services/shared\n\ngo 1.22\n")
		gitRun(t, repo, "add", "services/api/go.mod", "services/shared/go.mod")

		vs, err := EvaluateLinkedDeps(repo, repo,
			[]string{"services/api/go.mod", "services/shared/go.mod"}, nil)
		require.NoError(t, err)
		assert.Empty(t, vs, "../shared from services/api resolves to services/shared, inside repo")
	})

	t.Run("manifest staged for deletion is skipped", func(t *testing.T) {
		repo := initCommitRepo(t)
		writeFile(t, repo, "go.mod", "module example.com/app\n\ngo 1.22\n")
		gitRun(t, repo, "add", "go.mod")
		gitRun(t, repo, "commit", "-m", "seed go.mod")
		gitRun(t, repo, "rm", "go.mod")

		vs, err := EvaluateLinkedDeps(repo, repo, []string{"go.mod"}, []string{"go.mod"})
		require.NoError(t, err)
		assert.Empty(t, vs)
	})
}

func TestEvaluateLinkedDeps_GoWork(t *testing.T) {
	repo := initCommitRepo(t)
	writeFile(t, repo, "go.work", `go 1.22

use ./app
use ../outside
`)
	gitRun(t, repo, "add", "go.work")

	vs, err := EvaluateLinkedDeps(repo, repo, []string{"go.work"}, nil)
	require.NoError(t, err)
	require.Len(t, vs, 1)
	assert.Equal(t, LinkedDepKindGoWorkUse, vs[0].Kind)
	assert.Equal(t, "../outside", vs[0].Target)
}

func TestEvaluateLinkedDeps_PackageJSON(t *testing.T) {
	t.Run("file: escape is a violation", func(t *testing.T) {
		repo := initCommitRepo(t)
		writeFile(t, repo, "package.json", `{
  "name": "app",
  "dependencies": {
    "sibling": "file:../sibling"
  }
}
`)
		gitRun(t, repo, "add", "package.json")

		vs, err := EvaluateLinkedDeps(repo, repo, []string{"package.json"}, nil)
		require.NoError(t, err)
		require.Len(t, vs, 1)
		assert.Equal(t, LinkedDepKindPkgJSONFile, vs[0].Kind)
		assert.Equal(t, "sibling", vs[0].Name)
	})

	t.Run("workspace:* is clean", func(t *testing.T) {
		repo := initCommitRepo(t)
		writeFile(t, repo, "package.json", `{
  "name": "app",
  "dependencies": {
    "pkg-a": "workspace:*"
  }
}
`)
		gitRun(t, repo, "add", "package.json")

		vs, err := EvaluateLinkedDeps(repo, repo, []string{"package.json"}, nil)
		require.NoError(t, err)
		assert.Empty(t, vs)
	})

	t.Run("in-repo link: is clean", func(t *testing.T) {
		repo := initCommitRepo(t)
		writeFileInDir(t, repo, "packages/shared/package.json", `{"name":"shared"}`)
		writeFile(t, repo, "package.json", `{
  "name": "app",
  "devDependencies": {
    "shared": "link:./packages/shared"
  }
}
`)
		gitRun(t, repo, "add", "package.json", "packages/shared/package.json")

		vs, err := EvaluateLinkedDeps(repo, repo,
			[]string{"package.json", "packages/shared/package.json"}, nil)
		require.NoError(t, err)
		assert.Empty(t, vs)
	})

	t.Run("semver and git deps are clean", func(t *testing.T) {
		repo := initCommitRepo(t)
		writeFile(t, repo, "package.json", `{
  "name": "app",
  "dependencies": {
    "lodash": "^4.17.21",
    "custom": "git+ssh://git@example.com/x.git"
  }
}
`)
		gitRun(t, repo, "add", "package.json")

		vs, err := EvaluateLinkedDeps(repo, repo, []string{"package.json"}, nil)
		require.NoError(t, err)
		assert.Empty(t, vs)
	})
}

func staticLinkedDepDecider(d LinkedDepDecision) LinkedDepDecider {
	return func(context.Context, LinkedDepViolation) (LinkedDepChoice, error) {
		return LinkedDepChoice{Decision: d}, nil
	}
}

func TestRunLinkedDepsCheck_Unstage(t *testing.T) {
	repo := initCommitRepo(t)
	writeFile(t, repo, "go.mod", "module example.com/app\n\ngo 1.22\n\nreplace foo => ../../escaping\n")
	gitRun(t, repo, "add", "go.mod")

	outcome, err := RunLinkedDepsCheck(context.Background(), LinkedDepsParams{
		WorkDir:     repo,
		GitRoot:     repo,
		StagedFiles: []string{"go.mod"},
		Decider:     staticLinkedDepDecider(LinkedDepDecisionUnstage),
		Mode:        IgnoreCheckModePrompt,
	})
	require.NoError(t, err)
	assert.False(t, outcome.Cancelled)
	assert.Equal(t, []string{"go.mod"}, outcome.Unstaged)

	staged := gitOutput(t, repo, "diff", "--cached", "--name-only")
	assert.NotContains(t, staged, "go.mod")

	// Working-tree manifest is untouched.
	body := readFile(t, filepath.Join(repo, "go.mod"))
	assert.Contains(t, body, "replace foo => ../../escaping")
}

func TestAutoUnstageLinkedDepDecider(t *testing.T) {
	choice, err := autoUnstageLinkedDepDecider(context.Background(), LinkedDepViolation{})
	require.NoError(t, err)
	assert.Equal(t, LinkedDepDecisionUnstage, choice.Decision)
}

// applyLinkedDepsCheck with AssumeYes must unstage the offending manifest
// without ever consulting the interactive decider. We poison the decider so
// any prompt path fails the test loudly.
func TestApplyLinkedDepsCheck_AssumeYesUnstagesWithoutPrompt(t *testing.T) {
	prev := interactiveLinkedDepDecider
	t.Cleanup(func() { interactiveLinkedDepDecider = prev })
	interactiveLinkedDepDecider = func(_ context.Context, _ LinkedDepViolation) (LinkedDepChoice, error) {
		t.Fatal("interactive decider must not run when AssumeYes is set")
		return LinkedDepChoice{}, nil
	}

	repo := initCommitRepo(t)
	writeFile(t, repo, "go.mod", "module example.com/app\n\ngo 1.22\n\nreplace foo => ../../escaping\n")
	gitRun(t, repo, "add", "go.mod")

	source, err := readStagedSource(repo)
	require.NoError(t, err)
	require.Contains(t, source.Files, "go.mod")

	result, err := applyLinkedDepsCheck(context.Background(), Options{
		AssumeYes:     true,
		PrecommitMode: IgnoreCheckModePrompt,
		WorkDir:       repo,
	}, source)
	require.NoError(t, err)
	assert.NotContains(t, result.Files, "go.mod")

	staged := gitOutput(t, repo, "diff", "--cached", "--name-only")
	assert.NotContains(t, staged, "go.mod")
}

func TestRunLinkedDepsCheck_IgnoreKeepsFileStaged(t *testing.T) {
	repo := initCommitRepo(t)
	writeFile(t, repo, "go.mod", "module example.com/app\n\ngo 1.22\n\nreplace foo => ../../escaping\n")
	gitRun(t, repo, "add", "go.mod")

	outcome, err := RunLinkedDepsCheck(context.Background(), LinkedDepsParams{
		WorkDir:     repo,
		GitRoot:     repo,
		StagedFiles: []string{"go.mod"},
		Decider:     staticLinkedDepDecider(LinkedDepDecisionIgnore),
		Mode:        IgnoreCheckModePrompt,
	})
	require.NoError(t, err)
	assert.False(t, outcome.Cancelled)
	assert.Empty(t, outcome.Unstaged)

	staged := gitOutput(t, repo, "diff", "--cached", "--name-only")
	assert.Contains(t, staged, "go.mod")
}

func TestRunLinkedDepsCheck_Cancel(t *testing.T) {
	repo := initCommitRepo(t)
	writeFile(t, repo, "go.mod", "module example.com/app\n\ngo 1.22\n\nreplace foo => ../../escaping\n")
	gitRun(t, repo, "add", "go.mod")

	outcome, err := RunLinkedDepsCheck(context.Background(), LinkedDepsParams{
		WorkDir:     repo,
		GitRoot:     repo,
		StagedFiles: []string{"go.mod"},
		Decider:     staticLinkedDepDecider(LinkedDepDecisionCancel),
		Mode:        IgnoreCheckModePrompt,
	})
	require.NoError(t, err)
	assert.True(t, outcome.Cancelled)

	staged := gitOutput(t, repo, "diff", "--cached", "--name-only")
	assert.Contains(t, staged, "go.mod", "cancel must leave staging as-is")
}

func TestRunLinkedDepsCheck_UpgradeRewritesAndStagesGoModAndGoSum(t *testing.T) {
	withStubbedLookup(t, "v1.4.2", nil)
	tidyCalls := withStubbedTidy(t, func(modDir string) error {
		// Simulate `go mod tidy` writing a go.sum entry.
		return os.WriteFile(filepath.Join(modDir, "go.sum"),
			[]byte("github.com/flanksource/foo v1.4.2 h1:fake\n"), 0o644)
	})

	repo := initCommitRepo(t)
	writeFile(t, repo, "go.mod", `module example.com/app

go 1.22

require github.com/flanksource/foo v0.0.0

replace github.com/flanksource/foo => ../../foo
`)
	gitRun(t, repo, "add", "go.mod")

	outcome, err := RunLinkedDepsCheck(context.Background(), LinkedDepsParams{
		WorkDir:     repo,
		GitRoot:     repo,
		StagedFiles: []string{"go.mod"},
		Decider: func(_ context.Context, v LinkedDepViolation) (LinkedDepChoice, error) {
			return LinkedDepChoice{Decision: LinkedDepDecisionUpgrade, UpgradeVersion: "v1.4.2"}, nil
		},
		Mode: IgnoreCheckModePrompt,
	})
	require.NoError(t, err)
	assert.False(t, outcome.Cancelled)
	assert.Equal(t, []string{"go.mod"}, outcome.Upgraded)
	require.Len(t, outcome.PendingRestores, 1)
	assert.Equal(t, "go.mod", outcome.PendingRestores[0].GoModFile)
	require.Len(t, outcome.PendingRestores[0].Replaces, 1)
	assert.Equal(t, "../../foo", outcome.PendingRestores[0].Replaces[0].OldTarget)
	assert.Equal(t, "v1.4.2", outcome.PendingRestores[0].Replaces[0].NewVersion)
	assert.Equal(t, 1, *tidyCalls)

	body := readFile(t, filepath.Join(repo, "go.mod"))
	assert.NotContains(t, body, "replace github.com/flanksource/foo")
	assert.Contains(t, body, "github.com/flanksource/foo v1.4.2")

	staged := gitOutput(t, repo, "diff", "--cached", "--name-only")
	assert.Contains(t, staged, "go.mod")
	assert.Contains(t, staged, "go.sum")
}

func TestRunLinkedDepsCheck_UpgradeBatchesMultipleReplaces(t *testing.T) {
	withStubbedLookup(t, "v2.0.0", nil)
	tidyCalls := withStubbedTidy(t, func(modDir string) error {
		return os.WriteFile(filepath.Join(modDir, "go.sum"), []byte("sum\n"), 0o644)
	})

	repo := initCommitRepo(t)
	writeFile(t, repo, "go.mod", `module example.com/app

go 1.22

require (
	github.com/flanksource/foo v0.0.0
	github.com/flanksource/bar v0.0.0
)

replace github.com/flanksource/foo => ../../foo
replace github.com/flanksource/bar => ../../bar
`)
	gitRun(t, repo, "add", "go.mod")

	outcome, err := RunLinkedDepsCheck(context.Background(), LinkedDepsParams{
		WorkDir:     repo,
		GitRoot:     repo,
		StagedFiles: []string{"go.mod"},
		Decider: func(_ context.Context, v LinkedDepViolation) (LinkedDepChoice, error) {
			return LinkedDepChoice{Decision: LinkedDepDecisionUpgrade, UpgradeVersion: "v2.0.0"}, nil
		},
		Mode: IgnoreCheckModePrompt,
	})
	require.NoError(t, err)
	require.Len(t, outcome.PendingRestores, 1, "both replaces batched under one file")
	assert.Len(t, outcome.PendingRestores[0].Replaces, 2)
	assert.Equal(t, 1, *tidyCalls, "tidy invoked once per file, not per replace")

	body := readFile(t, filepath.Join(repo, "go.mod"))
	assert.NotContains(t, body, "replace github.com/flanksource/foo")
	assert.NotContains(t, body, "replace github.com/flanksource/bar")
	assert.Contains(t, body, "github.com/flanksource/foo v2.0.0")
	assert.Contains(t, body, "github.com/flanksource/bar v2.0.0")
}

func TestRunLinkedDepsCheck_UpgradeNoTaggedVersionsRePrompts(t *testing.T) {
	// First call to lookup fails (no tags). The prompt decider must re-show
	// the top-level menu so the user can pick a different action.
	withStubbedLookup(t, "", ErrNoTaggedVersions)

	repo := initCommitRepo(t)
	writeFile(t, repo, "go.mod", "module example.com/app\n\ngo 1.22\n\nreplace foo => ../../escaping\n")
	gitRun(t, repo, "add", "go.mod")

	// Drive the real prompt decider directly so the re-prompt loop is
	// exercised. The Decider field in LinkedDepsParams skips the interactive
	// path; we want to hit the interactive path, so we plug in
	// runPromptLinkedDepDecider explicitly and stub the underlying prompt.
	calls := 0
	prevPrompt := promptSelectFunc
	promptSelectFunc = func(options []promptSelectOption, title string) (promptSelectOption, bool) {
		calls++
		switch calls {
		case 1:
			// Top-level menu: pick Upgrade (index 1 — Unstage is 0).
			return options[1], true
		case 2:
			// Re-prompt after lookup fails: pick Ignore.
			for _, o := range options {
				if strings.Contains(o.Label, "Ignore") {
					return o, true
				}
			}
		}
		return promptSelectOption{}, false
	}
	t.Cleanup(func() { promptSelectFunc = prevPrompt })

	outcome, err := RunLinkedDepsCheck(context.Background(), LinkedDepsParams{
		WorkDir:     repo,
		GitRoot:     repo,
		StagedFiles: []string{"go.mod"},
		Decider:     runPromptLinkedDepDecider,
		Mode:        IgnoreCheckModePrompt,
	})
	require.NoError(t, err)
	assert.False(t, outcome.Cancelled)
	assert.Empty(t, outcome.Upgraded)
	assert.Empty(t, outcome.Unstaged)
	assert.GreaterOrEqual(t, calls, 2, "must have re-prompted after lookup failure")
}

func TestBuildLinkedDepMenu_HidesUpgradeForGoWork(t *testing.T) {
	v := LinkedDepViolation{
		File:     "go.work",
		Kind:     LinkedDepKindGoWorkUse,
		Name:     "../outside",
		Target:   "../outside",
		Resolved: "/somewhere/outside",
	}
	items, _ := buildLinkedDepMenu(v)
	for _, item := range items {
		assert.NotContains(t, item, "Upgrade", "go.work violations must not show Upgrade")
	}
}

func TestBuildLinkedDepMenu_HidesUpgradeForPackageJSON(t *testing.T) {
	v := LinkedDepViolation{
		File: "package.json",
		Kind: LinkedDepKindPkgJSONFile,
		Name: "sibling",
	}
	items, _ := buildLinkedDepMenu(v)
	for _, item := range items {
		assert.NotContains(t, item, "Upgrade", "package.json violations must not show Upgrade")
	}
}

func TestBuildLinkedDepMenu_ShowsUpgradeForGoModReplace(t *testing.T) {
	v := LinkedDepViolation{
		File: "go.mod",
		Kind: LinkedDepKindGoModReplace,
		Name: "github.com/flanksource/foo",
	}
	items, decisions := buildLinkedDepMenu(v)
	require.Equal(t, len(items), len(decisions))
	foundUpgrade := false
	for i, item := range items {
		if strings.Contains(item, "Upgrade") {
			foundUpgrade = true
			assert.Equal(t, LinkedDepDecisionUpgrade, decisions[i])
		}
	}
	assert.True(t, foundUpgrade)
}

func TestPickLatestVersion_PrereleaseFallback(t *testing.T) {
	assert.Equal(t, "v1.2.0", pickLatestVersion([]string{"v1.0.0", "v1.2.0", "v1.1.0-rc.1"}))
	assert.Equal(t, "v0.5.0-rc.2", pickLatestVersion([]string{"v0.5.0-rc.1", "v0.5.0-rc.2"}))
	assert.Equal(t, "v3.0.0", pickLatestVersion([]string{"v2.9.9", "v3.0.0"}))
}

func TestRealLatestGoVersion_ParsesOutput(t *testing.T) {
	prev := goVersionsCommand
	goVersionsCommand = func(ctx context.Context, module string) ([]byte, error) {
		return []byte("github.com/flanksource/foo v0.1.0 v0.2.0 v1.0.0-rc.1 v1.0.0\n"), nil
	}
	t.Cleanup(func() { goVersionsCommand = prev })

	got, err := realLatestGoVersion(context.Background(), "github.com/flanksource/foo")
	require.NoError(t, err)
	assert.Equal(t, "v1.0.0", got)
}

func TestRealLatestGoVersion_EmptyReturnsErrNoTaggedVersions(t *testing.T) {
	prev := goVersionsCommand
	goVersionsCommand = func(ctx context.Context, module string) ([]byte, error) {
		return []byte("github.com/flanksource/foo\n"), nil
	}
	t.Cleanup(func() { goVersionsCommand = prev })

	_, err := realLatestGoVersion(context.Background(), "github.com/flanksource/foo")
	assert.ErrorIs(t, err, ErrNoTaggedVersions)
}

func withStubbedLookup(t *testing.T, version string, retErr error) {
	t.Helper()
	prev := lookupLatestGoVersion
	lookupLatestGoVersion = func(context.Context, string) (string, error) {
		if retErr != nil {
			return "", retErr
		}
		return version, nil
	}
	t.Cleanup(func() { lookupLatestGoVersion = prev })
}

func withStubbedTidy(t *testing.T, sideEffect func(modDir string) error) *int {
	t.Helper()
	calls := 0
	prev := runGoModTidy
	runGoModTidy = func(modDir string) error {
		calls++
		if sideEffect != nil {
			return sideEffect(modDir)
		}
		return nil
	}
	t.Cleanup(func() { runGoModTidy = prev })
	return &calls
}

func TestRunLinkedDepsCheck_FailMode(t *testing.T) {
	repo := initCommitRepo(t)
	writeFile(t, repo, "go.mod", "module example.com/app\n\ngo 1.22\n\nreplace foo => ../../escaping\n")
	gitRun(t, repo, "add", "go.mod")

	_, err := RunLinkedDepsCheck(context.Background(), LinkedDepsParams{
		WorkDir:     repo,
		GitRoot:     repo,
		StagedFiles: []string{"go.mod"},
		Mode:        IgnoreCheckModeFail,
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "outside the git root")
	assert.Contains(t, err.Error(), "go.mod")
}

func TestRunLinkedDepsCheck_SkipMode(t *testing.T) {
	repo := initCommitRepo(t)
	writeFile(t, repo, "go.mod", "module example.com/app\n\ngo 1.22\n\nreplace foo => ../../escaping\n")
	gitRun(t, repo, "add", "go.mod")

	outcome, err := RunLinkedDepsCheck(context.Background(), LinkedDepsParams{
		WorkDir:     repo,
		GitRoot:     repo,
		StagedFiles: []string{"go.mod"},
		Mode:        IgnoreCheckModeSkip,
	})
	require.NoError(t, err)
	assert.False(t, outcome.Cancelled)
	assert.Empty(t, outcome.Unstaged)
}

func TestRunLinkedDepsCheck_NoViolationsIsNoOp(t *testing.T) {
	repo := initCommitRepo(t)
	writeFile(t, repo, "go.mod", "module example.com/app\n\ngo 1.22\n")
	gitRun(t, repo, "add", "go.mod")

	outcome, err := RunLinkedDepsCheck(context.Background(), LinkedDepsParams{
		WorkDir:     repo,
		GitRoot:     repo,
		StagedFiles: []string{"go.mod"},
		Mode:        IgnoreCheckModePrompt,
	})
	require.NoError(t, err)
	assert.Empty(t, outcome.Unstaged)
	assert.False(t, outcome.Cancelled)
}

func TestRunLinkedDepsCheck_NonTTYEscalatesToFail(t *testing.T) {
	prevTTY := stdinIsTerminal
	stdinIsTerminal = func() bool { return false }
	t.Cleanup(func() { stdinIsTerminal = prevTTY })

	repo := initCommitRepo(t)
	writeFile(t, repo, "go.mod", "module example.com/app\n\ngo 1.22\n\nreplace foo => ../../escaping\n")
	gitRun(t, repo, "add", "go.mod")

	_, err := RunLinkedDepsCheck(context.Background(), LinkedDepsParams{
		WorkDir:     repo,
		GitRoot:     repo,
		StagedFiles: []string{"go.mod"},
		Mode:        IgnoreCheckModePrompt,
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "outside the git root")
}

func TestRunLinkedDepsCheck_IgnoresViolationsAlreadyPresentOnHEAD(t *testing.T) {
	repo := initCommitRepo(t)
	writeFile(t, repo, "go.mod", `module example.com/app

go 1.22

require github.com/flanksource/commons v1.0.0

replace github.com/flanksource/commons => ../../commons
`)
	gitRun(t, repo, "add", "go.mod")
	gitRun(t, repo, "commit", "-m", "seed existing linked dep")

	writeFile(t, repo, "go.mod", `module example.com/app

go 1.22

require (
	github.com/flanksource/commons v1.0.0
	example.com/newdep v1.2.3
)

replace github.com/flanksource/commons => ../../commons
`)
	gitRun(t, repo, "add", "go.mod")

	source, err := readStagedSource(repo)
	require.NoError(t, err)

	outcome, err := RunLinkedDepsCheck(context.Background(), LinkedDepsParams{
		WorkDir:     repo,
		GitRoot:     repo,
		StagedFiles: source.Files,
		Changes:     source.Changes,
		Decider: func(context.Context, LinkedDepViolation) (LinkedDepChoice, error) {
			t.Fatal("pre-existing HEAD violation should not prompt")
			return LinkedDepChoice{Decision: LinkedDepDecisionCancel}, nil
		},
		Mode: IgnoreCheckModePrompt,
	})
	require.NoError(t, err)
	assert.False(t, outcome.Cancelled)
	assert.Empty(t, outcome.Unstaged)
}

func TestRunLinkedDepsCheck_FlagsChangedViolationRelativeToHEAD(t *testing.T) {
	repo := initCommitRepo(t)
	writeFile(t, repo, "go.mod", `module example.com/app

go 1.22

replace foo => ../../escaping
`)
	gitRun(t, repo, "add", "go.mod")
	gitRun(t, repo, "commit", "-m", "seed existing linked dep")

	writeFile(t, repo, "go.mod", `module example.com/app

go 1.22

replace foo => ../../../new-escaping
`)
	gitRun(t, repo, "add", "go.mod")

	source, err := readStagedSource(repo)
	require.NoError(t, err)

	_, err = RunLinkedDepsCheck(context.Background(), LinkedDepsParams{
		WorkDir:     repo,
		GitRoot:     repo,
		StagedFiles: source.Files,
		Changes:     source.Changes,
		Mode:        IgnoreCheckModeFail,
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "../../../new-escaping")
}
