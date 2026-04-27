package runners

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMatchesBuildContext(t *testing.T) {
	dir := t.TempDir()

	cases := []struct {
		name     string
		filename string
		body     string
		want     bool
	}{
		{
			name:     "plain test file matches",
			filename: "plain_test.go",
			body:     "package x\n",
			want:     true,
		},
		{
			name:     "//go:build never excludes the file",
			filename: "never_test.go",
			body:     "//go:build never\n\npackage x\n",
			want:     false,
		},
		{
			name:     "//go:build linux on darwin runners would not match — use a tag we do not pass",
			filename: "tagged_test.go",
			body:     "//go:build buildtag_that_definitely_isnt_set\n\npackage x\n",
			want:     false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(dir, tc.filename)
			if err := os.WriteFile(path, []byte(tc.body), 0o644); err != nil {
				t.Fatalf("write: %v", err)
			}
			got := matchesBuildContext(path)
			if got != tc.want {
				t.Errorf("matchesBuildContext(%s) = %v, want %v", tc.filename, got, tc.want)
			}
		})
	}
}

func TestMatchesBuildContextOnUnreadableReturnsTrue(t *testing.T) {
	// Nonexistent file: MatchFile errors out. We treat error as "permissive"
	// so the file isn't silently skipped on edge cases.
	got := matchesBuildContext("/does/not/exist/foo_test.go")
	if !got {
		t.Errorf("matchesBuildContext on missing file = false, want true (permissive default)")
	}
}
