package commit

import (
	"context"
	"path/filepath"
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
	return func(context.Context, LinkedDepViolation) (LinkedDepDecision, error) { return d, nil }
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
