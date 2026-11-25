package commit

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

type azurePushTransport http.HandlerFunc

func (transport azurePushTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	recorder := httptest.NewRecorder()
	transport(recorder, r)
	return recorder.Result(), nil
}

func mockAzurePush(handler http.HandlerFunc) {
	GinkgoT().Setenv("AZURE_DEVOPS_EXT_PAT", "fixture-pat")
	previous := http.DefaultTransport
	http.DefaultTransport = azurePushTransport(handler)
	DeferCleanup(func() { http.DefaultTransport = previous })
}

var _ = Describe("Azure commit push validation", func() {
	It("surfaces Azure preflight errors before changing staged files or HEAD", func() {
		mockAzurePush(func(w http.ResponseWriter, r *http.Request) {
			Expect(r.Method).To(Equal(http.MethodGet))
			w.WriteHeader(http.StatusForbidden)
		})
		repo := GinkgoT().TempDir()
		runGit(repo, "init", "-b", "feature")
		runGit(repo, "commit", "--allow-empty", "-m", "root")
		runGit(repo, "remote", "add", "origin", "https://dev.azure.com/acme/product/_git/service")
		Expect(os.WriteFile(filepath.Join(repo, "feature.txt"), []byte("feature"), 0600)).To(Succeed())
		runGit(repo, "add", "feature.txt")
		before, err := exec.Command("git", "-C", repo, "status", "--porcelain=v1").Output()
		Expect(err).NotTo(HaveOccurred())
		_, err = Run(context.Background(), Options{WorkDir: repo, Push: true, AutoMerge: true, Stage: StageStaged, Message: "feat: feature"})
		Expect(err).To(MatchError(ContainSubstring("HTTP 403")))
		Expect(os.ReadFile(filepath.Join(repo, "feature.txt"))).To(Equal([]byte("feature")))
		after, err := exec.Command("git", "-C", repo, "status", "--porcelain=v1").Output()
		Expect(err).NotTo(HaveOccurred())
		Expect(after).To(Equal(before))
		head, err := exec.Command("git", "-C", repo, "log", "-1", "--format=%s").Output()
		Expect(err).NotTo(HaveOccurred())
		Expect(string(head)).To(Equal("root\n"))
	})

	It("enables auto-complete on the newly created PR in the selected Azure remote", func(ctx SpecContext) {
		const repoPath = "/acme/product/_apis/git/repositories/"
		const actor = "342bacfe-06de-43fc-83a4-b9732983482c"
		var requests []string
		mockAzurePush(func(w http.ResponseWriter, r *http.Request) {
			request := r.Method + " " + r.URL.Path
			requests = append(requests, request)
			switch request {
			case "GET " + repoPath + "service":
				fmt.Fprint(w, `{"id":"repo-id","name":"service","project":{"id":"project-id","name":"product"}}`)
			case "GET " + repoPath + "repo-id/pullrequests":
				fmt.Fprint(w, `{"value":[]}`)
			case "POST " + repoPath + "repo-id/pullrequests":
				fmt.Fprint(w, `{"pullRequestId":27,"status":"active"}`)
			case "GET /acme/_apis/connectionData":
				fmt.Fprintf(w, `{"authenticatedUser":{"id":%q}}`, actor)
			case "PATCH " + repoPath + "repo-id/pullrequests/27":
				var payload map[string]any
				Expect(json.NewDecoder(r.Body).Decode(&payload)).To(Succeed())
				Expect(payload).To(Equal(map[string]any{
					"autoCompleteSetBy": map[string]any{"id": actor},
					"completionOptions": map[string]any{"mergeStrategy": "squash", "bypassPolicy": false, "deleteSourceBranch": false},
				}))
				fmt.Fprintf(w, `{"pullRequestId":27,"status":"active","autoCompleteSetBy":{"id":%q}}`, actor)
			default:
				Fail("unexpected request: " + request)
			}
		})
		repo := GinkgoT().TempDir()
		runGit(repo, "init", "-b", "main")
		runGit(repo, "commit", "--allow-empty", "-m", "root")
		runGit(repo, "remote", "add", "origin", "https://github.com/acme/service.git")
		runGit(repo, "remote", "add", "zim", "https://dev.azure.com/acme/product/_git/service")
		runGit(repo, "update-ref", "refs/remotes/zim/main", "HEAD")
		runGit(repo, "checkout", "-b", "feature")
		runGit(repo, "commit", "--allow-empty", "-m", "feat: change")
		opts := Options{WorkDir: repo, Push: true, Target: "zim/main:feature", AutoMerge: true, MergeType: "squash"}
		Expect(opts.resolveTarget()).To(Succeed())
		deps, err := providerPushDeps(ctx, opts)
		Expect(err).NotTo(HaveOccurred())
		deps.rebaseOnto = func(_, _ string) error { return nil }
		deps.gitPush = func(_, _ string) error { return nil }
		deps.generatePRPrompt = func(context.Context, PRContentInput) (PRContent, error) {
			return PRContent{Title: "feat: change"}, nil
		}
		Expect(pushWithDeps(ctx, opts, &Result{}, deps)).To(Succeed())
		Expect(requests).To(Equal([]string{
			"GET " + repoPath + "service", "GET " + repoPath + "repo-id/pullrequests", "POST " + repoPath + "repo-id/pullrequests",
			"GET /acme/_apis/connectionData", "PATCH " + repoPath + "repo-id/pullrequests/27",
		}))
	})
})
