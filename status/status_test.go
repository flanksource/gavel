package status

import (
	"github.com/flanksource/repomap"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseStatusPorcelain(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want []FileStatus
	}{
		{
			name: "modified unstaged",
			raw:  " M cmd/gavel/lint.go\x00",
			want: []FileStatus{{
				Path:     "cmd/gavel/lint.go",
				State:    StateUnstaged,
				WorkKind: KindModified,
			}},
		},
		{
			name: "modified staged",
			raw:  "M  cmd/gavel/lint.go\x00",
			want: []FileStatus{{
				Path:       "cmd/gavel/lint.go",
				State:      StateStaged,
				StagedKind: KindModified,
			}},
		},
		{
			name: "added staged",
			raw:  "A  cmd/gavel/lint_summary.go\x00",
			want: []FileStatus{{
				Path:       "cmd/gavel/lint_summary.go",
				State:      StateStaged,
				StagedKind: KindAdded,
			}},
		},
		{
			name: "deleted staged",
			raw:  "D  commit/ai-commit-group.md\x00",
			want: []FileStatus{{
				Path:       "commit/ai-commit-group.md",
				State:      StateStaged,
				StagedKind: KindDeleted,
			}},
		},
		{
			name: "both modified",
			raw:  "MM cmd/gavel/lint.go\x00",
			want: []FileStatus{{
				Path:       "cmd/gavel/lint.go",
				State:      StateBoth,
				StagedKind: KindModified,
				WorkKind:   KindModified,
			}},
		},
		{
			name: "untracked",
			raw:  "?? cmd/gavel/new.go\x00",
			want: []FileStatus{{
				Path:       "cmd/gavel/new.go",
				State:      StateUntracked,
				StagedKind: KindUntracked,
				WorkKind:   KindUntracked,
			}},
		},
		{
			name: "conflict UU",
			raw:  "UU commit/planner.go\x00",
			want: []FileStatus{{
				Path:           "commit/planner.go",
				State:          StateConflict,
				StagedKind:     KindModified,
				WorkKind:       KindModified,
				ConflictReason: ConflictReasonUnmerged,
			}},
		},
		{
			name: "conflict AA",
			raw:  "AA newfile.go\x00",
			want: []FileStatus{{
				Path:           "newfile.go",
				State:          StateConflict,
				StagedKind:     KindAdded,
				WorkKind:       KindAdded,
				ConflictReason: ConflictReasonUnmerged,
			}},
		},
		{
			name: "rename staged",
			raw:  "R  commit/new.go\x00commit/old.go\x00",
			want: []FileStatus{{
				Path:         "commit/new.go",
				PreviousPath: "commit/old.go",
				State:        StateStaged,
				StagedKind:   KindRenamed,
			}},
		},
		{
			name: "copy staged",
			raw:  "C  b.go\x00a.go\x00",
			want: []FileStatus{{
				Path:         "b.go",
				PreviousPath: "a.go",
				State:        StateStaged,
				StagedKind:   KindCopied,
			}},
		},
		{
			name: "multiple records",
			raw:  "M  a.go\x00 M b.go\x00?? c.go\x00",
			want: []FileStatus{
				{Path: "a.go", State: StateStaged, StagedKind: KindModified},
				{Path: "b.go", State: StateUnstaged, WorkKind: KindModified},
				{Path: "c.go", State: StateUntracked, StagedKind: KindUntracked, WorkKind: KindUntracked},
			},
		},
		{
			name: "empty output",
			raw:  "",
			want: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parseStatusPorcelain([]byte(tc.raw))
			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestGatherFromRepo(t *testing.T) {
	repo := initStatusRepo(t)
	require.NoError(t, os.WriteFile(filepath.Join(repo, "staged.go"), []byte("package x\n"), 0o644))
	gitRun(t, repo, "add", "staged.go")
	require.NoError(t, os.WriteFile(filepath.Join(repo, "README.md"), []byte("# test\nchanged\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(repo, "untracked.go"), []byte("package x\n"), 0o644))

	restore := stubFileMap(func(path, commit string) (*repomap.FileMap, error) {
		return &repomap.FileMap{Path: path, Language: "go"}, nil
	})
	defer restore()

	result, err := Gather(repo, Options{})
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Len(t, result.Files, 3)

	byPath := map[string]FileStatus{}
	for _, f := range result.Files {
		byPath[f.Path] = f
	}
	assert.Equal(t, StateStaged, byPath["staged.go"].State)
	assert.Equal(t, StateUnstaged, byPath["README.md"].State)
	assert.Equal(t, StateUntracked, byPath["untracked.go"].State)
	assert.Equal(t, "go", byPath["staged.go"].FileMap.Language)
}

func TestGatherHidesGitIgnoredTrackedFiles(t *testing.T) {
	repo := initStatusRepo(t)
	gitRun(t, repo, "config", "core.excludesFile", "/dev/null")
	require.NoError(t, os.WriteFile(filepath.Join(repo, ".gitignore"), []byte("dist/*\n!dist/keep.js\n"), 0o644))
	gitRun(t, repo, "add", ".gitignore")

	require.NoError(t, os.MkdirAll(filepath.Join(repo, "dist"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(repo, "dist", "bundle.js"), []byte("v1\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(repo, "dist", "keep.js"), []byte("v1\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(repo, "app.go"), []byte("package x\n"), 0o644))
	gitRun(t, repo, "add", "-f", "dist/bundle.js")   // force-track the ignored bundle
	gitRun(t, repo, "add", "dist/keep.js", "app.go") // keep.js is !-negated, so a plain add works
	gitRun(t, repo, "commit", "-m", "baseline")

	// Modify all three; leave them unstaged.
	require.NoError(t, os.WriteFile(filepath.Join(repo, "dist", "bundle.js"), []byte("v2\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(repo, "dist", "keep.js"), []byte("v2\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(repo, "app.go"), []byte("package x\n// edit\n"), 0o644))

	result, err := Gather(repo, Options{NoRepomap: true})
	require.NoError(t, err)
	paths := pathsOf(result.Files)
	assert.NotContains(t, paths, "dist/bundle.js", "force-tracked .gitignore'd file must be hidden")
	assert.Contains(t, paths, "dist/keep.js", "!-negated file must remain visible")
	assert.Contains(t, paths, "app.go", "normal tracked modification must remain visible")
}

func TestGatherKeepsStagedGitIgnoredFiles(t *testing.T) {
	repo := initStatusRepo(t)
	gitRun(t, repo, "config", "core.excludesFile", "/dev/null")
	require.NoError(t, os.WriteFile(filepath.Join(repo, ".gitignore"), []byte("dist/\n"), 0o644))
	gitRun(t, repo, "add", ".gitignore")
	require.NoError(t, os.MkdirAll(filepath.Join(repo, "dist"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(repo, "dist", "bundle.js"), []byte("v1\n"), 0o644))
	gitRun(t, repo, "add", "-f", "dist/bundle.js")
	gitRun(t, repo, "commit", "-m", "baseline")

	// Modify and stage the ignored bundle: an explicit `git add` keeps it visible.
	require.NoError(t, os.WriteFile(filepath.Join(repo, "dist", "bundle.js"), []byte("v2\n"), 0o644))
	gitRun(t, repo, "add", "-f", "dist/bundle.js")

	result, err := Gather(repo, Options{NoRepomap: true})
	require.NoError(t, err)
	byPath := map[string]FileStatus{}
	for _, f := range result.Files {
		byPath[f.Path] = f
	}
	f, ok := byPath["dist/bundle.js"]
	require.True(t, ok, "manually staged .gitignore'd file must stay visible")
	assert.Equal(t, StateStaged, f.State)
}

func TestGatherWithFolderFilter(t *testing.T) {
	repo := initStatusRepo(t)

	// Commit a baseline so each file path is tracked. Without this, git
	// porcelain output collapses entirely-untracked directories into a
	// single "sub/" entry rather than reporting each file individually.
	require.NoError(t, os.WriteFile(filepath.Join(repo, "top.go"), []byte("package x\n"), 0o644))
	require.NoError(t, os.MkdirAll(filepath.Join(repo, "sub"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(repo, "sub", "inner.go"), []byte("package x\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(repo, "sub", "doc.md"), []byte("# sub\n"), 0o644))
	require.NoError(t, os.MkdirAll(filepath.Join(repo, "other"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(repo, "other", "thing.go"), []byte("package x\n"), 0o644))
	gitRun(t, repo, "add", "top.go", "sub/inner.go", "sub/doc.md", "other/thing.go")
	gitRun(t, repo, "commit", "-m", "baseline")

	// Now modify each file so they reappear in `git status --porcelain`.
	require.NoError(t, os.WriteFile(filepath.Join(repo, "top.go"), []byte("package x\n// top\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(repo, "sub", "inner.go"), []byte("package x\n// inner\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(repo, "sub", "doc.md"), []byte("# sub\nupdated\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(repo, "other", "thing.go"), []byte("package x\n// other\n"), 0o644))

	restore := stubFileMap(func(path, commit string) (*repomap.FileMap, error) {
		return &repomap.FileMap{Path: path, Language: "go"}, nil
	})
	defer restore()

	t.Run("subfolder prefix keeps only matching files", func(t *testing.T) {
		result, err := Gather(repo, Options{FolderFilter: "sub"})
		require.NoError(t, err)
		assert.ElementsMatch(t,
			[]string{"sub/inner.go", "sub/doc.md"},
			pathsOf(result.Files),
		)
	})

	t.Run("empty filter keeps everything", func(t *testing.T) {
		result, err := Gather(repo, Options{FolderFilter: ""})
		require.NoError(t, err)
		assert.ElementsMatch(t,
			[]string{"top.go", "sub/inner.go", "sub/doc.md", "other/thing.go"},
			pathsOf(result.Files),
		)
	})

	t.Run("dot filter keeps everything", func(t *testing.T) {
		result, err := Gather(repo, Options{FolderFilter: "."})
		require.NoError(t, err)
		assert.Len(t, result.Files, 4)
	})

	t.Run("exact file path keeps only that file", func(t *testing.T) {
		result, err := Gather(repo, Options{FolderFilter: "sub/inner.go"})
		require.NoError(t, err)
		assert.Equal(t, []string{"sub/inner.go"}, pathsOf(result.Files))
	})

	t.Run("prefix does not match similarly named sibling", func(t *testing.T) {
		// "sub" must not match a sibling directory like "subother/".
		require.NoError(t, os.MkdirAll(filepath.Join(repo, "subother"), 0o755))
		require.NoError(t, os.WriteFile(filepath.Join(repo, "subother", "x.go"), []byte("package x\n"), 0o644))
		gitRun(t, repo, "add", "subother/x.go")
		gitRun(t, repo, "commit", "-m", "subother baseline")
		require.NoError(t, os.WriteFile(filepath.Join(repo, "subother", "x.go"), []byte("package x\n// upd\n"), 0o644))

		result, err := Gather(repo, Options{FolderFilter: "sub"})
		require.NoError(t, err)
		for _, f := range result.Files {
			assert.False(t, strings.HasPrefix(f.Path, "subother/"),
				"folder filter %q must not match sibling %q", "sub", f.Path)
		}
	})
}

func pathsOf(files []FileStatus) []string {
	out := make([]string, len(files))
	for i, f := range files {
		out[i] = f.Path
	}
	return out
}

func initStatusRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	gitRun(t, dir, "init")
	gitRun(t, dir, "config", "user.email", "test@example.com")
	gitRun(t, dir, "config", "user.name", "Test User")
	gitRun(t, dir, "config", "commit.gpgsign", "false")
	require.NoError(t, os.WriteFile(filepath.Join(dir, "README.md"), []byte("# test\n"), 0o644))
	gitRun(t, dir, "add", "README.md")
	gitRun(t, dir, "commit", "-m", "initial commit")
	return dir
}

func gitRun(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, "git %v failed: %s", args, out)
}

func stubFileMap(fn func(path, commit string) (*repomap.FileMap, error)) func() {
	previous := fetchFileMapFunc
	fetchFileMapFunc = fn
	return func() {
		fetchFileMapFunc = previous
	}
}
