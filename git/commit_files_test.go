package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseCommitFiles(t *testing.T) {
	patch := strings.Join([]string{
		`diff --git a/added.go b/added.go`,
		`new file mode 100644`,
		`index 0000000..abc1234`,
		`--- /dev/null`,
		`+++ b/added.go`,
		`@@ -0,0 +1,2 @@`,
		`+package main`,
		`+func main() {}`,
		`diff --git a/modified.go b/modified.go`,
		`index 1111111..2222222 100644`,
		`--- a/modified.go`,
		`+++ b/modified.go`,
		`@@ -1,3 +1,3 @@`,
		` package main`,
		`-old line`,
		`+new line`,
		`diff --git a/deleted.go b/deleted.go`,
		`deleted file mode 100644`,
		`index 3333333..0000000`,
		`--- a/deleted.go`,
		`+++ /dev/null`,
		`@@ -1,2 +0,0 @@`,
		`-package main`,
		`-func gone() {}`,
		`diff --git a/old/name.go b/new/name.go`,
		`similarity index 80%`,
		`rename from old/name.go`,
		`rename to new/name.go`,
		`index 4444444..5555555 100644`,
		`--- a/old/name.go`,
		`+++ b/new/name.go`,
		`@@ -1,2 +1,2 @@`,
		` package main`,
		`-renamed old`,
		`+renamed new`,
		`diff --git a/img.png b/img.png`,
		`new file mode 100644`,
		`index 0000000..6666666`,
		`Binary files /dev/null and b/img.png differ`,
	}, "\n")

	got := parseCommitFiles(patch)
	byPath := make(map[string]CommitFile, len(got))
	for _, f := range got {
		byPath[f.Path] = f
	}

	want := map[string]CommitFile{
		"added.go":    {Path: "added.go", Status: "added", Adds: 2, Dels: 0},
		"modified.go": {Path: "modified.go", Status: "modified", Adds: 1, Dels: 1},
		"deleted.go":  {Path: "deleted.go", Status: "deleted", Adds: 0, Dels: 2},
		"new/name.go": {Path: "new/name.go", PreviousPath: "old/name.go", Status: "renamed", Adds: 1, Dels: 1},
		"img.png":     {Path: "img.png", Status: "added", Binary: true},
	}
	if len(got) != len(want) {
		t.Fatalf("parsed %d files, want %d: %+v", len(got), len(want), got)
	}
	for path, exp := range want {
		f, ok := byPath[path]
		if !ok {
			t.Fatalf("missing file %q in %+v", path, got)
		}
		if f.Status != exp.Status || f.Adds != exp.Adds || f.Dels != exp.Dels ||
			f.PreviousPath != exp.PreviousPath || f.Binary != exp.Binary {
			t.Fatalf("file %q = %+v, want %+v", path, f, exp)
		}
	}
}

func TestCommitFiles(t *testing.T) {
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

	write := func(name, content string) {
		t.Helper()
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}
	write("keep.go", "package main\n\nfunc keep() {}\n")
	write("gone.go", "package main\n\nfunc gone() {}\n")
	git("add", "keep.go", "gone.go")
	git("commit", "-q", "-m", "seed")

	// One commit that adds a file, edits another, and deletes a third.
	write("new.go", "package main\n\nfunc fresh() {}\n")
	write("keep.go", "package main\n\nfunc keep() { println(1) }\n")
	git("rm", "-q", "gone.go")
	git("add", "new.go", "keep.go")
	git("commit", "-q", "-m", "change")

	head := strings.TrimSpace(runGit(t, dir, "rev-parse", "HEAD"))
	files, err := CommitFiles(dir, head)
	if err != nil {
		t.Fatalf("CommitFiles: %v", err)
	}
	byPath := make(map[string]CommitFile, len(files))
	for _, f := range files {
		byPath[f.Path] = f
	}
	if got := byPath["new.go"].Status; got != "added" {
		t.Fatalf("new.go status = %q, want added", got)
	}
	if got := byPath["keep.go"].Status; got != "modified" {
		t.Fatalf("keep.go status = %q, want modified", got)
	}
	if got := byPath["gone.go"].Status; got != "deleted" {
		t.Fatalf("gone.go status = %q, want deleted", got)
	}
	// keep.go swaps one line for another: one add, one delete.
	if f := byPath["keep.go"]; f.Adds != 1 || f.Dels != 1 {
		t.Fatalf("keep.go counts = +%d/-%d, want +1/-1", f.Adds, f.Dels)
	}

	if _, err := CommitFiles(dir, "not-a-hash"); err == nil {
		t.Fatal("expected error for invalid hash")
	}
}

func TestCommitDiffFileScopesToOnePath(t *testing.T) {
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
	if err := os.WriteFile(filepath.Join(dir, "a.txt"), []byte("a\n"), 0o644); err != nil {
		t.Fatalf("write a: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "b.txt"), []byte("b\n"), 0o644); err != nil {
		t.Fatalf("write b: %v", err)
	}
	git("add", "a.txt", "b.txt")
	git("commit", "-q", "-m", "two files")

	head := strings.TrimSpace(runGit(t, dir, "rev-parse", "HEAD"))
	diff, _, err := CommitDiff(dir, head, "a.txt")
	if err != nil {
		t.Fatalf("CommitDiff(file): %v", err)
	}
	if !strings.Contains(diff, "a.txt") {
		t.Fatalf("scoped diff missing a.txt:\n%s", diff)
	}
	if strings.Contains(diff, "b.txt") {
		t.Fatalf("scoped diff should exclude b.txt:\n%s", diff)
	}
}
