package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestTrailerDiffStats(t *testing.T) {
	dir := t.TempDir()
	git := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	write := func(name, body string) {
		t.Helper()
		if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}
	git("init", "-q")
	git("config", "user.email", "test@example.com")
	git("config", "user.name", "Test User")
	git("config", "commit.gpgsign", "false")

	// Issue ABC123: two commits. First adds a 2-line file, second adds one line
	// to it and adds a new file — 3 files touched (a.txt counted once), 4 adds.
	write("a.txt", "one\ntwo\n")
	git("add", "a.txt")
	git("commit", "-m", "feat: first\n\nGavel-Issue-Id: ABC123")
	write("a.txt", "one\ntwo\nthree\n")
	write("b.txt", "b\n")
	git("add", "a.txt", "b.txt")
	git("commit", "-m", "feat: second\n\nGavel-Issue-Id: ABC123")

	// Issue XYZ789: one commit deleting a line from a.txt.
	write("a.txt", "one\ntwo\n")
	git("add", "a.txt")
	git("commit", "-m", "fix: third\n\nGavel-Issue-Id: XYZ789")

	// An untagged commit must not contribute to any issue.
	write("c.txt", "c\n")
	git("add", "c.txt")
	git("commit", "-m", "chore: untagged")

	stats, err := TrailerDiffStats(dir, "Gavel-Issue-Id")
	if err != nil {
		t.Fatalf("TrailerDiffStats: %v", err)
	}

	abc := stats["ABC123"]
	// 2 commits; files a.txt + b.txt (a.txt touched by both, counted once) = 2;
	// adds: 2 (a.txt v1) + 1 (a.txt v2) + 1 (b.txt) = 4; dels: 0.
	if abc.Commits != 2 || abc.Files != 2 || abc.Adds != 4 || abc.Dels != 0 {
		t.Fatalf("ABC123 stats = %+v, want {Commits:2 Files:2 Adds:4 Dels:0}", abc)
	}

	xyz := stats["XYZ789"]
	if xyz.Commits != 1 || xyz.Files != 1 || xyz.Adds != 0 || xyz.Dels != 1 {
		t.Fatalf("XYZ789 stats = %+v, want {Commits:1 Files:1 Adds:0 Dels:1}", xyz)
	}

	if _, ok := stats[""]; ok {
		t.Fatalf("untagged commit leaked into stats under empty key: %+v", stats)
	}
}

func TestTrailerDiffStatsEmptyRepo(t *testing.T) {
	dir := t.TempDir()
	cmd := exec.Command("git", "init", "-q")
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git init: %v\n%s", err, out)
	}
	// A repo with no commits yet is a valid empty result, not an error.
	stats, err := TrailerDiffStats(dir, "Gavel-Issue-Id")
	if err != nil {
		t.Fatalf("TrailerDiffStats on empty repo: %v", err)
	}
	if len(stats) != 0 {
		t.Fatalf("empty repo returned %d stats, want 0", len(stats))
	}
}
