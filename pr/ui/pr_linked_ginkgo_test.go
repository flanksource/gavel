package ui

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"

	"github.com/flanksource/gavel/github"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const linkedPRFixture = `{"data":{"repository":{"pullRequest":{"id":"PR_node_27","number":27,"title":"Fix validation","state":"OPEN","headRefOid":"reviewed-head","headRefName":"fix/validation","baseRefName":"main","url":"https://github.com/acme/service/pull/27","author":{"login":"dev"},"mergeable":"MERGEABLE","reviewThreads":{"nodes":[]},"comments":{"nodes":[]}}}}}`

type linkedPRTransport func(*http.Request) (*http.Response, error)

func (f linkedPRTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

var _ = Describe("linked PR UI endpoints", func() {
	It("loads and completes the linked Azure PR through its native provider", func() {
		ref := "https://dev.azure.com/acme/product/_git/service/pullrequest/27"
		mutations := 0
		upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer GinkgoRecover()
			if r.Method == http.MethodPatch {
				mutations++
				Expect(r.URL.Path).To(Equal("/acme/product/_apis/git/repositories/repo-id/pullrequests/27"))
				var body map[string]any
				Expect(json.NewDecoder(r.Body).Decode(&body)).To(Succeed())
				Expect(body).To(Equal(map[string]any{
					"status": "completed", "lastMergeSourceCommit": map[string]any{"commitId": "reviewed-head"},
					"completionOptions": map[string]any{"mergeStrategy": "rebase", "deleteSourceBranch": false, "bypassPolicy": false},
				}))
				fmt.Fprint(w, `{"pullRequestId":27,"status":"completed"}`)
				return
			}
			switch {
			case strings.HasSuffix(r.URL.Path, "/repositories/service"):
				fmt.Fprint(w, `{"id":"repo-id","name":"service","project":{"id":"project-id","name":"product"}}`)
			case strings.HasSuffix(r.URL.Path, "/pullrequests/27"):
				fmt.Fprint(w, `{"pullRequestId":27,"title":"Fix validation","status":"active","mergeStatus":"succeeded","sourceRefName":"refs/heads/feature","targetRefName":"refs/heads/main","lastMergeSourceCommit":{"commitId":"reviewed-head"},"lastMergeTargetCommit":{"commitId":"base"},"lastMergeCommit":{"commitId":"merge"}}`)
			case strings.HasSuffix(r.URL.Path, "/refs"):
				fmt.Fprint(w, `{"value":[{"name":"refs/heads/feature","objectId":"reviewed-head"},{"name":"refs/heads/main","objectId":"base"}]}`)
			case strings.HasSuffix(r.URL.Path, "/threads"), strings.HasSuffix(r.URL.Path, "/evaluations"), strings.HasSuffix(r.URL.Path, "/builds"):
				fmt.Fprint(w, `{"value":[]}`)
			default:
				Fail("unexpected Azure request: " + r.URL.Path)
			}
		}))
		DeferCleanup(upstream.Close)
		endpoint, err := url.Parse(upstream.URL)
		Expect(err).NotTo(HaveOccurred())
		transport := http.DefaultTransport
		http.DefaultTransport = linkedPRTransport(func(r *http.Request) (*http.Response, error) {
			clone := r.Clone(r.Context())
			clone.URL.Scheme, clone.URL.Host = endpoint.Scheme, endpoint.Host
			return transport.RoundTrip(clone)
		})
		DeferCleanup(func() { http.DefaultTransport = transport })
		GinkgoT().Setenv("AZURE_DEVOPS_EXT_PAT", "fixture-pat")
		s := &Server{refreshCh: make(chan struct{}, 1)}
		response := httptest.NewRecorder()
		s.handleLinkedPR(response, httptest.NewRequest("GET", "/api/prs/linked?url="+url.QueryEscape(ref), nil))
		Expect(response.Code).To(Equal(http.StatusOK), response.Body.String())
		Expect(response.Body.String()).To(And(ContainSubstring(`"provider":"azuredevops"`), ContainSubstring(`"headRefOid":"reviewed-head"`), ContainSubstring(`"repo":"https://dev.azure.com/acme/product/_git/service"`)))
		response = httptest.NewRecorder()
		s.handlePRMerge(response, httptest.NewRequest("POST", "/api/prs/merge", strings.NewReader(`{"url":"`+ref+`","method":"rebase","headSha":"reviewed-head"}`)))
		Expect(response.Code).To(Equal(http.StatusOK), response.Body.String())
		Expect(mutations).To(Equal(1))
		Expect(s.refreshCh).To(HaveLen(1))
	})
	It("documents URL-based status lookup and merge input from the wire types", func() {
		recorder := httptest.NewRecorder()
		(&Server{}).handleOpenAPI(recorder, httptest.NewRequest(http.MethodGet, "/api/openapi.json", nil))
		Expect(recorder.Code).To(Equal(http.StatusOK), recorder.Body.String())
		var document map[string]any
		Expect(json.Unmarshal(recorder.Body.Bytes(), &document)).To(Succeed())
		paths := document["paths"].(map[string]any)
		Expect(paths).To(HaveKey("/api/prs/linked"))
		merge := paths["/api/prs/merge"].(map[string]any)["post"].(map[string]any)
		content := merge["requestBody"].(map[string]any)["content"].(map[string]any)
		ref := content["application/json"].(map[string]any)["schema"].(map[string]any)["$ref"].(string)
		schemas := document["components"].(map[string]any)["schemas"].(map[string]any)
		request := schemas[strings.TrimPrefix(ref, "#/components/schemas/")].(map[string]any)
		Expect(request["properties"]).To(HaveKey("url"))
		Expect(request["properties"]).To(HaveKey("headSha"))
		Expect(request["properties"]).NotTo(HaveKey("nodeId"))
		get := paths["/api/prs/linked"].(map[string]any)["get"].(map[string]any)
		response := get["responses"].(map[string]any)["200"].(map[string]any)["content"].(map[string]any)
		ref = response["application/json"].(map[string]any)["schema"].(map[string]any)["$ref"].(string)
		Expect(schemas[strings.TrimPrefix(ref, "#/components/schemas/")].(map[string]any)["properties"]).To(And(HaveKey("pr"), HaveKey("repo")))
	})
	It("loads a linked PR outside the configured dashboard repositories", func() {
		upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer GinkgoRecover()
			var body struct {
				Variables map[string]any `json:"variables"`
			}
			Expect(json.NewDecoder(r.Body).Decode(&body)).To(Succeed())
			Expect(body.Variables).To(ContainElements("acme", "service", float64(27)))
			fmt.Fprint(w, linkedPRFixture)
		}))
		DeferCleanup(upstream.Close)
		GinkgoT().Setenv("GITHUB_API_URL", upstream.URL)
		s := &Server{ghOpts: github.Options{Token: "fixture-token"}}
		response := httptest.NewRecorder()
		s.handleLinkedPR(response, httptest.NewRequest("GET", "/api/prs/linked?url="+url.QueryEscape("https://github.com/acme/service/pull/27"), nil))
		Expect(response.Code).To(Equal(http.StatusOK))
		var body struct {
			Repo string `json:"repo"`
			PR   struct {
				Number int    `json:"number"`
				Head   string `json:"headRefOid"`
			} `json:"pr"`
		}
		Expect(json.Unmarshal(response.Body.Bytes(), &body)).To(Succeed())
		Expect(body.Repo).To(Equal("acme/service"))
		Expect(body.PR.Number).To(Equal(27))
		Expect(body.PR.Head).To(Equal("reviewed-head"))
	})

	It("rejects non-hosted linked URLs before contacting a provider", func() {
		s := &Server{}
		for _, ref := range []string{"", "https://example.com/acme/service/pull/27", "http://github.com/acme/service/pull/27"} {
			response := httptest.NewRecorder()
			s.handleLinkedPR(response, httptest.NewRequest("GET", "/api/prs/linked?url="+url.QueryEscape(ref), nil))
			Expect(response.Code).To(Equal(http.StatusBadRequest))
		}
	})

	It("merges by the linked PR identity and reviewed revision, then requests refresh", func() {
		mutations := 0
		upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer GinkgoRecover()
			var body struct {
				Query     string         `json:"query"`
				Variables map[string]any `json:"variables"`
			}
			Expect(json.NewDecoder(r.Body).Decode(&body)).To(Succeed())
			if !strings.HasPrefix(body.Query, "mutation") {
				fmt.Fprint(w, linkedPRFixture)
				return
			}
			mutations++
			Expect(body.Variables).To(Equal(map[string]any{"prId": "PR_node_27", "method": "REBASE", "head": "reviewed-head"}))
			fmt.Fprint(w, `{"data":{"mergePullRequest":{"pullRequest":{"number":27,"state":"MERGED"}}}}`)
		}))
		DeferCleanup(upstream.Close)
		GinkgoT().Setenv("GITHUB_API_URL", upstream.URL)
		s := &Server{ghOpts: github.Options{Token: "fixture-token"}, refreshCh: make(chan struct{}, 1)}
		response := httptest.NewRecorder()
		s.handlePRMerge(response, httptest.NewRequest("POST", "/api/prs/merge", strings.NewReader(`{"url":"https://github.com/acme/service/pull/27","method":"rebase","headSha":"reviewed-head"}`)))
		Expect(response.Code).To(Equal(http.StatusOK))
		Expect(mutations).To(Equal(1))
		Expect(s.refreshCh).To(HaveLen(1))
		response = httptest.NewRecorder()
		s.handlePRMerge(response, httptest.NewRequest("POST", "/api/prs/merge", strings.NewReader(`{"url":"https://github.com/acme/service/pull/27","method":"rebase","headSha":"old-head"}`)))
		Expect(response.Code).To(Equal(http.StatusBadGateway))
		Expect(response.Body.String()).To(ContainSubstring("PR head changed"))
		Expect(mutations).To(Equal(1))
	})
})
