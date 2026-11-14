package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/flanksource/gavel/pr/blame"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

type prBlameTransport struct {
	server    *url.URL
	transport http.RoundTripper
}

func (t prBlameTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	clone := request.Clone(request.Context())
	endpoint := *request.URL
	endpoint.Scheme, endpoint.Host = t.server.Scheme, t.server.Host
	clone.URL = &endpoint
	return t.transport.RoundTrip(clone)
}

var _ = Describe("gavel pr blame Azure targets", func() {
	It("routes origin, explicit repository, modern and legacy PR URLs through the ADO provider", func() {
		dir := GinkgoT().TempDir()
		git := func(args ...string) string {
			cmd := exec.Command("git", args...)
			cmd.Dir = dir
			out, err := cmd.CombinedOutput()
			Expect(err).NotTo(HaveOccurred(), "%s", out)
			return strings.TrimSpace(string(out))
		}
		git("init", "-q", "-b", "feature")
		git("config", "commit.gpgsign", "false")
		git("remote", "add", "origin", "https://dev.azure.com/acme/product/_git/service")
		Expect(os.WriteFile(filepath.Join(dir, "source.txt"), []byte("before\n"), 0644)).To(Succeed())
		git("add", "--", "source.txt")
		git("-c", "user.name=Alice", "-c", "user.email=alice@example.com", "commit", "-qm", "base")
		base := git("rev-parse", "HEAD")
		Expect(os.WriteFile(filepath.Join(dir, "source.txt"), []byte("after\n"), 0644)).To(Succeed())
		git("add", "--", "source.txt")
		git("-c", "user.name=Author", "-c", "user.email=author@example.com", "commit", "-qm", "change")
		head := git("rev-parse", "HEAD")
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer GinkgoRecover()
			Expect(r.Method).To(Equal(http.MethodGet))
			Expect(r.Header.Get("Authorization")).To(HavePrefix("Basic "))
			switch {
			case strings.HasSuffix(r.URL.Path, "/pullrequests/27"):
				fmt.Fprintf(w, `{"pullRequestId":27,"title":"Update","status":"active","sourceRefName":"refs/heads/feature","targetRefName":"refs/heads/main","lastMergeSourceCommit":{"commitId":%q},"lastMergeTargetCommit":{"commitId":%q},"createdBy":{"id":"author-id","uniqueName":"author@example.com"}}`, head, base)
			case strings.HasSuffix(r.URL.Path, "/pullrequests"):
				fmt.Fprint(w, `{"value":[{"pullRequestId":27}]}`)
			case strings.HasSuffix(r.URL.Path, "/refs"):
				if r.URL.Query().Get("filter") == "heads/feature" {
					fmt.Fprintf(w, `{"value":[{"name":"refs/heads/feature","objectId":%q}]}`, head)
				} else {
					fmt.Fprintf(w, `{"value":[{"name":"refs/heads/main","objectId":%q}]}`, base)
				}
			case strings.HasSuffix(r.URL.Path, "/evaluations"):
				fmt.Fprint(w, `{"value":[]}`)
			default:
				Expect(r.URL.Path).To(Equal("/acme/product/_apis/git/repositories/service"))
				fmt.Fprint(w, `{"id":"repo-id","name":"service","project":{"id":"project-id","name":"product"}}`)
			}
		}))
		DeferCleanup(server.Close)
		endpoint, err := url.Parse(server.URL)
		Expect(err).NotTo(HaveOccurred())
		original := http.DefaultTransport
		http.DefaultTransport = prBlameTransport{server: endpoint, transport: original}
		DeferCleanup(func() { http.DefaultTransport = original })
		previous := workingDir
		workingDir = dir
		DeferCleanup(func() { workingDir = previous })
		GinkgoT().Setenv("AZURE_DEVOPS_EXT_PAT", "fixture-pat")
		for _, selection := range []struct {
			args []string
			repo string
		}{
			{nil, ""}, {[]string{"27"}, ""},
			{[]string{"27"}, "https://dev.azure.com/acme/product/_git/service"},
			{[]string{"https://dev.azure.com/acme/product/_git/service/pullrequest/27"}, ""},
			{[]string{"https://acme.visualstudio.com/product/_git/service/pullrequest/27"}, ""},
		} {
			value, err := runPRBlame(PRBlameOptions{Args: selection.args, Repo: selection.repo, Limit: 5, HistoryLimit: 100})
			Expect(err).NotTo(HaveOccurred())
			result := value.(*blame.Recommendations)
			Expect(result.Provider).To(Equal("azuredevops"))
			Expect(result.Reviewers).To(HaveLen(1))
			Expect(result.Reviewers[0].Email).To(Equal("alice@example.com"))
			Expect(result.Reviewers[0].Lines).To(Equal(1))
		}
		value, err := runPRBlame(PRBlameOptions{Args: []string{"27"}, Limit: 5, HistoryLimit: 100, Since: "2100-01-01"})
		Expect(err).NotTo(HaveOccurred())
		result := value.(*blame.Recommendations)
		Expect(result.Reviewers).To(BeEmpty())
		Expect(result.Since).NotTo(BeNil())
		Expect(result.Since.Format("2006-01-02")).To(Equal("2100-01-01"))
	})
})
