package git

import (
	"os/exec"
	"testing"
)

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

func TestCommitsWithTrailer(t *testing.T) {
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

	// Two commits share issue ABC123; one belongs to a different issue; one has
	// no trailer at all.
	git("commit", "--allow-empty", "-m", "feat: first\n\nGavel-Issue-Id: ABC123")
	git("commit", "--allow-empty", "-m", "feat: second\n\nGavel-Issue-Id: XYZ789")
	git("commit", "--allow-empty", "-m", "fix: third\n\nGavel-Issue-Id: ABC123")
	git("commit", "--allow-empty", "-m", "chore: untagged")

	commits, err := CommitsWithTrailer(dir, "Gavel-Issue-Id", "ABC123")
	if err != nil {
		t.Fatalf("CommitsWithTrailer: %v", err)
	}
	if len(commits) != 2 {
		t.Fatalf("got %d commits, want 2: %+v", len(commits), commits)
	}
	// Newest-first ordering: the "third" commit precedes "first". NewCommit
	// strips the conventional-commit type, so Subject is the bare description.
	if commits[0].Subject != "third" || commits[1].Subject != "first" {
		t.Fatalf("unexpected order/subjects: %q, %q", commits[0].Subject, commits[1].Subject)
	}
	for _, c := range commits {
		if c.Trailers["Gavel-Issue-Id"] != "ABC123" {
			t.Fatalf("commit %s has trailer %q, want ABC123", c.Hash, c.Trailers["Gavel-Issue-Id"])
		}
	}

	// An unmatched issue id yields no commits, not an error.
	none, err := CommitsWithTrailer(dir, "Gavel-Issue-Id", "NOPE")
	if err != nil {
		t.Fatalf("CommitsWithTrailer(NOPE): %v", err)
	}
	if len(none) != 0 {
		t.Fatalf("got %d commits for unmatched id, want 0", len(none))
	}

	// An empty value short-circuits to no commits.
	if got, err := CommitsWithTrailer(dir, "Gavel-Issue-Id", ""); err != nil || len(got) != 0 {
		t.Fatalf("empty value: got %d commits, err %v", len(got), err)
	}
}
