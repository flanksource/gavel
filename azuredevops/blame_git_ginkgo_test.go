package azuredevops

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Azure reviewer checkout", func() {
	It("fetches a missing fork head without moving refs, FETCH_HEAD, the index or working files", func() {
		root := GinkgoT().TempDir()
		source, target, checkout := filepath.Join(root, "fork"), filepath.Join(root, "target.git"), filepath.Join(root, "checkout")
		git := func(dir string, args ...string) string {
			cmd := exec.Command("git", args...)
			cmd.Dir = dir
			out, err := cmd.CombinedOutput()
			Expect(err).NotTo(HaveOccurred(), "%s", out)
			return strings.TrimSpace(string(out))
		}
		Expect(os.Mkdir(source, 0755)).To(Succeed())
		git(source, "init", "-q", "-b", "feature")
		git(source, "config", "commit.gpgsign", "false")
		Expect(os.WriteFile(filepath.Join(source, "source.txt"), []byte("base\n"), 0644)).To(Succeed())
		git(source, "add", "--", "source.txt")
		git(source, "-c", "user.name=Alice", "-c", "user.email=alice@example.com", "commit", "-qm", "base")
		base := git(source, "rev-parse", "HEAD")
		Expect(os.WriteFile(filepath.Join(source, "source.txt"), []byte("head\n"), 0644)).To(Succeed())
		git(source, "add", "--", "source.txt")
		git(source, "-c", "user.name=Author", "-c", "user.email=author@example.com", "commit", "-qm", "head")
		head := git(source, "rev-parse", "HEAD")
		git(root, "clone", "-q", "--bare", source, target)
		git(target, "update-ref", "refs/heads/main", base)
		git(target, "update-ref", "-d", "refs/heads/feature")
		git(root, "clone", "-q", "--no-local", "--single-branch", "-b", "main", target, checkout)
		git(checkout, "config", "url."+source+".insteadOf", "https://dev.azure.com/acme/product/_git/fork")
		Expect(os.WriteFile(filepath.Join(checkout, "source.txt"), []byte("local edits\n"), 0644)).To(Succeed())
		refs, index := git(checkout, "show-ref"), git(checkout, "ls-files", "--stage")
		_, err := blameGit(context.Background(), checkout, "cat-file", "-e", head+"^{commit}")
		Expect(err).To(HaveOccurred())
		client := New(Repository{Organization: "acme", Project: "product", Name: "service"})
		var metadata repositoryInfo
		Expect(json.Unmarshal([]byte(fixtureRepository), &metadata)).To(Succeed())
		client.info = &metadata
		var pr pullRequest
		Expect(json.Unmarshal([]byte(fmt.Sprintf(`{"pullRequestId":27,"sourceRefName":"refs/heads/feature","targetRefName":"refs/heads/main","lastMergeSourceCommit":{"commitId":%q},"lastMergeTargetCommit":{"commitId":%q},"forkSource":{"repository":{"id":"fork-id","name":"fork","project":{"id":"project-id","name":"product"}}}}`, head, base)), &pr)).To(Succeed())
		Expect(client.blameCommits(context.Background(), blameCheckout{Dir: checkout, Remote: "origin", PR: pr})).To(Succeed())
		Expect(git(checkout, "rev-parse", "HEAD")).To(Equal(base))
		Expect(git(checkout, "show-ref")).To(Equal(refs))
		Expect(git(checkout, "ls-files", "--stage")).To(Equal(index))
		Expect(git(checkout, "cat-file", "-t", head)).To(Equal("commit"))
		content, err := os.ReadFile(filepath.Join(checkout, "source.txt"))
		Expect(err).NotTo(HaveOccurred())
		Expect(string(content)).To(Equal("local edits\n"))
		_, err = os.Stat(filepath.Join(checkout, ".git", "FETCH_HEAD"))
		Expect(os.IsNotExist(err)).To(BeTrue())
	})
})
