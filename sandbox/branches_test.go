// Tests branch lookups: a repository's branches and its default, and the commit a branch a member
// picked is at, refusing names that aren't branches of the repository.
package sandbox

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBranchesListTheRepositorysBranches(t *testing.T) {
	dataDir := t.TempDir()
	org, repo := uuid.New(), uuid.New()
	bare := filepath.Join(dataDir, "repos", org.String(), repo.String()+".git")
	work := newWork(t)
	commitFile(t, work, "README.md", "main\n")
	git(t, work, "checkout", "-q", "-b", "feat/x")
	commitFile(t, work, "README.md", "feature\n")
	git(t, work, "checkout", "-q", "main")
	git(t, work, "clone", "-q", "--bare", work, bare)
	sources := &Sources{DataDir: dataDir}
	target := Repo{ID: repo, Org: org, Source: "push", Name: "app"}

	branches, err := sources.Branches(context.Background(), target)
	require.NoError(t, err)
	assert.Equal(t, Branches{Default: "main", Names: []string{"feat/x", "main"}}, branches)

	sha, err := sources.BranchCommit(context.Background(), target, "feat/x")
	require.NoError(t, err)
	assert.Equal(t, git(t, work, "rev-parse", "feat/x"), sha)

	_, err = sources.BranchCommit(context.Background(), target, "nope")
	require.EqualError(t, err, `unknown branch "nope": app has no such branch`)
	require.ErrorIs(t, err, ErrUnknownBranch)
	_, err = sources.BranchCommit(context.Background(), target, "--upload-pack=x")
	require.EqualError(t, err, `unknown branch "--upload-pack=x": not a valid branch name`)
	_, err = sources.BranchCommit(context.Background(), target, "")
	require.EqualError(t, err, `unknown branch "": not a valid branch name`)
}
