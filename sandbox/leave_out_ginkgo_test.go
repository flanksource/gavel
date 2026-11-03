package sandbox

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func leaveOutGit(dir string, args ...string) string {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GIT_AUTHOR_NAME=fixture", "GIT_AUTHOR_EMAIL=fixture@example.com", "GIT_COMMITTER_NAME=fixture", "GIT_COMMITTER_EMAIL=fixture@example.com")
	out, err := cmd.CombinedOutput()
	Expect(err).NotTo(HaveOccurred(), "git %v: %s", args, out)
	return strings.TrimSpace(string(out))
}

var _ = Describe("Literal landing exclusions", func() {
	for _, format := range []string{"sha1", "sha256"} {
		DescribeTable("restores exactly the selected filename with "+format, func(name string, existed bool) {
			work, remotes := GinkgoT().TempDir(), GinkgoT().TempDir()
			leaveOutGit(work, "init", "-q", "-b", "main", "--object-format="+format)
			for _, path := range []string{name, "a.txt"} {
				if path == name && !existed {
					continue
				}
				Expect(os.WriteFile(filepath.Join(work, path), []byte("before "+path), 0600)).To(Succeed())
			}
			leaveOutGit(work, "--literal-pathspecs", "add", "--all")
			leaveOutGit(work, "commit", "-q", "-m", "base")
			base := leaveOutGit(work, "rev-parse", "HEAD")
			remote := filepath.Join(remotes, "acme", "app.git")
			leaveOutGit(work, "clone", "-q", "--bare", work, remote)
			for _, path := range []string{name, "a.txt"} {
				Expect(os.WriteFile(filepath.Join(work, path), []byte("after "+path), 0600)).To(Succeed())
			}
			leaveOutGit(work, "--literal-pathspecs", "add", "--all")
			leaveOutGit(work, "commit", "-q", "-m", "run")
			head := leaveOutGit(work, "rev-parse", "HEAD")
			leaveOutGit(work, "push", "-q", remote, "HEAD:refs/heads/gavel/run")
			sources := &Sources{DataDir: GinkgoT().TempDir(), GitHubURL: "file://" + remotes,
				Token: func(context.Context, int64) (string, error) { return "fixture-token", nil }}
			repo := Repo{ID: uuid.New(), Org: uuid.New(), Source: "github", Name: "acme/app", GitHubRepoID: 1, InstallationID: 1}
			tip, err := sources.LeaveOut(context.Background(), repo, "gavel/run", base, head, []string{name})
			Expect(err).NotTo(HaveOccurred())
			Expect(leaveOutGit(remote, "diff", "--name-only", base, tip)).To(Equal("a.txt"))
			Expect(leaveOutGit(remote, "show", tip+":a.txt")).To(Equal("after a.txt"))
		},
			Entry("brackets", "[a].txt", true),
			Entry("pathspec magic", ":(glob)*.txt", true),
			Entry("option-shaped", "-option.txt", true),
			Entry("comma", "comma,name.txt", true),
			Entry("tab", "tab\tname.txt", true),
			Entry("newline", "line\nname.txt", true),
			Entry("Unicode", "例.txt", true),
			Entry("added brackets", "[added].txt", false),
			Entry("added option-shaped", "-added.txt", false),
		)
	}
})
