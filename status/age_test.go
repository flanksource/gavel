package status

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEnrichWithModTime(t *testing.T) {
	repo := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(repo, "live.go"), []byte("package x\n"), 0o644))

	want := time.Now().Add(-2 * time.Hour)
	require.NoError(t, os.Chtimes(filepath.Join(repo, "live.go"), want, want))

	files := []FileStatus{
		{Path: "live.go", State: StateUnstaged, WorkKind: KindModified},
		{Path: "missing.go", State: StateStaged, StagedKind: KindDeleted},
		{Path: "ghost.go", State: StateUnstaged, WorkKind: KindModified},
	}

	enrichWithModTime(repo, files)

	assert.WithinDuration(t, want, files[0].ModifiedAt, time.Second,
		"existing file should record its mtime")
	assert.True(t, files[1].ModifiedAt.IsZero(),
		"deleted files should be skipped without statting")
	assert.True(t, files[2].ModifiedAt.IsZero(),
		"missing files should leave ModifiedAt zero, not crash")
}
