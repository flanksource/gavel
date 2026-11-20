// Tests the PR client of a repository a git connection reaches: status with merge readiness and
// resolvable threads, resolving a thread once, merging the reviewed head, and opening a PR on the
// default branch it was given.
package provider

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"

	dbcontext "github.com/flanksource/commons-db/context"
	"github.com/flanksource/commons-db/git"
	"github.com/flanksource/commons-db/models"
	"github.com/flanksource/gavel/pr/model"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("A git connection's PR client", func() {
	const head = "1111111111111111111111111111111111111111"
	const pulls = "/acme/Platform/_apis/git/repositories/rules/pullrequests"
	var (
		mu    sync.Mutex
		calls []string
		sent  map[string]map[string]any
	)
	serve := func(routes map[string]string) Client {
		calls, sent = nil, map[string]map[string]any{}
		server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := r.Method + " " + r.URL.Path
			mu.Lock()
			calls = append(calls, key)
			var body map[string]any
			if json.NewDecoder(r.Body).Decode(&body) == nil {
				sent[key] = body
			}
			mu.Unlock()
			reply, ok := routes[key]
			if !ok {
				http.Error(w, "no route "+key, http.StatusNotFound)
				return
			}
			_, _ = w.Write([]byte(reply))
		}))
		DeferCleanup(server.Close)
		client, err := git.FromConnection(dbcontext.New(), models.Connection{Name: "acme", Type: models.ConnectionTypeAzureDevops,
			URL: server.URL + "/acme", Password: "azure-pat", InsecureTLS: true})
		Expect(err).NotTo(HaveOccurred())
		remote, err := client.Repository(server.URL + "/acme/Platform/_git/rules")
		Expect(err).NotTo(HaveOccurred())
		return NewConnectionClient(remote, "main")
	}
	pull := `{"pullRequestId":9,"status":"active","title":"Tighten","mergeStatus":"succeeded","sourceRefName":"refs/heads/tighten",
		"targetRefName":"refs/heads/main","lastMergeSourceCommit":{"commitId":"` + head + `"},"repository":{"id":"rules-id","name":"rules","project":{"name":"Platform"}}}`
	statusRoutes := map[string]string{
		"GET " + pulls + "/9":                                     pull,
		"GET /acme/_apis/projects/Platform":                       `{"id":"platform-id"}`,
		"GET /acme/Platform/_apis/policy/evaluations":             `{"value":[]}`,
		"GET " + pulls + "/9/reviewers":                           `{"value":[]}`,
		"GET /acme/Platform/_apis/git/repositories/rules-id/refs": `{"value":[{"name":"refs/heads/tighten","objectId":"` + head + `"}]}`,
		"GET " + pulls + "/9/threads":                             `{"value":[{"id":5,"status":"active","comments":[{"id":1,"content":"Hm.","commentType":"text","author":{"displayName":"Grace Hopper"}}]}]}`,
		"GET /acme/Platform/_apis/build/builds":                   `{"value":[]}`,
	}

	It("names its kind and reads a pull request with its readiness and resolvable threads", func(ctx SpecContext) {
		client := serve(statusRoutes)
		Expect(client.Kind()).To(Equal("azuredevops"))
		status, err := client.Status(ctx, 9, model.StatusOptions{Threads: true})
		Expect(err).NotTo(HaveOccurred())
		Expect(status.PR.Title).To(Equal("Tighten"))
		Expect(status.PR.State).To(Equal("OPEN"))
		Expect(status.PR.MergeReadiness).To(Equal(&model.MergeReadiness{State: "ready"}))
		Expect(status.PR.ReviewThreads).To(HaveLen(1))
		Expect(status.PR.ReviewThreads[0].CanResolve).To(BeTrue())
	})

	It("resolves an open thread, and leaves a resolved one alone", func(ctx SpecContext) {
		client := serve(map[string]string{
			"GET " + pulls + "/9/threads/5":   `{"id":5,"status":"active","comments":[{"id":1,"content":"Hm.","commentType":"text","author":{"displayName":"Grace Hopper"}}]}`,
			"PATCH " + pulls + "/9/threads/5": `{"id":5,"status":"fixed","comments":[{"id":1,"content":"Hm.","commentType":"text","author":{"displayName":"Grace Hopper"}}]}`,
			"GET " + pulls + "/9/threads/6":   `{"id":6,"status":"closed","comments":[{"id":1,"content":"Ok.","commentType":"text","author":{"displayName":"Grace Hopper"}}]}`,
		})
		Expect(client.ResolveThread(ctx, 9, "5")).To(Succeed())
		Expect(client.ResolveThread(ctx, 9, "6")).To(Succeed())
		Expect(calls).To(Equal([]string{"GET " + pulls + "/9/threads/5", "PATCH " + pulls + "/9/threads/5", "GET " + pulls + "/9/threads/6"}))
	})

	It("merges the reviewed head, and refuses auto-merge", func(ctx SpecContext) {
		client := serve(map[string]string{"PATCH " + pulls + "/9": `{"pullRequestId":9,"status":"completed"}`})
		Expect(client.MergePR(ctx, 9, model.MergeInput{Method: "squash", HeadSHA: head})).To(Succeed())
		Expect(sent["PATCH "+pulls+"/9"]["lastMergeSourceCommit"]).To(Equal(map[string]any{"commitId": head}))
		Expect(client.MergePR(ctx, 9, model.MergeInput{Method: "squash", HeadSHA: head, Auto: true})).
			To(MatchError(ContainSubstring("auto-merge isn't offered for azuredevops pull requests")))
	})

	It("opens a pull request onto the default branch it was given", func(ctx SpecContext) {
		client := serve(map[string]string{"POST " + pulls: `{"pullRequestId":12,"status":"active","title":"Land run","sourceRefName":"refs/heads/run",
			"targetRefName":"refs/heads/main","repository":{"name":"rules","project":{"name":"Platform"}}}`})
		branch, err := client.DefaultBranch(ctx)
		Expect(err).NotTo(HaveOccurred())
		Expect(branch).To(Equal("main"))
		created, err := client.CreatePR(ctx, model.CreatePRInput{Title: "Land run", Head: "run"})
		Expect(err).NotTo(HaveOccurred())
		Expect(created.Number).To(Equal(12))
		Expect(created.Base).To(Equal("main"))
		Expect(sent["POST "+pulls]["targetRefName"]).To(Equal("refs/heads/main"))
	})
})
