package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"

	"github.com/flanksource/gavel/prwatch"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Azure PR status targets", func() {
	DescribeTable("accepts hosted PR URLs and repository URLs", func(args []string, repo string) {
		target, err := parseStatusTarget(args)
		Expect(err).NotTo(HaveOccurred())
		Expect(target).To(Equal(statusTarget{Repo: repo, PR: 27}))
	}, Entry("modern", []string{"https://dev.azure.com/acme/product/_git/service/pullrequest/27"}, "https://dev.azure.com/acme/product/_git/service"),
		Entry("legacy", []string{"https://acme.visualstudio.com/product/_git/service/pullrequest/27"}, "https://dev.azure.com/acme/product/_git/service"),
		Entry("repo and number", []string{"https://dev.azure.com/acme/product/_git/service", "27"}, "https://dev.azure.com/acme/product/_git/service"))
	It("rejects unsupported Azure flags before requesting credentials", func() {
		_, err := runPRStatus(PRStatusOptions{Args: []string{"https://dev.azure.com/acme/product/_git/service/pullrequest/27"}, AIFix: true})
		Expect(err).To(MatchError(ContainSubstring("--ai-fix is not supported")))
	})
})

var _ = Describe("Azure PR status output", func() {
	BeforeEach(func() {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/acme/product/_apis/git/repositories/service":
				fmt.Fprint(w, `{"id":"repo-id","name":"service","project":{"id":"project-id","name":"product"}}`)
			case "/acme/product/_apis/git/repositories/repo-id/pullrequests/27":
				fmt.Fprint(w, `{"pullRequestId":27,"title":"Update parser","status":"active","sourceRefName":"refs/heads/feature","targetRefName":"refs/heads/main","mergeStatus":"succeeded","lastMergeSourceCommit":{"commitId":"head-sha"},"createdBy":{"displayName":"Casey Author","uniqueName":"casey@example.com"},"reviewers":[{"id":"alice-id","uniqueName":"alice@example.com","displayName":"Alice","vote":10,"isRequired":true},{"id":"bob-id","displayName":"Bob","vote":0}]}`)
			case "/acme/product/_apis/git/repositories/repo-id/pullrequests/27/threads":
				fmt.Fprint(w, `{"value":[
					{"id":7,"status":"active","comments":[
						{"id":1,"commentType":"text","content":"Check the parser","author":{"displayName":"Alice Reviewer","uniqueName":"alice@example.com"}},
						{"id":2,"parentCommentId":1,"commentType":"text","content":"I added coverage","author":{"uniqueName":"casey@example.com"}}
					]},
					{"id":8,"status":"fixed","threadContext":{"filePath":"/parser.go","rightFileStart":{"line":12}},"comments":[
						{"id":1,"commentType":1,"content":"Clarify the error","author":{"displayName":"Bob"}}
					]}
				]}`)
			case "/acme/product/_apis/git/repositories/repo-id/refs":
				fmt.Fprint(w, `{"value":[{"name":"refs/heads/feature","objectId":"head-sha"}]}`)
			case "/acme/product/_apis/policy/evaluations", "/acme/product/_apis/build/builds":
				fmt.Fprint(w, `{"value":[]}`)
			default:
				http.NotFound(w, r)
			}
		}))
		DeferCleanup(server.Close)
		endpoint, err := url.Parse(server.URL)
		Expect(err).NotTo(HaveOccurred())
		original := http.DefaultTransport
		http.DefaultTransport = azureTestTransport{server: endpoint, transport: original}
		DeferCleanup(func() { http.DefaultTransport = original })
		previous := exitCode
		DeferCleanup(func() { exitCode = previous })
		GinkgoT().Setenv("AZURE_DEVOPS_EXT_PAT", "fixture-pat")
	})

	status := func() *prwatch.PRWatchResult {
		value, err := runPRStatus(PRStatusOptions{Args: []string{"https://dev.azure.com/acme/product/_git/service/pullrequest/27"}, Interval: "1s"})
		Expect(err).NotTo(HaveOccurred())
		Expect(value).To(BeAssignableToTypeOf(&prwatch.PRWatchResult{}))
		Expect(exitCode).To(BeZero())
		return value.(*prwatch.PRWatchResult)
	}

	It("shows the PR author and attributes discussion comments and replies by default", func() {
		result := status()
		Expect(result.Pretty().String()).To(And(
			ContainSubstring("| Casey Author |"),
			ContainSubstring("Alice Reviewer: Check the parser"),
			ContainSubstring("casey@example.com: I added coverage"),
			ContainSubstring("Bob: Clarify the error (resolved)"),
		))
		Expect(result.Comments).To(HaveLen(3))
		Expect(result.Comments[2].Line).To(Equal(12))
		data, err := json.Marshal(result.Comments[0])
		Expect(err).NotTo(HaveOccurred())
		Expect(string(data)).To(And(ContainSubstring(`"author":"alice@example.com"`), ContainSubstring(`"authorName":"Alice Reviewer"`)))
	})

	It("shows the approval and individual reviewer votes in terminal and JSON output", func() {
		result := status()
		Expect(result.Pretty().String()).To(And(ContainSubstring("Review: APPROVED"), ContainSubstring("Alice (required): Approved"), ContainSubstring("Bob: No vote")))
		data, err := json.Marshal(result)
		Expect(err).NotTo(HaveOccurred())
		var output struct {
			PR struct {
				ReviewDecision string          `json:"reviewDecision"`
				Reviewers      json.RawMessage `json:"reviewers"`
			} `json:"pr"`
		}
		Expect(json.Unmarshal(data, &output)).To(Succeed())
		Expect(output.PR.ReviewDecision).To(Equal("APPROVED"))
		Expect(output.PR.Reviewers).To(MatchJSON(`[
			{"id":"alice-id","login":"alice@example.com","name":"Alice","state":"APPROVED","required":true},
			{"id":"bob-id","login":"","name":"Bob","state":"NO_VOTE","required":false}
		]`))
	})
})
