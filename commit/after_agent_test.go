package commit

import (
	"path/filepath"
	"testing"
)

func TestResolveAgentCommitDir(t *testing.T) {
	abs := filepath.Join(string(filepath.Separator), "abs", "path")
	cases := []struct {
		name    string
		workDir string
		cwd     string
		want    string
	}{
		{"empty cwd falls back to workDir", "/repo", "", "/repo"},
		{"relative cwd joins workDir", "/repo", "sub/dir", filepath.Clean("/repo/sub/dir")},
		{"absolute cwd is used verbatim", "/repo", abs, abs},
		{"relative cwd without workDir is cleaned", "", "sub/dir", filepath.Clean("sub/dir")},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := resolveAgentCommitDir(tc.workDir, tc.cwd); got != tc.want {
				t.Fatalf("resolveAgentCommitDir(%q, %q) = %q, want %q", tc.workDir, tc.cwd, got, tc.want)
			}
		})
	}
}
