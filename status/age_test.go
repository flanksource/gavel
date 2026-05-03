package status

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHumanAge(t *testing.T) {
	cases := []struct {
		name string
		d    time.Duration
		want string
	}{
		{"zero is empty", 0, ""},
		{"negative is empty", -time.Second, ""},
		{"sub-minute uses seconds", 30 * time.Second, "30s"},
		{"sub-hour uses minutes", 5 * time.Minute, "5m"},
		{"sub-day uses hours", 3 * time.Hour, "3h"},
		{"day boundary", 48 * time.Hour, "2d"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, HumanAge(tc.d))
		})
	}
}

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
