package status

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/flanksource/repomap"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEnrichWithConflictMarkers(t *testing.T) {
	dir := t.TempDir()

	mustWrite := func(name string, data []byte) string {
		t.Helper()
		path := filepath.Join(dir, name)
		require.NoError(t, os.WriteFile(path, data, 0o644))
		return name
	}

	cleanGo := mustWrite("clean.go", []byte("package x\n\nfunc Hello() {}\n"))
	allMarkers := mustWrite("conflicted.go", []byte(strings.Join([]string{
		"package x",
		"",
		"<<<<<<< HEAD",
		"a := 1",
		"=======",
		"a := 2",
		">>>>>>> theirs",
		"",
	}, "\n")))
	startOnly := mustWrite("partial.go", []byte("package x\n<<<<<<< HEAD\nstuff\n"))
	midOnly := mustWrite("docs.md", []byte("# Title\n\n=======\n\nbody\n"))
	binary := mustWrite("blob.bin", append([]byte{0x00, 0x01, 0x02},
		[]byte("\n<<<<<<< HEAD\n=======\n>>>>>>> theirs\n")...))
	oversize := mustWrite("huge.txt", append(bytes.Repeat([]byte("a"), conflictMaxFileBytes+1),
		[]byte("\n<<<<<<< HEAD\n=======\n>>>>>>> theirs\n")...))
	deletedPath := mustWrite("gone.go", []byte("<<<<<<< HEAD\n=======\n>>>>>>> theirs\n"))
	alreadyConflict := mustWrite("already.go", []byte("package x\n"))

	files := []FileStatus{
		{Path: cleanGo, State: StateUnstaged, WorkKind: KindModified},
		{Path: allMarkers, State: StateStaged, StagedKind: KindModified},
		{Path: startOnly, State: StateUnstaged, WorkKind: KindModified},
		{Path: midOnly, State: StateUnstaged, WorkKind: KindModified},
		{Path: binary, State: StateUntracked, StagedKind: KindUntracked, WorkKind: KindUntracked},
		{Path: oversize, State: StateUnstaged, WorkKind: KindModified},
		{Path: deletedPath, State: StateStaged, StagedKind: KindDeleted},
		{
			Path: alreadyConflict, State: StateConflict, StagedKind: KindModified,
			WorkKind: KindModified, ConflictReason: ConflictReasonUnmerged,
		},
	}

	enrichWithConflictMarkers(dir, files)

	byPath := map[string]FileStatus{}
	for _, f := range files {
		byPath[f.Path] = f
	}

	assert.Equal(t, StateUnstaged, byPath[cleanGo].State, "clean file should not be promoted")
	assert.Empty(t, string(byPath[cleanGo].ConflictReason))

	assert.Equal(t, StateConflict, byPath[allMarkers].State, "file with all three markers should be promoted")
	assert.Equal(t, ConflictReasonMarker, byPath[allMarkers].ConflictReason)

	assert.Equal(t, StateUnstaged, byPath[startOnly].State, "single marker should not trigger")
	assert.Equal(t, StateUnstaged, byPath[midOnly].State, "lone ======= line should not trigger")
	assert.Equal(t, StateUntracked, byPath[binary].State, "binary file should be skipped")
	assert.Equal(t, StateUnstaged, byPath[oversize].State, "oversized file should be skipped")
	assert.Equal(t, StateStaged, byPath[deletedPath].State, "deleted file should be skipped")

	assert.Equal(t, StateConflict, byPath[alreadyConflict].State, "already-conflict file stays conflict")
	assert.Equal(t, ConflictReasonUnmerged, byPath[alreadyConflict].ConflictReason,
		"already-conflict file keeps its original reason")
}

func TestPrettyMarkerConflict(t *testing.T) {
	r := &Result{
		Branch: "main",
		Files: []FileStatus{
			{
				Path: "broken.go", State: StateConflict,
				StagedKind: KindModified, WorkKind: KindModified,
				ConflictReason: ConflictReasonMarker,
			},
		},
	}
	clean := stripANSI(r.Pretty().ANSI())
	assert.Contains(t, clean, "⚠ conflict (markers)")
	assert.Contains(t, clean, "=1")
}

func TestGatherPromotesMarkerFiles(t *testing.T) {
	repo := initStatusRepo(t)
	conflicted := strings.Join([]string{
		"package x",
		"",
		"<<<<<<< HEAD",
		"a := 1",
		"=======",
		"a := 2",
		">>>>>>> theirs",
		"",
	}, "\n")
	require.NoError(t, os.WriteFile(filepath.Join(repo, "broken.go"), []byte(conflicted), 0o644))
	gitRun(t, repo, "add", "broken.go")

	restore := stubFileMap(func(path, _ string) (*repomap.FileMap, error) {
		return &repomap.FileMap{Path: path, Language: "go"}, nil
	})
	defer restore()

	result, err := Gather(repo, Options{NoRepomap: true})
	require.NoError(t, err)
	require.Len(t, result.Files, 1)

	got := result.Files[0]
	assert.Equal(t, "broken.go", got.Path)
	assert.Equal(t, StateConflict, got.State)
	assert.Equal(t, ConflictReasonMarker, got.ConflictReason)
	assert.Equal(t, 1, result.Counts().Conflict)

	clean := stripANSI(result.Pretty().ANSI())
	assert.Contains(t, clean, "=1")
	assert.Contains(t, clean, "⚠ conflict (markers)")
}
