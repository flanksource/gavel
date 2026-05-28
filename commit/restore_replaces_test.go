package commit

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRestoreLocalReplaces_PutsReplaceBackUnstaged(t *testing.T) {
	tidyCalls := withStubbedTidy(t, nil)

	repo := initCommitRepo(t)
	// Simulate a post-commit state: go.mod has only the upgraded require,
	// no local replace. The commit is already in HEAD.
	writeFile(t, repo, "go.mod", `module example.com/app

go 1.22

require github.com/flanksource/foo v1.4.2
`)
	gitRun(t, repo, "add", "go.mod")
	gitRun(t, repo, "commit", "-m", "upgrade foo to v1.4.2")

	restoreLocalReplaces(repo, []pendingRestore{{
		GoModFile: "go.mod",
		Replaces: []goModUpgrade{{
			OldPath:    "github.com/flanksource/foo",
			OldVersion: "",
			OldTarget:  "../foo",
			NewVersion: "v1.4.2",
		}},
	}})

	body := readFile(t, filepath.Join(repo, "go.mod"))
	assert.Contains(t, body, "replace github.com/flanksource/foo")
	assert.Contains(t, body, "../foo")
	assert.Contains(t, body, "require github.com/flanksource/foo v1.4.2",
		"require line must stay; replace coexists and overrides locally")

	// go.mod is modified locally but NOT staged.
	stagedNames := gitOutput(t, repo, "diff", "--cached", "--name-only")
	assert.NotContains(t, stagedNames, "go.mod", "restore must leave go.mod unstaged")

	unstagedNames := gitOutput(t, repo, "diff", "--name-only")
	assert.Contains(t, unstagedNames, "go.mod", "working tree must show the restored replace as a local edit")

	assert.Equal(t, 1, *tidyCalls)
}

func TestRestoreLocalReplaces_TidyFailureDoesNotFailRestore(t *testing.T) {
	withStubbedTidy(t, func(string) error {
		return errors.New("simulated tidy failure")
	})

	repo := initCommitRepo(t)
	writeFile(t, repo, "go.mod", `module example.com/app

go 1.22

require github.com/flanksource/foo v1.4.2
`)
	gitRun(t, repo, "add", "go.mod")
	gitRun(t, repo, "commit", "-m", "upgrade")

	// Must not panic or propagate the tidy error — restore is best-effort.
	restoreLocalReplaces(repo, []pendingRestore{{
		GoModFile: "go.mod",
		Replaces: []goModUpgrade{{
			OldPath:   "github.com/flanksource/foo",
			OldTarget: "../foo",
		}},
	}})

	body := readFile(t, filepath.Join(repo, "go.mod"))
	assert.Contains(t, body, "replace github.com/flanksource/foo",
		"replace must still be written even if tidy fails")
}

func TestRestoreLocalReplaces_EmptyIsNoOp(t *testing.T) {
	repo := initCommitRepo(t)
	require.NotPanics(t, func() {
		restoreLocalReplaces(repo, nil)
	})
}
