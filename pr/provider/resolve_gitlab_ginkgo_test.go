// Tests the PR client a GitLab repository reference or checkout resolves to: gitlab.com by its
// host, a self-managed GitLab when a gitlab connection names it, over HTTPS or SSH.
package provider

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os/exec"

	"github.com/flanksource/commons-db/git"
	"github.com/flanksource/commons-db/models"
	"github.com/flanksource/gavel/gitlab"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("A GitLab repository's PR client", func() {
	owner := func(connection *models.Connection) {
		previous := gitlab.FindConnection
		gitlab.FindConnection = func(_ context.Context, repo git.RepoURL) (*models.Connection, error) {
			if connection != nil && (repo.Host == "gitlab.example.com" || repo.Host == "127.0.0.1") {
				return connection, nil
			}
			return nil, nil
		}
		DeferCleanup(func() { gitlab.FindConnection = previous })
	}
	checkout := func(origin string) string {
		dir := GinkgoT().TempDir()
		for _, args := range [][]string{{"init", "-q"}, {"remote", "add", "origin", origin}} {
			out, err := exec.Command("git", append([]string{"-C", dir}, args...)...).CombinedOutput()
			Expect(err).NotTo(HaveOccurred(), string(out))
		}
		return dir
	}

	DescribeTable("resolves a GitLab repository reference", func(ref string) {
		GinkgoT().Setenv("GITLAB_TOKEN", "glpat-env")
		owner(&models.Connection{Name: "lab", Type: models.ConnectionTypeGitlab, URL: "https://gitlab.example.com", Password: "glpat"})
		client, err := Resolve(Options{Repo: ref})
		Expect(err).NotTo(HaveOccurred())
		Expect(client.Kind()).To(Equal("gitlab"))
		Expect(client.(connectionClient).remote.URL.FullName()).To(Equal("acme/platform/api"))
	}, Entry("gitlab.com HTTPS", "https://gitlab.com/acme/platform/api.git"),
		Entry("gitlab.com SSH", "git@gitlab.com:acme/platform/api.git"),
		Entry("self-managed HTTPS a connection names", "https://gitlab.example.com/acme/platform/api"),
		Entry("self-managed SSH a connection names", "git@gitlab.example.com:acme/platform/api.git"))

	It("refuses a self-managed host no gitlab connection names, as no GitHub repository either", func() {
		owner(nil)
		_, err := Resolve(Options{Repo: "git@gitlab.example.com:acme/platform/api.git"})
		Expect(err).To(MatchError("repository git@gitlab.example.com:acme/platform/api.git is on gitlab.example.com, " +
			"which is neither github.com, gitlab.com, Azure DevOps nor a host a gitlab connection names"))
	})

	It("reads the default branch of a checkout whose SSH origin a gitlab connection names", func(ctx SpecContext) {
		server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.EscapedPath() != "/api/v4/projects/acme%2Fapi" {
				http.Error(w, "no route "+r.URL.EscapedPath(), http.StatusNotFound)
				return
			}
			_, _ = w.Write([]byte(`{"id":1,"path":"api","path_with_namespace":"acme/api","default_branch":"trunk"}`))
		}))
		DeferCleanup(server.Close)
		owner(&models.Connection{Name: "lab", Type: models.ConnectionTypeGitlab, URL: server.URL, Password: "glpat", InsecureTLS: true})
		parsed, err := url.Parse(server.URL)
		Expect(err).NotTo(HaveOccurred())

		client, err := Resolve(Options{WorkDir: checkout("git@" + parsed.Hostname() + ":acme/api.git"), Context: ctx})
		Expect(err).NotTo(HaveOccurred())
		Expect(client.Kind()).To(Equal("gitlab"))
		Expect(client.DefaultBranch(ctx)).To(Equal("trunk"))
	})

	It("refuses opening a merge request on another project than the origin pushes go to", func(ctx SpecContext) {
		GinkgoT().Setenv("GITLAB_TOKEN", "glpat-env")
		owner(nil)
		err := PreflightCreate(ctx, Options{WorkDir: checkout("git@gitlab.com:acme/api.git"), Repo: "https://gitlab.com/acme/web"})
		Expect(err).To(MatchError("GitLab PR repository does not match origin; pushes target origin"))
	})
})
