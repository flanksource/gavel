package commit

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolveCommitFiles(t *testing.T) {
	cases := []struct {
		name    string
		gitRoot string
		baseDir string
		args    []string
		want    []string
		wantErr bool
	}{
		{
			name:    "relative arg from git root",
			gitRoot: "/repo",
			baseDir: "/repo",
			args:    []string{"pkg/a.go"},
			want:    []string{"pkg/a.go"},
		},
		{
			name:    "relative arg from a subdirectory resolves against baseDir",
			gitRoot: "/repo",
			baseDir: "/repo/pkg",
			args:    []string{"a.go", "b.go"},
			want:    []string{"pkg/a.go", "pkg/b.go"},
		},
		{
			name:    "absolute arg inside the repo",
			gitRoot: "/repo",
			baseDir: "/repo/pkg",
			args:    []string{"/repo/cmd/main.go"},
			want:    []string{"cmd/main.go"},
		},
		{
			name:    "dot passes through",
			gitRoot: "/repo",
			baseDir: "/repo",
			args:    []string{"."},
			want:    []string{"."},
		},
		{
			name:    "blank args are skipped",
			gitRoot: "/repo",
			baseDir: "/repo",
			args:    []string{"  ", "", "a.go"},
			want:    []string{"a.go"},
		},
		{
			name:    "path escaping the repo is rejected",
			gitRoot: "/repo",
			baseDir: "/repo",
			args:    []string{"../outside.go"},
			wantErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ResolveCommitFiles(tc.gitRoot, tc.baseDir, tc.args)
			if tc.wantErr {
				require.ErrorIs(t, err, ErrFilesOutsideRepo)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestRunCommitFilesCommitsOnlyNamedPaths(t *testing.T) {
	repo := initCommitRepo(t)
	writeFile(t, repo, "a.txt", "one\n")
	writeFile(t, repo, "b.txt", "two\n")
	writeFile(t, repo, "c.txt", "three\n")

	t.Setenv(testEnvVar, "1")

	result, err := Run(context.Background(), Options{WorkDir: repo, Files: []string{"a.txt", "b.txt"}})
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{"a.txt", "b.txt"}, result.Staged)

	assert.ElementsMatch(t, []string{"a.txt", "b.txt"}, committedFiles(t, repo))
	// c.txt was never named, so it stays an untracked working-tree change.
	assert.Contains(t, gitOutput(t, repo, "status", "--short"), "?? c.txt")
}

func TestRunCommitFilesOverridesPreStagedIndex(t *testing.T) {
	repo := initCommitRepo(t)
	writeFile(t, repo, "a.txt", "one\n")
	writeFile(t, repo, "other.txt", "other\n")
	gitRun(t, repo, "add", "other.txt")

	t.Setenv(testEnvVar, "1")

	result, err := Run(context.Background(), Options{WorkDir: repo, Files: []string{"a.txt"}})
	require.NoError(t, err)
	assert.Equal(t, []string{"a.txt"}, result.Staged)

	assert.Equal(t, []string{"a.txt"}, committedFiles(t, repo))
	// other.txt was reset out of the index and left untracked, not committed.
	assert.Contains(t, gitOutput(t, repo, "status", "--short"), "?? other.txt")
}

func TestRunCommitFilesWithCommitAllGroupsOnlyNamedPaths(t *testing.T) {
	repo := initCommitRepo(t)
	writeFileInDir(t, repo, "alpha/a.txt", "one\n")
	writeFileInDir(t, repo, "beta/b.txt", "two\n")
	writeFileInDir(t, repo, "gamma/c.txt", "three\n")

	t.Setenv(testEnvVar, "1")
	stubGroupPerFile(t)

	result, err := Run(context.Background(), Options{
		WorkDir:   repo,
		CommitAll: true,
		Files:     []string{"alpha/a.txt", "beta/b.txt"},
	})
	require.NoError(t, err)
	require.Len(t, result.Commits, 2)
	assert.Equal(t, []string{"alpha/a.txt"}, result.Commits[0].Files)
	assert.Equal(t, []string{"beta/b.txt"}, result.Commits[1].Files)
	// gamma/c.txt was not named, so grouping never saw it.
	assert.Contains(t, gitOutput(t, repo, "status", "--short"), "?? gamma/")
}

func TestRunCommitFilesRejectsIncompatibleFlags(t *testing.T) {
	cases := []struct {
		name string
		opts Options
		want error
	}{
		{name: "interactive", opts: Options{Interactive: true}, want: ErrFilesWithInteractive},
		{name: "since", opts: Options{Since: "HEAD~1"}, want: ErrFilesWithSince},
		{name: "fixup", opts: Options{Fixup: "abc123"}, want: ErrFilesWithFixup},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo := initCommitRepo(t)
			opts := tc.opts
			opts.WorkDir = repo
			opts.Files = []string{"a.txt"}

			_, err := Run(context.Background(), opts)
			require.ErrorIs(t, err, tc.want)
		})
	}
}

func TestRunCommitFilesUnknownPathspecFailsLoud(t *testing.T) {
	repo := initCommitRepo(t)
	t.Setenv(testEnvVar, "1")

	_, err := Run(context.Background(), Options{WorkDir: repo, Files: []string{"nope.txt"}})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "nope.txt")
}

// committedFiles returns the file paths changed by HEAD.
func committedFiles(t *testing.T, repo string) []string {
	t.Helper()
	out := gitOutput(t, repo, "diff-tree", "--no-commit-id", "--name-only", "-r", "HEAD")
	return splitLines(strings.TrimSpace(out))
}
