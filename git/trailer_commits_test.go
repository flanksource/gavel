package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// ansiEscape matches the CSI color sequences `git ... --color=always` emits, so
// tests can assert on diff text without depending on where git splits its spans.
var ansiEscape = regexp.MustCompile(`\x1b\[[0-9;]*m`)

func stripANSI(s string) string {
	return ansiEscape.ReplaceAllString(s, "")
}

func TestRemoteToWebURL(t *testing.T) {
	cases := []struct {
		name   string
		remote string
		want   string
	}{
		{"ssh scp form", "git@github.com:flanksource/gavel.git", "https://github.com/flanksource/gavel"},
		{"ssh scp no suffix", "git@github.com:flanksource/gavel", "https://github.com/flanksource/gavel"},
		{"https with suffix", "https://github.com/flanksource/gavel.git", "https://github.com/flanksource/gavel"},
		{"https no suffix", "https://github.com/flanksource/gavel", "https://github.com/flanksource/gavel"},
		{"ssh scheme userinfo", "ssh://git@github.com/flanksource/gavel.git", "https://github.com/flanksource/gavel"},
		{"gitlab subgroup", "git@gitlab.com:group/sub/repo.git", "https://gitlab.com/group/sub/repo"},
		{"empty", "", ""},
		{"local path", "/srv/git/repo.git", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := remoteToWebURL(tc.remote); got != tc.want {
				t.Fatalf("remoteToWebURL(%q) = %q, want %q", tc.remote, got, tc.want)
			}
		})
	}
}

func TestIsValidCommitHash(t *testing.T) {
	valid := []string{"abc1", "0123456789abcdef", "ABCDEF0123456789ABCDEF0123456789ABCDEF01"}
	for _, h := range valid {
		if !IsValidCommitHash(h) {
			t.Fatalf("IsValidCommitHash(%q) = false, want true", h)
		}
	}
	invalid := []string{"", "abc", "xyz1", "abc123; rm -rf", "--all", "abc 123"}
	for _, h := range invalid {
		if IsValidCommitHash(h) {
			t.Fatalf("IsValidCommitHash(%q) = true, want false", h)
		}
	}
}

func TestCommitDiff(t *testing.T) {
	dir := t.TempDir()
	git := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	git("init", "-q")
	git("config", "user.email", "test@example.com")
	git("config", "user.name", "Test User")
	git("config", "commit.gpgsign", "false")
	if err := os.WriteFile(filepath.Join(dir, "hello.txt"), []byte("hi\n"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}
	git("add", "hello.txt")
	git("commit", "-m", "feat: add hello")

	head := strings.TrimSpace(runGit(t, dir, "rev-parse", "HEAD"))
	diff, truncated, err := CommitDiff(dir, head, "")
	if err != nil {
		t.Fatalf("CommitDiff: %v", err)
	}
	if truncated {
		t.Fatalf("small diff reported truncated")
	}
	// The diffstat names the file and the patch adds its content. Strip the
	// forced ANSI coloring first: git colors the "+" and "hi" as separate spans,
	// so the raw bytes never contain a contiguous "+hi".
	plain := stripANSI(diff)
	if !strings.Contains(plain, "hello.txt") || !strings.Contains(plain, "+hi") {
		t.Fatalf("diff missing expected content:\n%s", diff)
	}

	if _, _, err := CommitDiff(dir, "not-a-hash", ""); err == nil {
		t.Fatal("expected error for invalid hash")
	}
}

func runGit(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("git %v: %v", args, err)
	}
	return string(out)
}
