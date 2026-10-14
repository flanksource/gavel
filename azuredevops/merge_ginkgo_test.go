package azuredevops

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/flanksource/gavel/pr/model"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Azure PR merge actions", func() {
	DescribeTable("completes the selected PR without bypassing policies or deleting its branch", func(method, strategy string) {
		var payload map[string]any
		c := testClient(func(w http.ResponseWriter, r *http.Request) {
			defer GinkgoRecover()
			if !strings.Contains(r.URL.Path, "/pullrequests/") {
				fmt.Fprint(w, fixtureRepository)
				return
			}
			Expect(r.URL.Path).To(Equal("/acme/product/_apis/git/repositories/repo-id/pullrequests/27"))
			Expect(r.Method).To(Equal(http.MethodPatch))
			Expect(json.NewDecoder(r.Body).Decode(&payload)).To(Succeed())
			fmt.Fprint(w, `{"pullRequestId":27,"status":"completed"}`)
		})
		Expect(c.MergePR(context.Background(), 27, model.MergeInput{Method: method, HeadSHA: "reviewed-head"})).To(Succeed())
		Expect(payload).To(Equal(map[string]any{
			"status":                "completed",
			"lastMergeSourceCommit": map[string]any{"commitId": "reviewed-head"},
			"completionOptions":     map[string]any{"mergeStrategy": strategy, "bypassPolicy": false, "deleteSourceBranch": false},
		}))
	}, Entry("rebase", "rebase", "rebase"), Entry("squash", "squash", "squash"), Entry("merge commit", "merge", "noFastForward"), Entry("normalized method", " REBASE ", "rebase"))

	It("rejects missing revision, unsupported methods and auto-merge before making requests", func() {
		c := testClient(func(http.ResponseWriter, *http.Request) { Fail("invalid action reached Azure") })
		for _, input := range []model.MergeInput{{Method: "rebase"}, {Method: "invalid", HeadSHA: "head"}, {Method: "merge", HeadSHA: "head", Auto: true}} {
			Expect(c.MergePR(context.Background(), 27, input)).NotTo(Succeed())
		}
	})

	It("surfaces branch policy rejection", func() {
		c := testClient(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodGet {
				fmt.Fprint(w, fixtureRepository)
				return
			}
			w.WriteHeader(http.StatusConflict)
			fmt.Fprint(w, `{"message":"Required policy is pending"}`)
		})
		Expect(c.MergePR(context.Background(), 27, model.MergeInput{Method: "rebase", HeadSHA: "head"})).To(MatchError(ContainSubstring("Required policy is pending")))
	})

	It("rejects an unexpected successful response", func() {
		c := testClient(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodGet {
				fmt.Fprint(w, fixtureRepository)
				return
			}
			fmt.Fprint(w, `{"pullRequestId":28,"status":"active"}`)
		})
		Expect(c.MergePR(context.Background(), 27, model.MergeInput{Method: "rebase", HeadSHA: "head"})).NotTo(Succeed())
	})
	It("accepts an asynchronous completion queued by the host", func() {
		c := testClient(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodGet {
				fmt.Fprint(w, fixtureRepository)
				return
			}
			fmt.Fprint(w, `{"pullRequestId":27,"status":"active","mergeStatus":"succeeded","completionQueueTime":"2026-10-13T14:00:00Z"}`)
		})
		Expect(c.MergePR(context.Background(), 27, model.MergeInput{Method: "rebase", HeadSHA: "head"})).To(Succeed())
	})
})
