package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// prCloseStub serves the GraphQL calls `gavel pr close` makes: the by-number PR
// fetch that resolves the node ID, the optional addComment, and the close
// mutation. It records each mutation's variables and the order they arrived in,
// so specs can assert both that the PR was closed and — for the states and
// failures that must stop early — that it was not.
type prCloseStub struct {
	server *httptest.Server
	// commentFails makes addComment return a GraphQL error, so specs can assert
	// a failed comment aborts before the PR is closed.
	commentFails bool

	closeVars    map[string]any
	closeCalls   int
	commentVars  map[string]any
	commentCalls int
	// mutations records mutation names in arrival order; the comment must land
	// on the PR while it is still open.
	mutations []string
}

func newPRCloseStub(prState string) *prCloseStub {
	stub := &prCloseStub{}
	stub.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var payload struct {
			Query     string         `json:"query"`
			Variables map[string]any `json:"variables"`
		}
		Expect(json.Unmarshal(body, &payload)).To(Succeed())

		switch {
		case strings.Contains(payload.Query, "addComment"):
			stub.commentCalls++
			stub.commentVars = payload.Variables
			stub.mutations = append(stub.mutations, "addComment")
			if stub.commentFails {
				_, _ = w.Write([]byte(`{"errors":[{"message":"Resource not accessible by integration"}]}`))
				return
			}
			_, _ = w.Write([]byte(`{"data":{"addComment":{"commentEdge":{"node":{"url":"https://github.com/acme/widgets/pull/7#issuecomment-1"}}}}}`))

		case strings.Contains(payload.Query, "closePullRequest"):
			stub.closeCalls++
			stub.closeVars = payload.Variables
			stub.mutations = append(stub.mutations, "closePullRequest")
			_, _ = w.Write([]byte(`{"data":{"closePullRequest":{"pullRequest":{"number":7,"state":"CLOSED"}}}}`))

		default:
			_, _ = fmt.Fprintf(w, `{"data":{"repository":{"pullRequest":{
				"id":"PR_node_7","number":7,"title":"Add widget cache","state":%q,
				"url":"https://github.com/acme/widgets/pull/7","headRefName":"feat/cache","baseRefName":"main"
			}}}}`, prState)
		}
	}))
	return stub
}

var _ = Describe("gavel pr close", func() {
	var stub *prCloseStub

	newStub := func(state string) {
		stub = newPRCloseStub(state)
		DeferCleanup(stub.server.Close)
		GinkgoT().Setenv("GITHUB_API_URL", stub.server.URL)
		GinkgoT().Setenv("GITHUB_TOKEN", "test-token")
	}

	It("closes an open PR and reports its new state", func() {
		newStub("OPEN")

		result, err := runPRClose(PRCloseOptions{Repo: "acme/widgets", Args: []string{"7"}})
		Expect(err).ToNot(HaveOccurred())

		Expect(stub.closeCalls).To(Equal(1))
		Expect(stub.closeVars["prId"]).To(Equal("PR_node_7"))
		Expect(result).To(Equal(&PRCloseResult{
			Repo:   "acme/widgets",
			Number: 7,
			Title:  "Add widget cache",
			State:  "CLOSED",
			URL:    "https://github.com/acme/widgets/pull/7",
		}))
	})

	It("accepts a full PR URL in place of a number", func() {
		newStub("OPEN")

		_, err := runPRClose(PRCloseOptions{Args: []string{"https://github.com/acme/widgets/pull/7"}})
		Expect(err).ToNot(HaveOccurred())
		Expect(stub.closeCalls).To(Equal(1))
	})

	It("refuses to close a merged PR without calling the mutation", func() {
		newStub("MERGED")

		_, err := runPRClose(PRCloseOptions{Repo: "acme/widgets", Args: []string{"7"}})
		Expect(err).To(MatchError(ContainSubstring("already merged")))
		Expect(stub.closeCalls).To(BeZero())
	})

	It("refuses to close an already-closed PR without calling the mutation", func() {
		newStub("CLOSED")

		_, err := runPRClose(PRCloseOptions{Repo: "acme/widgets", Args: []string{"7"}})
		Expect(err).To(MatchError(ContainSubstring("already closed")))
		Expect(stub.closeCalls).To(BeZero())
	})

	Describe("--comment", func() {
		It("posts the comment before closing and reports it", func() {
			newStub("OPEN")

			result, err := runPRClose(PRCloseOptions{
				Repo:    "acme/widgets",
				Args:    []string{"7"},
				Comment: "superseded by #9",
			})
			Expect(err).ToNot(HaveOccurred())

			Expect(stub.commentVars["subjectId"]).To(Equal("PR_node_7"))
			Expect(stub.commentVars["body"]).To(Equal("superseded by #9"))
			// The comment has to land while the PR is still open.
			Expect(stub.mutations).To(Equal([]string{"addComment", "closePullRequest"}))
			Expect(result.(*PRCloseResult).Comment).To(Equal("superseded by #9"))
		})

		It("does not comment when the flag is unset", func() {
			newStub("OPEN")

			_, err := runPRClose(PRCloseOptions{Repo: "acme/widgets", Args: []string{"7"}})
			Expect(err).ToNot(HaveOccurred())
			Expect(stub.commentCalls).To(BeZero())
			Expect(stub.mutations).To(Equal([]string{"closePullRequest"}))
		})

		It("leaves the PR open when the comment fails", func() {
			newStub("OPEN")
			stub.commentFails = true

			_, err := runPRClose(PRCloseOptions{
				Repo:    "acme/widgets",
				Args:    []string{"7"},
				Comment: "superseded by #9",
			})
			Expect(err).To(MatchError(ContainSubstring("Resource not accessible by integration")))
			Expect(stub.closeCalls).To(BeZero())
		})

		It("rejects a blank comment before touching the PR", func() {
			newStub("OPEN")

			_, err := runPRClose(PRCloseOptions{
				Repo:    "acme/widgets",
				Args:    []string{"7"},
				Comment: "   ",
			})
			Expect(err).To(MatchError(ContainSubstring("--comment")))
			Expect(stub.commentCalls).To(BeZero())
			Expect(stub.closeCalls).To(BeZero())
		})
	})
})
