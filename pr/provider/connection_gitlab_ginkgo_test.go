// Tests the PR client of a GitLab project a git connection reaches: a merge request's status with
// its readiness and its latest pipeline's jobs in the dashboard's states, and the default branch
// its project names.
package provider

import (
	"context"
	"net/http"
	"net/http/httptest"

	dbcontext "github.com/flanksource/commons-db/context"
	"github.com/flanksource/commons-db/git"
	"github.com/flanksource/commons-db/models"
	"github.com/flanksource/gavel/gitlab"
	"github.com/flanksource/gavel/pr/model"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("A GitLab connection's PR client", func() {
	const head = "2222222222222222222222222222222222222222"
	serve := func(routes map[string]string) Client {
		server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			reply, ok := routes[r.Method+" "+r.URL.EscapedPath()]
			if !ok {
				http.Error(w, "no route "+r.Method+" "+r.URL.EscapedPath(), http.StatusNotFound)
				return
			}
			_, _ = w.Write([]byte(reply))
		}))
		DeferCleanup(server.Close)
		client, err := git.FromConnection(dbcontext.New(), models.Connection{Name: "acme", Type: models.ConnectionTypeGitlab,
			URL: server.URL, Password: "glpat", InsecureTLS: true})
		Expect(err).NotTo(HaveOccurred())
		remote, err := client.Repository(server.URL + "/acme/api.git")
		Expect(err).NotTo(HaveOccurred())
		return NewConnectionClient(remote, "")
	}
	mr := `{"iid":5,"state":"opened","title":"Speed up the API","source_branch":"fast","target_branch":"main","sha":"` + head + `",
		"source_project_id":1,"target_project_id":1,"detailed_merge_status":"ci_must_pass","author":{"name":"Grace Hopper"}}`
	job := func(id, name, status string) string {
		return `{"id":` + id + `,"name":"` + name + `","status":"` + status + `","web_url":"https://gitlab.example.com/acme/api/-/jobs/` + id + `"}`
	}

	It("reads a merge request with its readiness and its latest pipeline's jobs", func(ctx SpecContext) {
		client := serve(map[string]string{
			"GET /api/v4/projects/acme%2Fapi/merge_requests/5":           mr,
			"GET /api/v4/projects/acme%2Fapi/merge_requests/5/pipelines": `[{"id":102,"project_id":1,"sha":"` + head + `","ref":"fast","status":"failed"}]`,
			"GET /api/v4/projects/1/pipelines/102/jobs": `[` + job("7001", "build", "success") + `,` + job("7002", "test", "failed") + `,` +
				job("7003", "deploy", "manual") + `,` + job("7004", "lint", "running") + `,` + job("7005", "docs", "created") + `]`,
			"GET /api/v4/projects/acme%2Fapi/jobs/7002/trace": "go test ./...\n--- FAIL: TestSpeed\nFAIL\n",
		})
		Expect(client.Kind()).To(Equal("gitlab"))
		status, err := client.Status(ctx, 5, model.StatusOptions{Logs: true, TailLogs: 2})
		Expect(err).NotTo(HaveOccurred())
		Expect(status.PR.Title).To(Equal("Speed up the API"))
		Expect(status.PR.HeadRefOID).To(Equal(head))
		Expect(status.PR.MergeReadiness).To(Equal(&model.MergeReadiness{State: "blocked", Reasons: []string{"pipeline must succeed"}}))
		Expect(status.Runs).To(HaveKey(int64(102)))
		run := status.Runs[102]
		Expect(run.Status).To(Equal("COMPLETED"))
		Expect(run.Conclusion).To(Equal("FAILURE"))
		states := map[string][2]string{}
		for _, job := range run.Jobs {
			states[job.Name] = [2]string{job.Status, job.Conclusion}
		}
		Expect(states).To(Equal(map[string][2]string{
			"build": {"COMPLETED", "SUCCESS"}, "test": {"COMPLETED", "FAILURE"}, "deploy": {"WAITING", ""},
			"lint": {"IN_PROGRESS", ""}, "docs": {"QUEUED", ""},
		}))
		Expect(run.Jobs[1].Logs).To(Equal("--- FAIL: TestSpeed\nFAIL"), "a failed job's own log, without steps")
		Expect(status.PR.StatusCheckRollup).To(HaveLen(5))
	})

	It("refuses a pipeline state it doesn't know instead of marking it successful", func(ctx SpecContext) {
		client := serve(map[string]string{
			"GET /api/v4/projects/acme%2Fapi/merge_requests/5":           mr,
			"GET /api/v4/projects/acme%2Fapi/merge_requests/5/pipelines": `[{"id":102,"project_id":1,"sha":"` + head + `","status":"exploded"}]`,
			"GET /api/v4/projects/1/pipelines/102/jobs":                  `[]`,
		})
		_, err := client.Status(ctx, 5, model.StatusOptions{})
		Expect(err).To(MatchError(`pipeline execution 102: unexpected GitLab pipeline or job status "exploded"`))
	})

	It("is what a local gavel resolves a merge request's URL to, through the connection that owns it", func(ctx SpecContext) {
		var server *httptest.Server
		server = httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.EscapedPath() {
			case "/api/v4/projects/acme%2Fapi/merge_requests/5":
				_, _ = w.Write([]byte(mr))
			case "/api/v4/projects/acme%2Fapi/merge_requests/5/pipelines":
				_, _ = w.Write([]byte(`[]`))
			default:
				http.NotFound(w, r)
			}
		}))
		DeferCleanup(server.Close)
		previous := gitlab.FindConnection
		gitlab.FindConnection = func(context.Context, git.RepoURL) (*models.Connection, error) {
			return &models.Connection{Name: "acme", Type: models.ConnectionTypeGitlab, URL: server.URL, Password: "glpat", InsecureTLS: true}, nil
		}
		DeferCleanup(func() { gitlab.FindConnection = previous })

		client, number, err := ResolveURL(Options{Context: ctx}, server.URL+"/acme/api/-/merge_requests/5")
		Expect(err).NotTo(HaveOccurred())
		Expect(number).To(Equal(5))
		Expect(client.Kind()).To(Equal("gitlab"))
		status, err := client.Status(ctx, number, model.StatusOptions{})
		Expect(err).NotTo(HaveOccurred())
		Expect(status.PR.Title).To(Equal("Speed up the API"))
	})

	It("reaches a gitlab.com project locally through GITLAB_TOKEN when no connection owns it", func(ctx SpecContext) {
		previous := gitlab.FindConnection
		gitlab.FindConnection = nil
		DeferCleanup(func() { gitlab.FindConnection = previous })
		GinkgoT().Setenv("GITLAB_TOKEN", "env-token")
		remote, err := LocalRemote(ctx, models.ConnectionTypeGitlab, "https://gitlab.com/acme/platform/api")
		Expect(err).NotTo(HaveOccurred())
		Expect(remote.Provider.Type()).To(Equal(models.ConnectionTypeGitlab))
		Expect(remote.URL.Owner).To(Equal("acme/platform"))
		_, err = LocalRemote(ctx, models.ConnectionTypeGithub, "https://github.com/acme/api")
		Expect(err).To(MatchError(ContainSubstring("a local gavel reaches no github repository like")))
	})

	It("names the default branch its project does when it was given none", func(ctx SpecContext) {
		client := serve(map[string]string{"GET /api/v4/projects/acme%2Fapi": `{"id":1,"path":"api","path_with_namespace":"acme/api","default_branch":"trunk"}`})
		branch, err := client.DefaultBranch(ctx)
		Expect(err).NotTo(HaveOccurred())
		Expect(branch).To(Equal("trunk"))
	})
})
