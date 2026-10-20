package prfix

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/flanksource/gavel/verify"
)

func git(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v: %s", args, err, out)
	}
	return strings.TrimSpace(string(out))
}

// conflictingBranch builds a repo on feat/thing whose go.mod bump conflicts with
// main's, and returns it with main's tip, the base the rebase replays onto.
func conflictingBranch(t *testing.T) (dir, base string) {
	t.Helper()
	dir = t.TempDir()
	git(t, dir, "init", "-b", "main")
	git(t, dir, "config", "user.email", "gavel@example.com")
	git(t, dir, "config", "user.name", "Gavel Test")
	writeGoMod(t, dir, "require xz v0.5.12\n")
	git(t, dir, "add", "go.mod")
	git(t, dir, "commit", "-m", "base")
	git(t, dir, "switch", "-c", "feat/thing")
	writeGoMod(t, dir, "require xz v0.5.14\n")
	git(t, dir, "commit", "-am", "bump on the branch")
	git(t, dir, "switch", "main")
	writeGoMod(t, dir, "require xz v0.5.20\n")
	git(t, dir, "commit", "-am", "bump on main")
	base = git(t, dir, "rev-parse", "HEAD")
	git(t, dir, "switch", "feat/thing")
	return dir, base
}

func writeGoMod(t *testing.T, dir, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}

// startRebase stops the rebase on its conflict, as the pr-fix hook leaves it.
func startRebase(t *testing.T, dir, base string) {
	t.Helper()
	cmd := exec.Command("git", "rebase", base)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err == nil {
		t.Fatalf("the rebase must stop on the go.mod conflict: %s", out)
	}
}

// failingRebaseChecks runs the rendered rebase checks (every verify command but
// the pr status poll) in dir and returns the stderr of each one that fails.
func failingRebaseChecks(t *testing.T, dir, base string) []string {
	t.Helper()
	pr := conflictingPR()
	pr.BaseSHA = base
	commands := resolve(t, verify.PromptSpec{}, pr).Workflow.Verify.Commands
	var failures []string
	for _, command := range commands[:len(commands)-1] {
		cmd := exec.Command("sh", "-c", command)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			failures = append(failures, strings.TrimSpace(string(out)))
		}
	}
	return failures
}

func TestRebaseChecksAgainstRealBranches(t *testing.T) {
	tests := []struct {
		name string
		// arrange leaves the branch in the state a turn might end in.
		arrange func(t *testing.T, dir, base string)
		// want is the start of each failing check's output; nil means all pass.
		want []string
	}{
		{
			name:    "rebase stopped on its conflict",
			arrange: startRebase,
			want:    []string{"the rebase is still in progress"},
		},
		{
			name: "base merged in instead of rebasing",
			arrange: func(t *testing.T, dir, base string) {
				cmd := exec.Command("git", "merge", base)
				cmd.Dir = dir
				_ = cmd.Run()
				writeGoMod(t, dir, "require xz v0.5.20\n")
				git(t, dir, "commit", "-am", "merge main")
			},
			want: []string{"merge commits on the branch"},
		},
		{
			name: "rebased and resolved",
			arrange: func(t *testing.T, dir, base string) {
				startRebase(t, dir, base)
				writeGoMod(t, dir, "require xz v0.5.20\n")
				git(t, dir, "add", "go.mod")
				git(t, dir, "-c", "core.editor=true", "rebase", "--continue")
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			dir, base := conflictingBranch(t)
			tc.arrange(t, dir, base)

			failures := failingRebaseChecks(t, dir, base)

			if len(failures) != len(tc.want) {
				t.Fatalf("failing checks = %q, want %d starting with %q", failures, len(tc.want), tc.want)
			}
			for i, prefix := range tc.want {
				if !strings.HasPrefix(failures[i], prefix) {
					t.Errorf("failure %d = %q, want it to start with %q", i, failures[i], prefix)
				}
			}
		})
	}
}
