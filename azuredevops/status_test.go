package azuredevops

import (
	"context"
	"fmt"
	"github.com/flanksource/gavel/pr/model"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"net/http"
	"strings"
)

const fixturePR = `{"pullRequestId":27,"title":"Feature","status":"active","sourceRefName":"refs/heads/feature","targetRefName":"refs/heads/main","mergeStatus":"succeeded","lastMergeSourceCommit":{"commitId":"head-sha"},"lastMergeCommit":{"commitId":"merge-sha"}}`
const fixtureBuilds = `{"value":[{"id":101,"status":"completed","result":"failed","sourceBranch":"refs/pull/27/merge","sourceVersion":"merge-sha","definition":{"id":8,"name":"CI"}},{"id":99,"status":"completed","result":"succeeded","sourceBranch":"refs/pull/27/merge","sourceVersion":"old-sha","definition":{"id":8,"name":"CI"}}]}`
const fixtureTimeline = `{"records":[{"id":"job-uuid","name":"Test","type":"Job","state":"completed","result":"failed","order":1},{"id":"task-uuid","parentId":"job-uuid","name":"Run tests","type":"Task","state":"completed","result":"failed","order":2,"log":{"id":7}}]}`

var _ = Describe("pipeline status", func() {
	It("rejects policy evaluations missing their blocking configuration", func() {
		c := testClient(func(w http.ResponseWriter, r *http.Request) {
			switch {
			case strings.HasSuffix(r.URL.Path, "/pullrequests/27"):
				fmt.Fprint(w, fixturePR)
			case strings.HasSuffix(r.URL.Path, "/evaluations"):
				fmt.Fprint(w, `{"value":[{"evaluationId":"policy-id","status":"approved"}]}`)
			case strings.HasSuffix(r.URL.Path, "/refs"):
				fmt.Fprint(w, `{"value":[{"name":"refs/heads/feature","objectId":"head-sha"}]}`)
			case strings.HasSuffix(r.URL.Path, "/builds"):
				fmt.Fprint(w, `{"value":[]}`)
			default:
				fmt.Fprint(w, fixtureRepository)
			}
		})
		_, err := c.Status(context.Background(), 27, model.StatusOptions{})
		Expect(err).To(MatchError(ContainSubstring("policy configuration")))
	})
	It("queries a fork source repository instead of the target branch with the same name", func() {
		c := testClient(func(w http.ResponseWriter, r *http.Request) {
			defer GinkgoRecover()
			switch {
			case strings.HasSuffix(r.URL.Path, "/pullrequests/27"):
				fmt.Fprint(w, strings.TrimSuffix(fixturePR, "}")+`,"forkSource":{"repository":{"id":"fork-id","project":{"name":"fork-project"}}}}`)
			case strings.HasSuffix(r.URL.Path, "/refs"):
				Expect(r.URL.Path).To(Equal("/acme/fork-project/_apis/git/repositories/fork-id/refs"))
				fmt.Fprint(w, `{"value":[{"name":"refs/heads/feature","objectId":"head-sha"}]}`)
			case strings.HasSuffix(r.URL.Path, "/builds"):
				if r.URL.Query().Get("branchName") == "refs/heads/feature" {
					Expect(r.URL.Path).To(Equal("/acme/fork-project/_apis/build/builds"))
					Expect(r.URL.Query().Get("repositoryId")).To(Equal("fork-id"))
				}
				fmt.Fprint(w, `{"value":[]}`)
			case strings.HasSuffix(r.URL.Path, "/evaluations"):
				fmt.Fprint(w, `{"value":[]}`)
			default:
				fmt.Fprint(w, fixtureRepository)
			}
		})
		result, err := c.Status(context.Background(), 27, model.StatusOptions{})
		Expect(err).NotTo(HaveOccurred())
		Expect(result.PR.MergeReadiness.State).To(Equal("ready"))
	})
	It("excludes a trial merge after the target branch changes", func() {
		c := testClient(func(w http.ResponseWriter, r *http.Request) {
			switch {
			case strings.HasSuffix(r.URL.Path, "/pullrequests/27"):
				fmt.Fprint(w, strings.TrimSuffix(fixturePR, "}")+`,"lastMergeTargetCommit":{"commitId":"old-target"}}`)
			case strings.HasSuffix(r.URL.Path, "/refs"):
				if r.URL.Query().Get("filter") == "heads/main" {
					fmt.Fprint(w, `{"value":[{"name":"refs/heads/main","objectId":"new-target"}]}`)
				} else {
					fmt.Fprint(w, `{"value":[{"name":"refs/heads/feature","objectId":"head-sha"}]}`)
				}
			case strings.HasSuffix(r.URL.Path, "/evaluations"):
				fmt.Fprint(w, `{"value":[]}`)
			case strings.HasSuffix(r.URL.Path, "/builds"):
				fmt.Fprint(w, fixtureBuilds)
			default:
				fmt.Fprint(w, fixtureRepository)
			}
		})
		result, err := c.Status(context.Background(), 27, model.StatusOptions{})
		Expect(err).NotTo(HaveOccurred())
		Expect(result.Runs).To(BeEmpty())
		Expect(result.PR.MergeReadiness.WaitingForMerge).To(BeTrue())
	})
	It("does not reuse an old trial merge while the source branch advances", func() {
		c := testClient(func(w http.ResponseWriter, r *http.Request) {
			switch {
			case strings.HasSuffix(r.URL.Path, "/pullrequests/27"):
				fmt.Fprint(w, fixturePR)
			case strings.HasSuffix(r.URL.Path, "/refs"):
				fmt.Fprint(w, `{"value":[{"name":"refs/heads/feature","objectId":"new-sha"}]}`)
			case strings.HasSuffix(r.URL.Path, "/evaluations"):
				fmt.Fprint(w, `{"value":[]}`)
			case strings.HasSuffix(r.URL.Path, "/builds"):
				fmt.Fprint(w, fixtureBuilds)
			default:
				fmt.Fprint(w, fixtureRepository)
			}
		})
		result, err := c.Status(context.Background(), 27, model.StatusOptions{})
		Expect(err).NotTo(HaveOccurred())
		Expect(result.Runs).To(BeEmpty())
		Expect(result.PR.HeadRefOID).To(Equal("new-sha"))
		Expect(result.PR.MergeReadiness.WaitingForMerge).To(BeTrue())
	})
	DescribeTable("scopes builds to the current revision and loads failed task tails only on request", func(logs bool) {
		logRequests := 0
		c := testClient(func(w http.ResponseWriter, r *http.Request) {
			defer GinkgoRecover()
			switch {
			case strings.HasSuffix(r.URL.Path, "/pullrequests/27"):
				fmt.Fprint(w, fixturePR)
			case strings.HasSuffix(r.URL.Path, "/refs"):
				fmt.Fprint(w, `{"value":[{"name":"refs/heads/feature","objectId":"head-sha"}]}`)
			case strings.HasSuffix(r.URL.Path, "/evaluations"):
				Expect(r.URL.Query().Get("artifactId")).To(Equal("vstfs:///CodeReview/CodeReviewId/project-id/27"))
				fmt.Fprint(w, `{"value":[]}`)
			case strings.HasSuffix(r.URL.Path, "/builds"):
				Expect(r.URL.Query().Get("repositoryId")).To(Equal("repo-id"))
				fmt.Fprint(w, fixtureBuilds)
			case strings.HasSuffix(r.URL.Path, "/timeline"):
				fmt.Fprint(w, fixtureTimeline)
			case strings.HasSuffix(r.URL.Path, "/logs/7"):
				logRequests++
				Expect(r.URL.Query().Get("startLine")).To(Equal("3"))
				fmt.Fprint(w, "old\nline two\nline three\n")
			case strings.HasSuffix(r.URL.Path, "/logs"):
				logRequests++
				fmt.Fprint(w, `{"value":[{"id":7,"lineCount":5}]}`)
			default:
				fmt.Fprint(w, fixtureRepository)
			}
		})
		result, err := c.Status(context.Background(), 27, model.StatusOptions{Logs: logs, TailLogs: 2})
		Expect(err).NotTo(HaveOccurred())
		Expect(result.PR.MergeReadiness.State).To(Equal("ready"))
		Expect(result.Runs).To(HaveLen(1))
		Expect(result.Runs[101].Jobs[0].NativeID).To(Equal("job-uuid"))
		Expect(result.PR.StatusCheckRollup[0].RunID).To(Equal(int64(101)))
		Expect(result.PR.StatusCheckRollup[0].JobID).To(Equal("job-uuid"))
		if logs {
			Expect(result.Runs[101].Jobs[0].Steps[0].Logs).To(Equal("line two\nline three"))
			Expect(logRequests).To(Equal(2))
		} else {
			Expect(logRequests).To(BeZero())
		}
	}, Entry("with logs", true), Entry("without logs", false))
	It("surfaces timeline permission failures", func() {
		c := testClient(func(w http.ResponseWriter, r *http.Request) {
			switch {
			case strings.HasSuffix(r.URL.Path, "/pullrequests/27"):
				fmt.Fprint(w, fixturePR)
			case strings.HasSuffix(r.URL.Path, "/refs"):
				fmt.Fprint(w, `{"value":[{"name":"refs/heads/feature","objectId":"head-sha"}]}`)
			case strings.HasSuffix(r.URL.Path, "/evaluations"):
				fmt.Fprint(w, `{"value":[]}`)
			case strings.HasSuffix(r.URL.Path, "/builds"):
				fmt.Fprint(w, fixtureBuilds)
			case strings.HasSuffix(r.URL.Path, "/timeline"):
				w.WriteHeader(http.StatusForbidden)
			default:
				fmt.Fprint(w, fixtureRepository)
			}
		})
		_, err := c.Status(context.Background(), 27, model.StatusOptions{})
		Expect(err).To(MatchError(ContainSubstring("HTTP 403")))
	})
})
