package github

import (
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func initRepoWithRemotes(dir string, remotes map[string]string) {
	repo, err := git.PlainInit(dir, false)
	Expect(err).NotTo(HaveOccurred())
	for name, url := range remotes {
		_, err := repo.CreateRemote(&config.RemoteConfig{Name: name, URLs: []string{url}})
		Expect(err).NotTo(HaveOccurred())
	}
}

var _ = Describe("ReposFromRemotes", func() {
	It("lists the upstream repo before origin and skips non-GitHub remotes", func() {
		dir := GinkgoT().TempDir()
		initRepoWithRemotes(dir, map[string]string{
			"origin":   "git@github.com:acme-fork/widget.git",
			"upstream": "https://github.com/acme/widget.git",
			"mirror":   "ssh://git.example.com:2222/widget",
		})

		Expect(ReposFromRemotes(dir)).To(Equal([]string{"acme/widget", "acme-fork/widget"}))
	})

	It("ignores a non-GitHub origin", func() {
		dir := GinkgoT().TempDir()
		initRepoWithRemotes(dir, map[string]string{"origin": "ssh://git.example.com:2222/widget"})

		Expect(ReposFromRemotes(dir)).To(BeEmpty())
	})

	It("reports no repos for a directory outside any git repository", func() {
		Expect(ReposFromRemotes(GinkgoT().TempDir())).To(BeEmpty())
	})
})
